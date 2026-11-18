# Cómo contribuir

SIGTA lo desarrolla un equipo cerrado de cuatro integrantes. Esta guía resume la forma de trabajo.

## Flujo

1. Cada cambio parte de un ticket de Linear (acceso restringido al equipo).
2. Se trabaja en una rama `<tipo>/sig-<n>-<descripción>`, con el tipo de Conventional Commits y el número del ticket de Linear, por ejemplo `ci/sig-59-ci-on-develop`. No se usa el nombre que propone Linear.
3. Se abre un pull request contra `develop` con la plantilla del repositorio completa.
4. El pull request entra con la aprobación de otro integrante y la integración continua sin errores.

`develop` integra el trabajo del sprint; `main` recibe `develop` solo cuando hay una versión para mostrar, como una demo o una validación con la Secretaría.

## Convenciones

- **Idioma:** el código, los endpoints y los nombres técnicos van en inglés; la documentación, en español.
- **Comportamiento primero:** cada historia de usuario se especifica con escenarios Gherkin antes de implementarse. Van en inglés, en `features/`, con las etiquetas de la historia y del ticket (`@US14 @SIG-17`). Corren con godog dentro de `go test`, y sus pasos van en `services/backend/test/acceptance/`: un paso sin implementar hace fallar el pipeline, así que un `.feature` entra junto con sus pasos.
- **Contrato primero:** un cambio en la API empieza por la especificación OpenAPI.
- **Documentación en el mismo pull request:** si un cambio modifica el comportamiento, la arquitectura o la configuración de trámites, la documentación se actualiza junto con el código.
- **Decisiones:** toda decisión que condicione la arquitectura o el alcance se registra como ADR en [docs/04-decisiones-adr.md](docs/04-decisiones-adr.md). El README muestra solo el resultado vigente y se actualiza en el mismo pull request.
- **Datos:** nunca se suben datos personales reales; solo datos sintéticos o anonimizados.

## Modelo de tipo de trámite

Los tipos de trámite se van a editar desde la interfaz de SIGTA, sin tocar el código (ADR-013); la interfaz y su endpoint todavía no existen. El modelo está en `services/backend/src/procedure/domain/`: la definición y su validación en `definition/`, el contenido publicado y su versión en `version/`, y lo que muestra el catálogo en `catalog/` (ADR-016). Un tipo de trámite tiene un nombre, una descripción en lenguaje llano (`description`), sus plazos (`deadlines`) y condiciones de admisibilidad (`admissibility`), los datos (`fields`) y documentos (`documents`) que pide, y sus pasos; el primero es el inicio. Cada documento declara los tipos de archivo que acepta (`accepted_types`, como `application/pdf`) y su tamaño máximo en bytes (`max_size`), ambos obligatorios. Los datos y documentos son propios de cada tipo de trámite. Las áreas son del personal y las comparten todos los tipos: los pasos del personal las referencian por `id`, y los del titular no tienen área. Cada área tiene un nombre distinto. En producción los tipos de trámite se guardan en la base, como versiones. Los borradores de configuración, como los de los cinco trámites del relevamiento, no se suben al repositorio, porque describen procesos que la Secretaría no confirmó. Se validan desde `services/backend` con `PROCEDURE_DRAFTS=<carpeta> go test -run TestDrafts ./src/procedure/domain/definition/`, donde la carpeta tiene un JSON por tipo de trámite y las áreas que usan en `areas.json`.

Cada dato tiene un tipo: `text`, `long_text`, `number`, `date`, `option`, `degree` o `subject_list`. Un dato `option` lista sus opciones, cada una con `id` y nombre: lo que se guarda en un trámite es el `id`, así que corregir el nombre de una opción no cambia lo que significan las respuestas anteriores.

### Pasos

El trámite empieza en el primer paso. Cada paso tiene un rol (`role`), que dice quién lo hace, y una sola de cuatro acciones:

| Acción | Uso | Lleva |
| --- | --- | --- |
| `enter_data` | Cargar dato | `fields` y `optional_fields`, al menos uno |
| `upload_document` | Cargar documento | `documents` y `optional_documents`, al menos uno |
| `confirm_external_fact` | Confirmar un hecho ocurrido fuera de SIGTA, como una carátula o una firma en el expediente electrónico | `fields` y `optional_fields`, si hace falta registrar algo, como el número de expediente |
| `decide` | Elegir una salida | `outcomes`, al menos dos, cada una con su `next`; y, si hacen falta, los `fields` y `documents` que fundamentan la decisión, como el acta |

