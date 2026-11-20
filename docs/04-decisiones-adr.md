# Registro de decisiones (ADR)

Cada decisión que condiciona el producto, la arquitectura o el alcance se asienta acá. Formato: contexto, decisión, consecuencias, estado. Una decisión no se borra: se marca como reemplazada por otra.

---

## ADR-001: Los tipos de trámite se definen como configuración versionada

**Estado:** reemplazada por ADR-013
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

## ADR-002: Los tres usos de IA entran al MVP, con autonomía acotada

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

## ADR-003: El proveedor de identidad se abstrae detrás de una interfaz

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

## ADR-004: La interacción con sistemas externos se resuelve con puente humano

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

## ADR-005: Sin datos personales reales en ningún entorno

**Estado:** aceptada
**Modificada por:** ADR-027: la base guarda el correo y el nombre de las cuentas reales que ingresan con Google, en todos los entornos. Los trámites, sus documentos y los datos de las pruebas siguen siendo sintéticos.

**Contexto.**
El proyecto trabaja sobre trámites que en producción involucran datos personales. La disponibilidad de datos reales depende además de tiempos burocráticos que el proyecto no controla.

**Decisión.**
Todos los entornos del proyecto usan datos sintéticos o anonimizados. La base de conocimiento contiene requisitos, preguntas frecuentes y criterios, nunca expedientes ni datos de personas.

**Consecuencias.**
- Se requiere un generador de datos sintéticos como parte de la infraestructura base.
- Las instancias de validación con la Secretaría se hacen sobre casos sintéticos representativos, acordados de antemano.
- Se elimina la dependencia de los tiempos burocráticos para poder avanzar.

---

## ADR-006: Los cinco trámites se tratan como una familia, con cobertura incremental

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

## ADR-007: La documentación se produce dentro de cada ticket, no como historia aparte

**Estado:** aceptada

**Contexto.** El plan anterior tenía tres historias de documentación concentradas al final, por 18 SP. Documentación escrita al cierre sobre código de ocho meses atrás es documentación reconstruida de memoria, y llega justo cuando no queda margen.

**Decisión.** No hay historias de documentación en el backlog. La documentación se escribe dentro del ticket que la origina, forma parte de la DoD y el agente de desarrollo la produce y la actualiza en el mismo PR que el código. La única excepción es el informe final, que es un entregable académico.

**Consecuencias.**
- La DoD se endurece: un PR sin su documentación no se mergea.
- El instructivo de cómo editar la configuración de trámites, que ADR-001 volvía obligatorio y ADR-013 mantiene, se escribe en el pull request que toca el modelo, no en el sprint 19.
- La documentación funcional para la Secretaría se compila desde los escenarios Gherkin, que ya se escriben con el cliente.
- Se liberan 18 SP del cierre, que es exactamente donde hacía falta margen.

---

## ADR-008: Cierre en mayo de 2027 con alcance completo y recorte diferido

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

## ADR-009: La plataforma comunica el resultado de un trámite, no lo interpreta

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

## ADR-010: Se ingresa con la cuenta de Google de la UBA

**Estado:** aceptada
**Reemplaza:** ADR-003
**Modificada por:** ADR-027: el entorno local también ingresa con Google, y el adaptador de identidad falso queda solo para las pruebas.

**Contexto.** ADR-003 dejó la autenticación detrás de una interfaz, con una implementación local de credenciales propias por si el ingreso institucional no resultaba viable. En la reunión con la Subsecretaría de TICs del 9/10 se confirmó que SIGTA puede usar el ingreso con la cuenta de Google de la UBA, y que todos los titulares de los cinco trámites tienen cuenta UBA.

**Decisión.** Se ingresa solo con la cuenta de Google de la UBA, a través de un puerto de proveedor de identidad. SIGTA no guarda contraseñas ni ofrece otra forma de ingresar. Los roles, áreas y perfiles se administran en SIGTA: el proveedor solo dice quién es la persona.

**Consecuencias.**
- Se descartan la implementación local con credenciales propias y el enlace por correo para titulares sin cuenta institucional.
- Se descarta Keycloak o Zitadel como evolución: su disparador era que no se habilitara el ingreso institucional.
- Se cierran la decisión pendiente D-05 y el riesgo R-11, y se confirma el supuesto S-03.
- Los entornos local y de integración continua usan un adaptador de identidad falso, porque los datos sintéticos (ADR-005) no tienen cuenta UBA.
- Falta acordar con TICs cómo se registra SIGTA como aplicación y desde qué entornos se puede usar el ingreso real.

---

## ADR-011: Stack y arquitectura para la defensa

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

## ADR-012: El servicio asigna la versión de un tipo de trámite al publicarlo

**Estado:** aceptada

**Contexto.** La historia "Modelo de definición de trámite" pide que un trámite en curso no se rompa cuando cambia la configuración de su tipo de trámite. Si el número de versión lo escribe quien edita, nada impide cambiar el contenido de una versión ya publicada sin cambiar el número, y un trámite podría terminar con pasos distintos de los que tenía cuando empezó.

**Decisión.**
- Cada tipo de trámite se identifica por un `id` numérico que asigna la base, sin número de versión en la definición.
- Al publicar, el servicio codifica el contenido (la definición, con sus datos y documentos, más las áreas que usan sus pasos) y calcula el hash SHA-256 de esos bytes. Si difiere del de la última versión, publica la siguiente; si es igual, devuelve la última.
- La versión guarda el contenido byte por byte, tal como se hasheó, con el número de formato del contenido. El hash se verifica sobre esos bytes y nunca sobre una recodificación, para que un cambio en la codificación no invalide las versiones publicadas. El formato sube cada vez que cambia el contenido, y una clave desconocida hace fallar la lectura.
- Quien publica indica desde qué versión editó. Si ya no es la última, la publicación se rechaza, para que un cambio no pise a otro publicado en el medio.
- Una versión publicada no se modifica, y su hash se verifica cada vez que se lee: al publicar sobre ella y al leerla para un trámite.
- Cada trámite apunta a la versión con la que empezó.

**Consecuencias.**
- Queda constancia de la versión exacta que siguió cada trámite, incluidos los nombres de pasos, datos y documentos que vio el titular.
- Cualquier cambio visible genera una versión nueva, incluso corregir un nombre.
- El hash detecta un contenido alterado por error, no una alteración intencional: quien puede escribir el contenido puede recalcular el hash. Eso queda en la decisión abierta "Integridad del historial" del README.
- Si el hash de una versión no coincide, no se publica nada sobre ella ni se usa para un trámite hasta revisarla. Cómo se revisa queda para el ticket de la persistencia.
- Dos publicaciones simultáneas desde la misma base compiten por el mismo número y solo una lo guarda. La otra se rechaza como desactualizada, aunque haya publicado el mismo contenido: si vuelve a publicar, recibe la versión guardada.
- El contenido se guarda byte por byte. En PostgreSQL eso es `bytea` y no `jsonb`, porque `jsonb` reordena las claves y el hash dejaría de coincidir.
- Las versiones necesitan persistencia, que es de otro ticket. Hasta entonces solo existen en los tests, con un adaptador en memoria.
- Un trámite en curso no pasa a otra versión. Si hiciera falta migrarlo, la migración sería un evento auditado del trámite, con la versión de origen, la de destino y la correspondencia entre pasos.
- Un tipo de trámite se va a dar de baja marcándolo inactivo, y los trámites en curso terminan con su versión. Todavía no está implementado.

---

## ADR-013: Los tipos de trámite se editan desde la interfaz, con una persona que propone y otra que aprueba

**Estado:** aceptada por el equipo. A confirmar con los directores: contradice el anteproyecto, que pide archivos de configuración estáticos versionados en el repositorio, sin editor dinámico.
**Reemplaza:** ADR-001
**Modificada por:** ADR-027: ya no hay usuarios sintéticos con los dos permisos; nadie los tiene hasta que un administrador da las membresías.

**Contexto.** ADR-001 mantenía los tipos de trámite como archivos en el repositorio, editados por el equipo mediante pull request, y quedó en revisión por D-07, que evaluaba un editor dentro de la aplicación. Con las versiones guardadas en la base y validadas al publicar (ADR-012), un editor ya no exige un motor aparte: es una interfaz sobre el mismo modelo. Además, al terminar el proyecto el equipo deja de estar disponible para hacer pull requests.

**Decisión.**
- Los tipos de trámite se editan desde la interfaz de SIGTA, como borradores, y se publican como versiones nuevas.
- Cada versión la propone una persona y la aprueba otra. La versión guarda a las dos y la fecha de publicación.
- No se publica una definición inválida: el servicio la valida al publicar.

