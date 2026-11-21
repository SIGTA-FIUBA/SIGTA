@US14 @SIG-17
Feature: Start a procedure
  As an applicant,
  I want to start a procedure and attach its documents,
  so that I submit it without writing an email.

  The applicant fills in what the opening steps ask for, the steps of the
  applicant before the first step of the staff, and submits it all at once.
  Until then nothing is stored: the summary and replacing or removing a file
  before submitting happen in the browser.

  Background:
    Given a published procedure type with these fields:
      | field    | type         | options       |
      | reason   | long_text    |               |
      | credits  | number       |               |
      | since    | date         |               |
      | modality | option       | full, partial |
      | subjects | subject_list |               |
      | comment  | text         |               |
      | grade    | number       |               |
    And these documents:
      | document   | types                      | max size |
      | transcript | application/pdf, image/png | 1024     |
      | program    | application/pdf            | 1024     |
      | ruling     | application/pdf            | 1024     |
    And these steps:
      | step      | done by                   | mandatory                                  | optional |
      | request   | the applicant             | reason, credits, since, modality, subjects | comment  |
      | documents | the applicant             | transcript                                 | program  |
      | review    | student affairs, "review" | grade                                      |          |
      | ruling    | student affairs, "review" | ruling                                     |          |

  Scenario: a complete submission starts the procedure on the first step of the staff
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement
    Then the procedure is started with an identifier
    And its applicant is "applicant@fi.uba.ar"
    And it waits on the "review" step, in an office

  Scenario: what the applicant submits stays with the procedure
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement and:
      | requirement | value                 |
      | reason      | Cursé la materia afuera |
      | subjects    | Análisis I, Física I  |
      | program     | a PDF of 300 bytes    |
    Then the procedure records "reason" as "Cursé la materia afuera" on the "request" step
    And the procedure records "subjects" as "Análisis I, Física I" on the "request" step
    And the "program" of the procedure is uploaded by "applicant@fi.uba.ar" on the "documents" step

  Scenario: the optional requirements can be left out
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement
    Then the procedure is started with an identifier
    And the procedure records no "comment"
    And the procedure has no "program"

  Scenario Outline: a mandatory requirement cannot be left out
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement but the "<requirement>"
    Then the submission is rejected because the "<requirement>" is missing
    And no procedure is started

    Examples:
      | requirement |
      | reason      |
      | credits     |
      | since       |
      | modality    |
      | subjects    |
      | transcript  |

  Scenario Outline: each field is checked against its type
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement and:
      | requirement | value   |
      | <field>     | <value> |
    Then the submission is rejected because the "<field>" is not valid
    And no procedure is started

    Examples:
      | field    | value      |
      | reason   | " "        |
      | credits  | doce       |
      | since    | 31/12/2026 |
      | since    | 2026-02-30 |
      | modality | Full       |
      | modality | other      |
      | subjects | " , "      |

  Scenario Outline: each file is checked against its document
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement and:
      | requirement | value  |
      | transcript  | <file> |
    Then the submission is rejected because <reason>
    And no procedure is started

    Examples:
      | file                | reason                   |
      | a text file         | its type is not accepted |
      | a PDF of 1025 bytes | it is too large          |
      | an empty file       | it is empty              |

  Scenario Outline: what the opening steps do not ask for is rejected
    When "applicant@fi.uba.ar" submits the procedure with every mandatory requirement and:
      | requirement   | value   |
      | <requirement> | <value> |
    Then the submission is rejected because the "<requirement>" is not asked when starting
    And no procedure is started

    Examples:
      | requirement | value              |
      | grade       | 10                 |
      | ruling      | a PDF of 200 bytes |
      | unknown     | anything           |

  Scenario: the staff start procedures in their own name
    Given "reader@fi.uba.ar" is staff of "student affairs" with the "review" permission
    When "reader@fi.uba.ar" submits the procedure with every mandatory requirement
    Then the procedure is started with an identifier
    And its applicant is "reader@fi.uba.ar"

  Scenario: a procedure type that is not published cannot be started
    When "applicant@fi.uba.ar" submits a procedure of a type that is not published
    Then the procedure type is not found
    And no procedure is started
