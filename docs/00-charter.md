# Project Charter

**Proyecto:** Plataforma centralizada de gestión de trámites de la Secretaría de Gestión Académica (FIUBA), asistida por IA
**Versión:** 0.1 — borrador para revisión del equipo
**Estado:** en validación

---

## 1. Problema

La gestión de trámites de la Secretaría de Gestión Académica se realiza de forma descentralizada: los trámites se inician en Formularios de Google o por correo, los datos quedan dispersos en planillas y el seguimiento se hace por cadenas de mail individuales. No existe una base única que permita saber en qué estado está un trámite.

Consecuencias directas:

- El titular no tiene visibilidad del avance ni claridad sobre los requisitos, y canaliza esa incertidumbre como consultas a la Secretaría.
- La Secretaría recibe solicitudes desde fuentes distintas y centraliza la documentación de forma manual.
- Los tiempos de resolución se estiran sin que nadie pueda explicar dónde se detuvo el trámite.

## 2. Objetivo del proyecto

Reducir la fricción de los procesos de gestión y seguimiento de trámites, para el personal de la Secretaría y para los titulares, mediante una plataforma única donde el trámite se inicia, avanza y se consulta.

## 3. Producto

Una plataforma web con dos caras:

- **Secretaría:** bandeja de tareas pendientes, vista por tipo de trámite, detalle con paso a paso, estado, responsables y documentos, y registro de avance con trazabilidad.
- **Titular:** inicio de trámites, seguimiento del estado, historial y asistente que responde sobre requisitos y documentación.

Sobre ambas, tres usos de inteligencia artificial acotados (ver `04-decisiones-adr.md`, ADR-002).

## 4. Fuera de negociación

1. Ninguna acción sobre un trámite se toma sin aprobación humana.
2. Las respuestas del asistente se limitan a la base de conocimiento validada, citan su fuente y derivan a una persona cuando no hay cobertura.
3. No se procesan datos personales reales: todo dato de prueba es sintético o anonimizado.
4. La plataforma no reemplaza ni modifica los sistemas institucionales existentes.

## 5. Métricas de éxito

### 5.1 Del producto (requieren línea de base)

| Métrica | Cómo se obtiene | Línea de base |
|---|---|---|
| Tiempo promedio de resolución de un trámite | Relevamiento sobre el proceso actual vs. medición en la plataforma | A relevar con la Secretaría |
| Horas mensuales de trabajo manual evitadas | Estimación por paso automatizado o centralizado | A relevar con la Secretaría |
| Consultas recibidas por canales informales (mail) | Conteo declarado por la Secretaría | A relevar con la Secretaría |

> **Pendiente bloqueante:** sin línea de base, el impacto declarado en el anteproyecto no es demostrable en la defensa. Conviene pedir estos tres números en la primera reunión de relevamiento, aunque sean estimaciones a ojo del personal.

### 5.2 Del MVP (medibles sin la Secretaría)

| Métrica | Objetivo propuesto |
|---|---|
| Trámites con circuito completo end-to-end sobre datos sintéticos | 5 de 5 |
| Trámites de cobertura incorporados solo con configuración, sin código nuevo | 3 de 3 |
| Consultas del set de evaluación respondidas por el asistente con fuente citada | ≥ 80% |
| Consultas fuera de la base de conocimiento que derivan correctamente | 100% |
| Preactas generadas que el revisor aprueba sin editar campos estructurados | ≥ 90% |
| Escenarios de aceptación en verde en el pipeline | 100% al cierre de cada sprint |

## 6. Alcance en una línea

Los cinco trámites priorizados por la Secretaría (equivalencia entre asignaturas, reconocimiento de créditos, intercambios académicos, pase de universidad extranjera y pase de universidad dentro de la UBA), tratados como una familia, con sus definiciones mantenidas como configuración versionada e incorporación incremental. Detalle completo en `03-scope.md`.

## 7. Restricciones

| Restricción | Valor |
|---|---|
| Equipo | 4 integrantes, en rol de desarrolladores |
| Dedicación | 12 hs semanales por integrante (mínimo) |
| Esfuerzo total comprometido | 384 hs por integrante |
| Duración derivada | 32 semanas = **16 sprints de 2 semanas** |
| Capacidad teórica por sprint | 96 hs de equipo |
| Sprint 0 | 7 al 20 de septiembre de 2026 (definición y artefactos de gestión) |
| Sprint 1 | Arranca el 21 de septiembre de 2026 |
| Cierre aproximado | Julio 2027 (incluye preparación de defensa) |
| Datos | Sintéticos o anonimizados |
| Integraciones | Sin dependencia de SIU-Guaraní ni del expediente electrónico |
| Infraestructura | Opciones de la facultad o nube de bajo costo, a definir en la primera fase |

> La capacidad teórica de 96 hs por sprint es un techo, no un plan. La capacidad real se calibra con la velocidad de los dos primeros sprints (ver artefactos de proceso).

## 8. Organización

- **Equipo de proyecto:** diseño, implementación, testing, documentación y calidad del producto.
- **Directores:** supervisión, asesoría especializada en sistemas con IA, validación de entregables. Reunión quincenal.
- **Secretaría de Gestión Académica:** fuente del relevamiento y validación funcional. Contactos: Lucas Adrián Macías y Lucas Martín Diaz.

## 9. Fases

1. Relevamiento y definición
2. Diseño e infraestructura base
3. Desarrollo incremental del MVP
4. Validación y ajuste
5. Cierre y documentación

La asignación de sprints a cada fase se define en el roadmap una vez fijada la fecha de inicio del sprint 1.

## 10. Criterio de "proyecto terminado"

- Los cinco trámites se recorren completos sobre datos sintéticos, desde el inicio por parte del titular hasta el cierre registrado por la Secretaría.
- Los tres trámites de cobertura se incorporaron agregando configuración, sin código nuevo.
- Los tres usos de IA operan con aprobación humana en todos los puntos de decisión.
- Los escenarios de aceptación pasan automáticamente en el pipeline.
- La Secretaría validó el comportamiento en al menos una instancia formal por trámite.
- Documentación técnica y funcional entregada, informe final cerrado y defensa preparada.
