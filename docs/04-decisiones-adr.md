# Registro de decisiones (ADR)

Cada decisión que condiciona el producto, la arquitectura o el alcance se asienta acá. Formato: contexto, decisión, consecuencias, estado. Una decisión no se borra: se marca como reemplazada por otra.

---

## ADR-001 — Los tipos de trámite se definen como configuración versionada

**Estado:** aceptada, en revisión (D-07)
**Reemplaza:** el consenso de la reunión del 16/09 sobre incorporar una herramienta para que la Secretaría dé de alta, baja y modifique flujos

**Contexto.**
El anteproyecto aprobado establece que la definición de cada trámite se mantiene como archivos de configuración estáticos versionados en el repositorio, sin editor dinámico. En la reunión de planificación del 16/09 el equipo acordó lo contrario: una herramienta de edición de flujos para el personal administrativo. Se señaló allí mismo que esa funcionalidad es potente y que exige definir estrictamente qué rol tiene autorización para modificar un trámite.

**Decisión.**
Se mantiene la configuración versionada. Dar de alta o modificar un trámite es una tarea del equipo mediante pull request.

**Fundamento.**
- Un editor de flujos usable por personal no técnico implica motor de workflows, versionado de definiciones, validación de consistencia, permisos de edición y una interfaz propia. Es un producto aparte.
- El presupuesto del proyecto es de 14 sprints con carga, con 4 personas a dedicación parcial.
- La configuración versionada ya permite incorporar trámites nuevos sin tocar el código, que es el objetivo de fondo.
- Es lo que declara el documento aprobado por la cátedra.

**Consecuencias.**
- El alcance negativo incorpora el editor dinámico de flujos.
- Aparece una persona usuaria nueva: el administrador técnico (P4).
- Queda una expectativa de stakeholder que hay que desactivar explícitamente con la Secretaría (riesgo R-02).
- La documentación de cómo se edita la configuración pasa a ser un entregable obligatorio del cierre, porque al terminar el proyecto el equipo deja de estar disponible.
- El diseño del modelo de configuración debe ser lo bastante declarativo como para que un editor dinámico sea viable después, sin rediseñar.

---

## ADR-002 — Los tres usos de IA entran al MVP, con autonomía acotada

**Estado:** aceptada

**Contexto.**
El anteproyecto contempla tres usos de IA: asistente al titular, validación de documentación y generación de preactas. En la reunión del 16/09 se cuestionó el valor de la validación automática en la mayoría de los pasos, y se observó que el armado de la preacta conviene resolverlo de forma determinista mediante plantilla con campos variables. El anteproyecto compromete además que ninguna acción sobre un trámite se toma sin aprobación humana.

**Decisión.**
Los tres usos entran al MVP, cada uno con un nivel de autonomía acotado:

1. **Asistente RAG al titular:** responde sobre requisitos y documentación, cita la fuente de cada respuesta y deriva a una persona lo que no está cubierto por la base de conocimiento.
2. **Validación asistida:** compara la documentación cargada contra los requisitos configurados y produce un checklist con hallazgos. No aprueba, no rechaza y no cambia el estado del trámite.
3. **Generación de preactas:** la plantilla y los campos estructurados se completan de forma determinista desde los datos del trámite. El LLM interviene únicamente en campos de texto libre. El resultado es siempre un borrador.

**Orden de prioridad (decidido por el equipo).**

1. Asistente RAG al titular
2. Validación asistida
3. Generación de preactas

**Orden de recorte.** El inverso. Si la capacidad no alcanza, sale primero la generación de preactas, después la validación asistida. El asistente al titular no se recorta: es el componente que el anteproyecto usa para justificar la innovación del proyecto frente al estado del arte.

**Consecuencias.**
- Es el mayor riesgo de capacidad del proyecto (riesgo R-01). Los tres usos se secuencian en sprints distintos y cada uno se entrega completo antes de empezar el siguiente.
- El valor para la Secretaría llega más tarde que el valor para el titular, porque los dos usos que le ahorran trabajo al personal quedan en segundo y tercer lugar. Conviene tenerlo presente al planificar las instancias de validación: la primera demo le va a parecer poco a la Secretaría.
- Cada uso requiere su propio set de evaluación con casos esperados, incluyendo casos de inyección de prompt sobre documentos aportados por titulares.

