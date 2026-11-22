@US09 @SIG-15
Feature: Subject equivalence, the anchor procedure type
  As the team,
  I want subject equivalence fully defined in configuration,
  so that it can be followed from end to end.

  The configuration and its areas are the ones the team reviewed, in
  services/backend/test/acceptance/testdata/subject_equivalence/. Everyone
  signs in through the test identity provider and does their part through the
  API, as in SIGTA: nothing is loaded behind it.

  Background:
    Given SIGTA starts with only the Secretariat, and "admin@fi.uba.ar" signs in as its first administrator
    And the administrator creates the areas of subject equivalence, with their permissions and roles
    And these people sign in and the administrator grants them these memberships:
      | person           | area                            | roles                                          | permissions                                        |
      | ana@fi.uba.ar    | Secretaría de Gestión Académica |                                                | Editar tipos de trámite, Publicar tipos de trámite |
      | beto@fi.uba.ar   | Secretaría de Gestión Académica |                                                | Editar tipos de trámite, Publicar tipos de trámite |
      | carla@fi.uba.ar  | Secretaría de Gestión Académica |                                                | Publicar tipos de trámite                          |
      | dora@fi.uba.ar   | Mesa de Entradas                | Personal de Mesa de Entradas                   |                                                    |
      | ciro@fi.uba.ar   | Comisión Curricular             | Integrante de la Comisión Curricular           |                                                    |
      | alba@fi.uba.ar   | Dirección de Alumnos            | Personal de la Dirección de Alumnos            |                                                    |
      | bruno@fi.uba.ar  | Despacho del Consejo Directivo  | Personal del Despacho                          |                                                    |
      | elsa@fi.uba.ar   | Consejo Directivo               | Registro de las sesiones del Consejo Directivo |                                                    |
      | gael@fi.uba.ar   | Dirección General Académica     | Dirección General                              |                                                    |
    And "tomas@fi.uba.ar" signs in, with no membership

  Scenario: the configuration is published when someone who neither edited nor proposed it approves it
    When "ana@fi.uba.ar" creates the draft of subject equivalence with its configuration
    Then the draft of subject equivalence has no problems with the areas
    When "beto@fi.uba.ar" proposes the draft of subject equivalence
    And "carla@fi.uba.ar" approves the draft of subject equivalence
    Then version 1 of subject equivalence is published, proposed by "beto@fi.uba.ar" and approved by "carla@fi.uba.ar"
    And "tomas@fi.uba.ar" finds "Equivalencia entre asignaturas" in the catalog

  Scenario: whoever created or proposed the configuration does not approve it, even with the permission to publish
    Given "ana@fi.uba.ar" created the draft of subject equivalence and "beto@fi.uba.ar" proposed it
    When "beto@fi.uba.ar" approves the draft of subject equivalence
    Then subject equivalence is not published
    When "ana@fi.uba.ar" approves the draft of subject equivalence
    Then subject equivalence is not published

  Scenario: an applicant's procedure goes through every step until it is archived
    Given subject equivalence is published
    When "tomas@fi.uba.ar" starts subject equivalence with their data and their academic record
    Then the procedure has these steps, in order:
      | step                                                              | who       | area                           | action                |
      | Solicitud                                                         | applicant |                                | enter_data            |
      | Documentación                                                     | applicant |                                | upload_document       |
      | Expediente electrónico caratulado                                 | staff     | Mesa de Entradas               | confirm_external_fact |
      | Dictamen de la Comisión Curricular                                | staff     | Comisión Curricular            | upload_document       |
      | Proyecto de resolución                                            | staff     | Dirección de Alumnos           | upload_document       |
      | Resolución de Decano ad referéndum del Consejo Directivo firmada | staff     | Dirección de Alumnos           | confirm_external_fact |
      | Resolución en el orden del día del Consejo Directivo              | staff     | Despacho del Consejo Directivo | confirm_external_fact |
      | Resolución ratificada por el Consejo Directivo                    | staff     | Consejo Directivo              | upload_document       |
      | Equivalencias cargadas en SIU Guaraní                             | staff     | Dirección de Alumnos           | confirm_external_fact |
      | Visto bueno de la Dirección General Académica                     | staff     | Dirección General Académica    | confirm_external_fact |
      | Expediente en guarda temporal                                     | staff     | Mesa de Entradas               | confirm_external_fact |
    And the procedure waits on "Expediente electrónico caratulado"
    When the staff do each step from their inbox, in order:
      | step                                                              | person          | attaches                                       | external file number |
      | Expediente electrónico caratulado                                 | dora@fi.uba.ar  |                                                | EX-2026-00000001     |
      | Dictamen de la Comisión Curricular                                | ciro@fi.uba.ar  | Acta de la Comisión Curricular con su dictamen |                      |
      | Proyecto de resolución                                            | alba@fi.uba.ar  | Proyecto de resolución                         |                      |
      | Resolución de Decano ad referéndum del Consejo Directivo firmada | alba@fi.uba.ar  |                                                |                      |
      | Resolución en el orden del día del Consejo Directivo              | bruno@fi.uba.ar |                                                |                      |
      | Resolución ratificada por el Consejo Directivo                    | elsa@fi.uba.ar  | Resolución del Consejo Directivo               |                      |
      | Equivalencias cargadas en SIU Guaraní                             | alba@fi.uba.ar  |                                                |                      |
      | Visto bueno de la Dirección General Académica                     | gael@fi.uba.ar  |                                                |                      |
      | Expediente en guarda temporal                                     | dora@fi.uba.ar  |                                                |                      |
    Then the procedure ends closed, with no open step
    And its history has each step once, in the order of the procedure type
    And "tomas@fi.uba.ar" downloads the "Resolución del Consejo Directivo" that was attached
    And "tomas@fi.uba.ar" does not see the "Acta de la Comisión Curricular con su dictamen" nor the "Proyecto de resolución"

  Scenario: the external evaluation is a step of the Curricular Commission, which attaches its minutes with the ruling
    Given subject equivalence is published
    And "tomas@fi.uba.ar" started subject equivalence and "dora@fi.uba.ar" confirmed the case opening
    Then the procedure waits on "Dictamen de la Comisión Curricular"
    And only "ciro@fi.uba.ar" has it in their inbox to act on
    When "alba@fi.uba.ar" confirms the open step
    Then the confirmation is refused because they lack the permission
    When "ciro@fi.uba.ar" confirms the open step
    Then the confirmation is refused because a document is missing
    When "ciro@fi.uba.ar" attaches the "Acta de la Comisión Curricular con su dictamen"
    And "ciro@fi.uba.ar" confirms the open step
    Then the procedure waits on "Proyecto de resolución"
