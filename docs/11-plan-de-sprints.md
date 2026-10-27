# Planificación por sprint

**Versión:** 1.0 · revisión completa, historia por historia, con el equipo
**Qué es.** Las 53 historias del backlog repartidas en 14 sprints con carga, más dos reservados.
**Qué no es.** Un compromiso. Solo el sprint que arranca se compromete; el resto se reescala en el punto de control del sprint 2.

> **La curva.** 17, 19, 26, 30, 29, 17 en la primera mitad; 13, 11, 18, 24, 24, 14, 14, 8 en la segunda. Total 264 SP. Los sprints 3, 4 y 5 quedan por encima de la capacidad estimada de 24 SP: sostienen la demo de diciembre y no se recortaron. Es el riesgo consciente del plan.

---

## Sprint 1 · 21 sep – 4 oct 2026 · 17 SP

**Objetivo:** Entender los cinco trámites y dejar el proyecto en marcha.

| ID | Historia | SP |
|---|---|---|
| US62 | Relevamiento punta a punta de los cinco trámites | 3 |
| US71 | Solicitar ejemplos de trámites a la Secretaría | 1 |
| US63 | Inventario de acciones, estados y actores abstractos | 3 |
| US65 | Definición y setup del stack tecnológico | 5 |
| US01 | Repositorio e integración continua | 5 |

> El sprint más liviano de la primera mitad, a propósito: no hay velocidad medida. **Dos reuniones externas en la primera semana**, con la Secretaría y con TICs. Los ejemplos de trámites se piden el primer día.

## Sprint 2 · 5 oct – 18 oct · 19 SP

**Objetivo:** El modelo de dominio y quién puede entrar.

| ID | Historia | SP |
|---|---|---|
| US64 | Mapa de roles y permisos de la Secretaría | 3 |
| US08 | Modelo de definición de trámite | 3 |
| US02 | Configuración de herramientas agénticas de desarrollo | 5 |
| US06 | Autenticación con proveedor abstraído | 8 *(provisoria)* |

> El modelo de trámite arranca el primer día porque el inventario salió en el sprint 1. **Autenticación lleva estimación provisoria**: se confirma tras la reunión con TICs. **Punto de control:** velocidad medida contra 17 y 19 SP.

## Sprint 3 · 19 oct – 1 nov · 26 SP

**Objetivo:** El trámite ancla configurado y el titular iniciando solicitudes.

| ID | Historia | SP |
|---|---|---|
| US09 | Configuración del trámite ancla: equivalencia entre asignaturas | 8 |
| US07 | Perfiles y autorización | 3 |
| US15 | Carga y almacenamiento de documentos | 8 |
| US13 | Catálogo de trámites con requisitos | 2 |
| US14 | Iniciar una solicitud | 5 *(provisoria)* |

> **Orden interno:** configuración del ancla y perfiles → documentos → catálogo → iniciar una solicitud. **26 SP contra una capacidad estimada de 24**: el punto de control del sprint 2 decide si se sostiene.

## Sprint 4 · 2 nov – 15 nov · 30 SP

**Objetivo:** El trámite avanza, la Secretaría lo ve, y el segundo trámite prueba que el modelo generaliza.

| ID | Historia | SP |
|---|---|---|
| US10 | Avance de un trámite por confirmación de paso | 5 |
| US11 | Historial de pasos de un trámite | 5 |
| US19 | Vista de detalle del trámite | 3 |
| US17 | Bandeja y listado de trámites | 3 |
| US37 | Filtros y orden en el listado de trámites | 1 |
| US04 | Generador de datos sintéticos | 5 |
| US25 | Configuración del segundo trámite: reconocimiento de créditos | 8 |

> Sprint más cargado del proyecto, **30 SP**. **Orden interno:** avance por confirmación → historial → detalle → bandeja y listado → segundo trámite; datos sintéticos en paralelo. Avance por confirmación incluye la derivación al evaluador.

## Sprint 5 · 16 nov – 29 nov · 29 SP

**Objetivo:** La base del asistente y el entorno desplegado.

| ID | Historia | SP |
|---|---|---|
| US27 | Ingesta y versionado de la base de conocimiento | 8 |
| US28 | Recuperación de contexto relevante para la consulta | 8 |
| US68 | Despliegue automático | 8 |
| US16 | Visualización y descarga de documentos del trámite | 5 |

