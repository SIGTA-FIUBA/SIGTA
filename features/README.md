# Escenarios de aceptación

Comportamiento esperado de SIGTA, escrito en Gherkin y en inglés con el vocabulario de `AGENTS.md`. Cada historia tiene un archivo `.feature`, con las etiquetas `@USxx @SIG-n`, que se escribe antes de implementarla. Todos los escenarios corren con godog dentro de `go test`, en el pipeline; los pasos están en `services/backend/test/acceptance/`. Un paso sin implementar hace fallar el pipeline, así que un `.feature` entra junto con sus pasos. Los recorridos de interfaz se van a automatizar con Playwright, que todavía no está en el repositorio.

Para redactar los escenarios de una historia está la rutina `gherkin-scenarios` (`.claude/skills/`).

## Matriz de trazabilidad

Cada historia tiene al menos un escenario, y cada escenario referencia su historia.

| Historia | Ticket | Archivo | Escenarios |
| --- | --- | --- | --- |
| US06 | SIG-11, SIG-61 | [authentication.feature](authentication.feature) | 16 |
| US07 | SIG-14 | [profiles_and_authorization.feature](profiles_and_authorization.feature) | 29 |
| US07 | SIG-14 | [staff_areas.feature](staff_areas.feature) | 28 |
| US08 | SIG-12, SIG-65 | [procedure_type_definition.feature](procedure_type_definition.feature) | 12 |
| US10 | SIG-20 | [step_confirmation.feature](step_confirmation.feature) | 27 |
| US11 | SIG-21, SIG-81 | [procedure_history.feature](procedure_history.feature) | 9 |
| US12 | SIG-44 | [step_send_back.feature](step_send_back.feature) | 5 |
| US13 | SIG-16 | [procedure_catalog.feature](procedure_catalog.feature) | 10 |
| US14 | SIG-17 | [procedure_start.feature](procedure_start.feature) | 24 |
| US15 | SIG-18 | [document_upload.feature](document_upload.feature) | 9 |
| US16 | SIG-26 | [document_view.feature](document_view.feature) | 4 |
| US17 | SIG-22 | [procedure_inbox.feature](procedure_inbox.feature) | 14 |
| US22 | SIG-30 | [my_procedures.feature](my_procedures.feature) | 6 |
| US23 | SIG-31 | [my_procedure_detail.feature](my_procedure_detail.feature) | 8 |
| US24 | SIG-35 | [my_closed_procedures.feature](my_closed_procedures.feature) | 4 |
| US19 | SIG-23, SIG-81 | [procedure_detail.feature](procedure_detail.feature) | 6 |
| US33 | SIG-49 | [correction_request.feature](correction_request.feature) | 9 |
| US34 | SIG-50 | [correction_answer.feature](correction_answer.feature) | 11 |
| US37 | SIG-25 | [procedure_list_filters.feature](procedure_list_filters.feature) | 7 |
| US38 | SIG-53 | [procedure_receipt.feature](procedure_receipt.feature) | 4 |
| — | SIG-82 | [step_documents.feature](step_documents.feature) | 7 |
| US72 | SIG-74 | [procedure_type_drafts.feature](procedure_type_drafts.feature) | 14 |
| US74 | SIG-76 | [staff_administration.feature](staff_administration.feature) | 25 |
