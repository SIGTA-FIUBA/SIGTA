# User Story Mapping

**Versión:** 0.3 · orientado a roles

## Cómo se lee

Dos mapas, uno por cada rol de la plataforma. Cada mapa tiene tres niveles: **actividad** (el recorrido del rol en el tiempo), **tarea** (lo que el rol necesita lograr) y **detalle** (la forma concreta de resolverlo). Cada detalle apunta a la historia del backlog que lo implementa y al sprint en que cae.

| Marca | Significa |
|---|---|
| `US14 · S3` | Lo resuelve esa historia, en ese sprint |
| `—` | Fuera del alcance. Está en el mapa a propósito, para discutir alcance |
| `—*` | Fuera del alcance, pero es una duda pendiente con la Secretaría |

**Lo que no está en estos mapas.** Infraestructura, CI/CD, configuración de trámites, modelo de dominio y base de conocimiento no son tareas de ningún rol: viven solo en el backlog. El evaluador externo tampoco es un rol de la plataforma: sus pasos los registra el Secretario.

**Sobre el rol Secretario.** En el mapa es un solo rol. En permisos siguen existiendo dos perfiles: el **operativo** (recibe, verifica, deriva, registra) y el de **aprobación** (aprueba y firma). La columna Permiso marca qué detalles requieren el segundo.

---

# Mapa 1 — Titular del trámite

**Quién es.** Alguien que necesita resolver un trámite y entra pocas veces, espaciadas, probablemente desde el celular. Lo que le duele hoy: requisitos dispersos, no saber en qué está su trámite, y plazos que no ve hasta que los pierde.

| TA1. Descubrir | TA2. Preparar | TA3. Iniciar | TA4. Seguir | TA5. Responder | TA6. Recibir |
|---|---|---|---|---|---|
| Qué trámite le corresponde y qué pide | Conseguir la documentación | Presentar la solicitud | Saber en qué está sin preguntar | Completar lo que le piden | Obtener el resultado |

## TA1 — Descubrir

### TT1.1 — Encontrar el trámite que le corresponde

| ID | Detalle | Historia |
|---|---|---|
| TIT-1.1.a | Ver el catálogo de los cinco trámites en lenguaje llano | US13 · S3 |
| TIT-1.1.b | Ver en qué se diferencia cada trámite de los otros, porque los cinco se parecen y elegir mal cuesta semanas | US13 · S3 |
| TIT-1.1.c | Describir su situación al asistente y recibir el trámite que corresponde | US69 · S6 |
| TIT-1.1.d | Ver si ya tiene un trámite abierto del mismo tipo antes de iniciar otro | US22 · S6 |

### TT1.2 — Entender qué le piden

| ID | Detalle | Historia |
|---|---|---|
| TIT-1.2.a | Ver la lista de requisitos, con los obligatorios marcados | US13 · S3 |
| TIT-1.2.b | Ver el formato y el tamaño aceptados de cada documento | US13 · S3 |
| TIT-1.2.c | Ver plazos y condiciones de admisibilidad del trámite | US13 · S3 |
| TIT-1.2.d | Preguntarle al asistente qué significa un requisito escrito en lenguaje administrativo | US69 · S6 |
| TIT-1.2.e | Recibir con cada respuesta la normativa o el instructivo del que sale | US69 · S6 |

## TA2 — Preparar

### TT2.1 — Saber cómo conseguir cada documento

| ID | Detalle | Historia |
|---|---|---|
| TIT-2.1.a | Preguntarle al asistente dónde y cómo se obtiene cada documento | US69 · S6 |
| TIT-2.1.b | Saber qué documentación extranjera necesita legalización o apostilla | US52 · S16 |
| TIT-2.1.c | Ser derivado a una persona cuando la consulta excede la base de conocimiento | US69 · S6 |
| TIT-2.1.d | Que la plataforma tramite el documento ante otra institución | — |

### TT2.2 — Resolver dudas sin escribir un mail

| ID | Detalle | Historia |
|---|---|---|
| TIT-2.2.a | Conversar con el asistente sobre requisitos y documentación | US69 · S6 |
| TIT-2.2.b | Ver el historial de la conversación al volver | — |
| TIT-2.2.c | Que el asistente conozca el estado de su trámite y responda sobre él | — |

## TA3 — Iniciar

### TT3.1 — Entrar

| ID | Detalle | Historia |
|---|---|---|
| TIT-3.1.a | Iniciar sesión | US06 · S2 |
| TIT-3.1.b | Iniciar sesión con la cuenta de Google de la UBA | US06 · S2 |

### TT3.2 — Completar y adjuntar

| ID | Detalle | Historia |
|---|---|---|
| TIT-3.2.a | Completar el formulario de datos que define la configuración del trámite | US14 · S3 |
| TIT-3.2.b | Validación de campos obligatorios y formato antes de avanzar | US14 · S3 |
| TIT-3.2.c | Adjuntar un archivo por cada requisito documental | US15 · S3 |
| TIT-3.2.d | Ver qué requisito queda cubierto por cada adjunto | US15 · S3 |
| TIT-3.2.e | Reemplazar o eliminar un adjunto antes de enviar | US14 · S3 |
| TIT-3.2.f | Guardar la solicitud como borrador y retomarla después | — |