---

## ADR-003 — El proveedor de identidad se abstrae detrás de una interfaz

**Estado:** reemplazada por ADR-010

**Contexto.**
La reunión del 16/09 identificó el ingreso unificado de la facultad (FUA) como opción preferida de autenticación, con la salvedad de que nadie sabe si la integración es viable. Se acordó consultarlo. Esperar la respuesta bloquearía el primer sprint.

**Decisión.**
La autenticación se diseña contra una interfaz propia de proveedor de identidad, con dos implementaciones: la institucional, si resulta viable, y una local con credenciales propias como alternativa. Los roles se resuelven siempre en la base de datos del sistema, no se asumen provistos por el proveedor externo.

**Consecuencias.**
- El desarrollo no queda bloqueado por la respuesta institucional.
- Resolver roles localmente es necesario de todos modos: no hay expectativa razonable de que un proveedor institucional conozca los roles internos de la Secretaría.
- La consulta a la Subsecretaría de TICs sigue siendo una tarea abierta, con fecha límite: si no hay respuesta al cierre de la fase de diseño, se cierra con la implementación local.

---

## ADR-004 — La interacción con sistemas externos se resuelve con puente humano

**Estado:** aceptada

**Contexto.**
Varios pasos de los circuitos ocurren en sistemas ajenos a la plataforma: expediente electrónico, comisión curricular, otros sistemas institucionales. El anteproyecto establece que el diseño no depende de interfaces externas y que el puente lo realiza el personal de la Secretaría. La reunión del 16/09 llegó a la misma conclusión: la persona ejecuta la acción afuera y confirma el paso dentro de la plataforma.

**Decisión.**
Todo paso que ocurre fuera de la plataforma se modela como un paso de tipo *confirmación*, con responsable, fecha y trazabilidad de quién lo declaró cumplido.

**Consecuencias.**
- La plataforma es la fuente de verdad del **estado** del trámite, no de la ejecución de todos sus pasos.
- Existe la posibilidad de que un paso se declare cumplido sin haberlo sido. Queda registrado quién lo declaró, que es exactamente lo que hoy no existe.
- La automatización progresiva de pasos individuales queda habilitada sin rediseño: un paso de confirmación puede volverse automático más adelante.

---

## ADR-005 — Sin datos personales reales en ningún entorno

**Estado:** aceptada

**Contexto.**
El proyecto trabaja sobre trámites que en producción involucran datos personales. La disponibilidad de datos reales depende además de tiempos burocráticos que el proyecto no controla.

**Decisión.**
Todos los entornos del proyecto usan datos sintéticos o anonimizados. La base de conocimiento contiene requisitos, preguntas frecuentes y criterios, nunca expedientes ni datos de personas.

**Consecuencias.**
- Se requiere un generador de datos sintéticos como parte de la infraestructura base.
- Las instancias de validación con la Secretaría se hacen sobre casos sintéticos representativos, acordados de antemano.
- Se elimina la dependencia de los tiempos burocráticos para poder avanzar.

---

## ADR-006 — Los cinco trámites se tratan como una familia, con cobertura incremental

**Estado:** aceptada

**Contexto.**
La Secretaría priorizó cinco trámites: equivalencia entre asignaturas, reconocimiento de créditos, intercambios académicos, pase de universidad extranjera y pase de universidad dentro de la UBA. El equipo definió que los cinco entran al producto final. Los cinco resuelven la misma pregunta de fondo, el reconocimiento de estudios cursados fuera del plan actual del titular, y por lo tanto comparten estructura de requisitos, tipo de evaluador y tipo de resolución. Además, ninguno se resuelve íntegramente dentro del área: todos tienen interacción externa en mayor o menor grado.

**Decisión.**
Los cinco entran al producto final. Se implementan de forma incremental: un trámite ancla modelado a fondo, un segundo trámite que confirma que el modelo generaliza, y los tres restantes incorporados por configuración, uno por sprint y empezando por el más complejo. La derivación a actores externos se modela como caso normal del motor de pasos, no como excepción.

