# Contexto para agentes

SIGTA (Sistema Integral de Gestión de Trámites Académicos) es una plataforma web para iniciar, gestionar y seguir los trámites de la Secretaría de Gestión Académica de la FIUBA, con tres componentes de IA que proponen y nunca deciden. Es un Trabajo Profesional de Ingeniería en Informática: cuatro integrantes, sprints de dos semanas, cierre en mayo de 2027.

Este archivo resume lo que un agente necesita para trabajar en el repositorio. El detalle está en el [README](README.md), en [CONTRIBUTING.md](CONTRIBUTING.md) y en [docs/](docs/README.md). Se actualiza en el mismo PR que cambia una convención. Dónde está el proyecto (hitos, sprint actual y orden de lo que viene) no se repite acá: está en [docs/10-roadmap.md](docs/10-roadmap.md) y [docs/11-plan-de-sprints.md](docs/11-plan-de-sprints.md).

## Reglas que no se negocian

1. **La IA propone, una persona decide.** Ningún componente aprueba, rechaza ni cambia el estado de un trámite.
2. **El resultado de un trámite no se modela ni se interpreta** (ADR-009). Reconocimiento total, parcial y rechazo son el mismo evento: un dictamen que se adjunta y se comunica. No agregar estados ni campos de resultado.
3. **Sin datos personales reales** en ningún entorno, fixture, test ni commit (ADR-005). Solo datos sintéticos o anonimizados.
4. **Los trámites son configuración** que se va a editar desde la interfaz, y cada publicación es una versión que no se modifica (ADR-012, ADR-013). Sumar un trámite no debe requerir código; si lo requiere, se corrige la abstracción en vez de parchear (ADR-006).
5. **Los sistemas externos se registran, no se integran** (ADR-004). No hay integración con SIU Guaraní ni con el expediente electrónico: el personal confirma en SIGTA lo que hizo afuera.
6. **Todo documento aportado por un titular es contexto no confiable** para los modelos (riesgo R-07).
7. **No hay firma digital con validez legal** ni dictamen automático de equivalencias.

## Qué no se delega a un agente

Según la política de uso de US02, un agente puede proponer pero no define por su cuenta:

- La configuración de los tipos de trámite.
- Los prompts de producción de los componentes de IA.
- Los criterios de aceptación de las historias.

Todo lo que produce un agente pasa por la revisión de un par, y la autoría del PR es humana.

## Vocabulario

Usar los términos del [glosario](docs/01-glosario.md) en código, tickets, escenarios y documentación. Los que más se confunden:

| Término | En código y escenarios | Significa | No confundir con |
| --- | --- | --- | --- |
| Tipo de trámite | procedure type | Definición configurada de un trámite | — |
| Trámite | procedure | Instancia de un tipo de trámite, iniciada por un titular | "Request", que se reserva para HTTP |
| Titular | applicant | Quien inicia el trámite y a cuyo nombre corre | "Estudiante": no todo titular lo es |
| Secretario | staff | Rol del personal: quien tiene al menos una membresía vigente (ADR-020) | Titular: los roles se excluyen y el personal no inicia trámites |
| Membresía | membership | Pertenencia de una persona a un área, con algunos permisos de esa área, alcance opcional por tipo de trámite y vigencia | Una versión: la membresía no cambia al publicar |
| Administrador | administrator | Marca sobre una persona que le permite dar de alta membresías | Un rol: no da acceso a los trámites |
| Paso | step | Unidad de avance, con responsable y acción | Estado del trámite, que se deriva del paso abierto |
| Requisito | `field` o `document` del tipo de trámite | Dato, documento o condición que el trámite exige | Condición académica, que el modelo solo guarda como texto (`admissibility`) y nada aplica |
| Área | area | Parte de la organización a la que pertenece el personal, creada como dato; un paso del personal es de un área | Rol: el titular no tiene área, y cada paso dice su rol (`applicant` o `staff`) |
| Salida | outcome | Opción de una decisión, con el paso al que lleva | Resultado del trámite, que no se modela |
| Derivación | referral | Salida del trámite hacia un evaluador externo | Devolución |
| Retorno | referral return | Vuelta del trámite desde el evaluador, con el dictamen adjunto | Resultado, que no se modela |
| Dictamen | ruling | Lo que resuelve el evaluador; se adjunta y se comunica | Un estado del trámite |
| Devolución | send back | Vuelta a un paso anterior, decidida por la Secretaría | Subsanación |
| Subsanación | correction request | Pedido puntual al titular, sin cerrar ni reiniciar el trámite | Devolución |
| Preacta | draft record | Borrador generado desde una plantilla; siempre es borrador | Acta (record) |
| Evaluador externo | external evaluator | Comisión o unidad académica que dictamina; no usa la plataforma | Un rol de SIGTA |
| Hallazgo | finding | Resultado de la validación asistida sobre un requisito | Una decisión |

