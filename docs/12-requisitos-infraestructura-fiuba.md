# Requisitos de infraestructura de FIUBA

Qué cumple SIGTA de los requisitos que pide la Subsecretaría de TICs para desplegar un sistema en la infraestructura de la facultad. El documento de la Subsecretaría de TICs es "Requisitos para desplegar un sistema nuevo en la infraestructura FIUBA" ([original en Google Docs](https://docs.google.com/document/d/10q4lbU6mWsikQD9JoTu6fK5E3EHYFAfs/edit)). Este archivo no lo copia: resume cada requisito y lo contrasta con el estado de SIGTA al 2026-11-02.

**Estados.** Cumple: SIGTA ya lo hace. No cumple: SIGTA hace otra cosa o todavía no lo tiene. A confirmar: depende de algo que hay que preguntar o decidir. No aplica: el requisito no corresponde a SIGTA.

En resumen: los requisitos de autenticación y de contenedores ya se cumplen. La brecha que afecta a la arquitectura es pgvector sobre PostgreSQL 11.9 (riesgo R-17).

## Contenedores e infraestructura

La infraestructura de FIUBA corre sobre Docker Swarm, con Traefik como proxy reverso.

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Un Dockerfile de producción por componente, sin código montado en volúmenes | Cumple | `services/backend/Dockerfile` y `deploy/nginx/Dockerfile`, que construye el frontend y lo sirve con el proxy |
| Un compose de referencia que Swarm soporte: sin `build:`, sin `depends_on` con condiciones, sin `container_name` ni `links`, sin puertos publicados en el host para servicios web | No cumple | `compose.yaml` es solo para desarrollo. Los entornos de prueba corren en Cloud Run, así que no hay un compose de referencia para Swarm: se arma cuando se acuerde el despliegue con la Subsecretaría de TICs. Las migraciones (`backend migrate`) se corren como un paso del despliegue antes de actualizar el backend (ADR-022, [despliegue](15-despliegue.md)) |
| Cada servicio HTTP expone un único puerto interno y no termina TLS | Cumple | El proxy escucha en el 80 y el backend en el 8080, los dos sin TLS |
| Las imágenes las construye la Subsecretaría de TICs y se publican en su registry | A confirmar | Para los entornos de prueba, GitHub Actions las construye en cada push a `main` y a `develop` y las publica en Artifact Registry. Hay que acordar con la Subsecretaría de TICs quién las construye para producción |
| Configuración por variables de entorno, sin valores de entorno escritos en la imagen | Cumple | Todo el backend se configura con variables `SIGTA_*` (ver "Configuración del servicio de trámites" en el README) |
| Contraseñas y claves como secretos del despliegue, nunca en el repositorio | Cumple | Las credenciales de Google y las dos conexiones a la base (`SIGTA_DATABASE_URL` y `SIGTA_MIGRATIONS_DATABASE_URL`) son secretos del despliegue. Las contraseñas de `compose.yaml` son solo de desarrollo |
| Servicios sin estado, con sesiones, caché y colas en la base o en Redis, para correr más de una réplica | A confirmar | Las sesiones, las cuentas, las áreas, las membresías y los trámites se guardan en PostgreSQL y sobreviven a un reinicio. Los archivos de los documentos siguen en un volumen, y al arrancar el backend borra los que ningún documento nombra y tienen más de una hora, así que hoy supone una sola instancia |
| Datos persistentes solo en volúmenes declarados y documentados | A confirmar | Hay dos volúmenes: el de la base y el de los documentos (`/var/lib/sigta/documents`, `SIGTA_DOCUMENTS_DIR`). En los entornos de prueba, la base es Cloud SQL y los documentos van a un bucket de Cloud Storage por entorno, montado en esa ruta. Los trámites y el registro de sus documentos están en la base. Al arrancar, el backend borra los archivos que ningún documento registrado nombra y que tienen más de una hora: los de una subida que no llegó a registrarse. Por eso hoy supone una sola instancia |
| La SPA no tiene la URL del backend fija en la imagen | Cumple | El frontend llama a la API por la ruta relativa `/api` (la opción que prefiere la Subsecretaría de TICs) |

## Base de datos

La Subsecretaría de TICs ofrece MariaDB 10.1.45 o PostgreSQL 11.9, como mínimo. El contenedor de base del compose es solo para desarrollo.

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Funcionar con PostgreSQL 11.9 o superior, avisando de antemano si se usan funciones más nuevas | No cumple | Se desarrolla sobre PostgreSQL 18, y el asistente usa la extensión pgvector (`deploy/postgres/sql/roles.sql`). Las versiones recientes de pgvector ya no soportan PostgreSQL 11, y no se sabe si el servidor de la Subsecretaría de TICs tiene la extensión. Ver R-17. Las migraciones del servicio de trámites solo usan lo que existe en PostgreSQL 11, pero las pruebas corren sobre PostgreSQL 18: ninguna prueba corre contra la 11. Además, el backend se conecta con pgx v5, que soporta oficialmente PostgreSQL 14 y posteriores (sección "Supported Go and PostgreSQL Versions" de su README); que funcione con la 11.9 está a confirmar |
| Apuntar a un servidor de base externo por configuración | Cumple | `SIGTA_DATABASE_URL` |
| Esquema con migraciones versionadas en el repositorio | Cumple | Las migraciones son archivos `.sql` numerados en `services/backend/src/migrations/`, que van dentro del binario. Las aplica `backend migrate`, y el servidor no arranca si falta aplicar alguna o si se editó una ya aplicada (ADR-022). Las migraciones solo avanzan: no hay migraciones para deshacerlas, y un backend anterior no arranca contra una base con migraciones que no conoce. Para volver a una versión anterior se restauran los respaldos de la base y de la carpeta de los documentos tomados antes de migrar (ver "Ruta del volumen documentada para los respaldos") |
| Datos iniciales con scripts idempotentes, separados de los datos de prueba | No cumple | Las migraciones se pueden correr más de una vez: cada una se aplica una sola vez. Pero `deploy/postgres/init.sh` crea los roles y los esquemas solo al inicializar la base de desarrollo, y no se puede volver a correr. Los usuarios sintéticos y el tipo de trámite de ejemplo se cargan solo con `SIGTA_ENV=development` |
| Usuario de la aplicación con permisos solo sobre su base; documentar si las migraciones necesitan más | A confirmar | Cada servicio tiene su rol y su esquema (`backend` y `ai`). En el backend, el esquema y sus tablas son de un rol aparte, `backend_migrations`, que es el único que corre las migraciones y necesita crear tablas en el esquema; el rol de la aplicación solo tiene los permisos que le otorga cada migración: leer y agregar filas (de los administradores, por ahora solo leer), cambiar solo las columnas que cada una nombra y borrar solo sesiones y borradores de tipos de trámite, nunca eventos ni versiones (ADR-022). `backend` y `backend_migrations` son nombres fijos en las migraciones: si la Subsecretaría de TICs asigna otros usuarios, alcanza con que sean miembros de esos roles (el del servidor, de ningún otro con más permisos, ni superusuario: si no, no arranca): el servidor y `backend migrate` toman su rol con `SET ROLE` al conectarse, así que las tablas siguen siendo de `backend_migrations`. Como las migraciones crean las tablas sin nombrar el esquema, cada conexión usa el esquema `backend` salvo que su URL indique otro `search_path`. Hasta PostgreSQL 14, como la 11.9, cualquier rol puede crear tablas en el esquema `public`; `roles.sql` se lo quita con `REVOKE CREATE ON SCHEMA public FROM PUBLIC`, y en el servidor de la Subsecretaría de TICs hace falta lo mismo, a confirmar con ellos. Crear la extensión `vector` necesita permisos de superusuario |
| UTF-8 y zona horaria documentada | A confirmar | No está documentada la zona horaria que se asume |

## Autenticación y autorización

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Ingreso con Google restringido al dominio de FIUBA | Cumple | ADR-010: solo se ingresa con la cuenta de Google de la UBA |
| Dominio validado en el backend (claim `hd` y correo verificado) | Cumple | `auth/adapters/google/identity.go` rechaza el token sin correo verificado o con otro `hd` |
| Client ID, client secret y URL de redirección por variables de entorno | Cumple | `SIGTA_GOOGLE_CLIENT_ID`, `SIGTA_GOOGLE_CLIENT_SECRET` y `SIGTA_PUBLIC_URL` |
| Credenciales de Google de test y producción provistas por la Subsecretaría de TICs | A confirmar | Hoy se usa un cliente propio del equipo |
| Requisitos para un ingreso propio con contraseña | No aplica | SIGTA no guarda contraseñas ni ofrece otra forma de ingresar (ADR-010) |
| Roles y permisos verificados en el backend en cada endpoint | A confirmar | Los permisos son de cada área y se asignan en membresías, sueltos o en roles que crea el administrador (ADR-020). Se verifican en los casos de uso del servidor: la administración de áreas y membresías exige ser administrador, y publicar un tipo de trámite, los permisos de la Secretaría |

## Correo electrónico

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| SMTP de la facultad en test y producción; en desarrollo, una herramienta que capture los mensajes | No cumple | SIGTA todavía no envía correos: no hay adaptador de correo, ni en desarrollo ni fuera de él |
| Host, puerto, usuario, contraseña, cifrado y remitente por variables de entorno | No cumple | Se define con el adaptador SMTP |
| Envíos frecuentes a través de una cola, con el volumen estimado por día | A confirmar | Las notificaciones son por cambio de estado y pedido de acción, y la arquitectura prevé una cola de trabajos en PostgreSQL. Falta estimar el volumen |
| Sin servicios de envío externos pagos sin acordarlo | Cumple | No se usa ninguno |

## Archivos

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Estimación por tipo de archivo: formato, tamaño promedio y máximo, cantidad por año y retención | A confirmar | Cada tipo de trámite declara los formatos y el tamaño máximo de sus documentos, y un trámite guarda hasta 20 archivos por documento, número a confirmar con la Secretaría. Falta estimar cantidades y definir la retención con la Secretaría (D-09) |
| Tipo y tamaño validados en el backend | Cumple | El backend lee el tipo del contenido y corta el archivo que pasa del tamaño máximo, los dos según lo que declara el tipo de trámite para ese documento |
| Nombres generados por el sistema, fuera de la carpeta pública, y los privados servidos solo a quien corresponde | Cumple | Cada documento se guarda con el nombre de su hash SHA-256, que calcula el servidor, en una carpeta que el proxy no sirve. Solo lo descarga quien ve el trámite, a través de la API |
| Ruta del volumen documentada para los respaldos | Cumple | `/var/lib/sigta/documents` en la imagen, configurable con `SIGTA_DOCUMENTS_DIR` (ver "Configuración del servicio de trámites" en el README). Se respalda junto con la base, primero la base y después la carpeta: así cada documento que registra la base tiene su archivo, y los archivos de más los borra el backend al arrancar |

## Costos, mantenimiento y repositorio

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Lista de servicios externos con su costo y una alternativa gratuita | A confirmar | El asistente puede usar modelos por API, como Claude, que tienen costo. La alternativa prevista son modelos de pesos abiertos como Laya (ver ADR-018, "Inteligencia artificial" en el README y R-09) |
| Responsable de mantenimiento y soporte acordado antes de producción | A confirmar | Al terminar el proyecto, el equipo deja de estar disponible (ADR-013). Hay que acordarlo con la Subsecretaría de TICs y la Secretaría |
| Repositorio Git con acceso de lectura para la Subsecretaría de TICs, que se pueda copiar con su historial a su GitLab | Cumple | El repositorio es público en GitHub |
| Documentación de tecnologías, desarrollo, despliegue, arquitectura, operación y roles | No cumple | El README cubre tecnologías, arquitectura, configuración y roles. Faltan los recursos estimados, el despliegue, la operación (respaldos, logs y cómo verificar que el sistema está sano) y un `.env.example` |
| CHANGELOG o tags por versión | No cumple | No hay versiones etiquetadas |

## Datos personales (Ley 25.326)

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Relevamiento de los datos personales y sensibles que se guardan | No cumple | No está hecho. Los tipos de trámite piden, por ejemplo, el número de documento |
| Acceso restringido por rol y registro de quién consulta o modifica datos sensibles | A confirmar | La trazabilidad de los trámites está en el alcance. Falta definir si incluye las consultas |
| Derechos de acceso, rectificación y supresión | A confirmar | No está definido |
| Plazo de conservación y eliminación o anonimización | A confirmar | Pendiente con la Secretaría (D-09) |
| Sin datos reales en desarrollo y test | Cumple | ADR-005: en desarrollo y en las pruebas solo se usan datos sintéticos |
| Sin transferir datos personales a terceros, en especial fuera del país, sin acordarlo | A confirmar | ADR-018: a un modelo por API solo llega contenido académico, después de quitar los datos identificatorios dentro de SIGTA, y solo con proveedores que no entrenen con los datos y con retención nula o acotada. Falta elegir el proveedor y validarlo con la facultad |

## Buenas prácticas

No son condiciones excluyentes.

| Práctica | Estado | En SIGTA |
| --- | --- | --- |
| Imágenes oficiales con versión fija y multi-stage build | Cumple | `golang:1.26-alpine`, `node:24-alpine`, `nginx:1.30-alpine` y `distroless/static-debian13` |
| Procesos sin privilegios | No cumple | El backend corre como `nonroot`. El proxy usa la imagen de nginx por defecto, que corre como root |
| Endpoint de salud que verifique la conexión a la base | No cumple | No hay endpoint de salud. Cloud Run da por levantado el backend cuando responde `/procedure-types`, que lee la base |
| Logs a la salida estándar | Cumple | El backend escribe con `slog` en stdout |
| Tareas programadas y workers como servicios separados | No aplica | Todavía no hay tareas programadas ni colas |
| Sin secretos en el repositorio ni en el historial | A confirmar | Falta revisar el historial antes de entregarlo |
| CORS limitado a los dominios propios | Cumple | El frontend y la API comparten dominio, así que no hace falta habilitar CORS |
| Tests automatizados de las funciones principales | Cumple | La CI corre lint, tests, escenarios de aceptación y un arranque del sistema completo en cada pull request |
| Entornos de test y producción separados | A confirmar | Hay dos entornos de prueba en Google Cloud, cada uno con su base, su bucket de documentos, sus secretos y su cuenta de servicio: `prod` con `main` y `dev` con `develop` ([despliegue](15-despliegue.md)). Falta acordar los de FIUBA |
| Versionado semántico de entregas e imágenes | No cumple | No hay versiones |
| Accesibilidad WCAG 2.1 nivel AA y uso desde el celular | A confirmar | No está verificado |
