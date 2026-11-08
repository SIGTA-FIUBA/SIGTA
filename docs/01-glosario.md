# Glosario de dominio

Vocabulario único del proyecto. Se usa igual en el código, en los tickets, en los escenarios Gherkin y en las conversaciones con la Secretaría. Si una palabra no está acá, no se usa en un ticket.

Las entradas marcadas con **(V)** deben validarse con la Secretaría antes de codificarse: son interpretaciones nuestras de la planilla y de la reunión, no definiciones oficiales.

---

## Dominio del trámite

**Tipo de trámite**
Definición reutilizable de un procedimiento administrativo: su descripción en lenguaje llano, sus plazos y condiciones de admisibilidad (como texto que el titular lee), los datos y documentos que pide, la secuencia de pasos, y el área de cada paso con su perfil, salvo en los pasos del titular. Se va a editar desde la interfaz de SIGTA (ADR-013), y cada publicación genera una versión nueva (ADR-012).

**Versión de un tipo de trámite**
Contenido publicado de un tipo de trámite, con su número, su hash, quién la propuso, quién la aprobó y cuándo. No se modifica: un trámite sigue de punta a punta la versión con la que empezó.

**Borrador de un tipo de trámite**
Definición en edición que todavía no se publicó. No confundir con la preacta ni con los borradores de configuración que prepara un agente para que el equipo los revise.

**Trámite (o solicitud)**
Instancia concreta de un tipo de trámite, iniciada por un titular. Tiene un titular, un estado, un historial de pasos y un conjunto de documentos. Es la unidad de seguimiento del sistema.

**Circuito** **(V)**
Forma general que toma la secuencia de pasos de un tipo de trámite. **Ninguno de los cinco trámites se resuelve íntegramente dentro del área**: todos salen hacia al menos un evaluador externo y vuelven. Lo que diferencia un circuito de otro es cuántos evaluadores intervienen y cuántas idas y vueltas hay.

**Paso**
Unidad mínima de avance dentro de un trámite. Tiene un área responsable, un perfil salvo en los pasos del titular, una acción y lo que se registró al cumplirlo. Un paso no se considera cumplido hasta que alguien lo confirma en la plataforma.

**Acción de un paso**
Cada paso tiene una sola de cuatro acciones, derivadas de la planilla de circuitos:
- *Cargar dato* (`enter_data`): alguien completa uno o más datos.
- *Cargar documento* (`upload_document`): alguien adjunta uno o más archivos.
- *Confirmar hecho externo* (`confirm_external_fact`): alguien declara que algo pasó fuera de la plataforma, como una carátula, una firma o una carga en SIU Guaraní. Queda registrada la declaración, no el hecho. **La plataforma no implementa firma digital con validez legal.**
- *Decidir* (`decide`): alguien elige una salida, con los datos y documentos que la fundamentan. Una aprobación es un paso con perfil de aprobación: lo que la distingue es el perfil, no la acción. Es parte del proceso, no el resultado del trámite (ADR-009).

**Salida** (`outcome`)
Cada una de las opciones de una decisión, con el paso al que lleva. Es parte del proceso, no un resultado del trámite: no confundir con "Resultado del trámite".

**Área**
Parte de la organización a la que pertenece el personal, como la Dirección de Alumnos, Mesa de Entradas o una Comisión Curricular. Las áreas existen solo dentro del rol de personal: el titular no tiene área. No son fijas: se crean como datos de SIGTA, son comunes a todos los tipos de trámite y su personal se define con membresías (ADR-017). Cada paso dice su rol: los del titular (`applicant`) no tienen área, y los del personal (`staff`) son de un área. Las versiones publicadas antes del formato 3 modelaban al titular como un área de tipo `applicant`, y se convierten al leerlas.

**Pase**
Avance de un trámite de un paso a otro con un área distinta. No es un paso propio. No confundir con los trámites de pase de universidad.

**Estado del trámite**
Situación en la que se encuentra el trámite en un momento dado, derivada del tipo de área del paso abierto: esperando al titular, en el área, derivado o cerrado. La configuración del tipo de trámite no declara estados. A confirmar con la Secretaría: si un trámite cerrado guarda su motivo de cierre (resuelto y, según se defina, desistido o vencido). El motivo de cierre no sería el resultado del trámite.

**Resultado del trámite**
Lo que dictamina el evaluador externo: reconocimiento total, parcial o rechazo. **La plataforma no lo modela ni lo interpreta** (ADR-009): lo recibe como documento y lo comunica al titular. Para la plataforma, los tres casos son el mismo evento.

**Evaluador externo** **(V)**
Actor ajeno a la Secretaría que dictamina sobre el contenido académico del trámite: comisión curricular, departamento u otra unidad académica, según el trámite. No usa la plataforma (duda pendiente, ver alcance).

**Derivación**
Salida del trámite hacia un evaluador externo. En el modelo es pasar a un paso de otra área, que registra el destino, la fecha y, si corresponde, el número de expediente. Si el evaluador no usa SIGTA, el paso es del área que registra su dictamen.

**Retorno**
Vuelta del trámite desde el evaluador externo, con su dictamen adjunto. La plataforma registra el retorno y comunica el dictamen sin interpretarlo (ADR-009).

