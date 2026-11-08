# Escenarios de aceptación

Comportamiento esperado de SIGTA, escrito en Gherkin y en inglés con el vocabulario de `AGENTS.md`. Cada historia tiene un archivo `.feature`, con las etiquetas `@USxx @SIG-n`, que se escribe antes de implementarla. Todos los escenarios corren con godog dentro de `go test`, en el pipeline; los pasos están en `services/backend/test/acceptance/`. Un paso sin implementar hace fallar el pipeline, así que un `.feature` entra junto con sus pasos. Los recorridos de interfaz se van a automatizar con Playwright, que todavía no está en el repositorio.

Para redactar los escenarios de una historia está la rutina `gherkin-scenarios` (`.claude/skills/`).

## Matriz de trazabilidad

Cada historia tiene al menos un escenario, y cada escenario referencia su historia.

| Historia | Ticket | Archivo | Escenarios |
| --- | --- | --- | --- |
| US06 | SIG-11, SIG-61 | [authentication.feature](authentication.feature) | 16 |
| US07 | SIG-14 | [profiles_and_authorization.feature](profiles_and_authorization.feature) | 23 |
| US07 | SIG-14 | [staff_areas.feature](staff_areas.feature) | 9 |
| US08 | SIG-12 | [procedure_type_definition.feature](procedure_type_definition.feature) | 10 |
| US13 | SIG-16 | [procedure_catalog.feature](procedure_catalog.feature) | 10 |
