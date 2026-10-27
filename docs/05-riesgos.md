# Registro de riesgos

**Cómo se usa.** Se revisa en cada reunión quincenal con los directores. Cada riesgo tiene un dueño, que es quien reporta su estado. Un riesgo sin disparador definido no se puede monitorear: si no sabés qué señal mirar, te enterás cuando ya se materializó.

**Escalas.** Probabilidad y Severidad: Alta (3), Media (2), Baja (1). Exposición = P × S.
**Estados.** Abierto · Mitigándose · Materializado · Cerrado.

---

## Resumen por exposición

| ID | Riesgo | P | S | Exp. |
|---|---|---|---|---|
| R-01 | El alcance completo excede la capacidad del equipo | Alta | Alta | 9 |
| R-13 | Los sprints 3, 4 y 5 por encima de la capacidad estimada | Alta | Alta | 9 |
| R-02 | Expectativa desalineada sobre el editor de flujos | Media | Alta | 6 |
| R-03 | Los cinco trámites no quedan relevados a tiempo | Media | Alta | 6 |
| R-06 | Alucinaciones con consecuencia administrativa | Media | Alta | 6 |
| R-15 | Pérdida de contexto en el receso | Alta | Media | 6 |
| R-16 | La primera validación muestra una herramienta pensada para el titular | Alta | Media | 6 |
| R-04 | Las plantillas de documentos no se entregan | Media | Media | 4 |
| R-05 | Disponibilidad del equipo por debajo de lo comprometido | Media | Media | 4 |
| R-07 | Inyección de prompts a través de documentos | Media | Media | 4 |
| R-08 | Indisponibilidad de la Secretaría | Media | Media | 4 |
| R-14 | Las horas comprometidas no cierran con el cierre en mayo | Media | Media | 4 |
| R-10 | Los pasos no se reducen a acciones abstractas | Baja | Alta | 3 |
| R-09 | Costos de infraestructura a cargo del equipo | Media | Baja | 2 |
| R-12 | Retrasos en la comunicación | Baja | Media | 2 |

---

## Riesgos activos

### R-01 — El alcance completo excede la capacidad del equipo
**Categoría:** Alcance · **P:** Alta · **S:** Alta · **Exposición:** 9 · **Estado:** Abierto

El alcance son 264 SP en 14 sprints con carga: los cinco trámites, los tres usos de IA y la plataforma completa, con 4 personas a 12 hs semanales. Se decidió no recortar de antemano (ADR-008).

- **Disparador:** velocidad medida al cierre del sprint 2 más de un 25% por debajo de lo planificado, o cualquier punto de control del roadmap que no se cumpla.
- **Mitigación:** puntos de control decisorios en los sprints 2, 4, 6, 13, 14 y 16. Replanificar sobre la velocidad real, no sobre la nominal.
- **Contingencia:** orden de recorte fijado en ADR-008, de primero a último: generación de preactas, validación asistida, cobertura de cinco trámites reducida a tres. El asistente RAG no se recorta. La decisión la toma el equipo.

### R-02 — Expectativa desalineada con la Secretaría sobre el editor de flujos
**Categoría:** Stakeholders · **P:** Media · **S:** Alta · **Exposición:** 6 · **Estado:** Abierto

El equipo acordó internamente el 16/09 una herramienta de edición de flujos para el personal administrativo, y luego la decisión se revirtió (ADR-001). Si la idea llegó a circular con la Secretaría, queda una expectativa que no se va a cumplir.

- **Disparador:** cualquier mención de la Secretaría a "configurar trámites por su cuenta".
- **Mitigación:** está marcado como duda pendiente en el alcance. Se resuelve en el relevamiento del sprint 1 (US62) junto con el mecanismo alternativo (pedido al equipo, resuelto en días) y se asienta como ADR.
- **Contingencia:** ofrecer una vista de solo lectura de la configuración vigente, a costo bajo, que cubre buena parte de la necesidad real.

### R-03 — Los cinco trámites no quedan relevados a tiempo
**Categoría:** Dependencias · **P:** Media · **S:** Alta · **Exposición:** 6 · **Estado:** Abierto

El modelo de definición (sprint 2), la configuración del ancla (sprint 3) y la del segundo trámite (sprint 4) dependen de conocer el recorrido real de los cinco trámites.

- **Disparador:** cierre del sprint 1 sin el relevamiento de US62 validado por la Secretaría.
- **Mitigación:** partir de la documentación que ya entregó la Secretaría de Gestión Académica, no de cero. Pedir la reunión de relevamiento para la primera semana del sprint 1.
- **Contingencia:** el equipo completa el relevamiento con lo que la planilla permite inferir y lo lleva a validar, invirtiendo la carga: en vez de pedir una definición, pedir una confirmación.

