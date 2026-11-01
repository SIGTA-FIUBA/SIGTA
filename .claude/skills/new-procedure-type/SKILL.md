---
name: new-procedure-type
description: Prepara el borrador de configuración de un tipo de trámite nuevo de SIGTA (equivalencia, reconocimiento de créditos, pase extranjero, intercambios, pase UBA) usando solo el modelo existente. Usar cuando se pida configurar, agregar o dar de alta un trámite.
---

# Configuración de un tipo de trámite nuevo

Sumar un trámite tiene que ser solo configuración (ADR-006, ADR-013). La configuración de trámites **no se delega**: el agente prepara un borrador, el equipo lo revisa y la Secretaría lo valida contra su proceso real.

## Antes de empezar

- **Leer el modelo** en `CONTRIBUTING.md`, sección "Modelo de tipo de trámite", y el código en `services/backend/src/domain/procedure/`. No inventar campos que el modelo no tiene.
- **Leer el ticket del trámite en Linear** y sus criterios de aceptación.
- **Leer el relevamiento** del trámite y su borrador, si ya existe, con sus preguntas a confirmar. Los borradores no se suben al repositorio: describen procesos que la Secretaría no confirmó.
- **Revisar D-08** en `docs/04-decisiones-adr.md`: quién propone y quién aprueba una versión. El agente prepara borradores; nunca figura como quien propone ni como quien aprueba una versión.

## Pasos

1. **Usar las áreas existentes.** Las áreas son comunes a todos los trámites y se referencian por `id`; un área nueva se pide aparte. Los datos y documentos se declaran en el tipo de trámite, cada uno con un `id` en minúsculas ASCII, y todos tienen que usarse en algún paso.
2. **Modelar cada paso con una de las cuatro acciones:** `enter_data`, `upload_document`, `confirm_external_fact` o `decide`. Una decisión lleva los datos y documentos que la fundamentan, como el acta. Los pasos de un área de tipo `office` o `external` declaran su perfil (`staff_operative` o `staff_approver`; por convención, el de aprobación es para aprobar, y el código no lo valida); los del titular no llevan perfil y solo cargan datos o documentos.
3. **Modelar la derivación como un paso de un área `external`**, con su retorno.
4. **No agregar estados de resultado** (ADR-009). El dictamen se adjunta en el retorno y se comunica.
5. **Anotar los plazos y las condiciones de admisibilidad** del relevamiento como preguntas a confirmar: el modelo todavía no los tiene, y está a confirmar en qué ticket quedan (ver CONTRIBUTING).
6. **Validar la configuración:** guardar el borrador como `<nombre>.json` en una carpeta fuera del repositorio, con las áreas que usa en `areas.json` y sus preguntas en un README, y correr `PROCEDURE_DRAFTS=<carpeta> go test -run TestDrafts ./src/domain/procedure/` desde `services/backend`. El test valida todos los archivos de la carpeta y rechaza las claves desconocidas. Todavía no hay motor para recorrer el trámite de punta a punta.

## Alarma del ADR-006

Si el trámite no entra con el modelo existente (hace falta una acción nueva, un tipo de paso nuevo o cualquier cambio de código), **frenar**. No parchear el motor para este trámite. Avisar qué no encaja y por qué, para que el equipo corrija la abstracción y lo registre contra el riesgo R-10.

## Al entregar

- Aclarar que la configuración es un borrador generado por un agente y que necesita la revisión del equipo y la validación de la Secretaría.
- Si cambió el modelo de configuración, actualizar el instructivo en el mismo PR (ADR-007).
- Reportar el esfuerzo real, para compararlo con la estimación del ticket. No modificar el ticket de Linear sin que lo pidan.