- Un paso es del titular (`role: applicant`) o del personal (`role: staff`). Las áreas son solo del personal: un paso del titular no lleva `area` ni `permissions`, y solo carga datos o documentos. Un paso del personal lleva `area`, una de las áreas de SIGTA (la Secretaría, `secretariat`, o una que creó el administrador), y `permissions`, los permisos de esa área que exige (al menos uno): para hacerlo hay que tenerlos todos. Los permisos los crea el administrador en cada área, con el nombre que quiera, y el código no les da significado: lo que distingue una aprobación es el permiso que exige, no la acción. Un paso no puede exigir los permisos de plataforma de la Secretaría (`edit_procedure_types` y `publish_procedure_types`), que habilitan editar y publicar tipos de trámite. Publicar una versión exige que quien la propone tenga el primero y quien la aprueba, el segundo.
- Un paso sin `next` ni `outcomes` cierra el trámite.
- Un paso que pide documentos (`documents`) no se confirma mientras alguno no tenga un archivo adjuntado a ese mismo paso, cualquiera sea su acción: también una decisión que se fundamenta en un acta. Los opcionales (`optional_documents`) no frenan la confirmación. La API responde 400 con un mensaje que nombra los documentos que faltan (`engine.MissingDocuments`). Los pasos iniciales del titular ya lo controlan al enviar el trámite (`submission.MissingDocuments`).
- Una confirmación registra que alguien declaró el hecho, no el hecho: una firma que pasa en el expediente electrónico queda como la confirmación de quien la vio. A confirmar: si cada firma se registra como un paso aparte, y si la confirmación guarda la fecha del hecho además de la de quien confirma.
- Un paso del personal tiene un área fija. A confirmar: si un paso puede tomar el área según el trámite, por ejemplo la Comisión Curricular de la carrera del titular.
- Los plazos y las condiciones de admisibilidad son texto que el catálogo muestra: nada los aplica todavía. A confirmar: si el motor tiene que aplicarlos, como pedía "Plazos y condiciones de admisibilidad", que figura en Linear como duplicado de "Modelo de definición de trámite".

No hay estados en la configuración: el estado que ve el titular se calcula a partir del paso abierto (un paso del titular es "Esperando al titular", uno del personal es "En" seguido del nombre de su área, y sin paso abierto es "Cerrado"). Tampoco hay estados de resultado: el dictamen es un documento que se carga, y las salidas de una decisión son pasos del proceso, no resultados.

Los pases entre áreas y los avisos al titular no son pasos: pasar al paso siguiente es el pase, y el aviso sale cuando cambia el estado. A confirmar: si la notificación de la resolución al titular, que tiene fecha y medio, es un paso del tipo de trámite o un evento del sistema.

### Validación

No se publica una definición inválida. El error lista juntos los problemas de datos, documentos y pasos; los de recorrido (pasos inalcanzables o que no cierran) aparecen una vez corregidos esos. Un área del personal inválida impide publicar cualquier tipo de trámite.

Se rechaza:

- `id` vacío, repetido o fuera del formato: minúsculas ASCII, dígitos y guiones bajos, empezando por una letra. Así dos `id` que se ven iguales son el mismo.
- Nombres vacíos o con espacios al principio o al final.
- Tipos de dato o roles desconocidos, un paso del personal sin área, sin permisos o con un permiso que su área no tiene, y un paso del titular con área o con permisos.
- Opciones fuera de un dato `option`, un dato `option` sin opciones, u opciones con `id` o nombre repetido.
- Salidas de una decisión con `id` o nombre repetido.
- Referencias a áreas, datos, documentos o pasos que no existen, o referencias repetidas en un paso.
- Datos o documentos que ningún paso usa.
- Una acción con lo que no le corresponde, como un documento en `enter_data`.
- Una salida que lleva a su propio paso.
- Pasos inalcanzables desde el inicio o que nunca llegan a cerrar el trámite.
- Una definición sin descripción, o un plazo o una condición de admisibilidad en blanco.
- Un documento sin tipos aceptados o sin tamaño máximo, un tipo que no es un tipo de medio en minúsculas y sin parámetros (como `application/pdf`), un comodín (`image/*`, `*/*`), un tipo que la subida no puede reconocer en un archivo (ver Documentos) o un tipo repetido.