> Dos frentes en paralelo: el asistente y la infraestructura. El despliegue automático llega acá y no antes: nada necesita un entorno desplegado hasta la demo de diciembre, y TICs tiene dos meses para definir el destino. **El último día del sprint, el 29/11 a las 21 h, vence la entrega intermedia de la cátedra** (ver el roadmap, sección 6 bis): un video que incluye una demostración del producto, así que lo que se muestre tiene que estar mergeado antes.

## Sprint 6 · 30 nov – 13 dic · 17 SP

**Objetivo:** Cerrar 2026 con dos trámites y el agente respondiendo.

| ID | Historia | SP |
|---|---|---|
| US69 | Respuesta conversacional sobre el contexto recuperado | 13 |
| US23 | Detalle de mi trámite | 2 |
| US22 | Mis trámites en curso | 2 |

> Carga reducida por diciembre. **Demo ante los directores:** dos trámites configurados y el agente respondiendo con fuentes. La respuesta conversacional lleva 13 SP, fuera de la escala habitual.

---

## Sprints 7, 8 y 9 · 14 dic – 24 ene · sin carga

Once semanas sin compromiso de entrega. Antes de irse hay que dejar agendada la primera validación con la Secretaría y pedidas las plantillas de documentos.

---

## Sprint 10 · 25 ene – 7 feb 2027 · 13 SP

**Objetivo:** Volver del receso con el asistente evaluado y el retorno del evaluador.

| ID | Historia | SP |
|---|---|---|
| US31 | Set de evaluación del asistente en el pipeline | 8 |
| US21 | Registrar retorno de la evaluación externa | 5 *(provisoria)* |

> Carga baja. Primera validación con la Secretaría sobre el flujo de los dos trámites. Registrar retorno llega acá desde el sprint 5: su forma depende de si los evaluadores entran a la plataforma.

## Sprint 11 · 8 feb – 21 feb · 11 SP

**Objetivo:** Avisos por correo y el historial del titular.

| ID | Historia | SP |
|---|---|---|
| US36 | Correo de pedido de acción | 8 |
| US24 | Historial de trámites cerrados | 3 |

> Carga baja. Es la primera historia que manda correos, así que carga con la configuración del envío.

## Sprint 12 · 22 feb – 7 mar · 18 SP

**Objetivo:** La plataforma lee los documentos y los cruza con los requisitos.

| ID | Historia | SP |
|---|---|---|
| US40 | Lectura de los documentos cargados | 5 |
| US41 | Criterios verificables de cada requisito y cruce con los documentos | 8 |
| US66 | Verificación manual de requisitos | 5 |

> Arranca la validación asistida. La verificación manual baja desde diciembre para quedar junto al cruce que la asiste. **El cruce necesita que la Secretaría defina qué hace verificable a cada requisito.**

## Sprint 13 · 8 mar – 21 mar · 24 SP

**Objetivo:** Las preactas generándose y el asistente asegurado.

| ID | Historia | SP |
|---|---|---|
| US45 | Generación determinista de preacta desde plantilla | 8 |
| US32 | Pruebas de seguridad del asistente | 8 |
| US46 | Campos de texto libre asistidos por modelo | 5 |
| US44 | Carga y versionado de plantillas de documentos | 3 |

> Pico de la segunda mitad junto con el 14. Las plantillas llegan acá, en el mismo sprint que la preacta que las usa. **Depende de que la Secretaría haya entregado las plantillas.**

## Sprint 14 · 22 mar – 4 abr · 24 SP

**Objetivo:** Aprobación de preactas, el trámite más complejo y los avisos al titular.

| ID | Historia | SP |
|---|---|---|
| US47 | Revisión, edición y aprobación de la preacta | 5 |
| US43 | Set de evaluación de la validación asistida y de las preactas | 8 |
| US49 | Configuración de cobertura: pase de universidad extranjera | 5 |
| US35 | Correo en cambio de estado | 3 |
| US12 | Devolución de un paso | 3 |