### R-04 — Las plantillas de documentos no se entregan
**Categoría:** Dependencias · **P:** Media · **S:** Media · **Exposición:** 4 · **Estado:** Abierto

La generación de preactas depende de plantillas oficiales que la Secretaría todavía no entregó.

- **Disparador:** cierre del sprint 11 sin las plantillas. US44, que las necesita, está en el sprint 12.
- **Mitigación:** pedirlas por escrito en el primer contacto, para titulares y para gestión, y aceptar cualquier formato, incluso documentos ya emitidos como muestra.
- **Contingencia:** reconstruir una plantilla a partir de actas anteriores anonimizadas y validarla como propuesta.

### R-05 — Disponibilidad del equipo por debajo de lo comprometido
**Categoría:** Recursos · **P:** Media · **S:** Media · **Exposición:** 4 · **Estado:** Abierto

El compromiso es de 12 hs semanales por integrante. Ya se identificaron períodos de disponibilidad reducida por compromisos personales y festividades.

- **Disparador:** dos sprints consecutivos con horas registradas por debajo del 80% de lo comprometido.
- **Mitigación:** registrar horas por integrante y por sprint, y declarar la capacidad **real** de cada sprint en la planificación en vez de asumir la nominal.
- **Contingencia:** reducir el compromiso del sprint siguiente y ajustar el roadmap en la reunión quincenal, antes de que la deuda se acumule en silencio.

### R-06 — Alucinaciones del modelo con consecuencia administrativa
**Categoría:** Técnico / Producto · **P:** Media · **S:** Alta · **Exposición:** 6 · **Estado:** Mitigándose

Una respuesta incorrecta sobre requisitos o un dato inventado en una preacta puede hacer que alguien pierda un trámite o un plazo.

- **Disparador:** cualquier respuesta del set de evaluación sin fuente citada, o un campo estructurado de preacta no trazable al dato de origen.
- **Mitigación:** respuestas limitadas a la base de conocimiento validada, con cita de fuente y derivación a una persona fuera de ella. Campos estructurados de preacta completados de forma determinista. Ninguna acción sin aprobación humana.
- **Contingencia:** degradar el asistente a búsqueda con fragmentos citados, sin generación libre.

### R-07 — Inyección de prompts a través de documentos de titulares
**Categoría:** Seguridad · **P:** Media · **S:** Media · **Exposición:** 4 · **Estado:** Mitigándose

Los documentos que carga el titular pasan por componentes que los procesan con modelos de lenguaje.

- **Disparador:** un caso de ataque del set de pruebas que logre alterar el comportamiento del asistente o de la validación.
- **Mitigación:** tratar todo documento aportado como contexto no confiable, separar instrucciones de contenido, e incluir casos de ataque en la estrategia de pruebas automatizadas.
- **Contingencia:** restringir el procesamiento a extracción de texto sin interpretación por modelo, con validación por reglas.

### R-08 — Indisponibilidad de la Secretaría para relevamiento y validación
**Categoría:** Stakeholders · **P:** Media · **S:** Media · **Exposición:** 4 · **Estado:** Abierto

- **Disparador:** una instancia de validación de fase no concretada dentro de las dos semanas posteriores al cierre del entregable.
- **Mitigación:** acordar con anticipación las instancias de validación de cada fase y avanzar con datos sintéticos entre ellas.
- **Contingencia:** validación asincrónica: enviar un video de demo y un cuestionario corto con preguntas cerradas.

### R-09 — Costos de infraestructura a cargo del equipo
**Categoría:** Recursos · **P:** Media · **S:** Baja · **Exposición:** 2 · **Estado:** Abierto

Los servicios de nube y las herramientas de desarrollo se pagarían con medios personales de los integrantes.

- **Disparador:** un costo mensual proyectado que algún integrante no pueda sostener.
- **Mitigación:** evaluar primero la infraestructura disponible en la facultad; definir un techo de gasto mensual acordado antes de contratar nada; preferir servicios con capa gratuita.
- **Contingencia:** despliegue de costo cero, con el entorno de demo levantado solo durante las instancias de validación.

### R-10 — Los pasos de la planilla no se reducen a acciones abstractas
**Categoría:** Técnico · **P:** Baja · **S:** Alta · **Exposición:** 3 · **Estado:** Abierto

El modelo de configuración asume que todo paso cae en un conjunto chico de acciones (carga de documento, carga de dato, confirmación, aprobación, firma). Si la realidad es más irregular, la configuración se vuelve específica por trámite.

- **Disparador:** al modelar el segundo trámite aparecen más de dos tipos de acción nuevos.
- **Mitigación:** construir el inventario de acciones a partir de la planilla completa antes de fijar el modelo, no a partir de un solo trámite.
- **Contingencia:** admitir un tipo de paso genérico con instrucción libre y confirmación manual, que degrada la automatización pero no rompe el modelo.

