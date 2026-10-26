# Stakeholders y plan de comunicación

---

## 1. Mapa de stakeholders

| Stakeholder | Interés | Influencia | Qué necesita de nosotros | Qué necesitamos de ellos |
|---|---|---|---|---|
| Secretario de Gestión Académica (Lucas Adrián Macías) | Alto | Alta | Que el sistema resuelva un problema real de su área sin romper su operación | Priorización de trámites, acceso a su equipo, validación funcional |
| Contacto operativo de la Secretaría (Lucas Martín Diaz) | Alto | Media | Claridad sobre qué se le pide y cuándo | Detalle de circuitos, plantillas, coordinación de reuniones |
| Personal de las áreas que gestionan los trámites | Alto | Media | Que la herramienta les simplifique el día, no que les agregue pasos | Descripción del trabajo real, no del procedimiento formal |
| Directores del proyecto | Alto | Alta | Avance verificable, rigor metodológico, entregables a tiempo | Asesoría en sistemas con IA, validación de entregables, criterio académico |
| Cátedra de Proyectos Finales | Medio | Alta | Cumplimiento del anteproyecto aprobado y del formato de entrega | Reglas de entrega y fechas |
| Subsecretaría de TICs de la FIUBA | Bajo | Media | Pedidos concretos y acotados | Coordinación del entorno de CI/CD, hosting e identidad institucional |
| Titulares de trámites | Alto | Baja | Una plataforma que funcione | Feedback sobre usabilidad, si se consigue acceso a alguno |
| Equipo de proyecto | Alto | Alta | — | — |

**Lectura del mapa.** El cuadrante crítico es interés alto + influencia alta: secretario, directores y cátedra. El riesgo silencioso está en el personal de las áreas: interés alto, influencia media, y son quienes pueden hundir la adopción si la herramienta les cambia el trabajo sin haberlos consultado. La reunión del 16/09 lo señaló: hay que entender cómo trabajan hoy antes de irrumpirles el flujo.

---

## 2. Plan de comunicación

| Con quién | Canal | Frecuencia | Contenido | Registro |
|---|---|---|---|---|
| Directores | Google Meet | Quincenal, fija | Estado de tareas comprometidas, seguimiento de riesgos, demo del estado actual, compromiso del sprint siguiente | Minuta versionada en el repositorio |
| Directores | Gmail, hilo único | Según necesidad | Consultas puntuales asincrónicas | El propio hilo |
| Secretaría | Reunión (modalidad a acordar) | Al inicio de cada fase (relevamiento) y al cierre de cada entregable (validación) | Circuitos, requisitos, plantillas, pruebas de aceptación | Minuta versionada, enviada por correo para confirmación |
| Subsecretaría de TICs | Correo, vía contacto de la Secretaría | Puntual | Identidad institucional, hosting, permisos | Registro en decisiones pendientes |
| Equipo, sincrónico | Discord | Semanal + según necesidad | Planificación, revisión, desbloqueo | Decisiones relevantes al ADR |
| Equipo, asincrónico | WhatsApp | Continuo | Coordinación operativa | No se registra |
| Equipo, trabajo | Linear, GitHub | Continuo | Tickets, PRs, bugs | Herramienta |

**Regla de registro.** Toda conversación con un stakeholder externo que produzca una definición se asienta el mismo día: minuta en el repositorio y, si cambia el producto, entrada en el ADR. Una definición que solo vive en la memoria de quien participó no existe.

---

## 3. Ceremonias del proyecto

| Ceremonia | Cuándo | Quiénes | Salida |
|---|---|---|---|
| Planificación de sprint | Día 1 del sprint | Equipo | Sprint backlog comprometido, con capacidad real declarada |
| Sincronización | Semanal | Equipo | Bloqueos identificados |
| Revisión y demo | Último día del sprint | Equipo + directores (quincenal) | Incremento demostrado, feedback registrado |
| Retrospectiva | Último día del sprint | Equipo | Una a tres acciones concretas, con dueño |
| Revisión de riesgos | Reunión quincenal | Equipo + directores | Registro de riesgos actualizado |
| Refinamiento del backlog | Mitad del sprint | Equipo | Tickets del sprint siguiente con criterios de aceptación escritos |

---

## 4. Responsabilidades internas propuestas

Reparto tentativo, a confirmar por el equipo. La idea no es especializar el trabajo técnico, que se reparte, sino asignar **dueño único** a las cosas que sin dueño no pasan.

| Responsabilidad | Qué implica |
|---|---|
| Gestión del proyecto | Mantiene artefactos de seguimiento, indicadores de proceso, registro de riesgos y agenda de las reuniones quincenales |
| Canal con la Secretaría | Único interlocutor para pedidos y coordinación, redacta y envía las minutas |
| Arquitectura e infraestructura | Decisiones de stack, despliegue, CI/CD, almacenamiento |
| Dominio y calidad | Modelo de trámites, escenarios de aceptación, criterios de aceptación de los tickets |

**Por qué un solo interlocutor con la Secretaría.** Porque cuatro personas escribiendo al mismo contacto genera ruido, pedidos duplicados y respuestas contradictorias. La información entra por una puerta y se comparte adentro.

---

## 5. Escalamiento

1. Bloqueo técnico: se resuelve en el equipo dentro de las 48 hs.
2. Bloqueo por falta de definición del cliente: lo levanta el canal con la Secretaría, con fecha límite.
3. Si no hay respuesta en una semana: se lleva a la reunión quincenal con los directores.
4. Si compromete el alcance: se asienta como ADR con la decisión tomada bajo incertidumbre y el supuesto asumido.
