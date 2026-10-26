# Proceso, calidad y métricas

---

## 1. Definition of Ready

Un ticket entra a un sprint solo si cumple todo esto:

- [ ] Está escrito como necesidad de una persona usuaria identificable, no como tarea técnica suelta.
- [ ] Tiene criterios de aceptación escritos **antes** de empezar el desarrollo.
- [ ] Está estimado en Story Points por el equipo.
- [ ] No depende de una definición pendiente del cliente. Si depende, va acompañado del supuesto explícito bajo el cual se construye.
- [ ] No supera los 8 SP. Un 13 entra solo como excepción justificada y registrada en el backlog.
- [ ] Se sabe cómo se va a probar.

## 2. Definition of Done

Un ticket está terminado cuando:

- [ ] El código está en la rama principal.
- [ ] Fue revisado por al menos un par.
- [ ] Los tests unitarios y de integración pasan en el pipeline.
- [ ] Los escenarios de aceptación asociados pasan en el pipeline.
- [ ] La documentación asociada está actualizada **en el mismo PR**: docstrings, README del módulo, archivo de contexto del repositorio y, si cambió el modelo de configuración, el instructivo de cómo editarla.
- [ ] Está desplegado en el entorno y se puede demostrar.
- [ ] No deja bugs conocidos de criticidad alta o crítica abiertos.

> La diferencia entre "terminado" y "terminado pero" es lo que hace que un proyecto de 18 sprints llegue al último con deuda invisible. No hay estado intermedio.

## 2 bis. Documentación continua

No hay historias de documentación en el backlog (ADR-007). La documentación se produce dentro de cada ticket, como parte de la DoD, y el agente de desarrollo la escribe y la actualiza en el mismo PR que el código.

| Qué se documenta | Dónde vive | Cuándo se escribe |
|---|---|---|
| Comportamiento de un módulo | Docstrings y README del módulo | En el PR que lo crea o modifica |
| Decisiones de arquitectura | Registro de ADR | El día que se toma la decisión |
| Modelo de configuración de trámites y cómo editarlo | Instructivo junto a los archivos de configuración | En el PR que cambia el modelo |
| Comportamiento esperado del sistema | Escenarios Gherkin | Antes de desarrollar la funcionalidad |
| Uso funcional para la Secretaría | Se deriva de los escenarios Gherkin | Se compila al cierre, sin redactarse de cero |

**El control.** Un PR sin su documentación no cumple la DoD y no se mergea. Si al llegar al cierre hay que escribir documentación de cero, la DoD no se aplicó durante ocho meses y ese es el problema, no la falta de una historia.

**Lo único que queda como historia** es el informe final, que es un entregable académico con su propio formato, no documentación del producto.

## 3. Estimación

- **Unidad:** Story Points, escala Fibonacci: 1, 2, 3, 5, 8, 13.
- **Qué miden:** esfuerzo relativo, considerando complejidad, incertidumbre y volumen. No horas.
- **Referencias del equipo** (se fijan en el sprint 1 y no se tocan después):
  - **2 SP:** una vista de listado simple sobre datos que ya existen.
  - **5 SP:** un caso de uso completo con su capa de datos, su API y su interfaz.
  - **8 SP:** algo con incertidumbre técnica real o que toca varios componentes.
  - **13 SP:** excepción. Se admite solo si el equipo decide que partir la historia la haría menos clara, y queda marcada como excepción en el backlog. Hoy hay una: la respuesta conversacional del asistente.
- **Cómo:** estimación conjunta del equipo, sin la presencia de quien vaya a tomar el ticket, para no anclar en la habilidad individual.
- **Cuándo:** en el refinamiento de mitad de sprint, para el sprint siguiente.

## 4. Capacidad y compromiso

| Concepto | Valor |
|---|---|
| Capacidad nominal por sprint | 4 personas × 12 hs × 2 semanas = **96 hs** |
| Capacidad real | Se declara en cada planificación, descontando ausencias conocidas |
| Velocidad | Promedio móvil de los últimos 3 sprints |

**Regla para los primeros dos sprints.** Sin velocidad histórica, el sprint 1 se compromete al 60% de lo que el equipo cree poder hacer. Los 19 SP que el plan asigna al sprint 1 son una hipótesis de reparto, no el compromiso: el compromiso se fija en la planificación. Es preferible terminar antes y tomar un ticket más que arrastrar deuda desde el primer sprint. A partir del sprint 3 se compromete según velocidad medida.

