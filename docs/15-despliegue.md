# Despliegue

Cómo se despliega SIGTA en el entorno de prueba, el que se ve en https://sigta.lopeztancredi.com.ar. Es un ensayo de la infraestructura de FIUBA (Docker Swarm con Traefik, ver [requisitos](12-requisitos-infraestructura-fiuba.md)), sobre una máquina virtual de Google Cloud pagada con el crédito de prueba. Cumple "Despliegue automático" (SIG-29) para este entorno; el destino definitivo se acuerda con la Subsecretaría de TICs.

## Qué hay

| Pieza | Dónde | Qué es |
| --- | --- | --- |
| Máquina virtual | Google Cloud, proyecto `summer-function-513519-a1`, zona `southamerica-east1-b`, `sigta-staging` | `e2-small` (2 GB de memoria) con Debian 13 y Docker en modo Swarm, de un solo nodo |
| Red | Red `sigta`, sin direcciones públicas | Ninguna conexión entra desde internet. SSH solo a través de Identity-Aware Proxy (regla `sigta-allow-iap-ssh`). La máquina sale a internet por Cloud NAT |
| Base | Stack `sigta-db` (`deploy/swarm/database.yml`) | PostgreSQL 18 con pgvector, con los datos en el volumen `sigta-db_db-data` |
| Aplicación | Stack `sigta` (`deploy/swarm/sigta.yml`) | Traefik, el proxy con el frontend, el backend y `cloudflared`. Los documentos adjuntos van al volumen `sigta_documents` |
| Entrada | Túnel de Cloudflare `sigta` | `cloudflared` corre dentro del stack y entrega el tráfico a Traefik por la red interna |
| Imágenes | GHCR, `ghcr.io/sigta-fiuba/sigta/backend` y `ghcr.io/sigta-fiuba/sigta/proxy` | Una etiqueta por commit de `develop` y `develop` para la última |
| Presupuesto | Cuenta de facturación del proyecto, presupuesto "SIGTA staging" | 25 USD por mes, con aviso por correo al 50 %, al 90 % y al 100 % |

El entorno corre con `SIGTA_ENV=development`, como la demo de hoy: usuarios sintéticos e ingreso con Google (ADR-005). Las sesiones están en memoria, así que cada despliegue cierra las sesiones abiertas.

## Cómo se despliega

En cada push a `develop`, el workflow `Deploy` (`.github/workflows/deploy.yml`):

1. Construye las imágenes del backend y del proxy y las publica en GHCR, con el commit como etiqueta.
2. Obtiene un token de Google Cloud de corta duración por Workload Identity Federation. El proveedor `github/sigta` solo acepta ejecuciones de la rama `develop` de este repositorio (por su identificador, no por su nombre), y la cuenta de servicio `sigta-deploy` solo puede entrar por SSH a la máquina a través del proxy. GitHub no guarda ninguna clave de Google Cloud.
3. Copia `deploy/swarm` y `deploy/postgres` a la máquina y corre `deploy/swarm/deploy.sh` con el commit. El acceso al registro es el token de la ejecución, que vence con ella.
4. Comprueba que el frontend y la API respondan a través del túnel.

`deploy.sh` hace tres pasos y, si uno falla, no sigue:

1. **Base:** actualiza el stack `sigta-db` y espera a que la base esté sana.
2. **Migraciones:** corre `backend migrate` de la versión nueva como un servicio que no se reinicia, conectado a la red de la base, y lee su estado final (ADR-022). Swarm no tiene contenedores sueltos con secretos.
3. **Aplicación:** actualiza el stack `sigta`. Cada servicio arranca la versión nueva antes de bajar la anterior y la reemplaza solo cuando está sana; si no, Swarm vuelve a la anterior y el despliegue falla. El backend está sano cuando `backend probe` obtiene el catálogo, que lee la base.

Las migraciones solo avanzan: si la aplicación nueva no arranca después de migrar, sigue en línea la versión anterior, que ya estaba corriendo. Un backend anterior que se reinicie no arranca contra migraciones que no conoce; volver atrás es restaurar el respaldo (ADR-022).

## Secretos

Son secretos de Swarm: se crean una vez en la máquina y los servicios los leen como archivos en `/run/secrets/`. Ninguno está en el repositorio ni en GitHub.

| Secreto | Contenido | Lo usa |
| --- | --- | --- |
| `postgres_password`, `backend_db_password`, `backend_migrations_db_password`, `ai_db_password` | Contraseñas aleatorias de la base | La base, al crear el volumen |
| `database_url` y `migrations_database_url` | Conexiones con el rol de la aplicación y con el de migraciones | El backend y las migraciones |
| `google_client_id` y `google_client_secret` | El cliente OAuth de Google | El backend |
| `cloudflared_credentials` | Las credenciales del túnel | `cloudflared` |

El backend acepta cualquiera de sus variables secretas como archivo: `NAME_FILE` nombra el archivo con el valor de `NAME` (`SIGTA_GOOGLE_CLIENT_ID`, `SIGTA_GOOGLE_CLIENT_SECRET`, `SIGTA_DATABASE_URL` y `SIGTA_MIGRATIONS_DATABASE_URL`). `deploy/postgres/init.sh` hace lo mismo con las contraseñas.

## Armar el entorno desde cero

1. Con una cuenta dueña del proyecto: `CLOUDSDK_CORE_PROJECT=<proyecto> BILLING_ACCOUNT=<cuenta> sh deploy/gcp/setup.sh`. Crea la red, la máquina, la identidad de GitHub y el presupuesto, y saltea lo que ya existe. Al arrancar, la máquina instala Docker e inicia Swarm (`deploy/gcp/startup.sh`).
2. Entrar a la máquina: `gcloud compute ssh sigta-staging --zone southamerica-east1-b --tunnel-through-iap`.
3. Crear `/etc/sigta/deploy.env`, que no tiene secretos:

   ```sh
   SIGTA_HOST=sigta.lopeztancredi.com.ar
   SIGTA_ENV=development
   CLOUDFLARE_TUNNEL_ID=<identificador del túnel>
   ```

4. Copiar a la máquina las credenciales de Google y las del túnel, crear los secretos con `sudo sh deploy/swarm/secrets.sh <google.env> <túnel.json>` y borrar las copias.
5. Correr el workflow `Deploy` desde GitHub, o hacer un push a `develop`.

## Operación

| Para | Comando, en la máquina |
| --- | --- |
| Ver los servicios y su estado | `sudo docker service ls` y `sudo docker service ps sigta_backend` |
| Ver los logs | `sudo docker service logs --since 1h sigta_backend` |
| Volver a la versión anterior de un servicio | `sudo docker service rollback sigta_backend` |
| Respaldar los documentos | `sudo tar -czf documents.tgz -C /var/lib/docker/volumes/sigta_documents/_data .` |
| Respaldar la base | `sudo docker exec $(sudo docker ps -q -f label=com.docker.swarm.service.name=sigta-db_db) pg_dump -U postgres sigta > sigta.sql` |

## Costo

Con el crédito de prueba (300 USD por 90 días) no se paga nada. Sin él, el costo estimado es de unos 20 USD por mes: la máquina `e2-small`, su disco de 20 GB y Cloud NAT. El presupuesto avisa antes de pasar los 25 USD.