**Consecuencias.**
- Se cierra D-07. Queda D-08: quién puede proponer y quién aprobar.
- El editor deja de estar fuera del alcance; lo que queda afuera es publicar sin la aprobación de otra persona.
- La documentación del modelo, en CONTRIBUTING, sigue siendo un entregable del cierre, ahora como referencia para quien edita.
- La edición y su API salen en US72 (SIG-74): un borrador por tipo de trámite, que se edita, se propone, se rechaza con un motivo o se aprueba. Al aprobarlo se publica la versión, con quien la propuso y quien la aprobó; el borrador recuerda a todos los que lo editaron, ninguno de ellos ni quien lo propuso lo aprueba o lo rechaza, y el servidor controla los dos permisos. La pantalla de edición todavía no existe: sale en SIG-84.
- Los dos permisos son los de plataforma de la Secretaría, `edit_procedure_types` y `publish_procedure_types` (ADR-020), que en desarrollo tienen dos usuarios sintéticos distintos. A quién se le dan sigue en D-08.
- Los borradores de los cinco trámites se siguen validando fuera del repositorio, con el test `TestDrafts`, porque describen procesos que la Secretaría no confirmó.
- Cambian cuatro criterios de "Modelo de definición de trámite": el archivo versionado, los estados, la validación al arrancar y los requisitos compartidos entre trámites. A actualizar en Linear con el equipo.
- No aparece la persona administrador técnico (P4) que preveía ADR-001.
- El riesgo R-02 baja, pero sigue abierto hasta resolver D-08.

---

## ADR-014: El servicio de trámites se divide por módulo y, dentro de cada módulo, por componente

**Estado:** propuesta, a acordar con el equipo. ADR-021 reemplaza la regla de cómo un módulo usa a otro.

**Contexto.** ADR-011 eligió un monolito modular con puertos y adaptadores, pero el código se ordenó solo por componente: un `domain/`, un `application/`, un `ports/` y un `adapters/` para todo el servicio. Con dos módulos (ingreso y tipos de trámite) ya aparecían los problemas: los puertos de salida de usuarios, sesiones, enlaces de ingreso y correo compartían archivo; los adaptadores en memoria y REST de los dos módulos compartían paquete; y dos ramas que tocaban módulos distintos chocaban en los mismos archivos. Cada módulo nuevo (trámites iniciados, pasos, bandeja) iba a tocar las mismas carpetas.

**Decisión.**
- `services/backend/src/` se divide por módulo: `auth/` (ingreso y sesiones) y `procedure/` (tipos de trámite y versiones).
- Dentro de cada módulo se mantienen los componentes de puertos y adaptadores: `domain/`, `application/`, `ports/input/`, `ports/output/` y `adapters/`.
- Un módulo no importa el núcleo de otro; si lo necesita, lo hace a través de un puerto.
- El contrato sigue siendo uno solo. Su código generado vive en `api/`, junto con el montaje de la API, y lo importan los adaptadores REST de cada módulo. `cmd/` compone los handlers de todos los módulos en el servidor.

**Consecuencias.**
- Un módulo nuevo es una carpeta nueva con los mismos componentes, sin tocar las de los demás.
- Las ramas abiertas al momento del cambio tienen que moverse a la estructura nueva al integrarse.
- Partir el contrato por módulo queda como alternativa si el archivo único se vuelve difícil de mantener.

---

## ADR-015: La interfaz se arma con Mantine y toma su aspecto de un único tema

**Estado:** propuesta, a acordar con el equipo

**Contexto.** ADR-011 fijó React, TypeScript y Vite, pero no cómo se construye la interfaz. El equipo revisa con dificultad el código de frontend, y la mayor parte la va a escribir un agente. Las opciones evaluadas en noviembre de 2026 fueron: shadcn/ui, que copia los componentes al repositorio y los estiliza con cadenas largas de clases de Tailwind; Base UI y React Aria, que no traen estilo y dejan todo el CSS a cargo del equipo; Radix Themes, sin versiones nuevas desde enero de 2026; IBM Carbon, pesado y difícil de alejar del aspecto de IBM; y Mantine, una biblioteca completa (formularios, fechas, menús, notificaciones) que se configura con propiedades y un objeto de tema.

**Decisión.**
- La interfaz se construye con componentes de Mantine 9, sin Tailwind ni CSS propio.
- El aspecto (colores, tipografía, radios, valores por defecto de cada componente) se define solo en `web/src/theme/theme.ts`.
- Las pantallas usan valores del tema, nunca colores ni medidas escritos a mano. Lo controla el lint con `react/forbid-component-props`, `react/forbid-dom-props`, `no-restricted-imports` y una regla propia en `web/lint/design-rules.js`.
- Las reglas de diseño, con su motivo, están en `web/DESIGN.md`.
- La única excepción es el botón "Iniciar sesión con Google", que Google exige con colores y tipografía propios.

**Consecuencias.**
- Revisar una pantalla es revisar qué componentes usa y con qué propiedades. Un cambio de aspecto se revisa en un solo archivo.
- La paleta institucional queda a confirmar hasta que se defina la identidad visual. Mientras tanto se usa el azul de Mantine.
- El frontend depende de Mantine: cambiar de biblioteca implica reescribir las pantallas.
- El frontend compilado pesa 533 kB sin comprimir (166 kB comprimido), por encima del aviso de 500 kB de Vite. A revisar con división por rutas cuando crezca la cantidad de pantallas.

---

## ADR-016: El dominio de los tipos de trámite se divide en definición, versión y catálogo

**Estado:** aceptada

**Contexto.** ADR-014 dejó un solo paquete `domain/` por módulo. En `procedure/` ese paquete juntaba cuatro responsabilidades: el modelo de la definición y su validación, el reporte de los problemas encontrados, el versionado (contenido, hash y autoría) y los requisitos que muestra el catálogo. Los tests no seguían esa división (`validate_test.go` probaba varios archivos a la vez), y varios nombres decían qué tipo de dato eran y no para qué servían (`Definition`, `Version`, `Problem`, `Ref`, `Code`, `Kind`, `Encoded`).

**Decisión.**
- `procedure/domain/` se divide en tres paquetes:
  - `definition/`: la definición de un tipo de trámite (`ProcedureTypeDefinition`), sus áreas, datos, documentos, pasos y salidas, y su validación (`ValidationProblem`, `ValidationError`).
  - `version/`: el contenido que se publica (`Content`), su hash, la autoría y la versión publicada (`ProcedureTypeVersion`). Importa `definition/`.
  - `catalog/`: los requisitos que el catálogo muestra al titular. Importa `definition/` y `version/`.
- Los nombres pasan a decir qué representan dentro del dominio: `ProcedureTypeDefinition`, `ProcedureTypeVersion` (con `ProcedureTypeID`, `EncodedContent`, `VerifyDigest` y `DecodeContent`), `ValidationProblem`, `ValidationError`, `ProblemCode`, `ElementKind` y `Element`.
- Los tests de cada paquete van en el paquete externo (`definition_test`, `version_test`, `catalog_test`) y prueban la API pública. Lo no exportado que hace falta probar se expone en `export_test.go`, que solo se compila con los tests. La definición y las áreas de prueba que comparten están en `definition/definitiontest`.

**Consecuencias.**
- Los tags JSON no cambian, así que la codificación del contenido y el hash de las versiones ya publicadas siguen iguales. Lo comprueban los tests de `version/testdata/formatN.json`.
- La aplicación, los puertos y los adaptadores importan los subpaquetes que usan. `ApplicantRequirements` pasa de método de `Content` a función de `catalog/`.
- Un paquete del dominio no puede usar lo no exportado de otro: lo que se comparta entre ellos tiene que exportarse o quedarse en `definition/`.
- El módulo `auth/` mantiene un solo `domain/`. Si crece de la misma forma, se divide con el mismo criterio.

**Nota.** Los paquetes nuevos de `procedure/domain/` siguen el mismo criterio: uno por lo que se muestra o se calcula, que importa `definition/` y `version/` y no los otros paquetes hermanos. `detail/` arma los pasos que muestra el detalle de un trámite y les pone al lado sus archivos, y dónde está un trámite derivado. `engine/` decide el paso siguiente de una confirmación y arma el evento que la registra (ADR-023). `submission/` define los pasos iniciales de un trámite, lo que piden y cómo se validan los datos que envía el titular. Lo que comparten el catálogo y el detalle, como los requisitos de un paso (`FieldRequirement`, `DocumentRequirement`), va en `definition/`. El estado del trámite (`State`) va en `domain/`, junto al trámite, para que el detalle y la bandeja lo calculen igual; `inbox/` arma la lista de la bandeja (filtros, orden y opciones de filtro) sobre trámites que `application` ya sabe visibles, con las mismas reglas que el detalle.

---

## ADR-017: Los permisos del personal se asignan con roles configurables, administrados en SIGTA

**Estado:** reemplazada por ADR-020
**Reemplaza:** la hipótesis de dos perfiles fijos del Secretario (operativo y de aprobación)

**Contexto.** El modelo de permisos partía de un rol Secretario con dos perfiles fijos, y D-02 preguntaba si eran personas distintas (supuesto S-07). El código y el contrato ya los tienen fijos como `staff_operative` y `staff_approver`. Al preguntarle por D-02, la Secretaría respondió el 4/11 con otro esquema: un rol de administrador y usuarios internos de cada área que interviene, con permisos que asigna el administrador. Los permisos van por rol y, más adelante, por tipo de trámite. También pide que quede registrado quién hizo cada acción en cada expediente. Además ofreció una reunión con las áreas para relevar el circuito en detalle.