Cada problema es un dato, no un texto: un código estable (`Code`, por ejemplo `unknown_next_step`), un `Path` que lo ubica (el paso `review` y dentro de él la salida `complete`) y el `Value` involucrado. El dominio no define cómo se escriben en JSON: eso lo hará el adaptador REST según el contrato. Vienen en el error de la validación (`ValidationError`), que se obtiene con `errors.As`. El editor muestra su propio texto para cada código, así que el idioma de la interfaz no depende del dominio; el texto en inglés del dominio es solo para logs y pruebas. Los códigos están en `definition/problem.go`. Un elemento sin `id`, o con uno inválido, se ubica sin `id` y el problema lleva el valor inválido.

### Versiones

La versión no se escribe: la asigna el servicio al publicar (ADR-012).

- Al publicar, el servicio arma el contenido del tipo de trámite (su definición más las áreas que usan sus pasos), lo codifica y calcula el hash SHA-256 de esos bytes. Si el contenido es igual al de la última versión, no publica nada y devuelve la última, con su propia autoría. Para comparar vuelve a codificar la última, así que un cambio en la codificación no genera una versión que no cambia nada.
- Quien publica indica la versión desde la que editó (0 para la primera). Si en el medio se publicó otra, se rechaza: hay que volver a editar desde la última. Si dos personas publican a la vez desde la misma versión, una la guarda y la otra se rechaza como desactualizada, aunque haya publicado el mismo contenido: al volver a publicar recibe la versión guardada.
- Cada versión la propone una persona y la aprueba otra (ADR-013); se comparan sin distinguir mayúsculas. La versión guarda a las dos y la fecha de publicación, en microsegundos, fuera del contenido hasheado. Hoy son textos; con la autenticación van a ser los identificadores de las personas autenticadas. A confirmar: si la autoría y la fecha entran en el hash.
- Una versión publicada no cambia. Guarda su contenido byte por byte, tal como se hasheó, así que se lee igual aunque después se renombre un área. El contenido lleva su número de formato (`format`), que sube cada vez que cambia `Content`, y una clave desconocida hace fallar la lectura: así un código que no conoce un formato no lo lee a medias. Cada `version/testdata/formatN.json` guarda los bytes de una versión de ese formato con todas sus claves y no se edita nunca; si su test falla, cambió el formato. Se siguen leyendo los formatos anteriores, y se convierten al formato actual al leerlos (`version/legacy.go`): una versión de formato 1 no tiene descripción, plazos, condiciones ni tipos y tamaño de sus documentos, y el catálogo la muestra sin esos datos. En los formatos 1 y 2 el titular era un área de tipo `applicant` y el resto de las áreas (`office` y `external`) eran del personal: al leerlos, cada paso toma su rol y la versión conserva solo las áreas del personal. Los bytes guardados no cambian, así que el hash sigue verificando. A confirmar: si las claves que recibe un formato se controlan contra las de ese formato, porque hoy una versión de formato 1 con claves del formato 2 se lee sin error.
- El hash se verifica cada vez que se lee una versión: al publicar sobre la última y al leerla para un trámite. También se controla que la fila sea la del tipo de trámite y el número pedidos.
- Cualquier cambio que se vea en el trámite genera una versión, incluso corregir un nombre.
- Un trámite en curso termina con la versión con la que empezó; los nuevos toman la última. No se migra un trámite en curso a otra versión.
- Previsto: un tipo de trámite se da de baja marcándolo inactivo; no se inician trámites nuevos y los que están en curso terminan con su versión.
- Un trámite guarda el número de su versión y también su hash, y los compara al leerla (`version.Pin` y `Pinned`): si la versión se perdió y se publicó otra con el mismo número, por ejemplo al restaurar un respaldo, el trámite no la toma. Iniciar un trámite guarda el hash de la versión con la que empieza, y cada lectura del trámite pasa por `Pinned`.

Las versiones se guardan en PostgreSQL (ver "Base de datos"); sin `SIGTA_DATABASE_URL`, en desarrollo, quedan en memoria. Las áreas, que lee de auth (ADR-021), todavía solo existen en memoria. El catálogo (`GET /procedure-types`, público) muestra la última versión de cada tipo de trámite y, como requisitos, solo lo que piden los pasos del titular; un requisito obligatorio en alguno de esos pasos es obligatorio. En `submission` muestra lo que se envía para iniciarlo (ver "Iniciar un trámite"). Si una versión no coincide con su hash o no se puede leer, el catálogo entero falla. En desarrollo (`SIGTA_ENV=development`) el servidor publica un trámite de ejemplo inventado, para que el catálogo tenga qué mostrar, solo si ese tipo de trámite todavía no tiene versiones. Todavía no hay un endpoint para publicar: hoy solo publica ese arranque en desarrollo (`publishIfMissing`, en `cmd/http.go`).

