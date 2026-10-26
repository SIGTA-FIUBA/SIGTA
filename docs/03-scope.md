# Alcance del proyecto

**Versión:** 0.2
**Regla de uso:** si algo no está en el alcance positivo, está fuera. El alcance negativo lista solo lo que alguien podría razonablemente asumir que está incluido.

---

## 1. Trámites cubiertos

El producto final cubre los cinco trámites priorizados por la Secretaría de Gestión Académica:

1. Equivalencia entre asignaturas
2. Reconocimiento de créditos
3. Intercambios académicos
4. Pase de universidad extranjera
5. Pase de universidad dentro de la UBA

### 1.1 Los cinco son una familia

Los cinco resuelven la misma pregunta de fondo: **reconocer estudios cursados fuera del plan actual del titular**. Eso implica, como hipótesis a validar en el relevamiento:

- Requisitos documentales muy parecidos entre sí: certificado analítico, programas de las asignaturas, plan de estudios de origen, y en el caso extranjero probablemente legalización o apostilla.
- Un mismo tipo de actor evaluador: la unidad académica que dictamina sobre contenidos.
- Un mismo tipo de salida: una resolución que establece qué se reconoce y qué no.

**Consecuencia de diseño.** El costo del trámite 3, 4 y 5 es bajo si la abstracción del primero está bien hecha, y alto si no lo está. Por eso el orden de los trámites no se elige por volumen, sino por lo que cada uno le prueba al modelo: el primero lo construye, el segundo confirma que generaliza, y el tercero, el más complejo, lo pone a prueba de verdad.

### 1.2 Ninguno de los cinco trámites se resuelve íntegramente dentro del área

Todos tienen, en mayor o menor medida, interacción con actores o sistemas externos. La variable que diferencia a un trámite de otro no es "interno vs. externo", sino:

- Cuántos actores externos intervienen.
- Cuántas idas y vueltas hay con ellos.
- Cuánta documentación externa hay que verificar antes de derivar.

El modelo de pasos tiene que tratar la derivación externa como caso normal, no como excepción.

### 1.3 Secuenciación propuesta

| Tanda | Trámite | Por qué en esa posición |
|---|---|---|
| Ancla | Equivalencia entre asignaturas | Es el núcleo conceptual de la familia: los otros cuatro terminan apoyándose en un dictamen de equivalencia. Modelarlo primero es modelar el dominio |
| Generalización | Reconocimiento de créditos | Confirma que el modelo sirve para un segundo trámite sin código nuevo. Es probablemente el más cercano a equivalencias, así que valida la generalización en el caso menos exigente |
| Cobertura | Pase de universidad extranjera, intercambios académicos, pase dentro de la UBA | Se incorporan por configuración, uno por sprint. El pase extranjero es el más complejo de la familia y el primero en ponerla a prueba de verdad |

> **A confirmar con la Secretaría.** El orden es una decisión del equipo, basada en la estructura de la familia. Si el secretario tiene otro criterio de urgencia, cambia el orden, no el alcance.

---

## 2. Alcance positivo

### 2.1 Plataforma — Secretaría

- Autenticación con roles diferenciados.
- Bandeja de tareas pendientes, agrupada por tipo de accionable.
- Vista de todos los trámites, filtrable por tipo y estado.
- Vista de detalle con paso a paso, estado actual, responsables y documentos.
- Registro de avance de pasos con trazabilidad de quién confirmó qué y cuándo.
- Registro de derivaciones a actores externos, con responsable, fecha de salida y fecha de retorno.
- Revisión y aprobación de preactas.

### 2.2 Plataforma — Titular

- Autenticación.
- Catálogo de trámites disponibles con sus requisitos.
- Inicio de un trámite con carga de la documentación requerida.
- Trámites en curso como pantalla principal.
- Detalle con estado, paso actual y responsable.
- Pedidos de subsanación: el titular ve qué le falta y lo completa sin abrir un trámite nuevo.
- Historial de trámites cerrados y descarga de comprobantes.

### 2.3 Documentos

- Carga de archivos por parte del titular y de la Secretaría.
- Almacenamiento asociado al trámite.
- Generación de preactas a partir de plantillas con campos variables.

### 2.4 Inteligencia artificial

Orden de prioridad decidido por el equipo. El orden de recorte es el inverso.

| Prioridad | Uso | Alcance | Límite explícito |
|---|---|---|---|
| 1 | Asistente RAG al titular | Responde sobre requisitos, documentación y cómo obtener cada documento, desde la base de conocimiento validada | Cita fuente. Deriva a una persona lo no cubierto |
| 2 | Validación asistida de documentación | Compara lo cargado contra los requisitos configurados y produce un checklist con hallazgos | No aprueba ni rechaza. Decide una persona |
| 3 | Generación de preactas | Completa la plantilla con datos del trámite; el LLM solo en campos de texto libre | Siempre borrador, sujeto a revisión y aprobación |