**Decisión.**
- Un **permiso** habilita una acción del modelo de dominio (por ejemplo avanzar pasos, derivar, aprobar o editar tipos de trámite).
- Un **rol** es un conjunto de permisos que se define en SIGTA. A cada usuario del personal se le asignan uno o más roles.
- El rol **administrador** gestiona usuarios, roles y permisos. Lo crea el despliegue, no la interfaz.
- El titular sigue siendo un rol aparte, que no se configura.
- Los usuarios del personal pueden ser de cualquier área de la facultad. Cuáles entran se define en la reunión con las áreas.
- Los permisos por tipo de trámite quedan para después. El modelo no debe impedirlos.
- Cada acción queda en el historial con su autor, como ya preveía el historial de acciones.

**Consecuencias.**
- Se cierran D-02 y el supuesto S-07: lo que eran perfiles fijos pasan a ser roles iniciales que el administrador puede cambiar.
- SIG-14 ("Perfiles y autorización") se replantea, y el campo `role` del contrato pasa a ser una lista de roles. El cambio tiene que entrar antes de que más código dependa de `staff_operative` y `staff_approver`.
- El paso de un tipo de trámite deja de declarar un perfil fijo y declara el permiso que exige. Afecta el modelo de definición (SIG-12) y los borradores de los cinco trámites.
- La nota de SIG-13 ("solo entra la Secretaría de Gestión Académica") queda reemplazada. El mapa de qué área hace cada paso se completa en la reunión con las áreas.
- D-08 sigue abierta, pero ahora se puede resolver como un permiso para proponer y otro para aprobar.

---

## ADR-018: A un modelo de terceros solo llega contenido académico, sin datos identificatorios

**Estado:** aceptada

**Contexto.** Los trámites contienen datos personales de estudiantes. El 4/11 la Secretaría pidió que la arquitectura los trate conforme a la Ley 25.326, y que lo que se mande al servicio de IA sea solo contenido académico (programas, cargas horarias, contenidos mínimos), sin datos identificatorios. La Subsecretaría de TICs preguntó qué datos van a cada modelo, con qué políticas y si eso cumple la ley. Si se hostea un modelo propio, pide además sus requisitos de infraestructura. El volumen informado por la Secretaría es de unos 300 trámites por cuatrimestre, concentrados antes de cada cuatrimestre.

**Decisión.**
- Los componentes de IA empiezan con un **modelo de terceros por API**, detrás de un puerto. Un modelo propio no queda descartado: la elección se revisa cuando se conozcan el flujo de cada trámite y los costos de cada opción.
- Antes de cada llamada, un paso que corre dentro de SIGTA **quita los datos identificatorios** (nombre, documento, legajo, correo, firmas y cualquier dato de una persona). El puerto del modelo solo recibe contenido académico.
- Solo se usan proveedores cuyas condiciones garanticen que los datos enviados no se usan para entrenar y que su retención sea nula o acotada. El proveedor elegido y sus condiciones se registran.
- Cada llamada registra qué tipo de contenido salió, además del costo, la latencia y la versión del prompt.
- Sigue vigente ADR-005: en desarrollo y en la CI no hay datos reales, y los modelos usan adaptadores deterministas.

**Consecuencias.**
- La anonimización pasa a ser un componente propio, con su set de evaluación: un dato identificatorio que llega al modelo es un error grave.
- Afecta sobre todo a la lectura de documentos (SIG-37) y a la validación asistida (SIG-38). El analítico tiene datos personales; los programas, en general, no.
- Para empezar no hace falta pedirle a TICs una GPU. Un modelo propio va detrás del mismo puerto, y se evalúa con esos costos o si un proveedor deja de cumplir las condiciones.
- Quedan abiertos, con la Secretaría y el área legal, la base legal, la inscripción de la base y la retención de los documentos (D-09).

---

## ADR-019: El personal tiene un perfil por tipo de trámite

**Estado:** reemplazada por ADR-020

**Contexto.** El modelo de US07 tenía tres perfiles fijos por persona: titular, Secretario operativo y Secretario de aprobación. Cada persona tenía uno solo, así que nadie podía ser operativo en un trámite y de aprobación en otro, y el supuesto S-07 (que los dos perfiles son personas distintas) condicionaba el modelo. En su respuesta al relevamiento, la Secretaría propuso un rol de administrador, usuarios internos para cada área que interviene, permisos otorgados por el administrador "por rol y eventualmente por tipo de trámite", y el registro de quién hizo cada acción en cada expediente.

**Decisión.**
- Hay tres tipos de usuario, que se excluyen entre sí: **titular**, **personal** y **administrador**. Solo el titular inicia trámites, y un titular no recibe permisos de personal.
- Un **permiso** le da a una persona del personal un perfil (operativo o de aprobación) sobre un tipo de trámite. Una persona puede tener perfiles distintos en tipos distintos, y ninguno en otros.
- El permiso referencia el tipo de trámite y no una versión: publicar una versión nueva no lo cambia.
- El personal actúa sobre un paso de un área de la facultad (`office`) si tiene permiso sobre ese tipo de trámite y su perfil alcanza para el del paso. El de aprobación cubre también los pasos operativos. Los pasos del titular los hace solo el titular del trámite, y los de un evaluador externo no los hace nadie en SIGTA.
- El personal ve los trámites de los tipos sobre los que tiene permiso; el titular, solo los propios. A quien no puede ver un trámite se le responde como si no existiera.
- El administrador no actúa sobre trámites por ser administrador.
- La autorización se aplica en el caso de uso del servidor, y cada negativa queda registrada con quién, qué, sobre qué trámite y cuándo.
- El área no forma parte del permiso. Se suma si la Secretaría confirma que, en un mismo trámite, intervienen oficinas distintas con personas que no deben hacer los pasos de la otra (lo que era D-09, antes de que ese número pasara a la retención de documentos).

**Consecuencias.**
- S-07 deja de condicionar el modelo: una misma persona puede tener los dos perfiles, en el mismo tipo de trámite o en tipos distintos.
- Hace falta una historia para que el administrador dé de alta al personal y le asigne permisos. Mientras no exista, los permisos se cargan desde la configuración del despliegue.
- Durante la transición conviven dos modelos: el ingreso (`auth`) sigue con el rol único de cada usuario (`applicant`, `staff_operative` o `staff_approver`), que usan `/me` y `SIGTA_USERS_FILE`, y la autorización (`procedure`) usa los tipos de usuario y los permisos. Unificarlos cambia el contrato y la configuración, y se hace en un cambio aparte.

---

## ADR-020: El personal actúa según su membresía a un área y los roles y permisos de esa área que se le asignan

**Estado:** aceptada. Lo que falta confirmar con la Secretaría queda en D-08, D-10, D-11 y D-12.
**Reemplaza:** ADR-017 y ADR-019
**Modificada por:** ADR-027, que saca `SIGTA_USERS_FILE`: las membresías iniciales ya no se cargan desde la configuración del despliegue, la marca de administrador se guarda en la base y el primer administrador se da por consola, y la Secretaría la crea una migración.

**Contexto.** ADR-017 resolvió la misma respuesta de la Secretaría con roles configurables: conjuntos de permisos que define el administrador, y pasos que declaran el permiso que exigen. Esta decisión conserva esa idea, pero la ubica dentro de cada área: los permisos y los roles que los agrupan pertenecen a un área, y se asignan en la membresía. Además, ADR-019 ataba el permiso al tipo de trámite y separaba tres tipos de usuario excluyentes. En su respuesta al relevamiento, la Secretaría pidió un administrador, "usuarios internos del sistema para cada área interviniente con diferentes permisos otorgados por el administrador", un esquema "por rol (y eventualmente por tipo de trámite)" y el registro de quién hizo cada acción en cada expediente. Además, hoy la Secretaría avanza los pasos, pero más adelante pueden hacerlo otras áreas. Con ADR-019 quedaban tres problemas: el área no formaba parte del permiso; los pasos de un evaluador externo no los podía registrar nadie; y como cada persona tiene una sola cuenta UBA (ADR-010), quien administra no podía ser también personal. Por último, una persona del personal también puede ser alumna, y tiene que poder iniciar sus propios trámites con la misma cuenta.

