# Glosario de dominio

Vocabulario único del proyecto. Se usa igual en el código, en los tickets, en los escenarios Gherkin y en las conversaciones con la Secretaría. Si una palabra no está acá, no se usa en un ticket.

Las entradas marcadas con **(V)** deben validarse con la Secretaría antes de codificarse: son interpretaciones nuestras de la planilla y de la reunión, no definiciones oficiales.

---

## Dominio del trámite

**Tipo de trámite**
Definición reutilizable de un procedimiento administrativo: sus requisitos, los documentos que exige, la secuencia de pasos, los estados posibles y los roles responsables de cada paso. Vive como configuración versionada en el repositorio.

**Trámite (o solicitud)**
Instancia concreta de un tipo de trámite, iniciada por un titular. Tiene un titular, un estado, un historial de pasos y un conjunto de documentos. Es la unidad de seguimiento del sistema.

**Circuito** **(V)**
Forma general que toma la secuencia de pasos de un tipo de trámite. **Ninguno de los cinco trámites se resuelve íntegramente dentro del área**: todos salen hacia al menos un evaluador externo y vuelven. Lo que diferencia un circuito de otro es cuántos evaluadores intervienen y cuántas idas y vueltas hay.

**Paso**
Unidad mínima de avance dentro de un trámite. Tiene un responsable (rol), un tipo de acción y un resultado registrado. Un paso no se considera cumplido hasta que alguien lo confirma en la plataforma.

**Tipo de acción de un paso** **(V)**
Categorías abstractas derivadas de la planilla de circuitos:
- *Carga de documento:* alguien adjunta un archivo.
- *Carga de dato:* alguien completa uno o más campos.
- *Confirmación:* alguien declara que ejecutó una acción fuera de la plataforma ("listo, ya lo cargué").
- *Aprobación:* alguien con perfil de aprobación da conformidad a un paso y el trámite avanza. Es la aprobación de un paso del proceso, no el resultado del trámite (ADR-009).
- *Firma:* caso particular de aprobación sobre un documento. **La plataforma no implementa firma digital con validez legal.**

**Estado del trámite**
Situación en la que se encuentra el trámite en un momento dado, derivada del último paso confirmado. El conjunto de estados válidos es parte de la configuración del tipo de trámite.

**Resultado del trámite**
Lo que dictamina el evaluador externo: reconocimiento total, parcial o rechazo. **La plataforma no lo modela ni lo interpreta** (ADR-009): lo recibe como documento y lo comunica al titular. Para la plataforma, los tres casos son el mismo evento.

**Evaluador externo** **(V)**
Actor ajeno a la Secretaría que dictamina sobre el contenido académico del trámite: comisión curricular, departamento u otra unidad académica, según el trámite. No usa la plataforma (duda pendiente, ver alcance).

**Derivación**
Salida del trámite hacia un evaluador externo o un sistema externo (expediente electrónico). Se registra como un paso con responsable, destino, fecha y, si corresponde, número de expediente. La acción ocurre afuera.

**Retorno**
Vuelta del trámite desde el evaluador externo, con su dictamen adjunto. La plataforma registra el retorno y comunica el dictamen sin interpretarlo (ADR-009).

**Devolución**
Vuelta del trámite a un paso anterior, decidida por la Secretaría e indicando el motivo.

**Subsanación**
Pedido puntual al titular para que complete o corrija algo, sin cerrar ni reiniciar el trámite. También canaliza los pedidos de ampliación que hace un evaluador externo.

**Puente humano**
Persona de la Secretaría que ejecuta una acción en un sistema externo y la registra en la plataforma. Es el mecanismo elegido para no depender de integraciones (ADR-004).

**Rol**
Cada uno de los dos tipos de persona usuaria de la plataforma: **titular del trámite** y **Secretario**.

**Perfil**
Nivel de permisos dentro del rol Secretario. El **perfil operativo** recibe, verifica, deriva y registra; el **perfil de aprobación** además aprueba y firma. En el mapa de usuario son un solo rol; en permisos se distinguen.

**Titular del trámite**
Persona que inicia y a cuyo nombre corre el trámite. Se usa esta expresión y no "estudiante", porque no todos los trámites los inicia alguien con condición de estudiante activo.

## Documentos

**Requisito**
Condición que el trámite exige para poder avanzar. Puede ser un documento, un dato o una condición académica. Forma parte de la configuración del tipo de trámite y es la base contra la que trabaja la validación asistida.

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
Archivos que definen los tipos de trámite, mantenidos en el repositorio y modificados por el equipo mediante pull request. No existe editor dinámico (ADR-001); es una duda pendiente con la Secretaría.

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