Las acciones del modelo de dominio son **cargar dato** (`enter_data`), **cargar documento** (`upload_document`), **confirmar hecho externo** (`confirm_external_fact`) y **decidir** (`decide`). Una decisión elige una de sus salidas, cada una con su paso siguiente; continuar, devolver o pedir información son salidas posibles, no acciones. Derivar es pasar a un paso de otra área; si quien dictamina no usa SIGTA, el paso es del área que registra el dictamen.

## Arquitectura

Monolito modular en Go más un servicio de IA en Python, todo sobre PostgreSQL con pgvector. Cada servicio sigue **puertos y adaptadores**: el núcleo define interfaces y los adaptadores dependen de él, nunca al revés.

- **Entre módulos, la API pública** (ADR-021). Un módulo del backend usa a otro solo por su `ports/input/`, recibida por constructor, sin puerto ni adaptador propio; nunca importa su `domain/`, `application/` ni `adapters/` (lo rechaza `depguard`). `procedure` usa `auth`, nunca al revés.
- **Contrato primero.** Todo cambio de API empieza en `contracts/openapi.yaml`. Después se regenera con `make contracts`; la CI falla si el código generado no coincide con el contrato. No editar a mano `api.gen.go` ni `schema.d.ts`.
- **Historial de acciones** en una tabla que solo crece, escrita en la misma transacción que el estado. Sin event sourcing.
- **Cola de trabajos en PostgreSQL**, encolada en la misma transacción que el cambio y tomada con `FOR UPDATE SKIP LOCKED`. Cada trabajo debe ser idempotente.
- **Cada servicio es dueño de sus datos.** Comparten la instancia de PostgreSQL pero no los esquemas; ninguno lee las tablas del otro.
- **Identidad:** se ingresa solo con la cuenta de Google de la UBA (ADR-010). SIGTA no guarda contraseñas; áreas, permisos y membresías se administran en SIGTA. En local y en pruebas se usa un adaptador de identidad falso.
- **Interfaz:** componentes de Mantine, con el aspecto definido solo en `web/src/theme/theme.ts` (ADR-015). Las reglas de diseño y su motivo están en [web/DESIGN.md](web/DESIGN.md), y el lint rechaza estilos, colores y medidas fuera del tema.
- **No sumar componentes** (NATS, S3, gateway, observabilidad completa) sin que se cumpla su disparador, listado en la sección "Evolución" del README.

## Estructura del repositorio

| Ruta | Contenido |
| --- | --- |
| `contracts/openapi.yaml` | Contrato de la API pública, fuente del código generado |
| `services/backend/src/` | Servicio de trámites en Go, dividido por módulo (`auth/`, `procedure/`) y, dentro de cada uno, por componente: `domain/`, `application/` (casos de uso), `ports/input/` y `ports/output/`, `adapters/`. En `procedure/`, `domain/` se divide en `definition/`, `version/` y `catalog/` (ADR-016). Además, `api/` (código generado desde el contrato y montaje de la API) y `cmd/` (arranque y configuración) |
| `services/backend/test/acceptance/` | Pasos de los escenarios de `features/`, con godog |
| `web/src/` | Frontend React + TypeScript con Mantine: `api/` (cliente generado), `theme/` (el único lugar del aspecto), `components/`, `layout/` (encabezado y menú de la cuenta), `pages/`. Los tests van aparte, en `web/tests/`, con la misma estructura. Reglas en `web/DESIGN.md` |
| `deploy/` | nginx de desarrollo y scripts de inicialización de PostgreSQL (roles por servicio) |
| `docs/` | Artefactos de gestión: alcance, ADR, riesgos, backlog, plan. Ver [docs/README.md](docs/README.md) |
| `features/` | Escenarios Gherkin de cada historia y su matriz de trazabilidad |
| `.claude/skills/` | Rutinas para las tareas recurrentes (ver más abajo) |
| `compose.yaml` | Entorno local completo |

El servicio de IA en Python todavía no existe en el repositorio.

## Comandos