### TT3.3 — Enviar

| ID | Detalle | Historia |
|---|---|---|
| TIT-3.3.a | Ver un resumen de lo que está por enviar | US14 · S3 |
| TIT-3.3.b | No poder enviar si falta un requisito obligatorio | US14 · S3 |
| TIT-3.3.c | Recibir la confirmación con el identificador de la solicitud | US14 · S3 |
| TIT-3.3.d | Recibir la confirmación también por correo | US35 · S14 |

## TA4 — Seguir

### TT4.1 — Ver en qué está

| ID | Detalle | Historia |
|---|---|---|
| TIT-4.1.a | Ver sus trámites en curso como pantalla principal | US22 · S6 |
| TIT-4.1.b | Ver el paso actual y qué se espera ahora | US23 · S6 |
| TIT-4.1.c | Ver que su trámite está en evaluación académica fuera del área | US23 · S6 |
| TIT-4.1.d | Ver el recorrido completo de su trámite | US23 · S6 |

### TT4.2 — Enterarse sin entrar

| ID | Detalle | Historia |
|---|---|---|
| TIT-4.2.a | Recibir un correo cada vez que el trámite cambia de estado | US35 · S14 |

## TA5 — Responder

### TT5.1 — Completar una subsanación

| ID | Detalle | Historia |
|---|---|---|
| TIT-5.1.a | Recibir un correo cuando le piden algo | US36 · S11 |
| TIT-5.1.b | Ver qué le falta y por qué | US34 · S15 |
| TIT-5.1.c | Completarlo sobre el mismo trámite, sin abrir uno nuevo | US34 · S15 |

## TA6 — Recibir

### TT6.1 — Obtener el resultado

| ID | Detalle | Historia |
|---|---|---|
| TIT-6.1.a | Ver el estado final al entrar | US23 · S6 |
| TIT-6.1.b | Descargar el documento resolutivo | US16 · S5 |
| TIT-6.1.c | Descargar un comprobante del trámite completo | US38 · S16 |

### TT6.2 — Consultar el pasado

| ID | Detalle | Historia |
|---|---|---|
| TIT-6.2.a | Ver el listado de sus trámites cerrados | US24 · S11 |
| TIT-6.2.b | Buscar un trámite anterior por tipo o fecha | — |

---

# Mapa 2 — Secretario

**Quién es.** Personal de la Secretaría de Gestión Académica. Recibe solicitudes de fuentes dispersas, las centraliza a mano, deriva a evaluadores externos y redacta resoluciones. Lo que le duele hoy: no tener una bandeja única, no saber qué se está durmiendo, y redactar actas desde cero.

| SA1. Organizar | SA2. Verificar | SA3. Pedir | SA4. Derivar | SA5. Resolver | SA6. Cerrar |
|---|---|---|---|---|---|
| Saber qué requiere su intervención | Controlar que la solicitud esté en condiciones | Obtener lo que falta | Llevar el trámite al evaluador y traerlo de vuelta | Producir el documento formal | Dejar constancia |

## SA1 — Organizar

### ST1.1 — Ver qué requiere su intervención

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-1.1.a | Bandeja con lo accionable por su perfil | Operativo | US17 · S4 |
| SEC-1.1.b | Distinguir lo nuevo de lo ya visto | Operativo | US17 · S4 |
| SEC-1.1.c | Recibir un correo cuando entra algo a su ámbito | Operativo | US36 · S11 |

### ST1.2 — Priorizar

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-1.2.a | Ver todos los trámites por tipo y estado | Operativo | US17 · S4 |
| SEC-1.2.b | Ver la antigüedad de cada trámite, para detectar lo que se duerme | Operativo | US17 · S4 |
| SEC-1.2.c | Filtrar y ordenar por tipo, estado, antigüedad y responsable | Operativo | US37 · S4 |
| SEC-1.2.d | Ver destacados los trámites con plazo próximo a vencer | Operativo | US67 · S15 |

## SA2 — Verificar

### ST2.1 — Revisar la solicitud

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-2.1.a | Ver los documentos junto a la lista de requisitos del trámite | Operativo | US19 · S4 |
| SEC-2.1.b | Abrir y descargar cada documento | Operativo | US16 · S5 |
| SEC-2.1.c | Marcar cada requisito como cumplido o no cumplido | Operativo | US66 · S12 |

### ST2.2 — Verificar con asistencia

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-2.2.a | Recibir un checklist de hallazgos contra los requisitos configurados | Operativo | US41 · S12 |
| SEC-2.2.b | Ver de dónde sale cada hallazgo, para contrastarlo con el documento | Operativo | US41 · S12 |
| SEC-2.2.c | Aceptar o descartar cada hallazgo; la decisión siempre es suya | Operativo | US41 · S12 |
| SEC-2.2.d | Que la plataforma apruebe la documentación sin intervención humana | — | — |

