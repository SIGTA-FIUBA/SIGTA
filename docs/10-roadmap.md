# Roadmap de sprints

**Estado al 27 de septiembre de 2026:** el proyecto está en marcha. El sprint 1 arrancó el 21 de septiembre y va por su segunda semana.

**Sprint 0:** 7 al 20 de septiembre de 2026.
**Sprint 1:** arranca el lunes 21 de septiembre de 2026.
**Cierre del trabajo:** 30 de mayo de 2027 (sprint 18). Defensa en junio y julio.

---

## 1. La secuencia

El orden lo define el producto, no la comodidad de la curva:

1. **Primer trámite**, el trabajo pesado: modelo de dominio, motor de pasos, gestión documental y derivación externa.
2. **Segundo trámite**, que confirma que la abstracción generaliza.
3. **Asistente RAG mínimo funcionando antes de cerrar 2026**, sobre requisitos configurados y base de conocimiento.
4. **Cierre del asistente** apenas se vuelve del receso, en febrero.
5. **Validación asistida y preactas** en marzo, donde está la capacidad.
6. **Trámites 3, 4 y 5**, uno por sprint, con carga baja, ya con el inventario de acciones y la IA definidos.
7. **Validación final y cierre** en mayo.

## 2. Fases del anteproyecto

| Fase | Sprints | Qué la cierra |
|---|---|---|
| 1. Relevamiento y definición | 0 a 2 | Relevamiento de los cinco trámites, mapa de roles, stack, inventario de acciones |
| 2. Diseño e infraestructura base | 1 a 3 | Repositorio con CI/CD, herramientas agénticas, modelo de definición de trámite |
| 3. Desarrollo incremental del MVP | 3 a 16 | Los cinco trámites y los tres usos de IA |
| 4. Validación y ajuste | 17 | Pruebas de aceptación, además de las validaciones intermedias de los sprints 10 y 14 |
| 5. Cierre y documentación | 18 y defensa | Informe final; la documentación del producto ya existe (ADR-007) |

Las fases describen el tipo de trabajo y se superponen. La unidad de planificación son los bloques.

## 2 bis. Bloques

| Bloque | Sprints | SP | Qué cierra |
|---|---|---|---|
| B1 Esqueleto y asistente mínimo | 1 a 6 | 138 | Dos trámites configurados, trámite recorrido con derivación externa, asistente respondiendo |
| B2 Cierre del asistente | 10 y 11 | 24 | Citas, derivación a persona, set de evaluación e inyección de prompt |
| B3 Validación asistida, preactas y cobertura | 12 a 16 | 94 | Los tres usos de IA completos y los cinco trámites andando |
| B4 Validación y cierre | 17 y 18 | 8 más reserva | Pruebas con la Secretaría, correcciones e informe |
| Defensa | jun-jul 2027 | sin historias | Ventana reservada para preparar la defensa |
| **Total** | | **264** | |

## 3. Resultado de la revisión

Las historias se revisaron una por una con el equipo, con sus criterios de aceptación, dependencias y estimación. El proyecto pasó de 63 historias y 307 SP a **53 historias y 264 SP**. Diez historias desaparecieron: la mayoría no se perdió, se absorbió en otra. El informe final, los ajustes posteriores a la validación y la preparación de la defensa salieron del backlog: los sprints 17 y 18 y la ventana de junio y julio quedan reservados para eso, sin historias.

La curva quedó en **17, 19, 26, 30, 29, 17** en la primera mitad y **13, 11, 18, 24, 24, 14, 14, 8** en la segunda. Febrero bajó a 24 SP entre sus dos sprints, que es lo que corresponde a carga baja; marzo y abril se llenaron hasta el tope.

**Los sprints 3, 4 y 5 quedan por encima de la capacidad estimada de 24 SP.** Todo lo que tienen alimenta la demo de diciembre, así que no se recortó nada: el punto de control del sprint 2 decide, con velocidad medida, si se sostienen o se corre algo.

## 4. Curva

| Sprint | SP | Forma |
|---|---|---|
| 1 | 17 | ████████ |
| 2 | 19 | ██████████ |
| 3 | 26 | █████████████ |
| 4 | 30 | ███████████████ |
| 5 | 29 | ██████████████ |
| 6 | 17 | ████████ |
| 7 | — | · sin carga |
| 8 | — | · sin carga |
| 9 | — | · sin carga |
| 10 | 13 | ██████ |
| 11 | 11 | ██████ |
| 12 | 18 | █████████ |
| 13 | 24 | ████████████ |
| 14 | 24 | ████████████ |
| 15 | 14 | ███████ |
| 16 | 14 | ███████ |
| 17 | 8 | ████ |
| 18 | — | · reservado |

- **Pico temprano**, sprints 3 y 4, porque el primer trámite es el trabajo pesado y va al principio.
- **El pozo** es diciembre desde el 14, enero entero y carga baja en febrero.
- **La bajada final** son los sprints 16 a 18. El sprint 16 queda en 14 SP; el 17 compromete solo 8 SP y reserva el resto para absorber el feedback de la validación; el 18 no tiene historias y queda para el informe y los ajustes no bloqueantes.

## 4 bis. Ciclos en Linear

Los ciclos del equipo en Linear no arrancan el 21 de septiembre, porque ese día ya había pasado cuando se configuraron. El ciclo 1 cubre la semana que quedaba del sprint 1, del 28 de septiembre al 4 de octubre, y de ahí en adelante coinciden uno a uno: el ciclo 2 es el sprint 2 y así hasta el 18.

## 5. Calendario