**Decisión.**
- **Titular y personal.** No hay tipos de persona que se excluyan, y nada se guarda como campo. Al ingresar no hay forma de saber si una persona es alumna o personal, así que cualquiera puede iniciar trámites y es **titular** de los que inicia: titular es una relación con un trámite, no una condición de la persona. Es **personal** quien tiene al menos una membresía vigente. Una persona puede ser las dos cosas: el personal también inicia trámites. Las áreas, sus permisos y sus roles existen solo para el personal.
- **Áreas.** Dentro del personal no hay áreas fijas, salvo la Secretaría: el administrador crea las que hagan falta (Comisión Curricular X, Dirección de Alumnos, etc.) como datos de SIGTA y no de cada versión de un tipo de trámite. Las áreas no tienen tipo. Cada paso de un tipo de trámite es del titular o de un área. Un paso que hoy resuelve alguien de afuera (por ejemplo, el dictamen de una comisión) se asigna al área de quien lo registra en SIGTA.
- **La Secretaría.** No se sabe de antemano qué áreas va a haber, pero la Secretaría de Gestión Académica existe siempre: la trae SIGTA, con el identificador `secretariat`, y no la crea ni la da de baja nadie. Trae dos **permisos de plataforma**, `edit_procedure_types` y `publish_procedure_types`, que habilitan proponer y aprobar versiones de los tipos de trámite (ADR-013). Solo existen en la Secretaría, no se dan de baja y ningún paso los puede exigir. En todo lo demás es un área como cualquier otra: el administrador le suma permisos para sus pasos y roles, y otorga membresías en ella.
- **Permisos.** Cada área tiene su propio grupo de permisos, que el administrador crea a gusto como datos de SIGTA (por ejemplo, "Verificar documentación", "Registrar expediente" o "Aprobar dictamen"). Un permiso pertenece a un área y se identifica dentro de ella. Ninguno cubre a otro, y el nombre de un permiso no le da ningún significado para el sistema. La única excepción son los dos permisos de plataforma de la Secretaría.
- **Roles.** Un rol agrupa permisos de su área, para no asignarlos uno por uno. Lo crea el administrador como dato, con un identificador que no se reusa y al menos un permiso del área. La membresía referencia el rol: cambiar los permisos de un rol cambia en ese momento lo que puede hacer cada persona que lo tiene. Cada cambio queda como una revisión con quién y cuándo, así que se puede reconstruir qué permisos tenía alguien en una fecha. Los pasos exigen permisos, nunca roles.
- **Membresía.** Une a una persona con un área y le asigna algunos de los roles de esa área, algunos permisos sueltos, o ambos. Los permisos que da son la unión de los sueltos y los de sus roles. Opcionalmente, la membresía tiene un **alcance**, una lista de tipos de trámite que restringe dónde ve y actúa la persona dentro del área. Sin alcance, vale para todos los tipos, incluidos los que se publiquen después. El alcance referencia el tipo y no una versión.
- **Varias áreas.** Una persona puede tener membresías en varias áreas. En una misma área tiene una sola membresía vigente por alcance; si necesita permisos distintos según el tipo de trámite (por ejemplo, aprobar solo equivalencias), tiene dos membresías con alcances distintos, y en un trámite valen los permisos de todas las que lo cubren.
- **Pasos.** Cada paso de un tipo de trámite es del titular o del personal. Un paso del personal declara su área y los permisos de esa área que exige para hacerlo, al menos uno. Al publicar se valida que el área y los permisos existan; los permisos de plataforma no cuentan como permisos de un paso.
- **Vigencia.** La membresía tiene una fecha desde y una fecha hasta opcional, y registra quién la otorgó. Sin fecha hasta, vale hasta que se la da de baja. También se puede otorgar con una fecha hasta, por ejemplo para una suplencia. Cambiar los permisos sueltos o los roles de una membresía la cierra y abre otra. Una baja cierra la vigencia sin borrar nada. Las acciones ya hechas quedan en el historial del trámite con su autor.
- **Actuar en un paso.** Los pasos del titular los hace solo el titular de ese trámite. En un paso del personal se puede actuar si la persona tiene membresías vigentes en el área del paso que cubren el tipo de trámite y que, juntas, le dan todos los permisos que el paso exige, **salvo que sea el titular del trámite**: nadie avanza su propio trámite como personal, en ningún paso ni en ningún área. Se controla en el momento de cada acción: el trámite no se asigna a personas.
- **Ver trámites.** Cada persona ve los trámites que inició. El personal ve además los que pasan por sus áreas, recortados por el alcance. A quien no puede ver un trámite se le responde como si no existiera.
- **Publicar tipos de trámite.** Quien propone una versión necesita `edit_procedure_types` y quien la aprueba, `publish_procedure_types`, en membresías vigentes de la Secretaría que cubran ese tipo de trámite. Un tipo nuevo no está en ningún alcance, así que crearlo exige una membresía sin alcance. Tener permisos con los mismos identificadores en otra área no alcanza.
- **Administrador.** Es una marca sobre la persona y no un tipo de usuario ni una membresía. Da de alta las áreas, sus permisos, sus roles y las membresías. No le da acceso a los trámites: para actuar sobre ellos necesita una membresía.
- **Registro.** Cada cambio de membresía y de rol queda registrado con quién, qué y cuándo. La autorización se aplica en el caso de uso del servidor. Registrar los intentos sin permiso no es parte de esta decisión: queda para una historia de auditoría, junto con los cambios de áreas y membresías y los ingresos.

**Consecuencias.**
- Queda resuelto si en un mismo trámite intervienen oficinas distintas con personas que no deben hacer los pasos de la otra: el área forma parte del permiso.
- S-07 sigue sin condicionar el modelo: una persona puede tener cualquier combinación de permisos en cualquier área.
- El personal puede iniciar trámites con su cuenta. Si la única persona con el permiso de un paso es la titular del trámite, ese paso queda sin quien lo haga hasta que se otorgue otra membresía.
- Desaparecen los perfiles fijos `staff_operative` y `staff_approver`, y la regla de que uno cubre al otro. Las versiones publicadas antes se siguen leyendo: el perfil de cada paso se lee como un permiso con el mismo identificador en el área del paso, que tiene que existir para que alguien pueda hacerlo.
- Sumar un área, un permiso o un rol no requiere código. Exigir un permiso nuevo en un paso requiere publicar una versión del tipo de trámite. Cambiar los permisos de un rol, no.
- Cambiar el personal no genera versiones nuevas: la configuración define qué función hace cada paso, y la membresía define quién la cumple.
- Las áreas, sus permisos y roles son entidades globales del módulo `auth`, junto con las membresías y la marca de administrador, porque describen a las personas y a la organización y no a un trámite. Cada versión de un tipo de trámite sigue guardando una copia de las áreas y los permisos de pasos que usa (`content.Areas`), para leerse igual si cambian de nombre. Desaparece el tipo de área `external`.
- La Secretaría se crea al arrancar el servidor si no existe (`EnsureSecretariat`), y cuando haya PostgreSQL pasa a ser una migración con los mismos datos: es la 0007 (ADR-027). El identificador `secretariat` y sus dos permisos de plataforma están reservados. `auth` los expone en su API pública, para que `procedure` los pida al publicar.
- Que la Secretaría edite los tipos de trámite resuelve en parte D-08: quien no tenga membresía en la Secretaría no propone ni aprueba versiones, y el equipo de proyecto necesita una para hacerlo. Queda abierto a quién se le otorgan esos permisos.
- La derivación deja de definirse como "pasar a un paso de un área `external`" y pasa a ser un paso de otra área.
- Las reglas que dependen del trámite (qué ve y en qué pasos actúa cada persona, y quién publica) quedan en `procedure`, que le pregunta a `auth` por su API pública (ADR-021). `auth` no depende de `procedure`: por eso, al otorgar una membresía no se comprueba que los tipos de su alcance existan, y un tipo inexistente simplemente no coincide con ningún trámite.
- Las versiones ya publicadas con áreas `external` se siguen leyendo, y esas áreas se tratan como del personal. Con ADR-021, las áreas pierden el tipo: el titular deja de ser un área y cada paso declara su rol.
- `/me` informa si la persona es personal (`staff`) y si es administradora. `SIGTA_USERS_FILE` declara `administrator` y `memberships`, cada una con sus `roles` y `permissions`. Las membresías de cada persona se suman a `/me` cuando una pantalla las necesite.
- Un área, un permiso o un rol se da de baja sin borrarse, como una membresía, y su identificador no se reusa. Un permiso no se puede dar de baja mientras una membresía vigente o un rol lo tenga, y un paso que lo exige queda sin quién lo haga hasta publicar una versión que lo reemplace. Un rol no se puede dar de baja mientras una membresía vigente lo tenga. Un área deja de listarse y de aceptar membresías, sigue nombrada en el historial y no se puede dar de baja mientras tenga membresías sin terminar. `auth` no sabe qué tipos de trámite usan un área, así que dar de baja una que todavía usan deja sus pasos sin personal hasta publicar una versión que la reemplace.
- La API permite crear áreas (`POST /areas`), sus permisos y sus roles (`POST /areas/{areaId}/roles`), cambiar los permisos de un rol (`PUT /areas/{areaId}/roles/{roleId}/permissions`) y otorgar membresías (`POST /memberships`), solo a administradores. Con SIG-76 (US74) se suman la lista de áreas con sus permisos y roles (`GET /areas`), la baja de un área, un permiso o un rol (`DELETE` sobre cada uno), las membresías de un área (`GET /areas/{areaId}/memberships`), el fin de una membresía (`DELETE /memberships/{membershipId}`) y el cambio de sus permisos (`PUT /memberships/{membershipId}/permissions`, que la termina y abre otra con el mismo alcance), todo con la pantalla de administración. Cada rechazo dice por qué. Un área y un permiso guardan quién los creó y cuándo; una membresía, quién la otorgó, cuándo y quién la terminó.

---

## ADR-021: Un módulo llama a otro por su API pública, sin puertos entre módulos

