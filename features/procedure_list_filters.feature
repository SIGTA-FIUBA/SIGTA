@US37 @SIG-25
Feature: Filter and sort the procedure list
  As a staff member,
  I want to filter and sort the list,
  so that I find quickly what I am looking for.

  Background:
    Given these procedure types:
      | procedure type      | step               | done by                          |
      | subject equivalence | Record file number | Students office, "record"        |
      | subject equivalence | Ruling             | Curricular committee, "evaluate" |
      | credit recognition  | Requested document | the applicant                    |
      | credit recognition  | Record file number | Students office, "record"        |
    And these memberships:
      | staff               | area            | permissions | procedure types |
      | operative@fi.uba.ar | Students office | record      |                 |
    And these procedures in progress:
      | id  | procedure type      | applicant        | open step          | started days ago | moved days ago |
      | p-1 | subject equivalence | lucia@fi.uba.ar  | Record file number | 3                | 3              |
      | p-2 | credit recognition  | martin@fi.uba.ar | Record file number | 20               | 2              |
      | p-3 | subject equivalence | sofia@fi.uba.ar  | Ruling             | 30               | 12             |
      | p-4 | credit recognition  | tomas@fi.uba.ar  | Requested document | 8                | 1              |

  Scenario: filters by procedure type
    When "operative@fi.uba.ar" lists all the procedures of the "credit recognition" type
    Then it lists the procedures "p-2, p-4"

  Scenario: filters by state
    When "operative@fi.uba.ar" lists all the procedures in the state "with_applicant"
    Then it lists the procedures "p-4"

  Scenario: filters by the area responsible for the open step
    When "operative@fi.uba.ar" lists all the procedures in the hands of "Students office"
    Then it lists the procedures "p-2, p-1"

  Scenario: filters combine
    When "operative@fi.uba.ar" lists all the procedures of the "subject equivalence" type in the state "in_office"
    Then it lists the procedures "p-3, p-1"

  Scenario: sorts by age, oldest first
    When "operative@fi.uba.ar" lists all the procedures sorted by "oldest"
    Then it lists the procedures "p-3, p-2, p-4, p-1"

  Scenario: sorts by last movement, most recent first
    When "operative@fi.uba.ar" lists all the procedures sorted by "last_moved"
    Then it lists the procedures "p-4, p-2, p-1, p-3"

  Scenario: the filter options do not depend on the filter chosen
    When "operative@fi.uba.ar" lists all the procedures of the "credit recognition" type
    Then the type filter offers "credit recognition, subject equivalence"
    And the area filter offers "Curricular committee, Students office"