### Documentos

Cada documento que pide un paso declara sus tipos aceptados y su tamaño máximo (`accepted_types` y `max_size`). Quien actúa sobre el paso abierto adjunta el archivo con `POST /procedures/{procedureId}/documents`, un formulario multipart con dos partes, en este orden: `document`, el identificador del documento, y `file`, el archivo.

- El tipo sale del contenido, no del nombre ni del tipo que manda el navegador (`http.DetectContentType`). Por eso un documento solo puede aceptar los tipos que ese detector reconoce, listados en `definition.DetectedTypes`: entre otros `application/pdf`, `image/png`, `image/jpeg` y `text/plain`. Los de Word, Excel y OpenDocument todavía no, porque se detectan como `application/zip`. El tamaño se controla mientras se lee: un archivo que pasa del máximo se corta sin terminar de leerlo. Además, ningún pedido pasa de 25 MB (`api.MaxRequestBytes`), aunque el documento permita más, y el servidor deja de leer un pedido a los 5 minutos (`api.ReadTimeout`), así un cliente que manda el archivo de a un byte no ocupa la conexión indefinidamente. En el entorno local, nginx recibe el cuerpo completo antes de pasarlo al backend y corta si pasan 60 segundos sin recibir nada (sus valores por defecto); un despliegue sin ese proxy depende solo de `api.ReadTimeout`.
- El archivo se guarda en disco, en `SIGTA_DOCUMENTS_DIR`, con el nombre del hash SHA-256 de su contenido; el registro del documento guarda el hash, el tamaño, el tipo, el nombre original (limpio de rutas y de caracteres invisibles, y terminado en una extensión del tipo detectado, que reemplaza la de otro tipo: una imagen PNG llamada `foto.pdf` queda `foto.png`), el paso y quién lo subió. Un archivo nuevo para el mismo documento no borra el anterior: el actual es el último (`domain.Current`).
- Lista y descarga los documentos solo quien ve el trámite. La descarga le pide al navegador que guarde el archivo (`Content-Disposition: attachment`), que use el tipo detectado (`X-Content-Type-Options: nosniff`), que no ejecute nada si igual lo abre (`Content-Security-Policy: sandbox`) y que no lo guarde en caché (`Cache-Control: no-store`).
- La web los muestra sin salir de la página (`components/DocumentFile`): baja el archivo con la sesión y lo abre en un modal desde un blob, un PDF en un `iframe` y una imagen PNG o JPG en `Image`. Así el documento se ve sin salir de la página y la descarga no cambia. Solo se ofrece ver lo que el navegador muestra por sí mismo, un PDF solo si tiene visor (`navigator.pdfViewerEnabled`, que Chrome para Android no tiene), y lo que se muestra lo decide el tipo que manda el servidor, no el nombre del archivo. El resto se descarga.
- Adjuntar deja un evento en el historial (`ProcedureRepository.Record`), con el paso abierto, quién, cuándo y el documento que cubre (`domain.Event.Document`), sin avanzar el trámite. Todavía no se escribe en la misma transacción que el registro del documento: si el evento falla, el documento queda sin su entrada en el historial. SIG-75 los registra juntos (ADR-024).
- Los trámites y el registro de los documentos todavía están en memoria: se pierden al reiniciar. Al arrancar, el servidor borra los archivos que ningún documento registrado nombra y que tienen más de una hora, así que hasta que el registro se guarde en PostgreSQL borra los de la ejecución anterior. La hora deja fuera una subida que guardó el archivo y todavía no lo registró. Por ahora el servidor supone que es la única instancia: con dos réplicas, o con un despliegue que levanta la nueva antes de bajar la vieja, una borraría los archivos de los documentos que la otra tiene en memoria. Esta limpieza es para el disco local: con un storage aparte, los archivos siguen la saga de ADR-024 y no quedan sin documento.

### Iniciar un trámite

Quien tiene una sesión inicia un trámite a su nombre con `POST /procedures`, sea o no del personal. El trámite toma la última versión del tipo de trámite y lleva lo que piden sus **pasos iniciales**: los pasos del titular desde el primero, siguiendo `next`, hasta el primer paso del personal (`domain/submission/`). El catálogo los muestra aparte, en `submission`, porque `fields` y `documents` suman además los de pasos del titular posteriores, como una subsanación. Al enviarse, el trámite queda en el paso que sigue a los iniciales, el primero del personal.

