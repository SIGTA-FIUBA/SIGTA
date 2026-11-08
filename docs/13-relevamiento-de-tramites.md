# Relevamiento de los trámites

Cómo circulan hoy los cinco trámites, según el relevamiento que el equipo armó con la información de la Secretaría ("Flujos de los trámites"). Es la base para configurar cada tipo de trámite. Las dudas que quedaron abiertas, y cómo se resuelven por ahora, están en [dudas abiertas](14-dudas-abiertas.md).

**Estado.** La Secretaría todavía no validó este relevamiento. El 4/11 respondió sobre roles y datos personales y ofreció una reunión con las áreas para relevar el circuito en detalle; ahí se valida.

**Los estados no son del modelo.** Los nombres de la columna "Pasa a" (`EN_CC`, `APROBADO_CD`, etc.) son los que usa hoy el circuito en papel. En SIGTA el estado se deriva del paso abierto, y el resultado del trámite no se modela (ADR-009): no se copian al código.

---

## 1. Actores

| Actor | Qué hace | Usa SIGTA | En el modelo |
| --- | --- | --- | --- |
| Estudiante | Inicia el trámite y sube la documentación | Sí | Titular (área `applicant`) |
| Dirección de Alumnos (DA) | Registra, deriva, redacta y carga en el SIU | Sí | Área `office` |
| Directora General de Alumnos | Revisa el cierre | Sí | Área `office`, con un rol propio que asigna el administrador |
| Intercambios Académicos | Recibe y registra los intercambios | Sí | Área `office` |
| Decano y Secretarios | Firman resoluciones | Sí, pero firman fuera de SIGTA | La DA confirma la firma como hecho externo |
| Comisión Curricular (CC) | Dictamina | No | Área `external` |
| Despacho del Consejo Directivo | Arma la resolución y la lleva al Consejo | No | Área `external` |
| Consejo Directivo (CD) | Trata y aprueba en sesión | No | Área `external` |
| Rectorado (Dpto. de Alumnos de la UBA) | Recibe los pases y los manda a FIUBA | No | Área `external` |
| Ciclo Básico Común (CBC) | Adjunta el analítico en los pases dentro de la UBA | No | Área `external` |
| Mesa de Entrada | Archiva el expediente | No | Área `external` |

Los actores que no usan SIGTA no cambian nada en la plataforma: cuando el expediente vuelve, la DA confirma en SIGTA lo que pasó afuera (ADR-004).

## 2. Acciones

El relevamiento usa nueve acciones. Todas entran en las cuatro del modelo:

| Acción del relevamiento | Acción del modelo |
| --- | --- |
| Iniciar, Cargar archivo | `enter_data`, `upload_document` del titular |
| Pedir documentación | Una salida de `decide` que vuelve al paso del titular |
| Registrar expediente | `enter_data` (número de expediente) |
| Derivar | Pasar a un paso de un área `external` |
| Generar documento | `upload_document` de la DA (proyecto o resolución) |
| Firmar | `confirm_external_fact`: la firma ocurre fuera de SIGTA |
| Marcar como completado | `confirm_external_fact` (dictamen recibido, carga en el SIU, visto bueno) |
| Archivar | Último paso, de Mesa de Entrada (`external`) |

## 3. Circuitos

Los cinco terminan igual: la DA confirma la carga en el SIU, la Directora General da el visto bueno y Mesa de Entrada archiva.

### 3.1 Equivalencia entre asignaturas

**Entrada:** nota al Decano. **Salida:** resolución del CD cargada en el SIU.

| # | Actor | Qué pasa | Genera |
| --- | --- | --- | --- |
| 1 | Estudiante | Inicia y carga los datos y el listado de aprobadas | Solicitud |
| 2 | DA | Registra el expediente y lo pasa a la CC | Expediente |
| 3 | CC | Dictamina | Acta |
| 4 | DA | Sube el proyecto de resolución y lo pasa a Despacho | Proyecto de resolución |
| 5 | Despacho del CD | Arma la resolución definitiva y la pone en el orden del día | Resolución definitiva |
| 6 | CD | La trata y la aprueba en sesión | Resolución del CD |
| 7 | Despacho del CD | Vuelve a la DA | — |
| 8 | DA | Confirma la carga en el SIU | Registro en el SIU |
| 9 | Directora General | Da el visto bueno | — |
| 10 | Mesa de Entrada | Archiva | Expediente en guarda temporal |

**Qué pide la nota al Decano** (plantilla de la DA): fecha, objeto, nombre, legajo, carrera, plan, qué se solicita, firma y datos de contacto (teléfono, domicilio, localidad, DNI, año de ingreso y correo). Se adjunta el listado de asignaturas aprobadas (código, nombre, libro, folio, fecha y calificación), el de las que cursa y, si hace falta, una nota que amplíe el pedido.

**Cómo queda en SIGTA** (decidido por el equipo el 8/11, a confirmar con la Secretaría):

