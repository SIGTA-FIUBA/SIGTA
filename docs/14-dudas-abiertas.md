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
| DU-05 | ¿Quiénes de la DA usan el sistema? ¿La Directora General es otra área? | Todo el personal de la DA, de Intercambios y la Directora General. Cada persona tiene una membresía en su área, con los roles y permisos de esa área que le asigna el administrador (ADR-020). La Directora General es parte del área de la DA, con un rol propio para los pasos que exigen su permiso | Supuesto, en línea con la respuesta del 4/11 |
| DU-06 | ¿Quién sube los documentos al expediente electrónico? | La DA, fuera de SIGTA. SIGTA guarda solo el número de expediente | Supuesto (ADR-004) |
| DU-34 | Rectorado, el CBC y el Despacho del Consejo Directivo no usan SIGTA, pero tampoco evalúan. Con ADR-020 no hay áreas `external`: un paso que resuelve alguien de afuera es del área que lo registra en SIGTA. ¿Sus pasos los confirma la DA, como el archivo, o cada dependencia es un área con su propio permiso para registrarlos? ¿Cómo se muestra al titular que el trámite espera a otra dependencia? | Sus pasos son de la DA, que confirma lo que pasó afuera, como el archivo. Mesa de Entrada ya no es un área | Se espera, con el equipo, junto con SIG-71 |

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
| DU-39 | ¿El entorno de FIUBA registra quién opera la consola, es decir quién corre `backend migrate` y `backend grant-admin`? SIGTA no guarda quién dio el primer administrador (ADR-027) | Queda a cargo del entorno donde se despliegue | Se espera, con la Subsecretaría de TICs |

## 4. De permisos y configuración

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-40 | ¿Con qué áreas se configura el trámite ancla? SIG-83 nombra seis: Dirección Académica, Comisión Curricular, Despacho del Consejo Directivo, Consejo Directivo, Dirección General y Mesa de Entrada. No coincide con lo relevado: la DA es la Dirección de Alumnos; según DU-05 la Directora General es parte del área de la DA, con un rol propio; según DU-34 los pasos del Despacho del Consejo Directivo los confirma la DA y Mesa de Entrada ya no es un área; y en los actores del relevamiento ([docs/13](13-relevamiento-de-tramites.md)) la CC, el Despacho y el CD no usan SIGTA | Ninguna todavía. Son configuración: las carga un administrador desde la pantalla de SIG-76, y si cambian se editan ahí, sin código (ADR-027) | A decidir con el equipo antes de cargar las áreas, y a confirmar con la Secretaría |
| DU-32 | ¿Quién puede proponer y quién aprobar una versión de un tipo de trámite? (D-08) | En parte: son dos permisos de plataforma de la Secretaría, uno para proponer y otro para aprobar (ADR-020). Falta saber a quién se le otorgan y si el equipo tiene una membresía en la Secretaría | Se espera, con la Secretaría, antes de la interfaz de edición |
| DU-41 | ¿Hasta cuándo vale una membresía con fecha de fin? | Hasta el final de ese día en Buenos Aires: el "Hasta" de la pantalla de administración se guarda como las 23:59:59 de ese día, hora de Argentina (SIG-76) | Supuesto; a confirmar con la Secretaría |
| DU-42 | ¿Qué pasa con el alcance y la fecha de fin de una membresía cuando se le cambian los roles o permisos? | Se conservan: el cambio cierra la membresía y abre otra con el mismo alcance y el mismo fin, con los roles y permisos nuevos, para que quede registrado quién cambió qué y cuándo (SIG-76) | Supuesto; a confirmar con la Secretaría |
| DU-43 | ¿Un administrador puede quitarse la administración a sí mismo? | Sí, si queda otro administrador vigente; nunca el último, para que SIGTA no quede sin administradores (ADR-027) | Supuesto del equipo; a confirmar con la Secretaría |
| DU-44 | ¿Se puede cambiar el nombre de un área? | No: un área, un permiso o un rol no se renombran. Si cambia el nombre, se da de baja y se crea otra con otro identificador, porque el de una baja no se vuelve a usar. Eso exige terminar antes las membresías que la usan y volver a publicar los tipos de trámite que la nombran, aunque el nombre solo se muestra (SIG-76) | Supuesto del equipo; a confirmar con la Secretaría si prefiere poder corregir el nombre |
| DU-45 | ¿Cómo se elige el identificador de un área, un permiso o un rol? | Lo tipea a mano el administrador al crearlo: minúsculas, dígitos y guiones bajos, empezando con una letra. No se puede cambiar después ni volver a usar tras una baja. Los pasos de un tipo de trámite nombran los permisos por ese identificador (ADR-020) | Supuesto del equipo; a revisar si se propone a partir del nombre |
| DU-46 | ¿A quiénes puede ver un administrador? | A cualquier persona que haya entrado a SIGTA, por nombre y correo: la búsqueda para otorgar una membresía o la administración devuelve hasta 10 por pedido, sin otro límite ni registro de lo buscado (SIG-76, ADR-027) | Supuesto del equipo; a confirmar con la Secretaría y, si hace falta, con quien vele por los datos personales |