**Fundamento.**
- Si los cinco son una familia, el costo marginal del tercero en adelante es bajo, siempre que la abstracción del primero sea correcta.
- El riesgo no está en la cantidad de trámites sino en la calidad de la abstracción, y ese riesgo se paga temprano o tarde. Se elige pagarlo temprano.
- Probar la generalización temprano, con el segundo trámite en el sprint 4, permite corregir la abstracción con todo 2026 por delante.
- Poner el caso más complejo primero dentro de la cobertura, y no último, deja dos sprints para corregir si la abstracción no lo soporta.

**Consecuencias.**
- Ancla: equivalencia entre asignaturas. Segundo trámite: reconocimiento de créditos, por decisión del equipo. El pase de universidad extranjera, el caso más complejo, pasa a la cobertura: la prueba de generalización contra el caso difícil se corre de noviembre a abril.
- Se define un criterio de alarma explícito: si incorporar uno de los tres trámites de cobertura exige código nuevo y no solo configuración, se detiene y se corrige la abstracción en vez de parchear.
- El relevamiento debe buscar activamente lo que los cinco trámites tienen **en común**, no solo describir cada uno por separado.

---

## ADR-007 — La documentación se produce dentro de cada ticket, no como historia aparte

**Estado:** aceptada

**Contexto.** El plan anterior tenía tres historias de documentación concentradas al final, por 18 SP. Documentación escrita al cierre sobre código de ocho meses atrás es documentación reconstruida de memoria, y llega justo cuando no queda margen.

**Decisión.** No hay historias de documentación en el backlog. La documentación se escribe dentro del ticket que la origina, forma parte de la DoD y el agente de desarrollo la produce y la actualiza en el mismo PR que el código. La única excepción es el informe final, que es un entregable académico.

**Consecuencias.**
- La DoD se endurece: un PR sin su documentación no se mergea.
- El instructivo de cómo editar la configuración de trámites, que ADR-001 vuelve obligatorio, se escribe en el PR que toca el modelo, no en el sprint 19.
- La documentación funcional para la Secretaría se compila desde los escenarios Gherkin, que ya se escriben con el cliente.
- Se liberan 18 SP del cierre, que es exactamente donde hacía falta margen.

---

## ADR-008 — Cierre en mayo de 2027 con alcance completo y recorte diferido

**Estado:** aceptada

**Contexto.** El trabajo tiene que estar terminado en mayo de 2027, y diciembre es parcial, enero nulo y febrero bajo. Eso deja 14 sprints con carga. Con el relevamiento, el stack y la verificación manual incorporados al backlog, el alcance queda en 264 SP tras la revisión historia por historia: los tres usos de IA, los cinco trámites y el esqueleto completo.

**Decisión.** Se mantiene el alcance completo. No se recorta nada de antemano. El orden de recorte de ADR-002 queda armado y se activa únicamente en un punto de control, con velocidad medida sobre la mesa, y lo decide el equipo.

**Fundamento.** Recortar antes del sprint 2 es recortar sin datos: la velocidad del equipo es desconocida y cualquier estimación de capacidad es una conjetura. Un recorte hecho con velocidad medida es una decisión; hecho ahora es una apuesta que además empobrece el producto sin necesidad.

**Consecuencias.**
- La planificación exige unos 24 SP por sprint en los cinco sprints plenos del bloque 1, con pico de 30 en el sprint 4. Es un ritmo exigente y llega antes de que exista una medición de velocidad (riesgo R-13).
- El cierre conserva margen: el sprint 16 queda en 14 SP; el 17 compromete solo 8 SP y reserva el resto para absorber el feedback de la validación; el 18 no tiene historias y queda para el informe y los ajustes no bloqueantes.
- Los puntos de control de los sprints 2, 4, 6, 13, 14 y 16 (ver roadmap) dejan de ser informativos y pasan a ser decisorios. Si en alguno la velocidad no acompaña, ahí el equipo decide el recorte.
- **Orden de recorte vigente, de primero a último:** generación de preactas · validación asistida · cobertura de cinco trámites reducida a tres. El asistente RAG no se recorta.
- Si el recorte llega tarde, el costo no es perder una funcionalidad sino entregar tres a medias. Por eso los puntos de control tienen fecha.

---

## ADR-009 — La plataforma comunica el resultado de un trámite, no lo interpreta

**Estado:** aceptada

**Contexto.** Los cinco trámites terminan en un dictamen del evaluador externo que puede reconocer todo, una parte o nada de lo solicitado. Modelar esos resultados obligaría a la plataforma a entender el contenido académico de cada dictamen.

