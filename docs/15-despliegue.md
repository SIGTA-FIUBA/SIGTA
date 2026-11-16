# Despliegue

Cómo se despliega SIGTA en sus dos entornos de prueba, en una máquina virtual de Google Cloud pagada con el crédito de prueba. Es un ensayo de la infraestructura de FIUBA (Docker Swarm con Traefik, ver [requisitos](12-requisitos-infraestructura-fiuba.md)). Cumple "Despliegue automático" (SIG-29) para estos entornos; el destino definitivo se acuerda con la Subsecretaría de TICs.

## Entornos

| Entorno | Rama | Dirección | Para |
| --- | --- | --- | --- |
| `prod` | `main` | https://sigta.lopeztancredi.com.ar | Mostrar una versión cerrada: la que se presenta a la Secretaría y a los docentes |
| `dev` | `develop` | https://sigta-dev.lopeztancredi.com.ar | Ver integrado lo que ya entró a `develop` |

`main` solo recibe `develop` cuando hay una versión para mostrar (CONTRIBUTING.md), así que pasar a `prod` es un PR de `develop` a `main`.

Los dos corren con `SIGTA_ENV=development`, como la demo: usuarios sintéticos e ingreso con Google (ADR-005). Cada uno tiene su base, sus documentos y sus secretos; comparten la máquina, Traefik y el cliente de Google. Las sesiones están en memoria, así que cada despliegue cierra las sesiones abiertas del entorno.

## Qué hay

| Pieza | Dónde | Qué es |
| --- | --- | --- |
| Máquina virtual | Google Cloud, proyecto `summer-function-513519-a1`, zona `southamerica-east1-b`, `sigta` | `e2-medium` (4 GB de memoria, disco de 30 GB) con Debian 13 y Docker en modo Swarm, de un solo nodo |
| Red | Red `sigta`, IP fija `sigta` (35.199.120.196) | Entran solo el 80 y el 443 (regla `sigta-allow-web`). SSH solo a través de Identity-Aware Proxy (regla `sigta-allow-iap-ssh`) |
| DNS | Cloudflare, zona `lopeztancredi.com.ar` | Las dos direcciones son registros A a la IP fija, sin el proxy de Cloudflare |
| Entrada | Stack `sigta-edge` (`deploy/swarm/edge.yml`) | Traefik: manda cada dirección al proxy de su entorno y saca los certificados de Let's Encrypt |
| Base de cada entorno | Stack `sigta-<entorno>-db` (`deploy/swarm/database.yml`) | PostgreSQL 18 con pgvector, con los datos en el volumen `sigta-<entorno>-db_db-data` |
| Aplicación de cada entorno | Stack `sigta-<entorno>` (`deploy/swarm/app.yml`) | El proxy con el frontend y el backend. Los documentos adjuntos van al volumen `sigta-<entorno>_documents` |
| Imágenes | GHCR, `ghcr.io/sigta-fiuba/sigta/backend` y `ghcr.io/sigta-fiuba/sigta/proxy` | Una etiqueta por commit, y `main` y `develop` para la última de cada rama |
| Presupuesto | Cuenta de facturación del proyecto, presupuesto "SIGTA" | 60 USD por mes, con aviso por correo al 50 %, al 90 % y al 100 % |

## Cómo se despliega

En cada push a `main` o a `develop`, el workflow `Deploy` (`.github/workflows/deploy.yml`):

1. Construye las imágenes del backend y del proxy y las publica en GHCR, con el commit como etiqueta.
2. Obtiene un token de Google Cloud de corta duración por Workload Identity Federation. El proveedor `github/sigta` solo acepta ejecuciones de `main` y `develop` de este repositorio, y la cuenta de servicio `sigta-deploy` solo puede entrar por SSH a la máquina. GitHub no guarda ninguna clave de Google Cloud.
3. Copia `deploy/swarm` y `deploy/postgres` a la máquina y corre `deploy/swarm/deploy.sh` con el entorno, el commit y la dirección. Los secretos le llegan por la entrada estándar.
4. Comprueba que el frontend y la API del entorno respondan.

