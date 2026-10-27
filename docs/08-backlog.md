# Product Backlog

**Fuente de verdad:** [Linear, equipo SIG](https://linear.app/sigta/team/SIG). La descripción de cada historia, sus criterios de aceptación, su estimación y su estado están en el ticket. Este documento es solo el índice por épica, para ubicar cada historia y su ticket.

**Estimación:** Story Points, Fibonacci. Una historia lleva 13 SP como excepción; el resto no supera 8.
**Criterios de aceptación:** mínimo 3, máximo 8. **Un criterio solo verifica lo que entrega su propia historia**, apoyándose en historias ya cerradas; nunca en funcionalidad de sprints posteriores.
**Fuera del backlog:** el informe final, los ajustes posteriores a la validación y la preparación de la defensa no son historias. Los sprints 17 y 18 y la ventana de defensa quedan reservados para eso.

---

## Épicas

| ID | Épica | Historias | SP |
|---|---|---|---|
| EP00 | Relevamiento y modelado del dominio | 4 | 10 |
| EP01 | Identidad y acceso | 2 | 11 |
| EP02 | Configuración de trámites | 6 | 34 |
| EP03 | Inicio de solicitud | 2 | 7 |
| EP04 | Seguimiento del titular | 5 | 13 |
| EP05 | Bandeja y verificación de la Secretaría | 6 | 18 |
| EP06 | Motor de pasos y estados | 3 | 13 |
| EP07 | Gestión documental | 2 | 13 |
| EP08 | Derivación a evaluación externa | 2 | 8 |
| EP09 | Asistente RAG al titular | 6 | 48 |
| EP10 | Validación asistida | 3 | 21 |
| EP11 | Generación de preactas | 4 | 21 |
| EP12 | Notificaciones | 2 | 11 |
| EP13 | Infraestructura y CI/CD | 5 | 28 |
| EP14 | Cierre académico | 1 | 8 |
| | **Total** | **53** | **264** |

---

## Historias por épica

### EP00 — Relevamiento y modelado del dominio

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US62 | SIG-6 | Relevamiento punta a punta de los cinco trámites | 3 | 1 |
| US63 | SIG-7 | Inventario de acciones, estados y actores abstractos | 3 | 1 |
| US71 | SIG-9 | Solicitar ejemplos de trámites a la Secretaría | 1 | 1 |
| US64 | SIG-13 | Mapa de roles y permisos de la Secretaría | 3 | 2 |

### EP01 — Identidad y acceso

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US06 | SIG-11 | Autenticación con proveedor abstraído | 8 | 2 |
| US07 | SIG-14 | Perfiles y autorización | 3 | 3 |

### EP02 — Configuración de trámites

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US08 | SIG-12 | Modelo de definición de trámite | 3 | 2 |
| US09 | SIG-15 | Configuración del trámite ancla: equivalencia entre asignaturas | 8 | 3 |
| US25 | SIG-24 | Configuración del segundo trámite: reconocimiento de créditos | 8 | 4 |
| US49 | SIG-48 | Configuración de cobertura: pase de universidad extranjera | 5 | 14 |
| US50 | SIG-51 | Configuración de cobertura: intercambios académicos | 5 | 15 |
| US51 | SIG-54 | Configuración de cobertura: pase dentro de la UBA | 5 | 16 |

### EP03 — Inicio de solicitud

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US13 | SIG-16 | Catálogo de trámites con requisitos | 2 | 3 |
| US14 | SIG-17 | Iniciar una solicitud | 5 | 3 |

### EP04 — Seguimiento del titular

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US22 | SIG-30 | Mis trámites en curso | 2 | 6 |
| US23 | SIG-31 | Detalle de mi trámite | 2 | 6 |
| US24 | SIG-35 | Historial de trámites cerrados | 3 | 11 |
| US34 | SIG-50 | El titular completa una subsanación | 3 | 15 |
| US38 | SIG-53 | Descarga de comprobante del trámite | 3 | 16 |

### EP05 — Bandeja y verificación de la Secretaría

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US17 | SIG-22 | Bandeja y listado de trámites | 3 | 4 |
| US19 | SIG-23 | Vista de detalle del trámite | 3 | 4 |
| US37 | SIG-25 | Filtros y orden en el listado de trámites | 1 | 4 |
| US66 | SIG-39 | Verificación manual de requisitos | 5 | 12 |
| US33 | SIG-49 | Pedido de subsanación al titular | 3 | 15 |
| US67 | SIG-52 | Plazos próximos a vencer | 3 | 15 |

### EP06 — Motor de pasos y estados

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US10 | SIG-20 | Avance de un trámite por confirmación de paso | 5 | 4 |
| US11 | SIG-21 | Historial de pasos de un trámite | 5 | 4 |
| US12 | SIG-44 | Devolución de un paso | 3 | 14 |

### EP07 — Gestión documental

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US15 | SIG-18 | Carga y almacenamiento de documentos | 8 | 3 |
| US16 | SIG-26 | Visualización y descarga de documentos del trámite | 5 | 5 |

### EP08 — Derivación a evaluación externa

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US21 | SIG-33 | Registrar retorno de la evaluación externa | 5 | 10 |
| US53 | SIG-56 | Recordatorio de derivación sin retorno | 3 | 16 |

### EP09 — Asistente RAG al titular

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US27 | SIG-27 | Ingesta y versionado de la base de conocimiento | 8 | 5 |
| US28 | SIG-28 | Recuperación de contexto relevante para la consulta | 8 | 5 |
| US69 | SIG-32 | Respuesta conversacional sobre el contexto recuperado | 13 | 6 |
| US31 | SIG-34 | Set de evaluación del asistente en el pipeline | 8 | 10 |
| US32 | SIG-40 | Pruebas de seguridad del asistente | 8 | 13 |
| US52 | SIG-55 | Base de conocimiento de los cinco trámites | 3 | 16 |

### EP10 — Validación asistida

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US40 | SIG-37 | Lectura de los documentos cargados | 5 | 12 |
| US41 | SIG-38 | Criterios verificables de cada requisito y cruce con los documentos | 8 | 12 |
| US43 | SIG-46 | Set de evaluación de la validación asistida y de las preactas | 8 | 14 |

### EP11 — Generación de preactas

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US44 | SIG-41 | Carga y versionado de plantillas de documentos | 3 | 13 |
| US45 | SIG-42 | Generación determinista de preacta desde plantilla | 8 | 13 |
| US46 | SIG-43 | Campos de texto libre asistidos por modelo | 5 | 13 |
| US47 | SIG-47 | Revisión, edición y aprobación de la preacta | 5 | 14 |

### EP12 — Notificaciones

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US36 | SIG-36 | Correo de pedido de acción | 8 | 11 |
| US35 | SIG-45 | Correo en cambio de estado | 3 | 14 |

### EP13 — Infraestructura y CI/CD

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US01 | SIG-5 | Repositorio e integración continua | 5 | 1 |
| US65 | SIG-8 | Definición y setup del stack tecnológico | 5 | 1 |
| US02 | SIG-10 | Configuración de herramientas agénticas de desarrollo | 5 | 2 |
| US04 | SIG-19 | Generador de datos sintéticos | 5 | 4 |
| US68 | SIG-29 | Despliegue automático | 8 | 5 |

### EP14 — Cierre académico

| Historia | Ticket | Título | SP | Sprint |
|---|---|---|---|---|
| US54 | SIG-57 | Pruebas de aceptación con la Secretaría | 8 | 17 |

---

## Tickets sin historia del backlog

Tickets que se crearon en Linear al partir una historia o para seguir una dependencia externa. No suman al total de 264 SP.

| Ticket | Título | Origen |
|---|---|---|
| SIG-58 | Respuesta de la Secretaría: validación del relevamiento y ejemplos de trámites | Seguimiento de US62 y US71 |
| SIG-59 | Proyecto base | Sale de US01 y US65 |
| SIG-62 | Entrega intermedia de la cátedra | Entrega obligatoria del 29/11, sprint 5 |