**Decisión.** Para la plataforma, un reconocimiento total, uno parcial y un rechazo son el mismo evento: un dictamen que vuelve del evaluador, se adjunta al trámite y se comunica al titular. La plataforma no clasifica, no interpreta ni modela el resultado.

**Consecuencias.**
- No hay estados de resultado en el modelo de configuración (US08) ni en el inventario de acciones (US63).
- El retorno de la evaluación (US21) registra el dictamen como documento, sin clasificarlo.
- Un pedido de ampliación del evaluador no es un resultado sino un paso del proceso: se canaliza como pedido de subsanación (US33).
- En las preactas (US45), los campos que dependen del dictamen los completa la persona. La plataforma completa solo lo que sale de los datos del trámite.
- Refuerza la exclusión del dictamen automático de equivalencias.

---

## ADR-010 — Se ingresa con la cuenta de Google de la UBA

**Estado:** aceptada
**Reemplaza:** ADR-003

**Contexto.** ADR-003 dejó la autenticación detrás de una interfaz, con una implementación local de credenciales propias por si el ingreso institucional no resultaba viable. En la reunión con la Subsecretaría de TICs del 9/10 se confirmó que SIGTA puede usar el ingreso con la cuenta de Google de la UBA, y que todos los titulares de los cinco trámites tienen cuenta UBA.

**Decisión.** Se ingresa solo con la cuenta de Google de la UBA, a través de un puerto de proveedor de identidad. SIGTA no guarda contraseñas ni ofrece otra forma de ingresar. Los roles, áreas y perfiles se administran en SIGTA: el proveedor solo dice quién es la persona.

**Consecuencias.**
- Se descartan la implementación local con credenciales propias y el enlace por correo para titulares sin cuenta institucional.
- Se descarta Keycloak o Zitadel como evolución: su disparador era que no se habilitara el ingreso institucional.
- Se cierran la decisión pendiente D-05 y el riesgo R-11, y se confirma el supuesto S-03.
- Los entornos local y de integración continua usan un adaptador de identidad falso, porque los datos sintéticos (ADR-005) no tienen cuenta UBA.
- Falta acordar con TICs cómo se registra SIGTA como aplicación y desde qué entornos se puede usar el ingreso real.

---

## ADR-011 — Stack y arquitectura para la defensa

**Estado:** aceptada

**Contexto.** US65 pide decidir el stack con los criterios del anteproyecto: costo para la institución, integración con GitHub Actions y mantenimiento posterior. Las decisiones pendientes D-03 (stack) y D-06 (almacenamiento de archivos) dependían de eso. La decisión se tomó el 4/10/2026 y su detalle está en el README.

**Decisión.**
- Un monolito modular en Go para el servicio de trámites, y un único servicio aparte en Python para la inteligencia artificial.
- Frontend en React, TypeScript y Vite.
- PostgreSQL con pgvector para el estado, el historial, la cola de trabajos, la búsqueda de texto y los vectores.
- Documentos en el sistema de archivos, detrás de un puerto.
- Contrato primero con OpenAPI, y arquitectura de puertos y adaptadores en cada servicio.
- Todo en contenedores, autoalojable, con Docker Compose en desarrollo y GitHub Actions como integración continua.

**Consecuencias.**
- Se cierran D-03 y D-06.
- Cada componente adicional (bus de eventos, almacenamiento compatible con S3, observabilidad completa, API gateway) se suma solo cuando se cumple su disparador, listado en la sección "Evolución" del README.
- Las alternativas descartadas y su fundamento están en la sección "Alternativas descartadas" del README.

---

## Decisiones pendientes

| ID | Decisión pendiente | Necesaria antes de | Responsable |
|---|---|---|---|
| D-02 | Si los perfiles operativo y de aprobación son personas distintas | Diseñar permisos (US64, sprint 1) | Canal con Secretaría |
| D-04 | Infraestructura de despliegue y su financiamiento | Fin de la fase de diseño | Equipo |
| D-07 | Si la Secretaría da de alta y modifica los trámites por su cuenta, con un editor dentro de la aplicación, lo que revertiría el ADR-001 | Próxima reunión del equipo | Equipo y canal con la Secretaría |
