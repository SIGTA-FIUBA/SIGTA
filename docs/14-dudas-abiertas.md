# Dudas abiertas

Todo lo que el proyecto todavía no sabe y cómo se avanza mientras tanto. Cada duda tiene una forma de resolverla por ahora (un supuesto, una decisión del equipo o "se espera") y con quién o cuándo se cierra. Cuando una duda se cierra, se pasa a la sección "Cerradas" con la fecha y la respuesta, y se actualiza el documento que la usaba.

Las decisiones pendientes de arquitectura y alcance siguen en [ADR](04-decisiones-adr.md) y los supuestos de alcance en [alcance](03-scope.md); acá se listan para tener todo en un lugar, sin repetir su detalle.

**Cómo se resuelve por ahora:**

- **Supuesto:** el equipo toma una respuesta razonable y sigue. Si se cae, se corrige la configuración.
- **Decidido por el equipo:** no depende de nadie más, pero conviene que la Secretaría lo sepa.
- **Se espera:** no bloquea hoy; se resuelve antes de la fecha indicada.

---

## 1. Del relevamiento de los trámites

Salen de "Flujos de los trámites" ([relevamiento](13-relevamiento-de-tramites.md)). Se validan en la reunión con las áreas que ofreció la Secretaría el 4/11.

### Generales

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-01 | ¿Cómo se entera el estudiante de que terminó el trámite? No hay paso de notificación | SIGTA le avisa por correo al cerrarse. No hace falta un paso en el circuito | Supuesto |
| DU-02 | ¿El Secretario Académico y el de Gestión Académica son la misma persona? | Son cargos distintos, pero no importa para SIGTA: firman afuera y la DA confirma la firma | Supuesto |
| DU-03 | ¿Por qué algunos trámites pasan por el CD y otros no? | Lo define cada tipo de trámite en su configuración | Supuesto |
| DU-04 | En los pases, ¿FIUBA ve algo antes de que llegue de Rectorado? | El estudiante inicia en SIGTA al presentar en Rectorado; el trámite espera en un paso externo hasta que la DA lo recibe | Supuesto |
| DU-05 | ¿Quiénes de la DA usan el sistema? ¿La Directora General es otra área? | Todo el personal de la DA, de Intercambios y la Directora General, con los roles que asigna el administrador. La Directora General tiene un rol propio | Supuesto, en línea con la respuesta del 4/11 |
| DU-06 | ¿Quién sube los documentos al expediente electrónico? | La DA, fuera de SIGTA. SIGTA guarda solo el número de expediente | Supuesto (ADR-004) |
| DU-34 | Rectorado, el CBC y el Despacho del Consejo Directivo no usan SIGTA, pero tampoco evalúan. Si sus pasos son de un área `external`, el trámite se muestra "En evaluación externa" y cuenta días sin retorno, como una derivación. ¿Se separa "evaluador externo" de "otra dependencia", o sus pasos los confirma la DA, como el archivo? | Se modelan como `external`, como en el relevamiento. Mesa de Entrada ya no es un área: el archivo lo confirma la DA | Se espera, con el equipo, junto con SIG-71 |

### Equivalencia entre asignaturas

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-07 | ¿Qué documentación se pide además de la nota? ¿Se sigue pidiendo foto del DNI? | Los datos de la nota al Decano y el PDF del historial académico del SIU Guaraní. Sin domicilio, teléfono, DNI ni firma | Decidido por el equipo; a confirmar con la Secretaría |
| DU-08 | ¿Quién revisa que esté completo antes de caratular? | SIGTA, con reglas en código: no deja enviar si falta algo. No hay paso de revisión; si después la DA ve un problema, pide una subsanación. El contenido de los documentos lo revisa después la validación asistida, que propone hallazgos y no decide | Decidido por el equipo |
| DU-09 | Si son materias de varias carreras, ¿va a una CC o a varias? | A una, la de la carrera de destino. Para dos carreras, dos trámites | Decidido por el equipo |
| DU-10 | ¿Cuáles son las equivalencias automáticas y por dónde van? | Fuera de SIGTA por ahora | Decidido por el equipo |
| DU-11 | ¿El SIU Guaraní deja descargar el historial académico como PDF con texto? | Se asume que sí. Si no, el listado se carga como dato del formulario | Supuesto; a confirmar con la Secretaría |

### Reconocimiento de créditos

Se resuelven antes de configurar el trámite (SIG-24, sprint 4).

