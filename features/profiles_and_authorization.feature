@US07 @SIG-14
Feature: Profiles and authorization
  As the Secretaría,
  I want each person to see and do only what belongs to them,
  so that procedures and approval decisions are protected.

  Background:
    Given a published "subject equivalence" procedure type with these steps:
      | step    | profile         |
      | verify  | staff_operative |
      | approve | staff_approver  |
    And a published "credit recognition" procedure type with these steps:
      | step   | profile         |
      | verify | staff_operative |
    And SIGTA has these users:
      | user                | kind          |
      | applicant@fi.uba.ar | applicant     |
      | other@fi.uba.ar     | applicant     |
      | admin@fi.uba.ar     | administrator |
    And SIGTA has these staff permissions:
      | staff               | procedure type      | profile         |
      | operative@fi.uba.ar | subject equivalence | staff_operative |
      | approver@fi.uba.ar  | subject equivalence | staff_approver  |
      | approver@fi.uba.ar  | credit recognition  | staff_operative |

  Scenario Outline: SIGTA recognizes three kinds of user
    When a user is registered as "<kind>"
    Then the user is registered

    Examples:
      | kind          |
      | applicant     |
      | staff         |
      | administrator |

  Scenario: a kind of user outside the three is rejected
    When a user is registered as "supervisor"
    Then the user is not registered

  Scenario: staff cannot start a procedure
    When "operative@fi.uba.ar" starts a "subject equivalence" procedure
    Then the procedure is not started
    And a denied attempt is recorded with who tried, what they tried and when

  Scenario: an applicant cannot be granted a staff permission
    When a staff permission is granted to "applicant@fi.uba.ar" on "subject equivalence" with the "staff_operative" profile
    Then the permission is not granted

  Scenario: a staff permission needs a known profile
    When a staff permission is granted on "subject equivalence" with the "supervisor" profile
    Then the permission is not granted

  Scenario: a staff permission needs a published procedure type
    When a staff permission is granted on "unknown procedure" with the "staff_operative" profile
    Then the permission is not granted

  Scenario Outline: who can act on each step
    Given a "<procedure type>" procedure whose open step is "<step>"
    When "<user>" acts on that step
    Then the action is <result>

    Examples:
      | procedure type      | step    | user                | result  |
      | subject equivalence | verify  | operative@fi.uba.ar | allowed |
      | subject equivalence | verify  | approver@fi.uba.ar  | allowed |
      | subject equivalence | approve | approver@fi.uba.ar  | allowed |
      | subject equivalence | approve | operative@fi.uba.ar | denied  |
      | credit recognition  | verify  | approver@fi.uba.ar  | allowed |
      | credit recognition  | verify  | operative@fi.uba.ar | denied  |
      | subject equivalence | verify  | applicant@fi.uba.ar | denied  |
      | subject equivalence | verify  | admin@fi.uba.ar     | denied  |

  Scenario: the same person has a different profile in each procedure type
    Given a "credit recognition" procedure whose open step is "verify"
    And a "subject equivalence" procedure whose open step is "approve"
    When "approver@fi.uba.ar" acts on each open step
    Then both actions are allowed

  Scenario: staff only see procedures of the types they have a permission on
    Given a "subject equivalence" procedure
    And a "credit recognition" procedure
    When "operative@fi.uba.ar" lists the procedures
    Then they see 1 procedure
    And it is a "subject equivalence" procedure

  Scenario: an operative profile is not offered approval actions
    Given a "subject equivalence" procedure whose open step is "approve"
    When "operative@fi.uba.ar" opens that procedure
    Then the procedure is shown
    And no action is offered on the "approve" step

  Scenario: an applicant opens their own procedure
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure
    When "applicant@fi.uba.ar" opens that procedure
    Then the procedure is shown

  Scenario: an applicant cannot open a procedure of someone else
    Given "other@fi.uba.ar" has a "subject equivalence" procedure
    When "applicant@fi.uba.ar" opens that procedure by its identifier
    Then access is denied
    And the response does not reveal whether the procedure exists

  Scenario: an applicant only sees their own procedures
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure
    And "other@fi.uba.ar" has a "subject equivalence" procedure
    When "applicant@fi.uba.ar" lists their procedures
    Then they see 1 procedure
    And none of them belongs to "other@fi.uba.ar"

  Scenario: authorization is enforced by the server, not only by the interface
    Given a "subject equivalence" procedure whose open step is "approve"
    When "operative@fi.uba.ar" sends the approval directly to the API
    Then the action is denied
    And the procedure stays on the "approve" step

  Scenario: a permission keeps working when a new version is published
    Given a "subject equivalence" procedure whose open step is "verify"
    When a new version of "subject equivalence" is published
    And "operative@fi.uba.ar" acts on that step
    Then the action is allowed

  Scenario: a denied attempt is recorded
    Given "other@fi.uba.ar" has a "subject equivalence" procedure
    When "applicant@fi.uba.ar" opens that procedure by its identifier
    Then a denied attempt is recorded with who tried, what they tried, on which procedure and when

  Scenario: a denied action on a step is recorded
    Given a "subject equivalence" procedure whose open step is "approve"
    When "operative@fi.uba.ar" acts on that step
    Then a denied attempt is recorded with who tried, what they tried, on which procedure and when

  Scenario: an allowed action is not recorded as a denied attempt
    Given a "subject equivalence" procedure whose open step is "verify"
    When "operative@fi.uba.ar" acts on that step
    Then no denied attempt is recorded
