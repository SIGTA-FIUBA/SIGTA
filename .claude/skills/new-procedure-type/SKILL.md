---
name: new-procedure-type
description: Prepara el borrador de configuración de un tipo de trámite nuevo de SIGTA (equivalencia, reconocimiento de créditos, pase extranjero, intercambios, pase UBA) usando solo el modelo existente. Usar cuando se pida configurar, agregar o dar de alta un trámite.
---

# Configuración de un tipo de trámite nuevo

Sumar un trámite tiene que ser solo configuración (ADR-001, ADR-006). La configuración de trámites **no se delega**: el agente prepara un borrador, el equipo lo revisa y la Secretaría lo valida contra su proceso real.

## Antes de empezar

- **El modelo de configuración tiene que existir.** Se define en US08 (SIG-12), junto con un instructivo de cómo editarlo. Si todavía no está en el repositorio, frenar y avisar: no inventar un formato.
- **Leer el ticket del trámite en Linear** (por ejemplo SIG-24, SIG-48, SIG-51 o SIG-54) y sus criterios de aceptación.
- **Leer el relevamiento** del trámite (US62, SIG-6) y los tipos de trámite ya configurados, para reutilizar sus requisitos.
- **Revisar D-07** en `docs/04-decisiones-adr.md`: si se decidió que la Secretaría edite los trámites por su cuenta, esta rutina puede haber cambiado.

## Pasos

1. **Reutilizar antes de crear.** Cada requisito que ya exista en otro trámite se referencia, no se copia: si se modifica, cambia en todos.
2. **Modelar cada paso con una de las cuatro acciones:** `enter data`, `upload document`, `confirm external fact` o `decide`. Cada paso declara su perfil responsable (`operator` o `approver`).
3. **Modelar la derivación como un paso normal**, con su evaluador externo y su retorno. Ningún trámite se resuelve íntegramente dentro del área.
4. **No agregar estados de resultado** (ADR-009). El dictamen se adjunta en el retorno y se comunica.
5. **Declarar plazos y condiciones de admisibilidad** si el relevamiento los tiene.
6. **Validar la configuración** con el mecanismo que define US08 (la aplicación no levanta si la configuración es inconsistente) y recorrer el trámite de punta a punta sobre datos sintéticos.

## Alarma del ADR-006

Si el trámite no entra con el modelo existente (hace falta una acción nueva, un tipo de paso nuevo o cualquier cambio de código), **frenar**. No parchear el motor para este trámite. Avisar qué no encaja y por qué, para que el equipo corrija la abstracción y lo registre contra el riesgo R-10.

## En el PR

- Aclarar que la configuración es un borrador generado por un agente y que necesita la revisión del equipo y la validación de la Secretaría.
- Si cambió el modelo de configuración, actualizar el instructivo en el mismo PR (ADR-007).
- Reportar el esfuerzo real en la descripción del PR, para compararlo con la estimación del ticket. No modificar el ticket de Linear sin que lo pidan.
