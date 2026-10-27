---
name: gherkin-scenarios
description: Redacta los escenarios Gherkin de una historia de SIGTA a partir de sus criterios de aceptación en Linear. Usar cuando se pidan escenarios para un USxx o un SIG-n, o antes de implementar una historia que todavía no los tiene.
---

# Escenarios Gherkin desde los criterios de aceptación

Cada historia se especifica con escenarios antes de implementarse. Los escenarios son el borrador de un acuerdo con la Secretaría: el agente los redacta, pero el equipo los revisa y los criterios no se cambian por su cuenta.

## Pasos

1. **Leer la historia en Linear.** Con `get_issue` sobre el `SIG-n`, o buscando `**USxx**` en la descripción. Los criterios de aceptación del ticket son la única fuente: no usar versiones viejas de `docs/`.
2. **Leer el vocabulario** de `AGENTS.md` (columna "En código y escenarios") y, si un término no está, el [glosario](../../../docs/01-glosario.md). No inventar sinónimos.
3. **Escribir un archivo por historia** en `features/<slug-en-inglés>.feature`, en inglés, con las etiquetas `@USxx @SIG-n` en la línea de `Feature`.
4. **Cubrir cada criterio con al menos un escenario.** Con `Scenario Outline` cuando el mismo comportamiento varía solo en los datos.
5. **Actualizar la matriz** de [features/README.md](../../../features/README.md) con la historia, el ticket, el archivo y la cantidad de escenarios.
6. **Listar las dudas.** Si un criterio es ambiguo o no se puede verificar, no inventar el comportamiento: dejar la pregunta en la descripción del PR para que el equipo la resuelva.

## Cómo se escribe un escenario

- **Declarativo, no de interfaz.** Describe qué pasa en el dominio, no dónde se hace clic: `When they submit the procedure`, no `When they click "Send"`.
- **Un comportamiento por escenario**, con un nombre que lo diga: `Scenario: cannot submit without a mandatory requirement`.
- **Solo lo que entrega su historia.** Un escenario se apoya en historias ya cerradas y nunca en funcionalidad de sprints posteriores.
- **Datos sintéticos** y genéricos (`an applicant`, `the transcript`), nunca nombres ni documentos reales (ADR-005).
- **Sin resultados de trámite** (ADR-009): reconocimiento total, parcial o rechazo son el mismo evento, un dictamen que se adjunta. Ningún escenario espera un estado de resultado.
- **La IA propone y una persona decide:** ningún escenario de validación asistida, asistente o preactas termina con un cambio de estado hecho por un modelo.

## Ejemplo

```gherkin
@US14 @SIG-17
Feature: Start a procedure

  Background:
    Given an applicant signed in with their UBA account
    And the "subject equivalence" procedure type requires a transcript

  Scenario: cannot submit without a mandatory requirement
    Given a draft "subject equivalence" procedure without a transcript
    When the applicant submits it
    Then the procedure is not submitted
    And the applicant is told the transcript is missing

  Scenario: submitting a complete procedure returns its identifier
    Given a draft "subject equivalence" procedure with every mandatory requirement
    When the applicant submits it
    Then the procedure is submitted
    And the applicant receives the procedure identifier
```

## Lo que no hace esta rutina

- No modifica los criterios de aceptación en Linear.
- No implementa los pasos (`step definitions`) de godog ni de Playwright: eso es parte de la implementación de la historia.