**Estado:** propuesta, a acordar con el equipo. Reemplaza la regla de ADR-014 sobre cómo un módulo usa a otro.

**Contexto.** ADR-014 pide que un módulo use a otro a través de un puerto. Con ADR-020, `procedure` necesita de `auth` las áreas y las membresías, y eso llevó a definir los puertos `AreaCatalog` y `StaffDirectory` en `procedure`, dos adaptadores que los implementan (uno sobre `auth`, otro en memoria para los tests) y copias en `procedure` de conceptos de `auth`: el perfil, la membresía (`StaffMembership`) y la regla de que un perfil cubre a otro (`Covers`). Cada concepto tiene dos definiciones que hay que mantener iguales, y la decisión de si una persona alcanza para un paso queda en `procedure` aunque los perfiles, la vigencia y el alcance son de `auth`. Los puertos sirven para reemplazar una dependencia o simularla en las pruebas. Un módulo del mismo binario no se reemplaza, y simularlo no ahorra nada: sus adaptadores en memoria ya son baratos.

**Decisión.**
- Cada módulo expone una **API pública** en `ports/input/`: las interfaces de sus casos de uso y los tipos que reciben y devuelven.
- Otro módulo importa solo esa API y la recibe por constructor, sin definir un puerto propio ni un adaptador. Nunca importa `domain/`, `application/` ni `adapters/` de otro módulo.
- Los tipos de la API pública se usan tal cual en el módulo que la llama; no se copian ni se traducen.
- Cada regla vive en el módulo dueño de sus datos. `auth` responde si una persona tiene ciertos permisos en un área para un tipo de trámite (vigencia, alcance, roles y permisos sueltos) y qué áreas y alcances alcanza. `procedure` decide qué exige cada paso, quién es el titular de un trámite, qué trámites se muestran y quién publica un tipo de trámite, y le pregunta a `auth` lo que es de `auth`.
- Las dependencias entre módulos van en un solo sentido: `procedure` usa `auth`, nunca al revés.
- Los puertos de salida (`ports/output/`) quedan para lo que está fuera del módulo y puede cambiar o hay que simular: la base de datos, el proveedor de identidad, el correo y los modelos de IA.
- Las pruebas de un módulo usan el otro módulo real, armado con sus adaptadores en memoria.

**Consecuencias.**
- Desaparecen `procedure/adapters/auth/`, los puertos `AreaCatalog` y `StaffDirectory`, el catálogo de áreas en memoria de `procedure`, `Actor`, `StaffMembership` y `Profile.Covers`. Cada área y membresía se define una sola vez, en `auth`, que expone `StaffAuthorizer` (`Holds`, `Scopes`) y los identificadores de la Secretaría y sus permisos de plataforma.
- Cada versión de un tipo de trámite sigue guardando su copia de las áreas que usa (ADR-012, ADR-020). No es una traducción entre módulos: es lo que permite leer una versión publicada igual que el día en que se publicó.
- Los permisos que exige un paso forman parte del contenido versionado de `procedure`, que guarda una copia de su nombre, y se validan al publicar contra los de `auth`. Para autorizar, `procedure` le pasa a `auth` los identificadores y `auth` decide.
- Las áreas son solo del personal: cada paso dice su rol (`applicant` o `staff`) y solo los del personal tienen área. Desaparecen el área del titular y el identificador que `auth` reservaba para ella. El contenido de las versiones pasa al formato 3, y los formatos 1 y 2, que modelaban al titular como un área de tipo `applicant`, se convierten al leerlos.
- Un cambio en la API pública de `auth` obliga a tocar a quien la usa. Hoy el cambio quedaba en el adaptador.
- Los ciclos entre módulos los rechaza el compilador de Go. Que nadie importe el interior de otro módulo lo controla una regla de `depguard` en `golangci-lint`.
- Si un módulo se separa algún día en un servicio aparte, el puerto vuelve a hacer falta, con un adaptador sobre la red. Hasta entonces no se agrega.
- ADR-020 decía que `procedure` lee a `auth` a través de sus puertos; con esta decisión lo hace por la API pública de `auth`. El README y `AGENTS.md` se actualizan junto con la implementación.

---

## ADR-022: Las tablas del servicio de trámites son de un rol de migraciones, y las migraciones van dentro del binario

**Estado:** propuesta, a acordar con el equipo

**Contexto.** La persistencia de los tipos de trámite crea las primeras tablas del servicio de trámites. Una versión publicada no se tiene que poder modificar (ADR-012), y lo mismo vale para un evento de un trámite. Pero `deploy/postgres/sql/roles.sql` hacía al rol de la aplicación dueño de su esquema, y el dueño de una tabla siempre puede modificarla o borrarla, tenga o no permisos sobre ella. Además hacía falta una forma de crear y actualizar las tablas que funcione igual en Docker Compose y en la infraestructura de destino, donde la facultad ofrece PostgreSQL 11.9 y pide migraciones versionadas en el repositorio.

**Decisión.**
- El esquema `backend` y todas sus tablas son del rol `backend_migrations`. El rol de la aplicación, `backend`, solo usa el esquema y tiene los permisos que le otorga cada migración, sin `ALTER DEFAULT PRIVILEGES`: hoy solo `SELECT` e `INSERT`, y sobre ninguna tabla `UPDATE`, `DELETE` ni `TRUNCATE`. Si un cambio necesita actualizar una tabla, su migración otorga `UPDATE` solo sobre las columnas que cambian, nunca sobre `procedure_type_versions`, `procedure_events` ni la versión de un trámite. El servidor no arranca si su rol, el usuario con el que se conecta o un rol que cualquiera de los dos puede asumir puede modificar o borrar las versiones, los eventos o la versión de un trámite, por ejemplo porque es el dueño de las tablas o un superusuario: `SET ROLE` no le quita los permisos al usuario, que los recupera con `RESET ROLE`. La lista de esas tablas y ese control están en el paquete de las migraciones, junto al esquema que los promete.
- Las migraciones son archivos `.sql` numerados en `services/backend/src/migrations/`, embebidos en el binario, con un ejecutor propio. Corren en orden, en una sola transacción y con un bloqueo consultivo, así dos ejecuciones simultáneas no las aplican dos veces. La tabla `schema_migrations` guarda el nombre y el hash de cada una.
- Una migración aplicada no se edita: un cambio es un archivo nuevo. No hay migraciones hacia atrás.
- Las aplica el comando `migrate` del mismo binario, con `SIGTA_MIGRATIONS_DATABASE_URL`, antes de arrancar el servidor; en Docker Compose, un servicio que corre una vez antes del backend. El servidor se conecta como `backend` y solo verifica que no falte ninguna migración, que ninguna aplicada haya cambiado y que no haya aplicadas que el binario no conoce; si no, no arranca.
- Las migraciones solo usan lo que existe en PostgreSQL 11.

**Alternativas descartadas.**
- Una herramienta de migraciones (goose, golang-migrate, Atlas) suma una dependencia y su propio formato para lo que hoy son unas decenas de líneas, y no todas detectan que se editó una migración ya aplicada.
- Migrar al arrancar el servidor: el proceso que atiende pedidos tendría la credencial del dueño de las tablas.
- Dejar al rol de la aplicación como dueño y confiar en el código: un error o una credencial filtrada alcanzan para alterar una versión.

**Consecuencias.**
- Cambiar una versión o un evento exige la credencial del rol de migraciones. Eso avanza la decisión abierta "Integridad del historial" sin cerrarla: el dueño de las tablas todavía puede alterarlas. La base rechaza una versión cuyo hash no es el de su contenido, pero el dueño puede quitar esa restricción, y el hash solo detecta un cambio que no lo recalculó.
- El despliegue tiene que correr `migrate` antes de cada versión nueva del servidor, con una credencial que el servidor no recibe.
- `roles.sql` corre solo al crear el volumen de la base, así que una base de desarrollo creada antes de este cambio se recrea con `docker compose down --volumes`.
- Un error en una migración aplicada se corrige con otra migración.
- Cada tabla llega con el adaptador que la usa, así sus permisos se prueban con el código que la lee: las de borradores, personas, áreas y trámites llegaron después, cada una en su migración.
- Que los números de versión de un tipo de trámite no tengan huecos lo garantiza el adaptador, que agrega una versión solo si la anterior es la última: la base no tiene una restricción para eso, así que un `INSERT` directo con el rol de la aplicación puede dejar un hueco.
- La restricción `approver_is_not_proposer` compara con `lower()`, que depende de la configuración regional de la base: con `en_US.UTF-8`, `lower('İnés')` es `inés`, mientras que el servicio, que compara en Go, los toma como distintos. Si la base rechaza una versión que el servicio aceptó, el adaptador lo informa como una autoría inválida. La configuración regional del servidor de la Subsecretaría de TICs está a confirmar.
- Las pruebas de los adaptadores de PostgreSQL toman la base de `SIGTA_TEST_DATABASE_URL` y crean un esquema propio por prueba, con los permisos de `backend` y de `backend_migrations`. Sin la variable se saltean, salvo en la integración continua, donde fallan.

---

