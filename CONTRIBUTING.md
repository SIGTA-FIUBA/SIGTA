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
- Un documento sin tipos aceptados o sin tamaño máximo, un tipo que no es un tipo de medio en minúsculas y sin parámetros (como `application/pdf`), un comodín (`image/*`, `*/*`) o un tipo repetido.

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
- Un trámite guarda el número de su versión y también su hash, y los compara al leerla (`version.Pin` y `Pinned`): si la versión se perdió y se publicó otra con el mismo número, por ejemplo al restaurar un respaldo, el trámite no la toma. `Pinned` todavía no tiene quien la llame en producción: la va a usar la historia que inicia un trámite.

Las versiones se guardan en PostgreSQL (ver "Base de datos"); sin `SIGTA_DATABASE_URL`, en desarrollo, quedan en memoria. Las áreas, que lee de auth (ADR-021), todavía solo existen en memoria. El catálogo (`GET /procedure-types`, público) muestra la última versión de cada tipo de trámite y, como requisitos, solo lo que piden los pasos del titular; un requisito obligatorio en alguno de esos pasos es obligatorio. Si una versión no coincide con su hash o no se puede leer, el catálogo entero falla. En desarrollo (`SIGTA_ENV=development`) el servidor publica un trámite de ejemplo inventado, para que el catálogo tenga qué mostrar, solo si ese tipo de trámite todavía no tiene versiones. Todavía no hay un endpoint para publicar: hoy solo publica ese arranque en desarrollo (`publishDevProcedureType`, en `cmd/http.go`).

## Base de datos

Las tablas del servicio de trámites se crean con las migraciones de `services/backend/src/migrations/`, un archivo `.sql` numerado por migración que va dentro del binario (ADR-022). Las tablas son del rol `backend_migrations`; el rol de la aplicación, `backend`, solo tiene los permisos que le otorga cada migración.

- Una migración aplicada no se edita: un cambio es un archivo nuevo, con el número siguiente y un nombre como `0002_areas.sql`. Si se edita una ya aplicada, el servidor no arranca. Todas corren en una sola transacción, así que ninguna usa `COMMIT`, `ROLLBACK` ni `CONCURRENTLY`.
- Cada tabla nueva otorga a `backend` solo lo que la aplicación necesita. Hoy solo puede leer y agregar filas: ninguna tabla le da `UPDATE`, `DELETE` ni `TRUNCATE`. Si un cambio necesita actualizar una tabla, la migración otorga `UPDATE` solo sobre las columnas que cambian, nunca sobre `procedure_type_versions`, `procedure_events` ni la versión de un trámite. Ninguna clave foránea borra ni cambia filas por su cuenta (`ON DELETE CASCADE` y similares).
- Las tablas y columnas que la aplicación nunca cambia están en `immutable`, en `migrations/schema.go`, junto con `CheckApplicationRole`, el control que hace el servidor al arrancar. `TestOwnerAndPrivilegesOfEachRelation`, en `migrations/schema_test.go`, exige que cada tabla que solo crece esté en esa lista o en su mapa `granted`, con sus permisos y las columnas que la aplicación puede actualizar (`updatable`): una migración que otorga `UPDATE` sobre una columna suma esa columna ahí.
- Solo se usa lo que existe en PostgreSQL 11, la versión que ofrece la facultad (ver [requisitos de infraestructura](docs/12-requisitos-infraestructura-fiuba.md)).
- `procedure_type_versions` guarda el contenido en `bytea`, tal como se hasheó: `jsonb` reordena las claves. El hash se puede recalcular en la base con `'sha256:' || encode(sha256(content), 'hex')`. La fila de `procedure_types` se crea al publicar la versión 1. Que los números de versión no tengan huecos lo garantiza el adaptador, no la base.
- Un trámite (`procedures`) referencia su versión por tipo, número y hash, con una clave foránea. Sus eventos (`procedure_events`) solo se agregan: lo que pasa después es un evento nuevo. Por ahora las dos tablas tienen solo sus claves y la versión; el resto de sus columnas llega con las historias que inician y hacen avanzar un trámite.
- El servidor no migra: se conecta como `backend` y, si falta aplicar alguna migración o si su rol puede modificar o borrar las versiones, los eventos o la versión de un trámite, no arranca. Las aplica el comando `migrate` del mismo binario, con `SIGTA_MIGRATIONS_DATABASE_URL`; en Docker Compose lo corre el servicio `migrate` antes del backend.

`deploy/postgres/sql/roles.sql` crea los roles y esquemas de cada servicio solo al crear el volumen de la base. Si el volumen es anterior a la persistencia de los tipos de trámite, le falta el rol de migraciones: se recrea con `docker compose down --volumes`.

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

Si los roles no existen, las pruebas los crean sin permiso para iniciar sesión. `TestRolesOfTheDatabase`, que revisa lo que deja `roles.sql`, solo corre si la base tiene el esquema `backend`.

Sin la variable, esas pruebas se saltean; en la integración continua, donde está definida `CI`, fallan.

El contrato de un puerto de salida puede estar en `ports/output/outputtest`, como una suite que recibe una función que arma el adaptador, por ejemplo `outputtest.TestProcedureVersions`. Cada adaptador del puerto, en memoria o contra una base, la corre en su propio test.

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