### R-12 — Retrasos en la comunicación interna y con los directores
**Categoría:** Proceso · **P:** Baja · **S:** Media · **Exposición:** 2 · **Estado:** Mitigándose

- **Disparador:** una reunión quincenal cancelada sin reprogramar dentro de la semana.
- **Mitigación:** hilo único de correo con los directores, reunión quincenal fija en calendario y canal directo con la Secretaría.

### R-13 — Los sprints 3, 4 y 5 están planificados por encima de la capacidad estimada
**Categoría:** Planificación · **P:** Alta · **S:** Alta · **Exposición:** 9 · **Estado:** Abierto

Los sprints 3, 4 y 5 piden 26, 30 y 29 SP contra una capacidad estimada de 24 por sprint pleno. Todo lo que tienen alimenta la demo de diciembre, así que se decidió no recortar nada y asumir el exceso. El ritmo se planificó sin ninguna medición de velocidad, y el pico llega antes de tenerla.

- **Disparador:** velocidad del sprint 2 por debajo del 80% de los 26 SP planificados.
- **Mitigación:** comprometer el sprint 1 con la regla del 60% y reescalar la curva entera al cierre del sprint 2, antes de entrar al pico.
- **Contingencia:** correr al bloque 2 las historias del bloque 1 que no están en el camino crítico, protegiendo el segundo trámite y el asistente mínimo que requiere la demo de diciembre.

### R-15 — Pérdida de contexto durante el receso de diciembre a febrero
**Categoría:** Recursos · **P:** Alta · **S:** Media · **Exposición:** 6 · **Estado:** Abierto

Entre el 14 de diciembre y el 7 de febrero hay tres sprints sin carga y uno de carga baja. Al volver, el equipo tiene que reconstruir el estado del código, de las decisiones y de los compromisos con la Secretaría.

- **Disparador:** velocidad del sprint 10 por debajo del 70% de lo planificado.
- **Mitigación:** cerrar el sprint 6 con todo mergeado, sin ramas abiertas, con el archivo de contexto del agente de desarrollo actualizado (US02) y la primera validación con la Secretaría ya agendada. El sprint 10 está planificado con historias chicas y cerradas a propósito.
- **Contingencia:** tratar la primera semana del sprint 10 como reincorporación, sin compromiso de entrega.

### R-16 — La primera validación le muestra a la Secretaría una herramienta pensada para el titular
**Categoría:** Stakeholders · **P:** Alta · **S:** Media · **Exposición:** 6 · **Estado:** Abierto

El titular tiene casi todo su mapa resuelto en el bloque 1; el Secretario tiene la mitad del suyo en marzo y abril. En la validación del sprint 10, la Secretaría puede percibir que la herramienta no le resuelve su trabajo.

- **Disparador:** feedback de la primera validación centrado en lo que le falta al Secretario.
- **Mitigación:** preparar la demo sobre lo que el Secretario ya tiene (bandeja, verificación manual, derivación y retorno, historial) y mostrar explícitamente cuándo llegan validación asistida y preactas.
- **Contingencia:** adelantar al sprint 11 alguna historia barata del Secretario que hoy está en marzo o abril.

### R-14 — Las horas comprometidas no cierran con el cierre en mayo
**Categoría:** Recursos · **P:** Media · **S:** Media · **Exposición:** 4 · **Estado:** Abierto

El compromiso es de 384 horas por integrante. Con once semanas de carga baja o nula y el trabajo cerrando en mayo, las horas entre septiembre y mayo no alcanzan. El anteproyecto declara cierre aproximado en julio de 2027.

- **Disparador:** registro de horas acumulado al sprint 11 por debajo de la proyección necesaria.
- **Mitigación:** confirmar con la cátedra que junio y julio cuentan como trabajo del proyecto. Registrar horas por integrante desde el sprint 1.
- **Contingencia:** extender el trabajo de cierre y preparación de la defensa a junio y julio con horas registradas.

---

## Riesgos cerrados

### R-11 — El proveedor de identidad institucional no resulta viable
**Categoría:** Dependencias · **Estado:** Cerrado el 9/10/2026

La Subsecretaría de TICs confirmó que SIGTA puede usar el ingreso con la cuenta de Google de la UBA, y todos los titulares tienen cuenta UBA (ADR-010).

---

## Asignación de dueños

| Riesgo | Dueño propuesto |
|---|---|
| R-01, R-05, R-12, R-13, R-14, R-15 | Responsable de gestión del proyecto |
| R-02, R-03, R-04, R-08, R-16 | Responsable del canal con la Secretaría |
| R-06, R-07, R-10 | Responsable técnico de IA y dominio |
| R-09 | Responsable de arquitectura e infraestructura |

> Los nombres se completan una vez acordado el reparto interno de responsabilidades.