**Regla de replanificación.** El roadmap se ajusta con la velocidad real, no con la nominal. Al cierre del sprint 2 se reescala la curva entera con la velocidad medida. Si está más de un 25% por debajo de lo previsto, el equipo decide en la reunión quincenal si activa el orden de recorte del ADR-008 (preactas, validación asistida, cobertura a tres), y no dos meses después.

## 5. Estados del ticket en Linear

Linear es la fuente de verdad de las historias: descripción, criterios de aceptación, estimación y estado viven en el ticket. El [backlog](08-backlog.md) del repositorio es solo un índice.

`Backlog` → `Todo` (cumple DoR) → `In Progress` → `In Review` → `Done` (cumple DoD)

Estados adicionales: `Canceled` y `Duplicate`.

## 6. Etiquetas

| Etiqueta | Uso |
|---|---|
| `s01` a `s18` | Sprint planificado |
| `b1` `b2` `b3` `b4` | Bloque al que pertenece |
| `Feature`, `Bug`, `Improvement` | Tipo de ticket |

Las historias no tienen etiqueta propia: cada una se reconoce por la marca `**US62**` al principio de su descripción.

## 7. Criticidad de bugs

| Nivel | Definición | Plazo |
|---|---|---|
| Crítico | Impide usar el sistema o corrompe datos de un trámite | Se atiende inmediatamente, interrumpe el sprint |
| Alto | Un caso de uso principal no se puede completar | Dentro del sprint |
| Medio | Funciona con rodeo o afecta un caso secundario | Se prioriza en el sprint siguiente |
| Bajo | Cosmético o de baja frecuencia | Backlog |

## 8. Indicadores de proceso, por sprint

Se registran en la planilla de gestión y se revisan en cada reunión quincenal.

| Indicador | Cómo se calcula | Para qué sirve |
|---|---|---|
| SP comprometidos vs. completados | Suma de SP del sprint backlog vs. SP en `Terminado` | Detecta sobrecompromiso sistemático |
| Velocidad | SP completados, promedio móvil de 3 sprints | Base de la replanificación |
| Tickets reabiertos | Tickets que vuelven de `Terminado` | Señal de que la DoD no se está aplicando |
| Bugs abiertos por criticidad | Conteo por nivel al cierre del sprint | Salud del producto |
| Cobertura de pruebas | Reportada por el pipeline | Riesgo de regresión |
| Tiempo de ciclo | Días entre `En curso` y `Terminado`, mediana | Detecta tickets demasiado grandes |
| Horas por integrante | Registro propio, por sprint | Capacidad real y riesgo R-05 |
| Supuestos caídos | Conteo de supuestos del scope que se invalidaron | Estabilidad del alcance |

## 9. Planilla de seguimiento: estructura mínima

**Hoja `Sprints`**
`sprint` · `desde` · `hasta` · `capacidad_hs_declarada` · `sp_comprometidos` · `sp_completados` · `tickets_reabiertos` · `bugs_criticos` · `bugs_altos` · `cobertura_pct` · `cycle_time_mediano` · `notas`

**Hoja `Horas`**
`sprint` · `integrante` · `horas_registradas` · `horas_comprometidas`

**Hoja `Bugs`**
`id` · `titulo` · `criticidad` · `sprint_detectado` · `sprint_resuelto` · `estado`

**Hoja `Burnup`**
`sprint` · `sp_acumulados_completados` · `sp_totales_alcance` — dos series en un gráfico.

> **Por qué burnup y no burndown.** El burndown de sprint oculta el crecimiento del alcance; el burnup lo muestra como una línea superior que sube. En un proyecto de 18 sprints con alcance en negociación con un cliente externo, eso es exactamente lo que hay que poder ver.

## 10. Estrategia de pruebas

1. **Escenarios de aceptación en Gherkin**, escritos junto con la validación del cliente final, como documentación viva. Cada funcionalidad empieza por acá.
2. **Tests unitarios y de integración** en ciclos cortos, hasta completar la funcionalidad.
3. **Sets de evaluación de IA**, uno por cada uso: asistente, validación asistida y preactas. Cada uno con casos esperados y umbral de aprobación, ejecutados en el pipeline como cualquier otro test.
4. **Casos de inyección de prompt** sobre documentos aportados por titulares, tratados como contexto no confiable.
5. Todo lo anterior se ejecuta automáticamente en GitHub Actions.

**Matriz de trazabilidad:** cada User Story tiene al menos un escenario Gherkin asociado, y cada escenario referencia su US. Se mantiene en el repositorio junto a los archivos de escenarios.