| ID | Duda | Por ahora |
| --- | --- | --- |
| DU-12 | ¿Con qué se justifica el motivo? | Se espera |
| DU-13 | En el micrositio hay varios parecidos (créditos, créditos extracurriculares, cursado fuera del plan, créditos por intercambio). ¿Cuál es este? ¿Siguen el mismo circuito? | Se espera |

### Intercambios académicos

Se resuelven antes de configurar el trámite (sprint 15).

| ID | Duda | Por ahora |
| --- | --- | --- |
| DU-14 | ¿Intercambios Académicos depende de la Secretaría de Gestión Académica? | Se espera |
| DU-15 | La firma figura dos veces en la planilla. ¿Es un error o son dos firmas? | Se espera |
| DU-16 | ¿"Rectificada" quiere decir ratificada? ¿Se espera al CD para registrar en el SIU? | Se espera |
| DU-17 | ¿Se pide antes de viajar o a la vuelta? | Se espera |
| DU-18 | ¿Incluye las equivalencias o van aparte? | Se espera |

### Pase de universidad extranjera

Se resuelven antes de configurar el trámite (sprint 14).

| ID | Duda | Por ahora |
| --- | --- | --- |
| DU-19 | ¿Evalúa el Departamento o la CC? | Se espera |
| DU-20 | ¿Las CC van en serie? ¿Quién decide cuáles? | Se espera |
| DU-21 | Si una CC rechaza una parte, ¿se resuelve con lo aprobado? | Se espera |
| DU-22 | ¿Quién valida la documentación de origen, Rectorado o FIUBA? | Se espera |

### Pase de universidad dentro de la UBA

Se resuelven antes de configurar el trámite (sprint 16).

| ID | Duda | Por ahora |
| --- | --- | --- |
| DU-23 | ¿Es el mismo trámite que en el micrositio va por TAD UBA? | Se espera |
| DU-24 | ¿Pasa siempre por el CBC? | Se espera |

## 2. Del modelo de datos

Salen de la persistencia de los tipos de trámite (SIG-65).

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-25 | Si el historial de acciones solo crece, ¿cómo se suprimen los datos personales de un documento cuando alguien lo pide? | Se espera; depende de la retención de documentos | Se espera, con la Secretaría y el área legal |
| DU-26 | ¿De qué maneras puede terminar un trámite? Se suponen tres: llega al final y se archiva, el titular desiste o vence un plazo. No es el resultado (si se reconoció o no), que no se modela | Las tres de arriba. El cierre se guarda como una entrada más del historial ("se cerró porque el titular desistió"), con quién y cuándo, y no como un campo del trámite. Publicar una versión nueva o dar de baja el tipo de trámite no cierra trámites en curso: siguen con su versión hasta el final (ADR-012) | Supuesto; a confirmar con la Secretaría |
| DU-27 | ¿En qué pasos puede desistir el titular? | Se espera | Se espera, con la Secretaría |
| DU-28 | ¿Cuánto tiempo se conservan los documentos y hay que archivar los de trámites cerrados? | Se espera | Se espera, con la Secretaría |

## 3. De infraestructura

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-29 | ¿Dónde se despliega y quién lo financia? (D-04) | Se prepara para la infraestructura de FIUBA, según los requisitos de la Subsecretaría de TICs | Se espera, con la Subsecretaría de TICs |
| DU-30 | ¿El PostgreSQL de la Subsecretaría de TICs soporta pgvector? (R-17) | Se sigue con pgvector, detrás de un puerto | Se espera, antes de construir el asistente |

## 4. De permisos y configuración

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-32 | ¿Quién puede proponer y quién aprobar una versión de un tipo de trámite? (D-08) | Se espera; con roles configurables se puede resolver como dos permisos | Se espera, con la Secretaría, antes de la interfaz de edición |

---

## Cerradas

| ID | Duda | Respuesta | Fecha |
| --- | --- | --- | --- |
| — | ¿Hace falta un editor de flujos? | Sí: los tipos de trámite se editan desde la interfaz (ADR-013) | Antes del 8/11 |
| DU-31 | ¿Los perfiles operativo y de aprobación son personas distintas? (D-02) | No es así: la Secretaría pidió un administrador que asigna roles configurables (ADR-017, aceptado) | 8/11 |
| DU-33 | ¿Se mantiene la validación asistida con IA, si la completitud se puede revisar con reglas en código? | Sí, se mantiene (ADR-002). Las reglas en código revisan que esté todo; la IA revisa el contenido, propone hallazgos y decide una persona | 8/11 |
