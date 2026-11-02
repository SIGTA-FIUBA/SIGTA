@US13 @SIG-16
Feature: Browse the catalog of procedure types
  As an applicant,
  I want to see which procedure types exist and what each one asks for,
  so that I know whether it applies to me and what I have to gather.

  Background:
    Given a published "subject equivalence" procedure type described as "Pedí que te reconozcan una materia"

  Scenario: the catalog lists the published procedure types in plain language
    Given a published "credit recognition" procedure type described as "Pedí créditos por una actividad"
    When the catalog is opened
    Then it lists 2 procedure types
    And "subject equivalence" is described as "Pedí que te reconozcan una materia"
    And "credit recognition" is described as "Pedí créditos por una actividad"

  Scenario: requirements say which ones are mandatory
    When the catalog is opened
    Then "subject equivalence" shows the "full_name" field as mandatory
    And "subject equivalence" shows the "reason" field as optional
    And "subject equivalence" shows the "transcript" document as mandatory
    And "subject equivalence" shows the "program" document as optional

  Scenario: a requirement mandatory in any applicant step is mandatory
    Given the "reason" field is also mandatory in a later applicant step
    When the catalog is opened
    Then "subject equivalence" shows the "reason" field as mandatory

  Scenario: what the staff records is not asked of the applicant
    When the catalog is opened
    Then "subject equivalence" does not show the "file_number" field
    And "subject equivalence" does not show the "ruling" document

  Scenario: each document shows the accepted formats and size
    When the catalog is opened
    Then the "transcript" document accepts "application/pdf" up to 10485760 bytes

  Scenario: the catalog shows deadlines and admissibility conditions
    When the catalog is opened
    Then "subject equivalence" shows the deadline "Hasta 30 días antes de cursar"
    And "subject equivalence" shows the admissibility condition "Ser estudiante regular"

  Scenario: a new procedure type appears by publishing it
    Given the catalog was opened
    When a "degree transfer" procedure type is published
    And the catalog is opened
    Then it lists 2 procedure types

  Scenario: the catalog follows the latest version
    Given a new version of "subject equivalence" described as "Nueva descripción"
    When the catalog is opened
    Then it lists 1 procedure type
    And "subject equivalence" is described as "Nueva descripción"

  Scenario: the catalog fails if a version was altered
    Given the stored content of the latest version no longer matches its digest
    When the catalog is opened
    Then the catalog is not shown

  Scenario Outline: a definition the catalog cannot show is not published
    When a definition with <problem> is published
    Then the publication is rejected with the code "<code>"

    Examples:
      | problem                         | code                   |
      | no description                  | no_description         |
      | a document without formats      | no_accepted_types      |
      | a document with a wrong format  | invalid_accepted_type  |
      | a document without a size limit | no_max_size            |
      | a blank deadline                | blank_text             |
