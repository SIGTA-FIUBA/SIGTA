# Despliegue

Cómo se despliega SIGTA en sus dos entornos de prueba, en una máquina virtual de Google Cloud pagada con el crédito de prueba. Es un ensayo de la infraestructura de FIUBA (Docker Swarm con Traefik, ver [requisitos](12-requisitos-infraestructura-fiuba.md)). Cumple "Despliegue automático" (SIG-29) para estos entornos; el destino definitivo se acuerda con la Subsecretaría de TICs.

## Entornos

| Entorno | Rama | Dirección | Para |
| --- | --- | --- | --- |
| `prod` | `main` | https://sigta.lopeztancredi.com.ar | Mostrar una versión cerrada: la que se presenta a la Secretaría y a los docentes |
| `dev` | `develop` | https://sigta-dev.lopeztancredi.com.ar | Ver integrado lo que ya entró a `develop` |

`main` solo recibe `develop` cuando hay una versión para mostrar (CONTRIBUTING.md), así que pasar a `prod` es un PR de `develop` a `main`.

Los dos corren con `SIGTA_ENV=development`, como la demo: usuarios sintéticos e ingreso con Google (ADR-005). Cada uno tiene su base, sus documentos y sus secretos; comparten la máquina, Traefik, el túnel y el cliente de Google. Las sesiones están en memoria, así que cada despliegue cierra las sesiones abiertas del entorno.

## Qué hay

| Pieza | Dónde | Qué es |
| --- | --- | --- |
| Máquina virtual | Google Cloud, proyecto `summer-function-513519-a1`, zona `southamerica-east1-b`, `sigta` | `e2-medium` (4 GB de memoria, disco de 30 GB) con Debian 13 y Docker en modo Swarm, de un solo nodo |
| Red | Red `sigta`, sin direcciones públicas | Ninguna conexión entra desde internet. SSH solo a través de Identity-Aware Proxy (regla `sigta-allow-iap-ssh`). La máquina sale a internet por Cloud NAT |
| Entrada | Stack `sigta-edge` (`deploy/swarm/edge.yml`) | Traefik, que manda cada dirección al proxy de su entorno, y `cloudflared`, que trae el tráfico del túnel de Cloudflare `sigta` por la red interna `sigta-edge` |
| Base de cada entorno | Stack `sigta-<entorno>-db` (`deploy/swarm/database.yml`) | PostgreSQL 18 con pgvector, con los datos en el volumen `sigta-<entorno>-db_db-data` |
| Aplicación de cada entorno | Stack `sigta-<entorno>` (`deploy/swarm/app.yml`) | El proxy con el frontend y el backend. Los documentos adjuntos van al volumen `sigta-<entorno>_documents` |
| Imágenes | GHCR, `ghcr.io/sigta-fiuba/sigta/backend` y `ghcr.io/sigta-fiuba/sigta/proxy` | Una etiqueta por commit, y `main` y `develop` para la última de cada rama |
| Presupuesto | Cuenta de facturación del proyecto, presupuesto "SIGTA" | 60 USD por mes, con aviso por correo al 50 %, al 90 % y al 100 % |

## Cómo se despliega

En cada push a `main` o a `develop`, el workflow `Deploy` (`.github/workflows/deploy.yml`):

1. Construye las imágenes del backend y del proxy y las publica en GHCR, con el commit como etiqueta.
2. Obtiene un token de Google Cloud de corta duración por Workload Identity Federation. El proveedor `github/sigta` solo acepta ejecuciones de `main` y `develop` de este repositorio (por su identificador, no por su nombre), y la cuenta de servicio `sigta-deploy` solo puede entrar por SSH a la máquina a través del proxy. GitHub no guarda ninguna clave de Google Cloud.
3. Copia `deploy/swarm` y `deploy/postgres` a la máquina y corre `deploy/swarm/deploy.sh` con el entorno de la rama y el commit. El acceso al registro es el token de la ejecución, que vence con ella.
4. Comprueba que el frontend y la API del entorno respondan a través del túnel.

Se despliega de a uno, porque los dos entornos comparten la entrada.

`deploy.sh` hace cuatro pasos y, si uno falla, no sigue:

1. **Entrada:** actualiza el stack `sigta-edge`.
2. **Base:** actualiza el stack de la base del entorno y espera a que esté sana.
3. **Migraciones:** corre `backend migrate` de la versión nueva como un servicio que no se reinicia, conectado a la red de la base, y lee su estado final (ADR-022). Swarm no tiene contenedores sueltos con secretos.
4. **Aplicación:** actualiza el stack del entorno. Cada servicio arranca la versión nueva antes de bajar la anterior y la reemplaza solo cuando está sana; si no, Swarm vuelve a la anterior y el despliegue falla. El backend está sano cuando `backend probe` obtiene el catálogo, que lee la base.