| Tarea | Comando |
| --- | --- |
| Levantar todo | `docker compose up --build` (http://localhost:8000, API en `/api/`) |
| Regenerar código desde el contrato | `make contracts` |
| Backend: tests y lint | `cd services/backend && go test -race ./...` y `golangci-lint run` |
| Frontend: lint, tipos, tests | `cd web && npm run lint && npm run format:check && npm run typecheck && npm test` |
| Instalar el hook de commits | `npm install` en la raíz |

## Convenciones

- **Idioma:** código, endpoints y nombres técnicos en inglés; documentación en español.
- **Commits y títulos de PR:** Conventional Commits, en inglés, una línea de hasta 72 caracteres (`feat(engine): add step confirmation`). Los valida commitlint en el hook y en la CI. Los PR se integran con squash, así que el título es el commit final.
- **Ramas:** `<tipo>/sig-<n>-<descripción>`, con el tipo de Conventional Commits y el número del ticket (`ci/sig-59-ci-on-develop`, `docs/sig-10-agent-context`). No se usa el nombre que propone Linear. Los PR van contra `develop`; `main` solo recibe `develop` cuando hay una versión para mostrar.
- **Comportamiento primero:** cada historia se especifica con escenarios Gherkin antes de implementarse, en inglés, en `features/`, con las etiquetas `@USxx @SIG-n`. Corren con godog dentro de `go test`, y un paso sin implementar hace fallar el pipeline, así que un `.feature` entra junto con sus pasos.
- **Pruebas:** el dominio se prueba con adaptadores en memoria; cada adaptador, contra la dependencia real en un contenedor.
- **Documentación en el mismo PR** (ADR-007): si cambia comportamiento, arquitectura o configuración de trámites, se actualizan docs, README del módulo y este archivo. Un PR sin su documentación no cumple la definición de terminado.
- **Decisiones:** toda decisión que condicione la arquitectura o el alcance se registra como ADR en `docs/04-decisiones-adr.md`, y el README se actualiza con el resultado vigente en el mismo PR. Un ADR no se borra: se marca como reemplazado.
- **Componentes de IA:** cada llamada a un modelo registra costo, latencia y versión del prompt. Los modelos van detrás de puertos; en la CI se usan adaptadores deterministas.

## Rutinas

Las tareas recurrentes tienen una rutina en `.claude/skills/`. Seguirla cuando se pida la tarea correspondiente:

| Rutina | Para |
| --- | --- |
| `gherkin-scenarios` | Redactar los escenarios de una historia desde sus criterios de aceptación |
| `new-procedure-type` | Preparar el borrador de configuración de un tipo de trámite nuevo |
| `open-pr` | Armar la rama, los commits y el PR con la plantilla completa |
| `break-ui` | Probar una pantalla con datos extremos (nombres largos, listas vacías, un solo elemento) antes de darla por terminada |
| `web-design-guidelines` | Revisar la accesibilidad, el foco, los formularios y los textos de una pantalla |
| `mantine-form` | Armar formularios con `@mantine/form` |
| `mantine-combobox` | Armar selectores y buscadores a medida |
| `mantine-custom-components` | Crear componentes propios que tomen el aspecto del tema |

Las cinco últimas son de terceros: cada carpeta tiene su licencia y un `SOURCE.md` con el repositorio y el commit de origen. Si una choca con `web/DESIGN.md` (por ejemplo, en el idioma o las mayúsculas de los textos), vale `web/DESIGN.md`.

## Linear

Los tickets viven en Linear, en el equipo SIG del espacio `sigta` (https://linear.app/sigta/team/SIG). El acceso es por el servidor MCP de Linear, declarado en `.mcp.json`; cada integrante lo autoriza con su cuenta.

- Los tickets se identifican como `SIG-n`. Las historias del backlog llevan la marca `**USxx**` al principio de la descripción; los tickets que salen de partir una historia dicen "Sale de" y enlazan al original. Los criterios de aceptación están en la descripción, como checklist: **Linear es la fuente de verdad de las historias**, y [docs/08-backlog.md](docs/08-backlog.md) es solo el índice que relaciona cada `USxx` con su `SIG-n`.
- Cada épica es un proyecto de Linear (`EP00` a `EP14`), y cada sprint es el ciclo del mismo número.
- Etiquetas: `s01` a `s18` (sprint), `b1` a `b4` (bloque) y `Feature`, `Bug`, `Improvement`.
- Estados: `Backlog` → `Todo` → `In Progress` → `In Review` → `Done`, más `Canceled` y `Duplicate`.
- Un agente lee tickets libremente, pero no crea, cierra ni cambia el estado de un ticket sin que se lo pidan.