> **Segunda validación con la Secretaría**, y la instancia donde se agendan las pruebas de aceptación del sprint 17. El pase extranjero va primero en la cobertura: si la abstracción no lo cubre, quedan dos sprints para corregir.

## Sprint 15 · 5 abr – 18 abr · 14 SP

**Objetivo:** Subsanaciones y el cuarto trámite.

| ID | Historia | SP |
|---|---|---|
| US50 | Configuración de cobertura: intercambios académicos | 5 |
| US67 | Plazos próximos a vencer | 3 |
| US33 | Pedido de subsanación al titular | 3 |
| US34 | El titular completa una subsanación | 3 |

> La subsanación aparece recién acá, siendo el caso más frecuente en trámites documentales.

## Sprint 16 · 19 abr – 2 may · 14 SP

**Objetivo:** Los cinco trámites andando.

| ID | Historia | SP |
|---|---|---|
| US51 | Configuración de cobertura: pase dentro de la UBA | 5 |
| US53 | Recordatorio de derivación sin retorno | 3 |
| US38 | Descarga de comprobante del trámite | 3 |
| US52 | Base de conocimiento de los cinco trámites | 3 |

> Último sprint con historias. Al cierre, los cinco trámites funcionan de punta a punta.

## Sprint 17 · 3 may – 16 may · 8 SP

**Objetivo:** Que la Secretaría rompa el sistema y arreglar lo que rompa.

| ID | Historia | SP |
|---|---|---|
| US54 | Pruebas de aceptación con la Secretaría | 8 |

> **8 SP comprometidos y el resto del sprint reservado** para corregir los hallazgos bloqueantes de la validación.

## Sprint 18 · 17 may – 30 may · 0 SP

**Objetivo:** Cerrar.

> **Sin historias.** Sprint reservado para el informe final y los ajustes no bloqueantes.

---

## Decisiones pendientes de la revisión

| Tema | Qué hay que decidir |
|---|---|
| Evaluadores externos | Si entran a la plataforma. Define Registrar retorno (US21) y puede sumar historias. Sale del relevamiento y el mapa de roles, sprint 1 |
| Escala de estimación | La respuesta conversacional (US69) quedó en 13 SP, fuera de la escala. O se parte, o se cambia la regla en el documento de proceso |
| Sprints 3, 4 y 5 | Van por encima de la capacidad estimada. El punto de control del sprint 2 decide si se sostienen o se corre algo |
| Campos de texto libre (US46) | Si quedan campos de texto libre reales una vez que el dictamen está fuera del alcance. Se resuelve al ver las plantillas, sprint 13 |
| Plazo de retorno del evaluador | Si existe un plazo interno que la Secretaría le da al evaluador. Sin él, el recordatorio (US53) se cae |
| Información interna | Qué información de la Secretaría no ve el titular; debería salir del Mapa de roles (US64) |
| Recuperación de contexto (US28) | Qué tiempo máximo de respuesta se exige |
| Líneas de base | Pedirle a la Secretaría en el sprint 1 el tiempo promedio de resolución y las horas mensuales de trabajo manual. En mayo ya no se pueden medir hacia atrás |

## Resumen

| Bloque | Sprints | SP |
|---|---|---|
| B1 Esqueleto y asistente mínimo | 1 a 6 | 138 |
| B2 Cierre del asistente | 10 y 11 | 24 |
| B3 Validación asistida, preactas y cobertura | 12 a 16 | 94 |
| B4 Validación y cierre | 17 y 18 | 8 más reserva |
| **Total** | 14 sprints con carga | **264** |

## Los sprints que importan

| Sprint | Por qué |
|---|---|
| **1** | El relevamiento es el insumo de todo y no se recupera después. Dos reuniones externas la primera semana |
| **2** | Punto de control que sostiene el plan entero: sin velocidad medida acá, lo que sigue es conjetura |
| **4** | 30 SP, el sprint más cargado. El segundo trámite valida que los cinco son una familia |
| **6** | La demo de diciembre define cómo llega el equipo al receso |
| **14** | El pase extranjero prueba la abstracción contra el caso difícil, con dos sprints de margen |

## Criterios de aceptación

Los criterios de cada historia están en su ticket de Linear, como checklist. El [backlog](08-backlog.md) relaciona cada historia con su ticket.