## 5. Del inicio de un trámite

Salen de "Iniciar una solicitud" (SIG-17).

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-35 | ¿Cómo se carga la carrera (`degree`) y las asignaturas (`subject_list`)? ¿Hay un listado para elegir? | Texto libre: la carrera, un texto; las asignaturas, una lista de textos. No se controla contra ningún plan de estudios | Supuesto; a confirmar con la Secretaría |
| DU-36 | ¿Cuánto puede ocupar un dato de texto? | 200 caracteres para un texto corto y 5000 para uno largo; hasta 100 asignaturas | Decidido por el equipo |
| DU-37 | ¿Un número admite decimales y negativos, como un promedio o un ajuste? | Sí: un decimal con signo. La web acepta coma o punto y envía punto. Un tipo de trámite todavía no puede pedir solo enteros ni un rango | Supuesto; a confirmar con la Secretaría |

## 6. Del comprobante del trámite

Salen de "Descarga de comprobante del trámite" (SIG-53).

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-47 | ¿Qué valor le da la Secretaría al comprobante? No lleva firma digital ni un código para verificarlo, y se puede editar (no hay firma digital con validez legal) | Deja constancia del estado y del historial del trámite en SIGTA cuando se emitió, y aclara que no es una resolución. Lo que vale es lo que está en SIGTA | Se espera, con la Secretaría |
| DU-48 | ¿Cómo sale en el comprobante un nombre con caracteres fuera del español, como los de un titular extranjero? | El PDF usa las fuentes Helvetica que trae todo lector, sin embeber ninguna, en WinAnsiEncoding: cubre el español y las lenguas de Europa occidental, y lo demás sale como `?` ("Zoë Chen 陈伟" sale "Zoë Chen ??"). En la página del trámite el nombre sale completo | Decidido por el equipo; si la Secretaría lo necesita, se embebe una fuente con esos caracteres |

## 7. Del seguimiento del titular

Salen del historial de trámites cerrados (SIG-35) y del detalle del trámite (SIG-31).

| ID | Duda | Por ahora | Cómo |
| --- | --- | --- | --- |
| DU-49 | Lo que el personal adjunta como resolución, ¿es siempre el documento firmado? Podría subir un borrador sin firmar, y el titular lo ve en cuanto el trámite se cierra. Si después lo reemplaza, el historial del detalle sigue mostrando el archivo anterior y deja descargarlo | SIGTA no lo distingue: el titular ve la última versión de cada documento resolutivo y, en el historial, las que fue reemplazando. Que lo adjuntado sea el documento final queda a cargo de quien lo sube | Se espera, con la Secretaría: si el titular debe ver solo la versión final y si las reemplazadas tienen que quedar a su alcance |

---

## Cerradas

| ID | Duda | Respuesta | Fecha |
| --- | --- | --- | --- |
| — | ¿Hace falta un editor de flujos? | Sí: los tipos de trámite se editan desde la interfaz (ADR-013) | Antes del 8/11 |
| DU-31 | ¿Los perfiles operativo y de aprobación son personas distintas? (D-02) | La pregunta dejó de aplicar: no hay perfiles fijos. Cada área tiene sus permisos y roles, que el administrador asigna en las membresías, y una persona puede tener cualquier combinación (ADR-020, que reemplaza a ADR-017) | 8/11 |
| DU-33 | ¿Se mantiene la validación asistida con IA, si la completitud se puede revisar con reglas en código? | Sí, se mantiene (ADR-002). Las reglas en código revisan que esté todo; la IA revisa el contenido, propone hallazgos y decide una persona | 8/11 |