## ADR-023: Confirmar un paso lo avanza, cambiar de área lo deriva y volver al área de la que salió es el retorno

**Estado:** propuesta, a acordar con el equipo

**Contexto.** ADR-020 dejó de definir la derivación como pasar a un paso de un área `external`: ahora es pasar a un paso de otra área. Faltaba decir cuándo vuelve el trámite, desde cuándo se cuentan los días sin retorno y cómo se confirma un paso del personal, que es lo que hace avanzar al trámite (US10). Sin una regla, la vuelta de la comisión a la Secretaría también sería "pasar a otra área", y contaría como una derivación nueva.

**Decisión.**
- **Confirmar un paso.** Lo confirma quien tiene, en el área del paso, los permisos que el paso exige (ADR-020), y solo el paso abierto. Queda quién, cuándo y una observación, que es obligatoria. Quién y cuándo los pone el servidor, nunca el pedido. El trámite pasa al paso siguiente: el configurado o, en una decisión, el de la salida elegida; un paso sin siguiente cierra el trámite. En un paso `confirm_external_fact` queda registrada la declaración de quien confirma, no el hecho (ADR-004).
- **Los pasos del titular no se confirman.** Los completa el titular al cargar sus datos o documentos (US14 y US15). Al iniciar el trámite, cada paso inicial queda en el historial como hecho por el titular, sin observación, junto con el trámite.
- **Derivación.** Pasar de un paso del personal a uno de otra área deriva el trámite. Se registra de qué área sale, a cuál va, quién derivó, cuándo y, si lo hay, el número de expediente externo. Un número de expediente en una confirmación que no deriva se rechaza, para que nadie crea que quedó guardado.
- **Retorno.** Volver a un paso del área de la que salió la derivación es el retorno: la derivación termina. Si estando derivado pasa a una tercera área, se registra esa derivación, pero el trámite sigue fuera del área de la que salió primero, y los días sin retorno se cuentan desde entonces. Los pasos del titular no tienen área: ni derivan ni devuelven. Un trámite que se cierra estando derivado deja de estarlo: la derivación termina sin retorno, y el detalle ya no cuenta días sin retorno.
- **Estado.** El estado del trámite suma `referred`. Se decide en este orden: sin paso abierto, `closed`; paso abierto del titular, `with_applicant`, aunque haya una derivación; derivación abierta, `referred`; si no, `in_office`.
- **Días sin retorno.** Se cuentan por día de calendario desde que salió del área.
- **Atomicidad.** El trámite y su evento se guardan como un solo cambio, y solo si el trámite sigue en el paso confirmado. Si otra persona lo confirmó antes, la confirmación se rechaza sin cambiar nada.

**Alternativas descartadas.**
- Un área responsable por tipo de trámite, de la que todo paso de otra área sería una derivación: es más precisa, pero suma un campo a la configuración de cada tipo de trámite y una versión nueva del formato.
- Marcar en la configuración los pasos que son derivaciones: depende de que quien configura se acuerde, y no dice cuándo vuelve el trámite.
- Contar como derivación todo cambio de área, sin retorno: la vuelta de la comisión a la Secretaría sería otra derivación, y el trámite nunca dejaría de estar derivado.

**Consecuencias.**
- La regla no exige cambios en la configuración de los tipos de trámite ni en las versiones publicadas.
- El glosario llama pase al avance a un paso de otra área: con esta regla, todo pase es una derivación, también uno interno, por ejemplo de la Secretaría a la Dirección de Alumnos. Si la Secretaría los distingue, hace falta una de las alternativas descartadas (D-13).
- Registrar el dictamen que trae el retorno queda para "Registrar retorno de la evaluación externa" (US21): acá el retorno solo termina la derivación.
- Los trámites y sus eventos se guardan en PostgreSQL: la migración otorga `UPDATE` solo sobre las columnas del paso abierto, de la derivación y de la fecha del último movimiento (ADR-022).
- La bandeja del personal tiene que calcular el estado con `domain.StateOf`, para que `referred` signifique lo mismo en el detalle y en la bandeja.

## ADR-024: Un trámite y sus datos y documentos se registran en una transacción, y un storage aparte se coordina con outbox y saga

**Estado:** propuesta, a acordar con el equipo

**Contexto.** Iniciar un trámite (`SubmissionService.Submit`) escribe en tres lugares sin una transacción que los abarque: los archivos en disco (`output.Files`), el registro de cada documento, con un `Add` por documento, y el trámite con los valores de sus campos, que es lo único que se registra junto. Un documento es la fila que describe lo que se adjuntó (trámite, requisito, paso, nombre, tipo, tamaño y el hash de su archivo); el archivo son los bytes, guardados aparte con el hash como nombre. Si el inicio falla después de registrar algún documento y antes de registrar el trámite, queda un documento de un trámite que no existe, que nadie ve y nada borra: `RemoveUnreferencedFiles` solo borra los archivos que ningún documento nombra. Hoy no se nota porque los registros están en memoria, pero pasa a importar cuando se guarden en PostgreSQL (SIG-75). Además, el disparador de "Almacenamiento compatible con S3" (README, "Evolución") puede llevar los archivos a un servicio aparte, que tampoco entra en una transacción de la base.

**Decisión.**
- **Una transacción.** Iniciar un trámite registra en una sola transacción el trámite, los valores de sus campos y el registro de cada documento. Adjuntar un documento registra el documento y su evento del historial en otra.
  - El mecanismo, elegido en SIG-75: un método del repositorio por operación, en `ports/output`. `ProcedureRepository.Add` recibe el inicio completo (trámite, eventos, valores y documentos) y `Attach`, el documento con su evento; el repositorio de documentos solo lee. Se prefirió a un puerto de unidad de trabajo porque hoy son dos operaciones y las dos son del mismo agregado. La aplicación no abre transacciones con `pgx`, y el adaptador en memoria da la misma garantía, probada con la suite de contrato de `ports/output/outputtest`.
  - Los archivos se escriben antes de la transacción. Con el disco local, lo único que puede sobrar es un archivo sin documento, que `RemoveUnreferencedFiles` borra al arrancar.
- **Con un storage aparte, outbox y saga, subiendo antes de confirmar.**
  1. Reserva: una transacción corta registra cada archivo a subir, con su clave en el storage y el estado `pending`. Esa tabla es el log de la saga y la outbox.
  2. Subida del archivo con esa clave. La clave es propia de la reserva y no solo el hash, así compensar no borra un archivo que nombra el documento de otro trámite (D-09).
  3. Confirmación: la transacción del trámite registra trámite, valores y documentos, y cierra las reservas.
  4. Compensación: si el envío se rechaza o algo falla antes de confirmar, se borran del storage los archivos de sus reservas y las reservas quedan `compensated`. El titular recibe el error en el momento.
  5. Recuperación: si el servidor se cae a mitad de camino, un proceso que corre al arrancar y cada cierto tiempo toma las reservas `pending` más viejas que un plazo con `FOR UPDATE SKIP LOCKED`, para que dos instancias no compensen la misma, y las compensa.
  6. Si una compensación falla, la reserva sigue abierta y se reintenta. Es el único caso en que algo queda pendiente, y queda registrado.

  Así nunca hay un trámite con un documento sin archivo ni un archivo sin reserva. No hace falta la cola de trabajos; cuando exista, la recuperación puede pasar a ser uno de sus trabajos.

**Alternativas descartadas.**
- Confirmar y después subir, con los documentos "pendientes de subir" desde una copia local: el inicio no depende del storage, pero el trámite existe un tiempo con documentos sin archivo, y si la subida falla para siempre hay que deshacer un trámite ya iniciado.
- Recolectar por referencia también en el storage aparte: acepta huérfanos, recorre todo el storage y, con varias instancias, puede borrar un archivo que otra todavía no confirmó.
- Transacción distribuida (2PC, `PREPARE TRANSACTION`): necesita un coordinador, bloquea si el coordinador cae, requiere `max_prepared_transactions`, que la infraestructura de destino no garantiza, y un storage de objetos no participa.
- Guardar los archivos en PostgreSQL: entrarían en la transacción, pero agrandan la base y su respaldo, y contradicen el puerto `Files`.

**Consecuencias.**
- SIG-75 cambia los puertos de salida para que el inicio de un trámite se registre junto, y adjuntar registre documento y evento juntos.
- La saga se implementa con el adaptador del storage aparte, cuando se cumpla su disparador: suma la tabla de reservas, con su migración y sus permisos según ADR-022, y la recuperación, y para ese storage retira `RemoveUnreferencedFiles`.
- Con el storage aparte, cada envío con archivos suma una transacción corta, la de la reserva.

---

## Decisiones pendientes

