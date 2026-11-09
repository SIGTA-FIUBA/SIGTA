@US17 @SIG-22
Feature: Procedure inbox
  As a staff member,
  I want to see when I come in what needs my intervention,
  so that I do not go through spreadsheets or mailboxes.

  Background:
    Given these procedures in progress:
      | id  | procedure type      | open step          | area                 | area kind | profile         | started days ago | moved days ago |
      | p-1 | subject equivalence | Record file number | Students office      | office    | staff_operative | 3                | 3              |
      | p-2 | credit recognition  | Sign resolution    | Students office      | office    | staff_approver  | 20               | 2              |
      | p-3 | subject equivalence | Ruling             | Curricular committee | external  |                 | 30               | 12             |
      | p-4 | credit recognition  | Requested document | Applicant            | applicant |                 | 8                | 1              |

  Scenario: the inbox shows what an operative profile can act on
    When a staff member with the "staff_operative" profile opens their inbox
    Then it lists the procedures "p-1"

  Scenario: the approval profile also acts on the operative steps
    When a staff member with the "staff_approver" profile opens their inbox
    Then it lists the procedures "p-2, p-1"

  Scenario: each item shows its procedure type, pending step and age
    When a staff member with the "staff_operative" profile opens their inbox
    Then "p-1" shows the procedure type "subject equivalence"
    And "p-1" shows the pending step "Record file number" in "Students office"
    And "p-1" is 3 days old

  Scenario: every procedure in progress can be listed, not only the actionable ones
    When a staff member with the "staff_operative" profile lists all the procedures
    Then it lists the procedures "p-3, p-2, p-4, p-1"
    And "p-1" is marked as actionable
    And "p-2" is not marked as actionable

  Scenario Outline: each item states the state derived from its open step
    When a staff member with the "staff_operative" profile lists all the procedures
    Then "<id>" is in the state "<state>"

    Examples:
      | id  | state          |
      | p-1 | in_office      |
      | p-3 | referred       |
      | p-4 | with_applicant |

  Scenario: a referred procedure shows how long it has been without return
    When a staff member with the "staff_operative" profile lists all the procedures
    Then "p-3" has been referred for 12 days
    And "p-1" shows no referral days

  Scenario: a staff member sees only the procedure types they hold a profile in
    Given a staff member holds the "staff_operative" profile only in "credit recognition"
    When they list all the procedures
    Then it lists the procedures "p-2, p-4"

  Scenario: the profile held in each procedure type decides what needs action
    Given a staff member holds the "staff_operative" profile in "credit recognition" and the "staff_approver" profile in "subject equivalence"
    When they open their inbox
    Then it lists the procedures "p-1"

  Scenario: an applicant cannot open the inbox
    When an applicant asks for the inbox
    Then they are told the inbox is only for the staff

  Scenario: someone without a session cannot open the inbox
    When someone without a session asks for the inbox
    Then they are asked to sign in
