# Escenarios de aceptación

Comportamiento esperado de SIGTA, escrito en Gherkin y en inglés con el vocabulario de `AGENTS.md`. Cada historia tiene un archivo `.feature`, con las etiquetas `@USxx @SIG-n`, que se escribe antes de implementarla. Las reglas del motor se automatizan con godog y los recorridos de interfaz con Playwright.

Para redactar los escenarios de una historia está la rutina `gherkin-scenarios` (`.claude/skills/`).

## Matriz de trazabilidad

Cada historia tiene al menos un escenario, y cada escenario referencia su historia.

| Historia | Ticket | Archivo | Escenarios |
| --- | --- | --- | --- |