- Es un formulario multipart con las partes en este orden: `procedure_type`, `values` (un JSON con un texto por campo, o una lista de textos para `subject_list`) y, por cada archivo, `document` seguido de `file`, como al adjuntar. Un campo vacío no se envía.
- Un envío rechazado no crea nada. Primero se validan los datos (`submission.CheckValues`): falta uno obligatorio, no es del tipo del campo o no lo piden los pasos iniciales. Un dato que se envía tiene que ser válido aunque sea opcional. Si los datos están bien, cada archivo se controla y se guarda como al adjuntar (`storeFile`, en `application/documents.go`). Después se controla que no falte ningún documento obligatorio. El trámite se registra al final, así que un envío rechazado deja como mucho archivos sin registro, que se borran al arrancar. Todavía no es atómico: si el registro falla después de guardar algún documento y antes del trámite, quedan documentos de un trámite que no existe, que nada borra. SIG-75 lo resuelve registrando trámite, valores y documentos en una sola transacción (ADR-024).
- Un envío rechazado responde 400 con `problems`: cada uno con su código, si es de un campo o de un documento, y su identificador. La web los muestra junto al campo o documento.
- El formato de cada tipo de campo: `text` y `degree`, texto de hasta 200 caracteres; `long_text`, hasta 5000; `number`, un decimal con punto (la web acepta también coma y la envía como punto); `date`, `AAAA-MM-DD`; `option`, el id de una opción; `subject_list`, hasta 100 textos no vacíos. Los límites están en `submission/values.go` y en `web/src/submission.ts`, que tienen que coincidir.
- En la web, el formulario está en `/catalog/{id}/start` y tiene tres etapas: datos, documentos y resumen. Hasta confirmar el resumen no se envía nada, así que reemplazar o quitar un archivo antes de enviar ocurre en el navegador. Si quedan datos cargados, salir del formulario pide confirmación. Como ningún pedido pasa de 25 MB (`api.MaxRequestBytes`), la web controla además que los archivos sumen menos que eso, aunque cada uno entre en su documento; si igual se pasa, el servidor responde 400 con un mensaje que lo dice. En `npm run dev`, `/dev/start-procedure/1?data=demo|worst|empty|one` muestra el formulario con datos de prueba extremos (`web/src/dev/`), para revisar cómo se ve con nombres largos, muchas opciones o nada que pedir.
- Los datos se guardan con el trámite (`ProcedureRepository.Values`), con el paso que los pidió y quién los cargó, y el detalle los muestra. Iniciar deja en el historial un evento por cada paso inicial, hecho por el titular y sin observación (`engine.Open`), que se registra junto con el trámite.

### Detalle de un trámite

`GET /procedures/{procedureId}` muestra un trámite a quien lo ve: su titular y el personal de las áreas por las que pasa. Para cualquier otra persona el trámite no existe, con la misma respuesta que uno que no existe. Trae los pasos de la versión con la que se inició, en el orden en que los declara (`domain/detail/`). Ese orden no es necesariamente el recorrido, porque una salida puede volver a un paso anterior: el recorrido lo dicen `next` y las salidas (`outcomes`) de cada paso. Cada paso dice quién lo hace (el titular, o `area` con el identificador y el nombre del área) y los datos (`fields`) y documentos que pide; uno que el paso pide como obligatorio y como opcional es obligatorio. Junto a los documentos de cada paso va el archivo vigente de cada uno (`files`), el último que se adjuntó para ese documento, y junto a los datos el último valor cargado de cada uno (`values`): `text`, o `items` en una lista de asignaturas, con quién lo cargó y cuándo. Un documento sin archivo o un dato sin valor no aparece. El caso de uso es `ProcedureDetailService.Detail`, que decide quién ve y quién actúa igual que `ProcedureService` y cruza los documentos y los valores con `detail.WithFiles` y `detail.WithValues`, así que la web no necesita pedirlos aparte.

El detalle dice además el estado del trámite (`state`), que se deriva del paso abierto en el dominio (`domain.StateOf`): `with_applicant` si espera al titular, `in_office` si lo tiene el personal y `closed` si no tiene paso abierto. No es el resultado del trámite, que no se modela. Si está derivado a otra área es `referred` (ADR-023). Un trámite cerrado no trae `open_step`, y nadie puede actuar en él (`can_act`). Si el paso abierto no está en la versión del trámite, el servidor responde 500. El titular (`applicant`) y quien subió cada archivo (`uploaded_by`) van con su identificador y su nombre, que el adaptador REST pide a `auth` por `People` (`auth/ports/input/`); el nombre queda vacío si la persona no está registrada.

