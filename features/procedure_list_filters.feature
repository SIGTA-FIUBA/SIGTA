@US37 @SIG-25
Feature: Filter and sort the procedure list
  As a staff member,
  I want to filter and sort the list,
  so that I find quickly what I am looking for.

  Background:
    Given these procedures in progress:
      | id  | procedure type      | open step          | area                 | area kind | profile         | started days ago | moved days ago |
      | p-1 | subject equivalence | Record file number | Students office      | office    | staff_operative | 3                | 3              |
      | p-2 | credit recognition  | Record file number | Students office      | office    | staff_operative | 20               | 2              |
      | p-3 | subject equivalence | Ruling             | Curricular committee | external  |                 | 30               | 12             |
      | p-4 | credit recognition  | Requested document | Applicant            | applicant |                 | 8                | 1              |

  Scenario: filters by procedure type
    When a staff member lists all the procedures of the "credit recognition" type
    Then it lists the procedures "p-2, p-4"

  Scenario: filters by state
    When a staff member lists all the procedures in the state "referred"
    Then it lists the procedures "p-3"

  Scenario: filters by the area responsible for the open step
    When a staff member lists all the procedures in the hands of "Students office"
    Then it lists the procedures "p-2, p-1"

  Scenario: filters combine
    When a staff member lists all the procedures of the "subject equivalence" type in the state "in_office"
    Then it lists the procedures "p-1"

  Scenario: sorts by age, oldest first
    When a staff member lists all the procedures sorted by "oldest"
    Then it lists the procedures "p-3, p-2, p-4, p-1"

  Scenario: sorts by last movement, most recent first
    When a staff member lists all the procedures sorted by "last_moved"
    Then it lists the procedures "p-4, p-2, p-1, p-3"

  Scenario: the filter options do not depend on the filter chosen
    When a staff member lists all the procedures of the "credit recognition" type
    Then the type filter offers "credit recognition, subject equivalence"
