# Personas usuarias y actores

**Versión:** 0.2 · alineado al mapa de usuario por roles
Las afirmaciones marcadas **(supuesto)** son hipótesis nuestras y se validan en el relevamiento del sprint 1 (US62 y US64).

---

## 1. Roles de la plataforma

La plataforma tiene **dos roles**. Son los mismos dos que ordenan el User Story Mapping.

| Rol | Perfiles de permisos | Qué hace |
|---|---|---|
| **Titular del trámite** | Uno | Inicia, carga documentación, consulta estado, completa subsanaciones, conversa con el asistente |
| **Secretario** | **Operativo** y **de aprobación** | Recibe, verifica, deriva, registra retornos; el perfil de aprobación además aprueba y firma |

**Por qué un rol con dos perfiles.** En el trabajo diario es el mismo tipo de persona haciendo el mismo recorrido, así que el mapa de usuario lo trata como un solo rol. Pero no cualquiera puede aprobar o firmar, así que los permisos distinguen dos perfiles. Si en el relevamiento resulta que son siempre la misma persona, el modelo se simplifica a un perfil (supuesto S-07, decisión D-02).

## 2. Actores que no son roles de la plataforma

| Actor | Relación con la plataforma |
|---|---|
| **Evaluador externo** (comisión curricular, departamento u otra unidad académica) | No usa la plataforma. Recibe derivaciones y devuelve dictámenes fuera de ella; el Secretario registra la salida y el retorno (ADR-004). Que pudiera trabajar dentro de la plataforma es una **duda pendiente** |
| **Equipo de proyecto** | Mantiene la configuración de trámites y la base de conocimiento mediante pull request (ADR-001). No es un rol de la plataforma: por eso no tiene mapa de usuario y su trabajo vive solo en el backlog |
| **Subsecretaría de TICs de la FIUBA** | Interlocutor para el entorno de CI/CD, el hosting y la identidad institucional |
| **Directores del proyecto** | Supervisan y validan entregables |

> Ningún trámite se resuelve íntegramente dentro del área: los cinco pasan por al menos un evaluador externo. Por eso el Secretario funciona como puente permanente entre la plataforma y actores que no la usan.

---

## 3. Personas

### P1 — Titular del trámite

**Quién es.** Alguien de la facultad que necesita resolver un trámite de reconocimiento de estudios. Entra pocas veces por trámite y puede pasar meses sin volver.

**Qué necesita.**
- Saber qué trámite corresponde a su situación, porque los cinco se parecen y elegir mal cuesta semanas.
- Saber qué documentación reunir y **de dónde se saca** cada documento.
- Saber en qué estado está su trámite sin preguntarle a nadie.

**Qué le duele hoy.**
- Los requisitos están dispersos y en lenguaje administrativo.
- Inicia el trámite en un formulario o por mail y después no tiene forma de consultarlo.
- Para reconstruir qué pasó tiene que buscar mails viejos.
- Hay plazos que no ve hasta que los pierde. *(caso real mencionado en la reunión del 16/09)*

**Qué espera del asistente.** Orientación sobre cómo obtener los documentos que le piden. No espera que el asistente le diga en qué está su trámite: eso lo muestra la interfaz.

**Cómo se ve el éxito.** Inicia el trámite sin escribir un mail, y en cualquier momento entra y entiende en qué paso está y en manos de quién.

**Contexto de uso (supuesto).** Desde el celular, en sesiones cortas y espaciadas.

---

### P2 — Secretario, perfil operativo

**Quién es (supuesto).** Personal de la Secretaría que recibe las solicitudes, verifica que la documentación esté completa, deriva al evaluador externo y registra el retorno. En la planilla de circuitos aparece como "área de recepción".

**Qué necesita.**
- Una bandeja única con lo que requiere su intervención.
- Ver rápido si una solicitud está completa o le falta algo.
- Registrar que hizo algo fuera de la plataforma sin explicarlo por mail.
- Saber qué derivación se está durmiendo.

**Qué le duele hoy.**
- Las solicitudes llegan de fuentes distintas y la centralización es manual.
- Las consultas de los titulares saturan su canal.
- No hay registro de quién hizo qué paso y cuándo.

**Lo que no quiere.** Que el sistema decida por ella. La validación asistida le señala hallazgos; la decisión sigue siendo suya.

**Cómo se ve el éxito.** Abre la plataforma, ve qué tiene pendiente y lo resuelve sin salir a buscar información a otro lado.

---

### P3 — Secretario, perfil de aprobación

**Quién es (supuesto).** La persona de la Secretaría con autoridad para aprobar y firmar. Además de todo lo que hace el perfil operativo, revisa y aprueba las preactas.

**Qué necesita.**
- Ver lo que está pendiente de su aprobación.
- Preactas listas para revisar, con los campos que salen de los datos del trámite ya completados.
- Saber qué está trabado y dónde.

**Qué le duele hoy.**
- No tiene visión agregada del área.
- Redacta actas desde cero o copiando documentos anteriores.

**Cómo se ve el éxito.** Revisa un borrador, completa lo que depende del dictamen, corrige lo que haga falta y aprueba.

**Límite.** La plataforma no interpreta el dictamen del evaluador (ADR-009): los campos de la preacta que dependen del resultado los completa esta persona.

---

## 4. Permisos

| Perfil | Ve trámites | Avanza pasos operativos | Deriva y registra retornos | Aprueba y firma | Edita configuración |
|---|---|---|---|---|---|
| Titular | Solo los propios | No | No | No | No |
| Secretario operativo | Todos los de su ámbito | Sí | Sí | No | No |
| Secretario de aprobación | Todos | Sí | Sí | Sí | No |
| Equipo de proyecto | — | — | — | — | Sí, vía pull request |

> La tabla es la hipótesis mínima. Se confirma con el mapa de roles de US64, en el sprint 1.
