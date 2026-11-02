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

Los tipos de trámite se van a editar desde la interfaz de SIGTA, sin tocar el código (ADR-013); la interfaz y su endpoint todavía no existen. El modelo está en `services/backend/src/procedure/domain/`. Un tipo de trámite tiene un nombre, una descripción en lenguaje llano (`description`), sus plazos (`deadlines`) y condiciones de admisibilidad (`admissibility`), los datos (`fields`) y documentos (`documents`) que pide, y sus pasos; el primero es el inicio. Cada documento declara los tipos de archivo que acepta (`accepted_types`, como `application/pdf`) y su tamaño máximo en bytes (`max_size`), ambos obligatorios. Los datos y documentos son propios de cada tipo de trámite. Las áreas son de la facultad y las comparten todos los tipos: los pasos las referencian por `id`. Cada área tiene un nombre distinto, y hay una sola del titular (`applicant`). En producción los tipos de trámite se guardan en la base, como versiones. Los borradores de configuración, como los de los cinco trámites del relevamiento, no se suben al repositorio, porque describen procesos que la Secretaría no confirmó. Se validan desde `services/backend` con `PROCEDURE_DRAFTS=<carpeta> go test -run TestDrafts ./src/procedure/domain/`, donde la carpeta tiene un JSON por tipo de trámite y las áreas que usan en `areas.json`.

Cada dato tiene un tipo: `text`, `long_text`, `number`, `date`, `option`, `degree` o `subject_list`. Un dato `option` lista sus opciones, cada una con `id` y nombre: lo que se guarda en un trámite es el `id`, así que corregir el nombre de una opción no cambia lo que significan las respuestas anteriores.

### Pasos

El trámite empieza en el primer paso. Cada paso tiene un área responsable y una sola de cuatro acciones:

| Acción | Uso | Lleva |
| --- | --- | --- |
| `enter_data` | Cargar dato | `fields` y `optional_fields`, al menos uno |
| `upload_document` | Cargar documento | `documents` y `optional_documents`, al menos uno |
| `confirm_external_fact` | Confirmar un hecho ocurrido fuera de SIGTA, como una carátula o una firma en el expediente electrónico | `fields` y `optional_fields`, si hace falta registrar algo, como el número de expediente |
| `decide` | Elegir una salida | `outcomes`, al menos dos, cada una con su `next`; y, si hacen falta, los `fields` y `documents` que fundamentan la decisión, como el acta |

- Los pasos de un área de tipo `office` o `external` llevan `profile`. Por convención, `staff_operative` es para una tarea y `staff_approver` para una aprobación; el código no lo valida, y lo que distingue una aprobación es el perfil, no la acción. Los pasos del titular no llevan perfil y solo cargan datos o documentos.
- Un paso sin `next` ni `outcomes` cierra el trámite.
- Una confirmación registra que alguien declaró el hecho, no el hecho: una firma que pasa en el expediente electrónico queda como la confirmación de quien la vio. A confirmar: si cada firma se registra como un paso aparte, y si la confirmación guarda la fecha del hecho además de la de quien confirma.
- Un paso tiene un área fija. A confirmar: si un paso puede tomar el área según el trámite, por ejemplo la Comisión Curricular de la carrera del titular.
- Los plazos y las condiciones de admisibilidad son texto que el catálogo muestra: nada los aplica todavía. A confirmar: si el motor tiene que aplicarlos, como pedía "Plazos y condiciones de admisibilidad", que figura en Linear como duplicado de "Modelo de definición de trámite".

No hay estados en la configuración: el estado que ve el titular se calcula a partir del tipo de área del paso abierto (`applicant` es "Esperando al titular", `office` es "En el área", `external` es "Derivado", y sin paso abierto es "Cerrado"). Tampoco hay estados de resultado: el dictamen es un documento que se carga, y las salidas de una decisión son pasos del proceso, no resultados.

Los pases entre áreas y los avisos al titular no son pasos: pasar al paso siguiente es el pase, y el aviso sale cuando cambia el estado. A confirmar: si la notificación de la resolución al titular, que tiene fecha y medio, es un paso del tipo de trámite o un evento del sistema.

### Validación

No se publica una definición inválida. El error lista juntos los problemas de datos, documentos y pasos; los de recorrido (pasos inalcanzables o que no cierran) aparecen una vez corregidos esos. Un área inválida del catálogo impide publicar cualquier tipo de trámite.

Se rechaza:

- `id` vacío, repetido o fuera del formato: minúsculas ASCII, dígitos y guiones bajos, empezando por una letra. Así dos `id` que se ven iguales son el mismo.
- Nombres vacíos o con espacios al principio o al final.
- Tipos de dato o de área desconocidos, perfiles desconocidos o un paso de `office` o `external` sin perfil.
- Opciones fuera de un dato `option`, un dato `option` sin opciones, u opciones con `id` o nombre repetido.
- Salidas de una decisión con `id` o nombre repetido.
- Referencias a áreas, datos, documentos o pasos que no existen, o referencias repetidas en un paso.
- Datos o documentos que ningún paso usa.
- Una acción con lo que no le corresponde, como un documento en `enter_data`.
- Una salida que lleva a su propio paso.
- Pasos inalcanzables desde el inicio o que nunca llegan a cerrar el trámite.
- Una definición sin descripción, o un plazo o una condición de admisibilidad en blanco.
- Un documento sin tipos aceptados o sin tamaño máximo, un tipo que no es un tipo de medio en minúsculas y sin parámetros (como `application/pdf`), un comodín (`image/*`, `*/*`) o un tipo repetido.

Cada problema es un dato, no un texto: un `Code` estable (por ejemplo `unknown_next_step`), un `Path` que lo ubica (el paso `review` y dentro de él la salida `complete`) y el `Value` involucrado. El dominio no define cómo se escriben en JSON: eso lo hará el adaptador REST según el contrato. `ProblemsOf` los saca del error de la validación. El editor muestra su propio texto para cada código, así que el idioma de la interfaz no depende del dominio; el texto en inglés del dominio es solo para logs y pruebas. Los códigos están en `problem.go`. Un elemento sin `id`, o con uno inválido, se ubica sin `id` y el problema lleva el valor inválido.

### Versiones

La versión no se escribe: la asigna el servicio al publicar (ADR-012).

- Al publicar, el servicio arma el contenido del tipo de trámite (su definición más las áreas que usan sus pasos), lo codifica y calcula el hash SHA-256 de esos bytes. Si el contenido es igual al de la última versión, no publica nada y devuelve la última, con su propia autoría. Para comparar vuelve a codificar la última, así que un cambio en la codificación no genera una versión que no cambia nada.
- Quien publica indica la versión desde la que editó (0 para la primera). Si en el medio se publicó otra, se rechaza: hay que volver a editar desde la última. Si dos personas publican a la vez desde la misma versión, una la guarda y la otra se rechaza como desactualizada, aunque haya publicado el mismo contenido: al volver a publicar recibe la versión guardada.
- Cada versión la propone una persona y la aprueba otra (ADR-013); se comparan sin distinguir mayúsculas. La versión guarda a las dos y la fecha de publicación, en microsegundos, fuera del contenido hasheado. Hoy son textos; con la autenticación van a ser los identificadores de las personas autenticadas. A confirmar: si la autoría y la fecha entran en el hash.
- Una versión publicada no cambia. Guarda su contenido byte por byte, tal como se hasheó, así que se lee igual aunque después se renombre un área. El contenido lleva su número de formato (`format`), que sube cada vez que cambia `Content`, y una clave desconocida hace fallar la lectura: así un código que no conoce un formato no lo lee a medias. Cada `testdata/formatN.json` guarda los bytes de una versión de ese formato con todas sus claves y no se edita nunca; si su test falla, cambió el formato. Se siguen leyendo los formatos anteriores: una versión de formato 1 no tiene descripción, plazos, condiciones ni tipos y tamaño de sus documentos, y el catálogo la muestra sin esos datos. A confirmar: si un formato anterior se convierte al leerlo y si las claves que recibe se controlan contra las de su formato, porque hoy una versión de formato 1 con claves del formato 2 se lee sin error.
- El hash se verifica cada vez que se lee una versión: al publicar sobre la última y al leerla para un trámite. También se controla que la fila sea la del tipo de trámite y el número pedidos.
- Cualquier cambio que se vea en el trámite genera una versión, incluso corregir un nombre.
- Un trámite en curso termina con la versión con la que empezó; los nuevos toman la última. No se migra un trámite en curso a otra versión.
- Previsto: un tipo de trámite se da de baja marcándolo inactivo; no se inician trámites nuevos y los que están en curso terminan con su versión.

Hoy las versiones y las áreas solo existen en memoria. El catálogo (`GET /procedure-types`, público) muestra la última versión de cada tipo de trámite y, como requisitos, solo lo que piden los pasos del área del titular; un requisito obligatorio en alguno de esos pasos es obligatorio. Si una versión no coincide con su hash o no se puede leer, el catálogo entero falla. En desarrollo (`SIGTA_ENV=development`) el servidor publica un trámite de ejemplo inventado, para que el catálogo tenga qué mostrar. La persistencia en PostgreSQL y el almacenamiento de documentos son de otro ticket.

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