### ST2.3 — Registrar la recepción

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-2.3.a | Confirmar el paso y hacer avanzar el estado | Operativo | US10 · S4 |
| SEC-2.3.b | Que quede registrado quién confirmó y cuándo | Operativo | US10 · S4 |
| SEC-2.3.c | Registrar el número de expediente externo asociado | Operativo | US10 · S4 |
| SEC-2.3.d | Que la plataforma cree la carátula en el sistema de expediente | — | — |

## SA3 — Pedir

### ST3.1 — Obtener lo que falta

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-3.1.a | Devolver el trámite a un paso anterior indicando el motivo | Operativo | US12 · S14 |
| SEC-3.1.b | Pedir una subsanación puntual sin cerrar ni reiniciar el trámite | Operativo | US33 · S15 |
| SEC-3.1.c | Ver el historial de subsanaciones pedidas y resueltas | Operativo | US11 · S4 |

## SA4 — Derivar

### ST4.1 — Llevar el trámite al evaluador

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-4.1.a | Ver en una pantalla todo lo que el evaluador necesita | Operativo | US19 · S4 |
| SEC-4.1.b | Registrar que el trámite salió, a qué evaluador y cuándo | Operativo | US10 · S4 |
| SEC-4.1.c | Que el estado refleje que el trámite está fuera del área | Operativo | US10 · S4 |
| SEC-4.1.d | Que el evaluador reciba y responda dentro de la plataforma | — | —* |

### ST4.2 — Seguir la espera

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-4.2.a | Ver cuánto tiempo lleva derivado sin retorno | Operativo | US17 · S4 |
| SEC-4.2.b | Recibir un recordatorio cuando una derivación supera el plazo esperado | Operativo | US53 · S16 |

### ST4.3 — Registrar el retorno

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-4.3.a | Registrar el retorno adjuntando el dictamen, sin interpretarlo | Operativo | US21 · S10 |
| SEC-4.3.b | Que el trámite avance al estado siguiente | Operativo | US21 · S10 |
| SEC-4.3.c | Trasladar al titular un pedido de ampliación del evaluador, como subsanación | Operativo | US33 · S15 |

## SA5 — Resolver

### ST5.1 — Redactar el documento resolutivo

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-5.1.a | Adjuntar el documento resolutivo producido fuera de la plataforma | Operativo | US15 · S3 |
| SEC-5.1.b | Generar una preacta desde la plantilla con los datos del trámite | Aprobación | US45 · S13 |
| SEC-5.1.c | Que los campos estructurados se completen sin modelo de lenguaje | Aprobación | US45 · S13 |
| SEC-5.1.d | Recibir asistencia del modelo solo en los campos de texto libre | Aprobación | US46 · S13 |
| SEC-5.1.e | Que la plataforma decida qué asignatura equivale a cuál | — | — |

### ST5.2 — Revisar y aprobar

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-5.2.a | Ver y editar cualquier campo del borrador antes de aprobar | Aprobación | US47 · S14 |
| SEC-5.2.b | Ver de qué dato del trámite sale cada campo completado | Aprobación | US47 · S14 |
| SEC-5.2.c | Aprobar, con registro de quién aprobó y cuándo | Aprobación | US10 · S4 |
| SEC-5.2.d | Que solo un perfil con permiso de aprobación pueda hacerlo | Aprobación | US07 · S3 |
| SEC-5.2.e | Firmar digitalmente con validez legal | — | — |

### ST5.3 — Reflejar la resolución en sistemas externos

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-5.3.a | Confirmar que la carga en el sistema académico se hizo, con registro de quién lo declaró | Operativo | US10 · S4 |
| SEC-5.3.b | Que la plataforma cargue la resolución en el sistema académico | — | — |

## SA6 — Cerrar

### ST6.1 — Dejar constancia

| ID | Detalle | Permiso | Historia |
|---|---|---|---|
| SEC-6.1.a | Ver el historial completo de pasos, con responsable y fecha | Operativo | US11 · S4 |
| SEC-6.1.b | Que el historial no se pueda alterar una vez registrado | Operativo | US11 · S4 |
| SEC-6.1.c | Consultar resoluciones anteriores como antecedente de un caso nuevo | — | — |

---

## Lo que muestran los mapas

**1. La plataforma comunica resultados, no los interpreta (ADR-009).** Un reconocimiento total, uno parcial o un rechazo son para la plataforma el mismo evento: un dictamen que vuelve del evaluador y se comunica al titular. Por eso el mapa del Secretario no tiene detalles sobre resultados, y SA4 termina en registrar el retorno.

**2. Todos los detalles están resueltos.** Cada detalle del mapa apunta a una historia del backlog o está declarado fuera del alcance. La revisión del mapa contra el backlog agregó dos historias (US66 verificación manual de requisitos, US67 plazos próximos a vencer) y absorbió otros tres detalles en historias existentes.

**3. Los mapas son asimétricos.** El Titular tiene casi todo resuelto en el bloque 1, antes de diciembre. El Secretario tiene la mitad de su mapa en marzo y abril: validación asistida, preactas, filtros, subsanaciones y recordatorios. La primera validación con la Secretaría, en el sprint 10, les va a mostrar una herramienta pensada más para el titular que para ellos.
