# Despliegue

Cómo se despliega SIGTA en sus dos entornos de prueba, en Cloud Run y Cloud SQL de Google Cloud, pagados con el crédito de prueba. Cumple "Despliegue automático" (SIG-29) para estos entornos; el destino definitivo se acuerda con la Subsecretaría de TICs ([requisitos](12-requisitos-infraestructura-fiuba.md)).

## Entornos

| Entorno | Rama | Dirección | Para |
| --- | --- | --- | --- |
| `prod` | `main` | https://sigta.lopeztancredi.com.ar | Mostrar una versión cerrada: la que se presenta a la Secretaría y a los docentes |
| `dev` | `develop` | https://sigta-dev.lopeztancredi.com.ar | Ver integrado lo que ya entró a `develop` |

`main` solo recibe `develop` cuando hay una versión para mostrar (CONTRIBUTING.md), así que pasar a `prod` es un PR de `develop` a `main`.

Los dos corren con `SIGTA_ENV=development`, como la demo: usuarios sintéticos e ingreso con Google (ADR-005). Cada uno tiene su base, su bucket, sus secretos y su cuenta de servicio, así que `dev` no puede leer nada de `prod`. Comparten solo el cliente de Google.

Las sesiones, las personas, las áreas, los trámites y los datos de los documentos se guardan en la base y sobreviven a un despliegue o un reinicio. Al arrancar, el backend borra los archivos que ningún documento nombra y que tienen más de una hora, así que cada entorno sigue corriendo en una sola instancia.

## Qué hay

Todo está en el proyecto `summer-function-513519-a1`, en la región `us-east1`. Es la región más cercana con dominios propios en Cloud Run: São Paulo no los tiene. Los datos de prueba quedan en Estados Unidos.

| Pieza | Nombre | Qué es |
| --- | --- | --- |
| Aplicación | Servicio de Cloud Run `sigta-<entorno>` | Dos contenedores: `proxy` (nginx con el frontend) recibe el tráfico y manda `/api/` a `backend`, que corre al lado en `127.0.0.1:8080`. Una sola instancia, siempre prendida |
| Migraciones | Job de Cloud Run `sigta-<entorno>-migrate` | `backend migrate` con la imagen nueva, antes de cada despliegue (ADR-022) |
| Base | Instancia de Cloud SQL `sigta` (`prod`) y `sigta-dev` (`dev`) | PostgreSQL 18 con pgvector, `db-f1-micro`, base `sigta`, respaldo diario con los últimos 7. Se llega por el conector de Cloud SQL, sin redes autorizadas |
| Documentos | Bucket `summer-function-513519-a1-sigta-<entorno>-documents` | Montado en `/var/lib/sigta/documents` del backend |
| Secretos | Secret Manager | Ver abajo |
| Cuentas de servicio | `sigta-<entorno>-run` | Con la que corre cada entorno: solo su instancia, su bucket y sus secretos |
| Imágenes | Artifact Registry `us-east1-docker.pkg.dev/summer-function-513519-a1/sigta` | `backend` y `proxy`, con el commit como etiqueta. Se borran las de más de 7 días, salvo las 10 últimas |
| Dominio | Domain mapping de Cloud Run, y en Cloudflare un CNAME a `ghs.googlehosted.com` sin el proxy | Google saca y renueva el certificado. El dominio está verificado en Search Console con la cuenta dueña del proyecto |
| Presupuesto | Cuenta de facturación del proyecto, presupuesto "SIGTA" | 60 USD por mes, con aviso por correo al 50 %, al 90 % y al 100 % |

## Cómo se despliega

En cada push a `main` o a `develop`, el workflow `Deploy` (`.github/workflows/deploy.yml`):

1. Obtiene un token de Google Cloud de corta duración por Workload Identity Federation. El proveedor `github/sigta` solo acepta ejecuciones de `main` y `develop` de este repositorio. GitHub no guarda ninguna clave de Google Cloud.
2. Construye las imágenes del backend y del proxy y las publica en Artifact Registry.
3. Corre el job de migraciones con la imagen nueva y espera a que termine. Si falla, no sigue y queda en línea la versión anterior.
4. Despliega la versión nueva del servicio. Cloud Run la pone en línea cuando el backend obtiene el catálogo, que lee la base. Si no arranca, sigue la anterior.
5. Comprueba que el frontend y la API respondan.

Las migraciones solo avanzan: si la aplicación nueva no arranca después de migrar, sigue en línea la versión anterior. Volver atrás una migración es restaurar el respaldo (ADR-022).

## Secretos

| Secreto | Contenido |
| --- | --- |
| `sigta-<entorno>-database-url` | Conexión del servidor, con el rol `backend` |
| `sigta-<entorno>-migrations-database-url` | Conexión de las migraciones, con el rol `backend_migrations` |
| `sigta-google-client-id` y `sigta-google-client-secret` | El cliente OAuth de Google "SIGTA", con el origen y la redirección de las dos direcciones |

Las contraseñas de los roles se generan una vez, al crear los secretos. Cambiar una es cambiarla en la base (`ALTER ROLE`) y agregar una versión nueva del secreto. La del usuario `postgres` no se guarda: se cambia con `gcloud sql users set-password`.

## Armar los entornos desde cero

1. Con una cuenta dueña del proyecto, y `cloud-sql-proxy` y `psql` instalados: `CLOUDSDK_CORE_PROJECT=<proyecto> BILLING_ACCOUNT=<cuenta> sh deploy/gcp/setup.sh`. Pide el cliente de Google la primera vez. Crea el registro de imágenes, las bases con sus roles, los buckets, los secretos, las cuentas de servicio, la identidad de GitHub y el presupuesto, y saltea lo que ya existe.
2. En el cliente OAuth de Google, cargar cada dirección como origen y su `/api/auth/callback` como redirección.
3. Correr el workflow `Deploy` desde GitHub en cada rama, o hacer un push a `main` y a `develop`.
4. Verificar el dominio en Search Console con la misma cuenta, mapearlo con `gcloud beta run domain-mappings create --service sigta-<entorno> --domain <dirección> --region us-east1`, y en el DNS apuntar cada dirección a `ghs.googlehosted.com` con un CNAME. El certificado tarda entre 15 minutos y un día.

## Operación

En los comandos, `<entorno>` es `prod` o `dev`, y todos llevan `--region us-east1`.

| Para | Comando |
| --- | --- |
| Ver el estado y las revisiones | `gcloud run services describe sigta-<entorno>` y `gcloud run revisions list --service sigta-<entorno>` |
| Ver los logs | `gcloud logging read 'resource.labels.service_name="sigta-<entorno>"' --freshness 1h`, o la pestaña Logs del servicio en la consola |
| Volver a la versión anterior | `gcloud run services update-traffic sigta-<entorno> --to-revisions <revisión>=100` |
| Respaldar la base a pedido | `gcloud sql backups create --instance <instancia>` |
| Respaldar los documentos | `gcloud storage cp -r gs://summer-function-513519-a1-sigta-<entorno>-documents .` |

## Costo

Con el crédito de prueba (300 USD por 90 días) no se paga nada. Sin él, el costo estimado es de unos 40 USD por mes, con los precios de lista de `us-east1`, a confirmar con la facturación real: las dos instancias `db-f1-micro` con su disco (unos 18 USD), y los dos servicios de Cloud Run, cada uno con una instancia de 1 vCPU y 640 MB siempre prendida (unos 20 USD). El bucket, los secretos y las imágenes cuestan centavos. El presupuesto avisa antes de pasar los 60 USD.