El historial (`history`) son los pasos hechos, en el orden en que se registraron: los pasos iniciales, que el titular hace al iniciar el trámite, cada confirmación del personal, cualquiera sea su acción, y cada documento adjuntado después del inicio. Cada uno dice quién lo hizo, con su nombre, cuándo, la observación y, si corresponde, la salida elegida, la derivación, el retorno o el documento adjuntado (`document`, con el requisito que cubre y el nombre del archivo); el nombre del paso y su área están en `steps`. Son los eventos del trámite (`domain.Event`), que solo se agregan: la API no permite cambiarlos ni borrarlos, el repositorio devuelve copias y en PostgreSQL `procedure_events` no le da a la aplicación `UPDATE` ni `DELETE` (ADR-022).

En la web, el detalle está en `/procedures/{id}` (`pages/ProcedurePage.tsx`), adonde llevan la bandeja y, para el titular, el botón "Ver el trámite" al iniciarlo. Muestra el estado, el titular, el paso abierto con su responsable, la derivación con sus días sin retorno, los datos cargados (cada uno una vez, aunque lo pidan dos pasos), el historial (un `Timeline`, con la fecha y hora de Buenos Aires, donde un adjunto lleva su archivo) y los pasos con el archivo vigente de cada documento. Cada archivo se puede ver, si el navegador lo muestra, y descargar (ver Documentos). A quien puede hacer el paso abierto del personal (`can_act`) le muestra los documentos del paso, para adjuntar los que no tienen archivo de ese mismo paso (el único que lo deja confirmar) y reemplazar los que ya lo tienen, con `POST /procedures/{procedureId}/documents` (el archivo nuevo pasa a ser el vigente y el anterior queda en el historial); y le deja confirmarlo ahí mismo, con `POST /procedures/{procedureId}/steps/{stepId}/confirmation`: pide la observación, la salida si es una decisión y, solo si el paso siguiente es de otra área, el número de expediente externo. Al confirmar, vuelve a pedir el detalle y la bandeja, así el historial muestra la confirmación. Los nombres y colores del estado son los de la bandeja (`src/procedureState.ts`). En `npm run dev`, `/dev/procedures/dev-09?data=demo|worst|empty|one` muestra la página con datos de prueba extremos (`web/src/dev/`).

En desarrollo el servidor deja además dos trámites de ejemplo (`addDevExample`, en `cmd/http.go`): `dev-example`, esperando los documentos de su titular, para tener un trámite esperando al titular, que un trámite iniciado desde el formulario ya no tiene, y `dev-review`, esperando la revisión del personal, para tener un paso que confirmar, que puede confirmar `operativo@fi.uba.ar` con `POST /api/procedures/dev-review/steps/review/confirmation`. Con la identidad de desarrollo el titular es el de prueba; con Google es la primera cuenta que se registra después de arrancar, porque nadie entra como el titular de prueba. Si no se le puede dar el ejemplo, la cuenta entra igual, sin él, el servidor lo anota en el log y el ejemplo queda para la próxima cuenta que se registre.

Los trámites de ejemplo que ya pasaron la solicitud traen un motivo y un comentario cargados (`devValues`).

En desarrollo se publica además un segundo tipo de ejemplo, "Trámite de ejemplo con derivación" (`devReferralProcedureType`, tipo 2): su revisión es una decisión que deriva el trámite a la "Comisión de prueba" o lo devuelve al titular, el dictamen de la comisión lo devuelve al área de prueba (el retorno) y la comunicación del dictamen lo cierra. La comisión confirma con `comision@fi.uba.ar`. Es un tipo aparte, y no un cambio al primero, porque un ejemplo solo se publica si su tipo no tiene versiones (`publishIfMissing`): una base que ya tiene el primero lo conserva. Por lo mismo, un tercer tipo, "Trámite de ejemplo con dictamen" (`devRulingProcedureType`, tipo 3), es el de derivación con una comunicación que pide adjuntar el dictamen: `dev-09` espera el dictamen de la comisión, y después `operativo@fi.uba.ar` adjunta el documento desde la página del trámite.

### Mis trámites en curso