Las migraciones solo avanzan: si la aplicación nueva no arranca después de migrar, sigue en línea la versión anterior, que ya estaba corriendo. Un backend anterior que se reinicie no arranca contra migraciones que no conoce; volver atrás es restaurar el respaldo (ADR-022).

Compose no reemplaza variables en los nombres de las etiquetas, y las de Traefik nombran el router y el servicio con el entorno (`sigta-prod`, `sigta-dev`). Por eso `deploy.sh` completa `${ENVIRONMENT}` en `app.yml` antes de desplegarlo.

## Secretos

Son secretos de Swarm: se crean una vez en la máquina y los servicios los leen como archivos en `/run/secrets/`. Ninguno está en el repositorio ni en GitHub.

| Secreto | Contenido | Lo usa |
| --- | --- | --- |
| `<entorno>_postgres_password`, `<entorno>_backend_db_password`, `<entorno>_backend_migrations_db_password`, `<entorno>_ai_db_password` | Contraseñas aleatorias de la base del entorno | La base, al crear el volumen |
| `<entorno>_database_url` y `<entorno>_migrations_database_url` | Conexiones con el rol de la aplicación y con el de migraciones | El backend y las migraciones del entorno |
| `google_client_id` y `google_client_secret` | El cliente OAuth de Google "SIGTA", con el origen y la redirección de las dos direcciones | El backend de los dos entornos |
| `cloudflared_credentials` | Las credenciales del túnel | `cloudflared` |

El backend acepta cualquiera de sus variables secretas como archivo: `NAME_FILE` nombra el archivo con el valor de `NAME` (`SIGTA_GOOGLE_CLIENT_ID`, `SIGTA_GOOGLE_CLIENT_SECRET`, `SIGTA_DATABASE_URL` y `SIGTA_MIGRATIONS_DATABASE_URL`). `deploy/postgres/init.sh` hace lo mismo con las contraseñas.

## Armar los entornos desde cero

1. Con una cuenta dueña del proyecto: `CLOUDSDK_CORE_PROJECT=<proyecto> BILLING_ACCOUNT=<cuenta> sh deploy/gcp/setup.sh`. Crea la red, la máquina, la identidad de GitHub y el presupuesto, y saltea lo que ya existe. Al arrancar, la máquina instala Docker e inicia Swarm (`deploy/gcp/startup.sh`).
2. En Cloudflare, apuntar las dos direcciones al túnel `sigta` (`cloudflared tunnel route dns sigta <dirección>`). En el cliente OAuth de Google, cargar cada dirección como origen y su `/api/auth/callback` como redirección.
3. Entrar a la máquina: `gcloud compute ssh sigta --zone southamerica-east1-b --tunnel-through-iap`.
4. Crear en `/etc/sigta/` los archivos de configuración, que no tienen secretos:

   | Archivo | Contenido |
   | --- | --- |
   | `edge.env` | `CLOUDFLARE_TUNNEL_ID=<identificador del túnel>` |
   | `prod.env` | `SIGTA_HOST=sigta.lopeztancredi.com.ar` y `SIGTA_ENV=development` |
   | `dev.env` | `SIGTA_HOST=sigta-dev.lopeztancredi.com.ar` y `SIGTA_ENV=development` |

5. Copiar a la máquina las credenciales de Google y las del túnel, crear los secretos con `sudo sh deploy/swarm/secrets.sh prod <google.env> <túnel.json>` y `sudo sh deploy/swarm/secrets.sh dev`, y borrar las copias.
6. Correr el workflow `Deploy` desde GitHub en cada rama, o hacer un push a `main` y a `develop`.

## Operación

En los comandos, `<entorno>` es `prod` o `dev`.

| Para | Comando, en la máquina |
| --- | --- |
| Ver los servicios y su estado | `sudo docker service ls` y `sudo docker service ps sigta-<entorno>_backend` |
| Ver los logs | `sudo docker service logs --since 1h sigta-<entorno>_backend` |
| Volver a la versión anterior de un servicio | `sudo docker service rollback sigta-<entorno>_backend` |
| Respaldar los documentos | `sudo tar -czf documents.tgz -C /var/lib/docker/volumes/sigta-<entorno>_documents/_data .` |
| Respaldar la base | `sudo docker exec $(sudo docker ps -q -f label=com.docker.swarm.service.name=sigta-<entorno>-db_db) pg_dump -U postgres sigta > sigta.sql` |

## Costo

Con el crédito de prueba (300 USD por 90 días) no se paga nada. Sin él, el costo estimado es de unos 47 USD por mes, con los precios de lista de São Paulo: la máquina `e2-medium` (39 USD), su disco de 30 GB (4,50 USD), y Cloud NAT y el tráfico (unos 4 USD). Los tres meses del crédito usan unos 140 USD. El presupuesto avisa antes de pasar los 60 USD.