**Devolución**
Vuelta del trámite a un paso anterior, decidida por la Secretaría e indicando el motivo.

**Subsanación**
Pedido puntual al titular para que complete o corrija algo, sin cerrar ni reiniciar el trámite. También canaliza los pedidos de ampliación que hace un evaluador externo.

**Puente humano**
Persona de la Secretaría que ejecuta una acción en un sistema externo y la registra en la plataforma. Es el mecanismo elegido para no depender de integraciones (ADR-004).

**Rol**
Cada uno de los dos tipos de persona usuaria de la plataforma, que se excluyen entre sí: **titular del trámite** (`applicant`) y **personal** (`staff`, el Secretario del mapa de usuario). No se guardan: es personal quien tiene al menos una membresía vigente, y titular cualquier otra persona que ingresa (ADR-017). Solo el titular inicia trámites.

**Perfil**
Rango del personal dentro de un área, que da su membresía. El **perfil operativo** recibe, verifica, deriva y registra; el **perfil de aprobación** además aprueba y firma, y cubre los pasos operativos. En el mapa de usuario son un solo rol; en permisos se distinguen.

**Membresía**
Pertenencia de una persona a un área, con un perfil. Puede tener un **alcance**: una lista de tipos de trámite que restringe dónde actúa dentro del área, que referencia el tipo y no una versión. Tiene vigencia desde y, opcionalmente, hasta; una baja cierra la vigencia sin borrarla. La otorga el administrador (ADR-017).

**Administrador**
Marca sobre una persona que le permite otorgar y dar de baja membresías. No es un rol ni da acceso a los trámites: para actuar sobre ellos necesita una membresía.

**Titular del trámite**
Persona que inicia y a cuyo nombre corre el trámite. Se usa esta expresión y no "estudiante", porque no todos los trámites los inicia alguien con condición de estudiante activo.

## Documentos

**Requisito**
Condición que el trámite exige para poder avanzar. Puede ser un documento, un dato o una condición académica. En el modelo, los datos (`fields`) y documentos (`documents`) se declaran en cada tipo de trámite. Las condiciones académicas se guardan solo como texto (`admissibility`) y nada las aplica: a confirmar cómo se declaran para que se puedan verificar, porque son la base contra la que trabaja la validación asistida.

**Preacta**
Borrador de un documento formal generado por la plataforma a partir de una plantilla, con campos completados desde los datos del trámite. **Siempre es un borrador**: no tiene valor hasta que una persona lo revisa y lo aprueba.

**Acta / resolución**
Documento formal producido al cierre de un paso, una vez que una persona aprobó la preacta correspondiente.

**Plantilla de documento**
Archivo base provisto por la Secretaría, con texto fijo y campos variables, a partir del cual se genera una preacta. **Pendiente de entrega por parte de la Secretaría.**

**Carátula / expediente electrónico** **(V)**
Artefacto y sistema institucional externos, mencionados en la planilla de circuitos. Quedan fuera del alcance: la plataforma solo registra que el paso se realizó allí.

## Componentes del sistema

**Base de conocimiento**
Corpus validado que alimenta al asistente: requisitos y condiciones de trámites, preguntas frecuentes y resoluciones históricas. No contiene expedientes ni datos de personas.

**RAG (Generación Aumentada por Recuperación)**
Técnica por la cual el asistente recupera fragmentos relevantes de la base de conocimiento y responde a partir de ese contexto, citando la fuente.

**Asistente del titular**
Componente conversacional que responde consultas sobre requisitos y documentación a partir del contexto recuperado, cita su fuente y deriva a una persona lo que no está cubierto.

**Validación asistida**
Comparación automática de la documentación cargada contra los requisitos configurados del trámite, que produce un checklist con hallazgos para que una persona decida. No aprueba ni rechaza por sí sola.

**Configuración versionada**
Término reemplazado. Con ADR-001, los tipos de trámite iban a ser archivos en el repositorio editados por pull request; ADR-013 lo reemplazó por la edición desde la interfaz, con versiones en la base. Ver "Versión de un tipo de trámite".

**Ingreso institucional**
Inicio de sesión con la cuenta de Google de la UBA. Es la única forma de ingresar a SIGTA, confirmada con la Subsecretaría de TICs el 9/10 (ADR-010). En la reunión del 16/09 se había considerado el FUA, el ingreso unificado de la facultad.

## Proceso de trabajo

**Sprint**
Iteración de dos semanas que cierra con un incremento funcional.

**Bloque**
Agrupación de sprints consecutivos que cierra una parte del producto: B1 esqueleto y asistente mínimo (sprints 1 a 6), B2 cierre del asistente (10 y 11), B3 validación asistida, preactas y cobertura (12 a 16), B4 validación final y cierre (17 y 18). Es la unidad de planificación.

**Story Point (SP)**
Unidad relativa de esfuerzo de una User Story. No equivale a horas.

**Escenario**
Caso de comportamiento esperado escrito en Gherkin, acordado con el cliente final, que se ejecuta automáticamente en el pipeline y funciona como documentación viva.

**Definition of Ready / Definition of Done**
Condiciones que un ticket debe cumplir para entrar a un sprint y para considerarse terminado, respectivamente.