`GET /me/procedures` devuelve los trámites que inició quien tiene la sesión y que todavía no se cerraron, primero los que se movieron más recientemente. Vale también para el personal, que inicia trámites como cualquiera: ve los suyos, no los de sus áreas, que están en la bandeja. Cada uno dice su tipo de trámite, el paso abierto, el estado y cuándo empezó y se movió por última vez, y nada de lo que pasó dentro de cada área. Lo arma `ProcedureDetailService.Mine`, con el mismo estado (`domain.StateOf`) que el detalle y la bandeja.

En la web, es la primera sección del inicio (`pages/HomePage.tsx`), adonde se llega al ingresar. Cada trámite lleva a su detalle.

## Base de datos

Las tablas del servicio de trámites se crean con las migraciones de `services/backend/src/migrations/`, un archivo `.sql` numerado por migración que va dentro del binario (ADR-022). Las tablas son del rol `backend_migrations`; el rol de la aplicación, `backend`, solo tiene los permisos que le otorga cada migración.

- Una migración aplicada no se edita: un cambio es un archivo nuevo, con el número siguiente y un nombre como `0002_areas.sql`. Si se edita una ya aplicada, el servidor no arranca. Todas corren en una sola transacción, así que ninguna usa `COMMIT`, `ROLLBACK` ni `CONCURRENTLY`.
- Cada tabla nueva otorga a `backend` solo lo que la aplicación necesita. Hoy solo puede leer y agregar filas: ninguna tabla le da `UPDATE`, `DELETE` ni `TRUNCATE`. Si un cambio necesita actualizar una tabla, la migración otorga `UPDATE` solo sobre las columnas que cambian, nunca sobre `procedure_type_versions`, `procedure_events` ni la versión de un trámite. Ninguna clave foránea borra ni cambia filas por su cuenta (`ON DELETE CASCADE` y similares).
- Las tablas y columnas que la aplicación nunca cambia están en `immutable`, en `migrations/schema.go`, junto con `CheckApplicationRole`, el control que hace el servidor al arrancar. `TestOwnerAndPrivilegesOfEachRelation`, en `migrations/schema_test.go`, exige que cada tabla que solo crece esté en esa lista o en su mapa `granted`, con sus permisos y las columnas que la aplicación puede actualizar (`updatable`): una migración que otorga `UPDATE` sobre una columna suma esa columna ahí.
- Solo se usa lo que existe en PostgreSQL 11, la versión que ofrece la facultad (ver [requisitos de infraestructura](docs/12-requisitos-infraestructura-fiuba.md)).
- `procedure_type_versions` guarda el contenido en `bytea`, tal como se hasheó: `jsonb` reordena las claves. La base rechaza una fila cuyo hash no sea `'sha256:' || encode(sha256(content), 'hex')` (restricción `digest_is_of_content`). La fila de `procedure_types` se crea al publicar la versión 1. Que los números de versión no tengan huecos lo garantiza el adaptador, no la base.
- Un trámite (`procedures`) referencia su versión por tipo, número y hash, con una clave foránea. Sus eventos (`procedure_events`) solo se agregan: lo que pasa después es un evento nuevo. Por ahora las dos tablas tienen solo sus claves y la versión; el resto de sus columnas llega con las historias que inician y hacen avanzar un trámite.
- El servidor no migra: se conecta como `backend` (o con un usuario miembro de ese rol, que toma con `SET ROLE`, igual que `migrate` con `backend_migrations`) y, si falta aplicar alguna migración o si su rol o el usuario con el que se conecta pueden modificar o borrar las versiones, los eventos o la versión de un trámite, no arranca. Por eso ese usuario no puede ser un superusuario ni miembro de `backend_migrations`, aunque tome `backend` con `SET ROLE`. Las aplica el comando `migrate` del mismo binario, con `SIGTA_MIGRATIONS_DATABASE_URL`; en Docker Compose lo corre el servicio `migrate` antes del backend.

`deploy/postgres/sql/roles.sql` crea los roles y esquemas de cada servicio solo al crear el volumen de la base. Si el volumen es anterior a la persistencia de los tipos de trámite, le falta el rol de migraciones: se recrea con `docker compose down --volumes`. El comando `migrate` lo sugiere solo con `SIGTA_ENV=development`, porque en otro entorno el mismo error viene de una contraseña o un permiso mal configurado.

### Pruebas contra PostgreSQL

Los adaptadores de PostgreSQL y las migraciones se prueban contra una base real. Las pruebas toman la base de `SIGTA_TEST_DATABASE_URL`, con un usuario administrador, crean un esquema propio por prueba y consultan con los permisos de `backend` o de `backend_migrations`. Con Docker Compose:

```sh
docker compose up --detach --wait db
cd services/backend
SIGTA_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/sigta?sslmode=disable' go test -race ./...
```

`-race` es opcional y necesita un compilador de C; sin él, las pruebas corren igual sin esa opción.

Sin Docker sirve cualquier PostgreSQL local. Con los binarios de PostgreSQL en el `PATH`, desde la raíz del repositorio:

```sh
initdb -D /tmp/sigta-pg -U postgres --auth=trust
pg_ctl -D /tmp/sigta-pg -o '-p 5433 -k /tmp' -l /tmp/sigta-pg.log start
createdb -h /tmp -p 5433 -U postgres sigta
cd services/backend
SIGTA_TEST_DATABASE_URL='postgres://postgres@localhost:5433/sigta?host=/tmp' go test ./...
```

Si los roles no existen, las pruebas los crean sin permiso para iniciar sesión. Las que revisan el control del rol de la aplicación se conectan, como el servidor, con usuarios propios que solo son miembros de esos roles, y los borran al terminar. `TestRolesOfTheDatabase`, que revisa lo que deja `roles.sql`, solo corre si la base tiene el esquema `backend`.

Sin la variable, esas pruebas se saltean; en la integración continua, donde está definida `CI`, fallan.

El contrato de un puerto de salida puede estar en `ports/output/outputtest`, como una suite que recibe una función que arma el adaptador, por ejemplo `outputtest.TestProcedureTypeVersionRepository`. Cada adaptador del puerto, en memoria o contra una base, la corre en su propio test.

La bandeja del personal (`GET /procedures`) lee los mismos trámites que el detalle (`ProcedureRepository`) y decide quién ve y quién actúa con las mismas reglas (`canSee` y `canAct`, ADR-020), así que la bandeja y el detalle le responden lo mismo a la misma persona. El caso de uso es `application.InboxService`; lo que arma la lista está en `services/backend/src/procedure/domain/inbox/`:

- Solo la ve el personal, quien tiene al menos una membresía vigente: el resto recibe 403.
- Muestra los trámites en curso que pasan por las áreas de quien mira, dentro del alcance de sus membresías. Deja afuera los cerrados y los propios: nadie avanza su propio trámite como personal. Si el personal ve todo lo que pasa por su área o solo lo que tiene un paso suyo abierto sigue a confirmar (D-10).
- El estado es el del detalle (`domain.StateOf`): `with_applicant`, `in_office` o `referred` si está derivado a otra área (ADR-023). No es el resultado del trámite (ADR-009).
- Un trámite requiere acción si quien mira tiene los permisos que exige el paso abierto, en su área y para ese tipo de trámite.
- `view=mine` muestra lo que requiere acción; `view=all`, todo lo visible, marcando lo accionable. Los filtros (tipo, estado, área) se combinan, y las opciones de filtro salen de la vista sin filtrar. Un paso del titular no tiene área.
- La antigüedad se cuenta desde que el trámite empezó (`StartedAt`) y el último movimiento es cuando llegó al paso abierto (`MovedAt`). Los dos los pone el trámite al iniciarse; cuando los trámites se guarden en PostgreSQL, la tabla `procedures` necesita esas dos columnas.

En desarrollo el servidor carga además nueve trámites inventados de los tres tipos de ejemplo, en distintos pasos, con titulares sintéticos y con el historial de cómo llegaron ahí (`src/cmd/dev_procedures.go`), para que la bandeja tenga qué mostrar: `dev-07` espera la decisión de la revisión, y `dev-08` y `dev-09` están derivados a la comisión. La pantalla de la bandeja pide una sola vez todo lo visible y filtra, ordena y cuenta en el navegador; los filtros de la API quedan para otras vistas. Los filtros de la pantalla viven en la URL, así se conservan al ir a un trámite y volver.

## Commits

Los commits siguen [Conventional Commits](https://www.conventionalcommits.org/), en inglés y en una sola línea de hasta 72 caracteres:

```text
feat(engine): add step confirmation
fix(api): return 404 for unknown procedures
docs: describe the ports and adapters
```

Tipos: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore` y `revert`.

La convención la valida [commitlint](https://commitlint.js.org/) con la configuración convencional, mediante un hook de [husky](https://typicode.github.io/husky/) que se instala solo al correr `npm install` en la raíz del repositorio. En cada pull request, la integración continua valida el título con la misma configuración.

Los pull requests se integran con squash merge, así que el título del pull request es el commit que queda en `develop` y sigue la misma convención.