`deploy.sh` hace cuatro pasos y, si uno falla, no sigue:

1. **Entrada:** actualiza el stack `sigta-edge`.
2. **Base:** actualiza el stack de la base del entorno y espera a que esté sana.
3. **Migraciones:** corre `backend migrate` de la versión nueva como un servicio que no se reinicia y lee su estado final (ADR-022).
4. **Aplicación:** actualiza el stack del entorno. Cada servicio arranca la versión nueva antes de bajar la anterior y la reemplaza solo cuando está sana; si no, Swarm vuelve a la anterior y el despliegue falla. El backend está sano cuando `backend probe` obtiene el catálogo, que lee la base.

Las migraciones solo avanzan: si la aplicación nueva no arranca después de migrar, sigue en línea la versión anterior. Volver atrás una migración es restaurar el respaldo (ADR-022).

## Secretos

Son secretos de los Environments `prod` y `dev` de GitHub. Cada Environment solo acepta despliegues de su rama, así que un PR no los puede leer. Llegan a los servicios como variables de entorno.

| Secreto | Contenido |
| --- | --- |
| `POSTGRES_PASSWORD`, `BACKEND_DB_PASSWORD`, `BACKEND_MIGRATIONS_DB_PASSWORD`, `AI_DB_PASSWORD` | Contraseñas de la base del entorno, en hexadecimal porque van dentro de la URL de conexión. Solo se aplican al crear el volumen de la base: cambiarlas después requiere cambiarlas también en la base |
| `GOOGLE_CLIENT_ID` y `GOOGLE_CLIENT_SECRET` | El cliente OAuth de Google "SIGTA", con el origen y la redirección de las dos direcciones |

## Armar los entornos desde cero

1. Con una cuenta dueña del proyecto: `CLOUDSDK_CORE_PROJECT=<proyecto> BILLING_ACCOUNT=<cuenta> sh deploy/gcp/setup.sh`. Crea la red, la IP fija, la máquina (que instala Docker e inicia Swarm al arrancar), la identidad de GitHub y el presupuesto, y saltea lo que ya existe. Al final muestra la IP.
2. En el DNS, apuntar las dos direcciones a esa IP con registros A. En el cliente OAuth de Google, cargar cada dirección como origen y su `/api/auth/callback` como redirección.
3. En GitHub, crear los Environments `prod` (solo `main`) y `dev` (solo `develop`) con sus secretos. Las contraseñas se generan con `openssl rand -hex 24`.
4. Correr el workflow `Deploy` desde GitHub en cada rama, o hacer un push a `main` y a `develop`.

## Operación

En los comandos, `<entorno>` es `prod` o `dev`. Se entra a la máquina con `gcloud compute ssh sigta --zone southamerica-east1-b --tunnel-through-iap`.

| Para | Comando, en la máquina |
| --- | --- |
| Ver los servicios y su estado | `sudo docker service ls` y `sudo docker service ps sigta-<entorno>_backend` |
| Ver los logs | `sudo docker service logs --since 1h sigta-<entorno>_backend` |
| Volver a la versión anterior de un servicio | `sudo docker service rollback sigta-<entorno>_backend` |
| Respaldar los documentos | `sudo tar -czf documents.tgz -C /var/lib/docker/volumes/sigta-<entorno>_documents/_data .` |
| Respaldar la base | `sudo docker exec $(sudo docker ps -q -f label=com.docker.swarm.service.name=sigta-<entorno>-db_db) pg_dump -U postgres sigta > sigta.sql` |

## Costo

Con el crédito de prueba (300 USD por 90 días) no se paga nada. Sin él, el costo estimado es de unos 47 USD por mes, con los precios de lista de São Paulo: la máquina `e2-medium` (39 USD), su disco de 30 GB (4,50 USD), y la IP fija y el tráfico (unos 4 USD). Los tres meses del crédito usan unos 140 USD. El presupuesto avisa antes de pasar los 60 USD.