| Sprint | Desde | Hasta | Bloque | SP | Carga |
|---|---|---|---|---|---|
| 0 | 7 sep 2026 | 20 sep 2026 | Artefactos de gestión | — | Plena |
| 1 | 21 sep | 4 oct 2026 | B1 Esqueleto y RAG mínimo | 17 | Plena |
| 2 | 5 oct | 18 oct | B1 Esqueleto y RAG mínimo | 19 | Plena |
| 3 | 19 oct | 1 nov | B1 Esqueleto y RAG mínimo | 26 | Plena |
| 4 | 2 nov | 15 nov | B1 Esqueleto y RAG mínimo | 30 | Plena |
| 5 | 16 nov | 29 nov | B1 Esqueleto y RAG mínimo · **entrega intermedia** | 29 | Plena |
| 6 | 30 nov | 13 dic | B1 Esqueleto y RAG mínimo · **cierra B1** | 17 | Reducida (diciembre) |
| 7 | 14 dic | 27 dic | — | — | Nula |
| 8 | 28 dic | 10 ene 2027 | — | — | Nula |
| 9 | 11 ene | 24 ene | — | — | Nula |
| 10 | 25 ene | 7 feb | B2 Cierre del asistente | 13 | Baja (febrero) |
| 11 | 8 feb | 21 feb | B2 Cierre del asistente · **cierra B2** | 11 | Baja (febrero) |
| 12 | 22 feb | 7 mar | B3 Validación asistida, preactas y cobertura | 18 | Plena |
| 13 | 8 mar | 21 mar | B3 Validación asistida, preactas y cobertura | 24 | Plena |
| 14 | 22 mar | 4 abr | B3 Validación asistida, preactas y cobertura | 24 | Plena |
| 15 | 5 abr | 18 abr | B3 Validación asistida, preactas y cobertura | 14 | Plena |
| 16 | 19 abr | 2 may | B3 Validación asistida, preactas y cobertura · **cierra B3** | 14 | Plena |
| 17 | 3 may | 16 may | B4 Validación final y cierre | 8 | Plena |
| 18 | 17 may | 30 may | B4 Validación final y cierre · **cierra B4** | — | Plena |
| — | 31 may | julio 2027 | Defensa | 5 | Según disponibilidad |

## 6. Hitos

| Hito | Sprint | Qué tiene que estar |
|---|---|---|
| Relevamiento con la Secretaría | 1 | Recorrido real de los cinco trámites, roles del área, plantillas de documentos |
| Coordinación con el Subsecretario de TICs de la FIUBA | 1 | Entorno de CI/CD, hosting e identidad institucional |
| **Entrega intermedia de la cátedra** | 5 | Video de 8 a 10 minutos con el estado del trabajo, entregado el 29/11 antes de las 21 h. Ver la sección 6 bis |
| **Demo de diciembre ante los directores** | 6 | Dos trámites configurados y el esqueleto del agente de IA |
| Primera validación con la Secretaría | 10 | Feedback sobre el flujo de los dos trámites |
| Demo del asistente cerrado | 11 | Respuestas con fuentes citadas y derivación fuera de cobertura |
| Segunda validación con la Secretaría | 14 | Validación asistida y preactas, y agenda de las pruebas de aceptación |
| Validación final | 17 | Pruebas de aceptación sobre los cinco trámites |
| Informe final | 18 | Entregado |

## 6 bis. Entrega intermedia

Entrega obligatoria de la cátedra al cierre del primer cuatrimestre. Es un punto de control del estado del trabajo a esa fecha, así que conviene prepararla en la última semana del sprint 5 y no antes.

- **Vence:** domingo 29 de noviembre de 2026, a las 21 h. Coincide con el último día del sprint 5.
- **Formato:** video MP4 de 8 a 10 minutos, calidad media (no full HD), de menos de 128 MB. Estilo libre.
- **Expositores:** hablan los cuatro integrantes, por turnos.
- **Contenido mínimo:**
  - objetivo general del trabajo;
  - estado de avance, cualitativo o con indicadores (velocidad, SP completados, burnup);
  - breve demostración de la versión actual del producto;
  - lecciones aprendidas y riesgos observados (registro de riesgos).
- **Cómo se entrega:** el video se sube en forma privada a YouTube o Google Drive, y el enlace se carga en https://tpf.fiuba-misc.com.ar/, en el botón "Entrega intermedia". Para usar otra plataforma hay que avisarle antes por correo a la coordinación de la materia.
- Los lineamientos son generales: los directores y el equipo pueden adaptarlos.

## 7. Puntos de control

| Cuándo | Qué se evalúa | Qué se hace si falla |
|---|---|---|
| Fin de sprint 2 | Velocidad medida contra 17 y 19 SP | Se reescala la curva entera. Es el punto de control que sostiene todo el plan |
| Fin de sprint 4 | ¿El segundo trámite entró solo con configuración? | Si exigió código, se corrige la abstracción antes de seguir |
| Fin de sprint 6 | ¿Hay dos trámites y asistente para la demo? | Se corre el bloque 2 y sale la generación de preactas |
| Fin de sprint 13 | ¿Los tres usos de IA cerrados? | Sale lo que quede de preactas |
| Fin de sprint 14 | ¿El pase de universidad extranjera entró con sus 5 SP, solo con configuración? | Si superó los 8 SP reales, la abstracción no cubre el caso complejo: se corrige antes de intercambios y pase UBA |
| Fin de sprint 16 | ¿Los cinco trámites end-to-end? | La cobertura baja a tres y el resto se documenta como trabajo futuro |

## 8. Riesgo abierto del calendario

Con once semanas de carga baja o nula, las 384 horas por integrante entre septiembre y mayo no cierran solas: dependen de que junio y julio cuenten como trabajo del proyecto. El anteproyecto declara cierre aproximado en julio de 2027, no en mayo. Conviene verificarlo con la cátedra antes de fijar el calendario.