| ID | Decisión pendiente | Necesaria antes de | Responsable |
|---|---|---|---|
| D-04 | Infraestructura de despliegue y su financiamiento | Fin de la fase de diseño | Equipo |
| D-08 | A quién se le otorgan los permisos de la Secretaría para proponer y aprobar versiones de los tipos de trámite, y si el equipo de proyecto tiene una membresía en ella (ADR-020) | La interfaz de edición de tipos de trámite | Equipo y canal con la Secretaría |
| D-09 | Cuánto tiempo se conservan los documentos de un trámite cerrado. Mientras tanto, tres reglas provisorias tocan la retención: un trámite guarda hasta 20 archivos por documento, el servidor borra al arrancar los archivos que ningún documento nombra, y un mismo contenido se guarda una sola vez aunque lo adjunten trámites de titulares distintos, así que suprimir el de uno no puede borrar el archivo si otro documento lo nombra (CONTRIBUTING, sección Documentos) | La estimación de almacenamiento para TICs | Secretaría |
| D-10 | Qué ve el personal: todos los trámites que pasan por su área o solo los que tienen abierto un paso suyo (ADR-020) | La bandeja del personal | Canal con la Secretaría |
| D-11 | Si hay un administrador global o un responsable por área que da de alta a su propio personal (ADR-020) | La gestión de membresías | Canal con la Secretaría |
| D-12 | Si el alcance por tipo de trámite se usa o alcanza con áreas más finas. Agrupar permisos ya se resolvió con roles (ADR-020) | La gestión de membresías | Canal con la Secretaría |
| D-13 | Si un pase entre dos áreas del personal es siempre una derivación, o si hay pases internos que no cuentan días sin retorno (ADR-023) | La configuración del segundo trámite | Canal con la Secretaría |

---

## ADR-025: Una devolución es una salida de una decisión marcada como tal, y lleva a un paso anterior

**Estado:** propuesta, a acordar con el equipo

**Contexto.** US12 pide devolver un trámite a un paso anterior con un motivo, sin cerrarlo. Devolver ya era posible como cualquier otra salida de una decisión, pero el motor no la distinguía: nada exigía que el destino fuera un paso anterior, el historial no la diferenciaba de una aprobación y el titular no tenía cómo saber que lo involucraba.

**Decisión.**
- **Una marca en la salida.** Una salida de una decisión puede declararse devolución (`send_back`). Sigue siendo una salida, no una acción ni un estado (ADR-009): elige el paso al que lleva y el trámite sigue abierto.
- **El destino es un paso anterior.** Al validar la definición, una devolución tiene que llevar a un paso al que se llega desde el primer paso sin pasar por el que decide, y desde el que se vuelve a llegar a este. Así un ciclo no hace pasar por anterior a un paso que viene después. Una que va hacia adelante se rechaza (`send_back_forward`).
- **El motivo es la observación.** Confirmar una devolución exige la observación, como toda confirmación, y el evento queda marcado como devolución, con quién, cuándo y el motivo.
- **El titular.** Una devolución involucra al titular cuando el paso al que vuelve es suyo: lo ve en el historial y en un aviso mientras el trámite está de vuelta en ese paso.
- **Formato 4.** La marca sube el contenido de las versiones al formato 4. El formato 3 se sigue leyendo como contenido actual, sin devoluciones, así que ninguna versión publicada cambia: una salida que se publicó sin la marca sigue siendo una salida común hasta que se publique una versión nueva.
- **Quién decide.** La definición no valida que decida la Secretaría: lo determinan los permisos del paso, como en cualquier decisión.

**Alternativas descartadas.**
- Una acción `send_back` aparte: el modelo de acciones se acordó con cuatro, y devolver es una salida posible de una decisión.
- Tratar como devolución toda salida que va a un paso anterior, sin marca: un camino hacia atrás puede ser otra cosa, y la intención quedaría implícita en el grafo.
- Un campo propio para el motivo: la observación ya es obligatoria y queda en el historial.

**Consecuencias.**
- El editor de tipos de trámite (US72) tiene que ofrecer la marca en cada salida.
- La API suma `send_back` a las salidas de un paso y a los eventos del historial.

---

## ADR-026: Una subsanación es un pedido sobre el paso abierto, que no lo mueve y lo deja con el titular

**Estado:** propuesta, a acordar con el equipo

**Contexto.** US33 pide que la Secretaría le pida al titular algo puntual sin cerrar ni reiniciar el trámite, que el trámite quede en espera del titular, que se pueda pedir más de una vez y que canalice los pedidos de ampliación del evaluador externo (ADR-009). La devolución (ADR-025) no alcanza: vuelve a un paso anterior que fija la configuración, y solo desde una decisión que la prevé. En la equivalencia, además, no hay paso de revisión: si la DA ve un problema, pide la subsanación desde el paso en el que esté el trámite (relevamiento, 3.1).

**Decisión.**
- **Un pedido sobre el paso abierto.** Quien puede hacer el paso abierto del personal puede pedirle una subsanación al titular, cualquiera sea la acción del paso. No es una acción ni una salida del tipo de trámite: no se configura y vale para todos los tipos.
- **Qué falta y por qué.** El pedido nombra datos y documentos que piden los pasos del titular, al menos uno, y lleva el motivo, que es obligatorio. Así el titular completa lo pedido sobre el mismo trámite (US34) con requisitos que la plataforma ya sabe validar, y la respuesta queda en uno de sus pasos, donde la ve.
- **El trámite no se mueve.** El paso abierto sigue siendo el mismo, y el trámite guarda los pedidos pendientes. Mientras haya alguno, el estado es `with_applicant`, aunque esté derivado, el paso no se confirma y deja de requerir acción en la bandeja. Cuando el titular responde (US34), sigue en el mismo paso.
- **Más de una.** Se pueden pedir varias, también con otra pendiente.
- **Historial.** Cada pedido es un evento sobre el paso abierto, con quién, cuándo, el motivo en la observación y lo que pide.
- **Evaluador externo.** Su pedido de ampliación lo traslada quien registra lo que dice el evaluador, sobre el paso de esa área. No hace falta un caso aparte.
- **Sin cambio de formato.** La definición de los tipos de trámite no cambia, así que el contenido de las versiones sigue en el mismo formato.

**Alternativas descartadas.**
- Modelarla en la configuración, como una salida hacia un paso del titular que vuelve al que decide: lo que se pide quedaría fijo de antemano y solo se podría pedir desde una decisión que lo prevea. Cada tipo nuevo tendría que acordarse de agregarla.
- Un texto libre en lugar de requisitos: el titular no tendría dónde cargar la respuesta, ni la plataforma qué validar.
- Pedir cualquier dato o documento del tipo de trámite: lo que solo pide un paso del personal, como el acta de una comisión, quedaría en un paso que el titular no ve, y no vería lo que respondió.
- Una sola subsanación pendiente por vez: la Secretaría no podría sumar algo que notó después sin esperar la respuesta.

**Consecuencias.**
- El estado del trámite se deriva del paso abierto y de los pedidos pendientes (`domain.StateOf`).
- La API suma `POST /procedures/{procedureId}/steps/{stepId}/correction-requests`, los pedidos pendientes en el detalle (`correction_requests`) y lo que pide cada uno en los eventos del historial (`correction_request`).
- Cómo responde el titular (US34, SIG-50): carga de nuevo, de a uno, lo que piden las pendientes, cada cosa en el paso suyo que la pide y en el historial, y después envía la respuesta, que exige todo lo pedido y responde todas las pendientes juntas. Lo cargado cuenta si viene después, en el historial, del último pedido que lo nombra. La API suma `POST /procedures/{procedureId}/values`, `POST /procedures/{procedureId}/correction-answer` y, en los eventos, `value` y `correction_answer`; "Mis trámites" dice si hay una pendiente (`correction_pending`).
- Si un pedido necesita algo que los pasos del titular no piden, se corrige el tipo (ADR-006).

---

## ADR-027: Las personas, las membresías y los administradores viven solo en la base, sin archivo de usuarios ni datos de ejemplo, y la Secretaría viaja con el código

**Estado:** propuesta, a acordar con el equipo
**Modifica:** ADR-020, en la carga de las membresías iniciales, en dónde se guarda la marca de administrador y en cómo se crea la Secretaría; ADR-010, en la identidad del entorno local; ADR-013, en quién tiene los permisos de publicación; ADR-005, en las cuentas que guarda la base

