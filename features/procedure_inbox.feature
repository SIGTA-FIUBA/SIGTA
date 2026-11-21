@US17 @SIG-22
Feature: Procedure inbox
  As a staff member,
  I want to see when I come in what needs my intervention,
  so that I do not go through spreadsheets or mailboxes.

  Background:
    Given these procedure types:
      | procedure type      | step               | done by                          |
      | subject equivalence | Request            | the applicant                    |
      | subject equivalence | Record file number | Students office, "record"        |
      | subject equivalence | Ruling             | Curricular committee, "evaluate" |
      | credit recognition  | Requested document | the applicant                    |
      | credit recognition  | Sign resolution    | Students office, "sign"          |
    And these memberships:
      | staff                | area                 | permissions  | procedure types    |
      | operative@fi.uba.ar  | Students office      | record       |                    |
      | approver@fi.uba.ar   | Students office      | record, sign |                    |
      | credits@fi.uba.ar    | Students office      | record, sign | credit recognition |
      | commission@fi.uba.ar | Curricular committee | evaluate     |                    |
    And these procedures in progress:
      | id  | procedure type      | applicant         | open step          | started days ago | moved days ago |
      | p-1 | subject equivalence | lucia@fi.uba.ar   | Record file number | 3                | 3              |
      | p-2 | credit recognition  | martin@fi.uba.ar  | Sign resolution    | 20               | 2              |
      | p-3 | subject equivalence | sofia@fi.uba.ar   | Ruling             | 30               | 12             |
      | p-4 | credit recognition  | tomas@fi.uba.ar   | Requested document | 8                | 1              |

  Scenario: the inbox shows the procedures whose open step the staff member holds the permissions for
    When "operative@fi.uba.ar" opens their inbox
    Then it lists the procedures "p-1"

  Scenario: whoever holds the permissions of several steps sees all of them
    When "approver@fi.uba.ar" opens their inbox
    Then it lists the procedures "p-2, p-1"

  Scenario: each item shows its procedure type, applicant, pending step and age
    When "operative@fi.uba.ar" opens their inbox
    Then "p-1" shows the procedure type "subject equivalence"
    And "p-1" shows the applicant "lucia@fi.uba.ar"
    And "p-1" shows the pending step "Record file number" in "Students office"
    And "p-1" is 3 days old

  Scenario: every procedure in progress that goes through the area can be listed, not only the actionable ones
    When "operative@fi.uba.ar" lists all the procedures
    Then it lists the procedures "p-3, p-2, p-4, p-1"
    And "p-1" is marked as actionable
    And "p-2" is not marked as actionable

  Scenario Outline: each item states the state derived from its open step
    When "operative@fi.uba.ar" lists all the procedures
    Then "<id>" is in the state "<state>"

    Examples:
      | id  | state          |
      | p-1 | in_office      |
      | p-3 | in_office      |
      | p-4 | with_applicant |

  Scenario: a step of the applicant shows no area
    When "operative@fi.uba.ar" lists all the procedures
    Then "p-4" shows the pending step "Requested document" with no area

  Scenario: a membership limited to some procedure types sees only those
    When "credits@fi.uba.ar" lists all the procedures
    Then it lists the procedures "p-2, p-4"

  Scenario: an area sees only the procedure types that go through it
    When "commission@fi.uba.ar" lists all the procedures
    Then it lists the procedures "p-3, p-1"

  Scenario: a closed procedure is not in the inbox
    Given the procedure "p-1" has no open step
    When "operative@fi.uba.ar" lists all the procedures
    Then it lists the procedures "p-3, p-2, p-4"

  Scenario: a staff member does not find their own procedure in the inbox
    Given "operative@fi.uba.ar" has a procedure "p-5" of "subject equivalence" waiting on "Record file number"
    When "operative@fi.uba.ar" opens their inbox
    Then it lists the procedures "p-1"

  Scenario: someone with no membership cannot open the inbox
    When "lucia@fi.uba.ar" asks for the inbox
    Then they are told the inbox is only for the staff

  Scenario: someone without a session cannot open the inbox
    When someone without a session asks for the inbox
    Then they are asked to sign in