### 2.5 Configuración

- Definición de tipos de trámite (requisitos, documentos, pasos, estados, roles) como archivos versionados en el repositorio (ADR-001).
- Incorporar un trámite nuevo consiste en agregar configuración, sin tocar el código.
- Requisitos y pasos compartidos, reutilizables entre trámites de la familia.

### 2.6 Calidad y operación

- Pipeline de CI/CD con GitHub Actions que ejecuta escenarios de aceptación, tests unitarios y de integración.
- Entorno desplegado y accesible para las instancias de validación con la Secretaría.
- Datos de prueba sintéticos.
- Notificación por correo en cambios de estado y pedidos de acción.

---

## 3. Alcance negativo

| Fuera del alcance | Por qué |
|---|---|
| Editor dinámico de flujos para la Secretaría **\*** | ADR-001. **Duda pendiente:** la reunión del 16/09 lo había acordado y la decisión se revirtió sin consultarlo con la Secretaría |
| Integración con SIU-Guaraní | El anteproyecto declara que el diseño no depende de sus interfaces |
| Integración con el sistema de expediente electrónico | Se resuelve con puente humano (ADR-004) |
| Acceso de actores externos a la plataforma **\*** | La derivación se registra, no se ejecuta dentro del sistema. **Duda pendiente:** no se consultó a los evaluadores si aceptarían trabajar dentro de la plataforma |
| Firma digital con validez legal | El paso "firma" se modela como aprobación registrada |
| Aprobación o rechazo automático de documentación | Contradice el principio de aprobación humana |
| Interpretación o clasificación del resultado de un trámite | ADR-009: la plataforma comunica el dictamen, no lo interpreta |
| Dictamen automático de equivalencias | El sistema no decide qué asignatura equivale a cuál. Esa es la decisión académica y es humana |
| Migración de datos históricos reales | Se trabaja con datos sintéticos (ADR-005) |
| Trámites fuera de los cinco priorizados | Se incorporan por configuración, después del proyecto |
| Aplicación móvil nativa | La web responsive cubre el caso |
| Gestión de pagos o aranceles | No aparece en los circuitos relevados |
| Notificaciones configurables, digests, canales alternativos | Solo correo, en cambios de estado y pedidos de acción |
| Reportes e indicadores de gestión para la Secretaría | Valioso, pero no es el problema que se vino a resolver |
| Soporte multi-facultad | Fuera del contexto |

**\*** Marcadas como dudas pendientes: no son exclusiones cerradas, son decisiones tomadas sin consultar al stakeholder. Se resuelven en la reunión de relevamiento del sprint 1 y se asientan como ADR.

> **Aclaración sobre el dictamen de equivalencias.** Es la exclusión más importante de la lista y conviene decirla en voz alta con la Secretaría. La plataforma ordena, acompaña y documenta la decisión académica. No la toma.

---

## 4. Supuestos

| ID | Supuesto | Si se cae |
|---|---|---|
| S-01 | Los cinco trámites comparten estructura de requisitos y un mismo tipo de evaluador académico | El costo marginal por trámite deja de ser bajo y hay que recortar la cobertura a tres |
| S-02 | La Secretaría entrega las plantillas de los documentos a generar | La generación de preactas se reduce a un formato propio, con menor valor |
| S-03 | Existe un proveedor de identidad institucional utilizable | Confirmado el 9/10: cuenta de Google de la UBA (ADR-010) |
| S-04 | La Secretaría dispone de una instancia de validación por fase | Se acumula deuda de validación, con riesgo al cierre |
| S-05 | Hay infraestructura de despliegue disponible o financiable | Despliegue de costo cero, con impacto en la demo |
| S-06 | Los pasos se reducen a un conjunto chico de acciones abstractas | La configuración se vuelve específica por trámite y se pierde el bajo costo marginal |
| S-07 | Los perfiles operativo y de aprobación del Secretario corresponden a personas distintas | Se simplifica a un único perfil de Secretaría con todos los permisos |
| S-08 | Los actores externos aceptan seguir trabajando fuera de la plataforma | Habría que negociar acceso externo, que hoy está fuera del alcance |

---

## 5. Criterio de cambio de alcance

Todo pedido fuera del alcance positivo entra como ticket con etiqueta `fuera-de-alcance`, se evalúa en la reunión quincenal y, si se acepta, se asienta como ADR indicando **qué sale** a cambio.