**Contexto.** Mientras no había administración de membresías, `SIGTA_USERS_FILE` (PR #29) cargaba al arrancar las personas conocidas, sus membresías y la marca de administrador desde un JSON guardado con los secretos del despliegue. Con las personas y las membresías en PostgreSQL (SIG-75), el archivo pasó a competir con la base: en cada arranque volvía a dar las membresías que listaba, incluso las que se hubieran dado de baja, y daba de baja las que dejaba de listar. Cuando un administrador pueda dar de baja una membresía desde SIGTA, con su registro de quién y cuándo, esa decisión podría deshacerse al reiniciar con un cambio en un archivo que no deja ese registro. Los entornos de prueba nunca lo usaron: corrían con `SIGTA_ENV=development`, que en cada arranque cargaba usuarios sintéticos con sus membresías, dos áreas de prueba, tres tipos de trámite de ejemplo y trámites inventados para la bandeja, y en local ofrecía un adaptador de identidad que dejaba entrar como cualquier correo. Esos usuarios eran otra lista fija de personas y membresías, escrita en el código, con el mismo problema que el archivo.

**Decisión.**
- Se saca `SIGTA_USERS_FILE`. El servidor ya no lo lee, y en producción no lo pide.
- Se sacan los datos de ejemplo: los usuarios sintéticos, las áreas de prueba, los tipos de trámite de ejemplo y los trámites inventados. Nadie existe hasta que ingresa con Google, y cada prueba crea sus propios datos. Al arrancar, el servidor no agrega nada.
- Se ingresa con Google en todos los entornos, también en local, con un cliente que tenga registrada la dirección local. El adaptador de identidad falso queda solo para las pruebas, y se saca `SIGTA_IDENTITY_PROVIDER`.
- Los entornos desplegados corren con `SIGTA_ENV=production`. `development` solo pone valores por defecto para arrancar en local: las direcciones de `localhost`, el dominio y, sin base, todo en memoria.
- Las personas se registran al ingresar con su cuenta de Google, como ya pasaba con quien no figuraba en el archivo.
- Las membresías viven solo en la base y se cambian solo desde SIGTA, con la pantalla de administración (SIG-76).
- La marca de administrador se guarda en la base (SIG-83), y el servidor solo la lee. Cada administrador es un período, con su origen, quién lo dio y desde cuándo; una baja lo cierra con quién y cuándo, sin borrar la fila, y una persona tiene a lo sumo un período abierto.
- El primer administrador de un entorno lo da el comando de consola `backend grant-admin <correo>`, que corre aparte como `migrate`, con el rol dueño de las tablas, a una persona ya registrada, y solo si no hay ningún administrador vigente. Ese período es el alta inicial del sistema (`bootstrap`), el único tipo sin quién lo dio, y la base lo exige: el comando no pide un autor porque un dato que tipea quien lo corre no prueba nada. A quién se le dio y cuándo queda en la tabla de administradores. Quién corrió el comando no lo guarda SIGTA: queda a cargo del entorno donde se despliegue, en su registro de quién opera la consola, sin atarlo a un proveedor, a confirmar con la Subsecretaría de TICs (DU-39).
- Los demás administradores y las bajas se dan desde SIGTA (SIG-76: `GET` y `POST /administrators`, `DELETE /administrators/{administratorId}`), y quién dio o cerró un período sale siempre del usuario de la sesión, autenticado con Google, nunca de un dato del pedido, como ya pasa con las membresías. Solo se le da a alguien que ya ingresó, y no se cierra el último período vigente: SIGTA nunca queda sin administradores. El rol de la aplicación puede agregar períodos dados por un administrador y cerrar uno abierto, nunca dar el alta inicial ni reabrir un período (`0009_administration.sql`, con seguridad por filas).
- Para otorgar, el administrador elige a la persona por parte de su nombre o de su correo, sin distinguir mayúsculas ni acentos (`GET /people?q=`, desde dos letras y hasta diez personas, ordenadas por nombre), solo entre quienes ya ingresaron, y solo un administrador puede buscar. El servidor rechaza una membresía o una administración para alguien que nunca ingresó, con el motivo. Es trazabilidad por autenticación, no no repudio en sentido estricto: el ingreso con Google es a lo sumo firma electrónica, y la presunción de autoría es solo de la firma digital (Ley 25.506, arts. 5 y 7).
- Lo que el código necesita viaja con el código; la configuración de la organización es del administrador y se edita desde SIGTA; las personas nunca van al repositorio. Es el criterio que siguen, por ejemplo, Django, Odoo, Rails, GitLab y Keycloak (SIG-83).
- Por eso la Secretaría y sus dos permisos de plataforma, que el backend chequea por su identificador, los crea una migración versionada (`0007_secretariat.sql`), y el servidor solo controla al arrancar lo que el código necesita de ellos: si falta la Secretaría, si está dada de baja, o si ella o alguno de los dos permisos perdió la marca de sistema o se dio de baja, no arranca, en vez de taparlo. El nombre del área y de los permisos no se controla. En memoria, donde no hay migraciones, el servidor los agrega. El administrador no es un rol sino una marca con su tabla (`0006_administrators.sql`), y el primero se da con `backend grant-admin`, como alta inicial sin autor.
- Ninguna otra área, rol, permiso ni membresía va en una migración ni en la configuración del despliegue. Las áreas del trámite ancla, sus roles y permisos, y el personal de cada una se cargan desde la pantalla de SIG-76; si cambian, se editan ahí, sin código.
- `POST /api/areas/{id}/permissions` se mantiene. El backend chequea por nombre fijo solo los dos permisos de plataforma, y ese endpoint no los puede crear: rechaza sus identificadores en cualquier área, para que nadie dé en otra área un permiso que parece de plataforma y no lo es; la baja tampoco puede quitarlos. Los demás permisos son configuración: cada paso del personal nombra su área y los permisos de esa área que exige (ADR-020), y el backend los compara con los de las membresías sin conocerlos de antemano. Un tipo de trámite nuevo trae permisos nuevos, y crearlos no puede requerir código (ADR-006).

**Alternativas descartadas.**
- Mantener el archivo y documentar que prevalece sobre las bajas de un administrador: deja dos fuentes de verdad, y una de ellas sin trazabilidad.
- Leer el archivo solo la primera vez: una base vacía en otro entorno lo volvería a aplicar, y el archivo seguiría guardando datos personales entre los secretos.
- Cargar las áreas del trámite ancla en una migración o desde el código: son configuración de la organización, quedarían fuera del registro de quién y cuándo de la interfaz, y cambiarlas requeriría un cambio de código.
- Que el servidor siga creando o reparando la Secretaría en cada arranque: el rol de la aplicación escribiría lo que es parte del esquema, y un cambio a mano se taparía en vez de verse.
- Sacar el endpoint de creación de permisos: haría falta código para cada permiso de un tipo de trámite nuevo.
- Que `grant-admin` reciba quién lo da (`--by`): es un dato tipeado. Una variable del despliegue con los correos de los administradores: deja la marca fuera de la base y sin autor. Que la primera cuenta que ingresa sea administradora: cualquier cuenta del dominio podría ganar la carrera.
- Dejar los usuarios sintéticos solo en local: siguen siendo personas y membresías fijadas en el código que vuelven en cada arranque, y obligan a mantener un ingreso que ningún entorno real usa.

**Consecuencias.**
- Un entorno nuevo, también el local, arranca con la Secretaría y sin administradores ni personal, y el primer administrador se da con `backend grant-admin <correo>`, a alguien que ya ingresó con Google: en local, con `docker compose run --rm migrate grant-admin <correo>`; en los entornos desplegados, con el job de las migraciones, `gcloud run jobs execute sigta-<entorno>-migrate --args grant-admin,<correo> --wait` ([docs/15](15-despliegue.md)). Sin base no hay cómo, porque la marca se guarda en ella. Los demás los da un administrador desde SIGTA. Quien ingresa sin membresías es un titular, con el catálogo vacío hasta que alguien publique un tipo de trámite.
- Las áreas del trámite ancla y su personal se cargan desde la pantalla de SIG-76 en cada entorno; el tipo de trámite ancla queda pendiente de SIG-15.
- Se pierden la bandeja de ejemplo, los tipos y trámites de ejemplo y el recorrido con usuarios de prueba (entrar como el operativo o la comisión con `login_hint` y confirmar sus pasos). Lo que cubrían lo prueban los tests, cada uno con sus datos: los casos de uso con adaptadores en memoria, los adaptadores contra PostgreSQL, el arranque con su reinicio contra la base y los escenarios con godog. En `npm run dev` siguen las vistas con datos de prueba del inicio de un trámite y de su detalle (`web/src/dev/`), que no pasan por el servidor; la bandeja no tiene una. Para recorrer SIGTA a mano con el personal, se ingresa con Google, se da el primer administrador con `grant-admin`, el administrador da las membresías y se publica un tipo de trámite.
- Las bases de los entornos desplegados conservan los tipos de trámite de ejemplo ya publicados, porque las versiones no se borran, y sus pasos nombran áreas de prueba que ya no se crean. Se limpian recreando la base de cada entorno.
- Quitarle el acceso a alguien o la marca de administrador se hace desde SIGTA. Un primer administrador dado por error da la administración a quien correspondía y después esa persona se la quita. Si no puede hacerlo, se cierra su período en la base con el rol dueño de las tablas, con `valid_until` y `ended_by`, y se vuelve a correr `grant-admin` ([docs/15](15-despliegue.md)).
- Si alguien cambia la Secretaría a mano y el servidor deja de arrancar, `migrate` no la repara, porque la 0007 ya está aplicada: se repara con el rol dueño de las tablas, como dice [docs/15-despliegue.md](15-despliegue.md).
- Ya no hay un archivo con datos personales entre los secretos del despliegue.
- Las únicas personas que guarda la base son las cuentas reales que ingresan con Google (su correo, su nombre y sus sesiones), como ya pasaba en los entornos desplegados. Los trámites y los documentos siguen siendo sintéticos, cargados por quien prueba. Es una excepción a ADR-005, que pedía datos sintéticos en todos los entornos: las cuentas se guardan porque ingresar con Google es la única forma de entrar.
- El despliegue en Google Cloud pasa a `SIGTA_ENV=production`, que exige la base y la carpeta de documentos que ya tenía.