- La nota la arma SIGTA con los datos que carga el estudiante: qué solicita, legajo, carrera, plan y año de ingreso. Nombre y correo salen del ingreso con la cuenta de la UBA.
- No se piden domicilio, localidad, teléfono, DNI ni firma: no hacen falta para tramitar (Ley 25.326, recolectar solo lo necesario) y el ingreso con la cuenta de la UBA identifica a quien pide.
- El listado de aprobadas se sube como el PDF del historial académico que descarga el SIU Guaraní. Las materias que cursa se eligen en el formulario.
- La completitud la revisa SIGTA con reglas en código, sin IA: no deja enviar si falta un dato, si el PDF no tiene texto, si no tiene el legajo del estudiante o si no tiene filas de materias. Por eso no hay un paso de revisión de la DA antes de abrir el expediente. Si la DA ve un problema después, le pide una subsanación al estudiante.
- Dictamina una sola CC, la de la carrera de destino. Equivalencias para dos carreras son dos trámites.
- Las equivalencias automáticas quedan fuera por ahora.

### 3.2 Reconocimiento de créditos

**Entrada:** nota al Decano con el motivo. **Salida:** resolución de Decano cargada en el SIU.

| # | Actor | Qué pasa | Genera |
| --- | --- | --- | --- |
| 1 | Estudiante | Inicia y carga la solicitud | Solicitud |
| 2 | DA | Registra el expediente y lo pasa a la CC | Expediente |
| 3 | CC | Dictamina | Acta |
| 4 | DA | Sube la resolución de Decano | Resolución de Decano |
| 5 | Decano y Secretario Académico | Firman | Resolución firmada |
| 6 | DA | Confirma la carga en el SIU | Registro en el SIU |
| 7 | Directora General | Da el visto bueno | — |
| 8 | Mesa de Entrada | Archiva | Expediente en guarda temporal |

A diferencia de la equivalencia, no pasa por el CD: lo resuelve el Decano.

### 3.3 Intercambios académicos

**Entrada:** pedido a Intercambios Académicos por correo. **Salida:** resolución de Decano ratificada por el CD y cargada en el SIU.

| # | Actor | Qué pasa | Genera |
| --- | --- | --- | --- |
| 1 | Estudiante | Hace el pedido | Pedido |
| 2 | Intercambios Académicos | Pide datos personales y actividades en destino | Pedido de documentación |
| 3 | Estudiante | Sube lo pedido | Documentación |
| 4 | Intercambios Académicos | Registra el expediente y lo pasa a la CC | Expediente |
| 5 | CC | Dictamina | Acta |
| 6 | DA | Sube la resolución de Decano | Resolución de Decano |
| 7 | Decano y Secretario Académico | Firman | Resolución firmada |
| 8 | Despacho del CD | La pone en el orden del día para ratificarla | — |
| 9 | DA | Confirma la carga en el SIU | Registro en el SIU |
| 10 | Directora General | Da el visto bueno | — |
| 11 | Mesa de Entrada | Archiva | Expediente en guarda temporal |

### 3.4 Pase de universidad extranjera

**Entrada:** nota al Rector, presentada en Rectorado. **Salida:** resolución de Decano ratificada por el CD y cargada en el SIU.

| # | Actor | Qué pasa | Genera |
| --- | --- | --- | --- |
| 1 | Estudiante | Presenta la nota en Rectorado | Nota al Rector |
| 2 | Rectorado | Pasa el expediente a FIUBA | Expediente |
| 3 | DA | Adjunta las materias pedidas y lo pasa a las CC | Materias pedidas |
| 4 | CC | Dictamina. Si falta otra carrera, vuelve a pasar a la CC siguiente | Acta |
| 5 | DA | Sube el proyecto con todas las equivalencias | Proyecto de resolución |
| 6 | Decano y Secretario de Gestión Académica | Firman | Resolución de Decano |
| 7 | Despacho del CD | La pone en el orden del día para ratificarla | — |
| 8 | DA | Confirma la carga en el SIU | Registro en el SIU |
| 9 | Directora General | Da el visto bueno | — |
| 10 | Mesa de Entrada | Archiva | Expediente en guarda temporal |

Es el único circuito con varias CC, una después de otra.

### 3.5 Pase de universidad dentro de la UBA

**Entrada:** DNI, constancia de inscripción, analítico y nota, presentados en Rectorado. **Salida:** resolución de Decano cargada en el SIU.

| # | Actor | Qué pasa | Genera |
| --- | --- | --- | --- |
| 1 | Estudiante | Presenta la solicitud y la documentación en Rectorado | Solicitud y documentación |
| 2 | Rectorado | Acepta, valida y pasa al CBC | Resolución del Consejo Superior |
| 3 | CBC | Adjunta el analítico y pasa a FIUBA por Mesa de Entrada | Analítico |
| 4 | DA | Adjunta la documentación y pasa a las CC | — |
| 5 | CC | Dictaminan | Actas |
| 6 | DA | Sube el proyecto de resolución de Decano | Proyecto de resolución |
| 7 | Decano y Secretario de Gestión Académica | Firman | Resolución de Decano |
| 8 | DA | Confirma la carga en el SIU | Registro en el SIU |
| 9 | Directora General | Da el visto bueno | — |
| 10 | Mesa de Entrada | Archiva | Expediente en guarda temporal |

**Pases que empiezan en Rectorado.** El estudiante inicia el trámite en SIGTA cuando presenta en Rectorado. Mientras está afuera, el trámite queda en un paso de un área `external`, que la DA cierra cuando el expediente llega a FIUBA. Así el estudiante lo sigue desde el principio.
