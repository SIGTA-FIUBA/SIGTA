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

**Contexto.** ADR-001 mantenía los tipos de trámite como archivos en el repositorio, editados por el equipo mediante pull request, y quedó en revisión por D-07, que evaluaba un editor dentro de la aplicación. Con las versiones guardadas en la base y validadas al publicar (ADR-012), un editor ya no exige un motor aparte: es una interfaz sobre el mismo modelo. Además, al terminar el proyecto el equipo deja de estar disponible para hacer pull requests.

**Decisión.**
- Los tipos de trámite se editan desde la interfaz de SIGTA, como borradores, y se publican como versiones nuevas.
- Cada versión la propone una persona y la aprueba otra. La versión guarda a las dos y la fecha de publicación.
- No se publica una definición inválida: el servicio la valida al publicar.

**Consecuencias.**
- Se cierra D-07. Queda D-08: quién puede proponer y quién aprobar.
- El editor deja de estar fuera del alcance; lo que queda afuera es publicar sin la aprobación de otra persona.
- La documentación del modelo, en CONTRIBUTING, sigue siendo un entregable del cierre, ahora como referencia para quien edita.
- La interfaz de edición y su endpoint no tienen historia en el backlog. A confirmar: qué historia los incluye, con su estimación, y cómo se publica el trámite ancla mientras no existan. Hasta entonces, los borradores de los cinco trámites se validan fuera del repositorio, con el test `TestDrafts`, porque describen procesos que la Secretaría no confirmó.
- Cambian cuatro criterios de "Modelo de definición de trámite": el archivo versionado, los estados, la validación al arrancar y los requisitos compartidos entre trámites. A actualizar en Linear con el equipo.
- No aparece la persona administrador técnico (P4) que preveía ADR-001.
- El riesgo R-02 baja, pero sigue abierto hasta resolver D-08.

---

## ADR-014: El servicio de trámites se divide por módulo y, dentro de cada módulo, por componente

**Estado:** propuesta, a acordar con el equipo

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

---

## ADR-017: Los permisos del personal se asignan con roles configurables, administrados en SIGTA

**Estado:** propuesta, a acordar con el equipo
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

**Estado:** propuesta, a acordar con el equipo

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

## Decisiones pendientes

| ID | Decisión pendiente | Necesaria antes de | Responsable |
|---|---|---|---|
| D-04 | Infraestructura de despliegue y su financiamiento | Fin de la fase de diseño | Equipo |
| D-08 | Quién puede proponer y quién aprobar una versión de un tipo de trámite: el equipo, la Secretaría o ambos | La interfaz de edición de tipos de trámite | Equipo y canal con la Secretaría |
| D-09 | Cuánto tiempo se conservan los documentos de un trámite cerrado | La estimación de almacenamiento para TICs | Secretaría |
