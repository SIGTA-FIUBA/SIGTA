# Requisitos de infraestructura de FIUBA

Qué cumple SIGTA de los requisitos que pide la Subsecretaría de TICs (SSTIC) para desplegar un sistema en la infraestructura de la facultad. El documento de SSTIC es "Requisitos para desplegar un sistema nuevo en la infraestructura FIUBA" ([original en Google Docs](https://docs.google.com/document/d/10q4lbU6mWsikQD9JoTu6fK5E3EHYFAfs/edit)). Este archivo no lo copia: resume cada requisito y lo contrasta con el estado de SIGTA al 2026-11-02.

**Estados.** Cumple: SIGTA ya lo hace. No cumple: SIGTA hace otra cosa o todavía no lo tiene. A confirmar: depende de algo que hay que preguntar o decidir. No aplica: el requisito no corresponde a SIGTA.

En resumen: los requisitos de autenticación y de contenedores ya se cumplen. Las brechas que afectan a la arquitectura son pgvector sobre PostgreSQL 11.9 (riesgo R-17) y el compose de referencia para Swarm.

## Contenedores e infraestructura

La infraestructura de FIUBA corre sobre Docker Swarm, con Traefik como proxy reverso.

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Un Dockerfile de producción por componente, sin código montado en volúmenes | Cumple | `services/backend/Dockerfile` y `deploy/nginx/Dockerfile`, que construye el frontend y lo sirve con el proxy |
| Un compose de referencia que Swarm soporte: sin `build:`, sin `depends_on` con condiciones, sin `container_name` ni `links`, sin puertos publicados en el host para servicios web | No cumple | `compose.yaml` es solo para desarrollo: usa `build:`, `depends_on` con `condition: service_healthy` y publica el puerto 8000 en el host. Falta un compose de referencia con imágenes ya construidas |
| Cada servicio HTTP expone un único puerto interno y no termina TLS | Cumple | El proxy escucha en el 80 y el backend en el 8080, los dos sin TLS |
| Las imágenes las construye SSTIC y se publican en su registry | A confirmar | El README prevé construir las imágenes en GitHub Actions al integrar en `main`. Hay que acordar con SSTIC quién las construye |
| Configuración por variables de entorno, sin valores de entorno escritos en la imagen | Cumple | Todo el backend se configura con variables `SIGTA_*` (ver "Configuración del servicio de trámites" en el README) |
| Contraseñas y claves como secretos del despliegue, nunca en el repositorio | Cumple | Las credenciales de Google y el archivo de usuarios son secretos del despliegue. Las contraseñas de `compose.yaml` son solo de desarrollo |
| Servicios sin estado, con sesiones, caché y colas en la base o en Redis, para correr más de una réplica | No cumple | Las sesiones y las cuentas registradas se guardan en memoria del backend: se pierden al reiniciar y no se comparten entre réplicas. Se resuelve al guardar usuarios y sesiones en PostgreSQL |
| Datos persistentes solo en volúmenes declarados y documentados | A confirmar | Hoy el único volumen es el de la base de desarrollo. Los documentos que adjunta el titular todavía no se guardan |
| La SPA no tiene la URL del backend fija en la imagen | Cumple | El frontend llama a la API por la ruta relativa `/api` (la opción que SSTIC prefiere) |

## Base de datos

SSTIC ofrece MariaDB 10.1.45 o PostgreSQL 11.9, como mínimo. El contenedor de base del compose es solo para desarrollo.

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Funcionar con PostgreSQL 11.9 o superior, avisando de antemano si se usan funciones más nuevas | No cumple | Se desarrolla sobre PostgreSQL 18, y el asistente usa la extensión pgvector (`deploy/postgres/sql/roles.sql`). Las versiones recientes de pgvector ya no soportan PostgreSQL 11, y no se sabe si el servidor de SSTIC tiene la extensión. Ver R-17 |
| Apuntar a un servidor de base externo por configuración | Cumple | `SIGTA_DATABASE_URL` |
| Esquema con migraciones versionadas en el repositorio | A confirmar | Todavía no hay tablas: el catálogo y los usuarios están en memoria. Falta elegir la herramienta de migraciones |
| Datos iniciales con scripts idempotentes, separados de los datos de prueba | No cumple | `deploy/postgres/init.sh` crea los roles y los esquemas solo al inicializar la base de desarrollo. Los usuarios sintéticos se cargan solo con `SIGTA_ENV=development` |
| Usuario de la aplicación con permisos solo sobre su base; documentar si las migraciones necesitan más | A confirmar | Cada servicio tiene su rol y su esquema (`backend` y `ai`). Crear la extensión `vector` necesita permisos de superusuario |
| UTF-8 y zona horaria documentada | A confirmar | No está documentada la zona horaria que se asume |

## Autenticación y autorización

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Ingreso con Google restringido al dominio de FIUBA | Cumple | ADR-010: solo se ingresa con la cuenta de Google de la UBA |
| Dominio validado en el backend (claim `hd` y correo verificado) | Cumple | `auth/adapters/google/identity.go` rechaza el token sin correo verificado o con otro `hd` |
| Client ID, client secret y URL de redirección por variables de entorno | Cumple | `SIGTA_GOOGLE_CLIENT_ID`, `SIGTA_GOOGLE_CLIENT_SECRET` y `SIGTA_PUBLIC_URL` |
| Credenciales de Google de test y producción provistas por SSTIC | A confirmar | Hoy se usa un cliente propio del equipo |
| Requisitos para un ingreso propio con contraseña | No aplica | SIGTA no guarda contraseñas ni ofrece otra forma de ingresar (ADR-010) |
| Roles y permisos verificados en el backend en cada endpoint | A confirmar | Los perfiles existen, pero todavía no hay endpoints que los exijan |

## Correo electrónico

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| SMTP de la facultad en test y producción; en desarrollo, una herramienta que capture los mensajes | No cumple | Solo existe el correo de desarrollo, que escribe los mensajes en el log. Fuera de desarrollo no hay adaptador de correo |
| Host, puerto, usuario, contraseña, cifrado y remitente por variables de entorno | No cumple | Se define con el adaptador SMTP |
| Envíos frecuentes a través de una cola, con el volumen estimado por día | A confirmar | Las notificaciones son por cambio de estado y pedido de acción, y la arquitectura prevé una cola de trabajos en PostgreSQL. Falta estimar el volumen |
| Sin servicios de envío externos pagos sin acordarlo | Cumple | No se usa ninguno |

## Archivos

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Estimación por tipo de archivo: formato, tamaño promedio y máximo, cantidad por año y retención | A confirmar | Cada tipo de trámite declara los formatos y el tamaño máximo de sus documentos. Falta estimar cantidades y definir la retención con la Secretaría |
| Tipo y tamaño validados en el backend | A confirmar | Se define con la carga de documentos, que todavía no existe |
| Nombres generados por el sistema, fuera de la carpeta pública, y los privados servidos solo a quien corresponde | A confirmar | Igual que el anterior |
| Ruta del volumen documentada para los respaldos | A confirmar | Igual que el anterior |

## Costos, mantenimiento y repositorio

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Lista de servicios externos con su costo y una alternativa gratuita | A confirmar | El asistente puede usar modelos por API, como Claude, que tienen costo. La alternativa prevista son modelos de pesos abiertos como Laya (ver "Inteligencia artificial" en el README y R-09) |
| Responsable de mantenimiento y soporte acordado antes de producción | A confirmar | Al terminar el proyecto, el equipo deja de estar disponible (ADR-013). Hay que acordarlo con SSTIC y la Secretaría |
| Repositorio Git con acceso de lectura para SSTIC, que se pueda copiar con su historial a su GitLab | Cumple | El repositorio es público en GitHub |
| Documentación de tecnologías, desarrollo, despliegue, arquitectura, operación y roles | No cumple | El README cubre tecnologías, arquitectura, configuración y roles. Faltan los recursos estimados, el despliegue, la operación (respaldos, logs y cómo verificar que el sistema está sano) y un `.env.example` |
| CHANGELOG o tags por versión | No cumple | No hay versiones etiquetadas |

## Datos personales (Ley 25.326)

| Requisito | Estado | En SIGTA |
| --- | --- | --- |
| Relevamiento de los datos personales y sensibles que se guardan | No cumple | No está hecho. Los tipos de trámite piden, por ejemplo, el número de documento |
| Acceso restringido por rol y registro de quién consulta o modifica datos sensibles | A confirmar | La trazabilidad de los trámites está en el alcance. Falta definir si incluye las consultas |
| Derechos de acceso, rectificación y supresión | A confirmar | No está definido |
| Plazo de conservación y eliminación o anonimización | A confirmar | No está definido |
| Sin datos reales en desarrollo y test | Cumple | ADR-005: no hay datos personales reales en ningún entorno |
| Sin transferir datos personales a terceros, en especial fuera del país, sin acordarlo | A confirmar | Un modelo por API que reciba documentos de titulares sería una transferencia al exterior. Hay que acordarlo con la facultad o usar un modelo local |

## Buenas prácticas

No son condiciones excluyentes.

| Práctica | Estado | En SIGTA |
| --- | --- | --- |
| Imágenes oficiales con versión fija y multi-stage build | Cumple | `golang:1.26-alpine`, `node:24-alpine`, `nginx:1.30-alpine` y `distroless/static-debian13` |
| Procesos sin privilegios | No cumple | El backend corre como `nonroot`. El proxy usa la imagen de nginx por defecto, que corre como root |
| Endpoint de salud que verifique la conexión a la base | No cumple | No hay endpoint de salud |
| Logs a la salida estándar | Cumple | El backend escribe con `slog` en stdout |
| Tareas programadas y workers como servicios separados | No aplica | Todavía no hay tareas programadas ni colas |
| Sin secretos en el repositorio ni en el historial | A confirmar | Falta revisar el historial antes de entregarlo |
| CORS limitado a los dominios propios | Cumple | El frontend y la API comparten dominio, así que no hace falta habilitar CORS |
| Tests automatizados de las funciones principales | Cumple | La CI corre lint, tests, escenarios de aceptación y un arranque del sistema completo en cada pull request |
| Entornos de test y producción separados | A confirmar | Hoy hay un solo entorno de demo |
| Versionado semántico de entregas e imágenes | No cumple | No hay versiones |
| Accesibilidad WCAG 2.1 nivel AA y uso desde el celular | A confirmar | No está verificado |
