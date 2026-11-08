@US07 @SIG-14
Feature: Profiles and authorization
  As the Secretaría,
  I want each person to see and do only what belongs to them,
  so that procedures and approval decisions are protected.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And a published "subject equivalence" procedure type with these steps:
      | step     | area                  | profile         |
      | verify   | student affairs       | staff_operative |
      | evaluate | curricular commission | staff_operative |
      | approve  | student affairs       | staff_approver  |
    And a published "credit recognition" procedure type with these steps:
      | step   | area            | profile         |
      | verify | student affairs | staff_operative |
    And "admin@fi.uba.ar" is an administrator
    And SIGTA has these memberships:
      | staff                | area                  | profile         | scope              |
      | operative@fi.uba.ar  | student affairs       | staff_operative |                    |
      | approver@fi.uba.ar   | student affairs       | staff_approver  |                    |
      | commission@fi.uba.ar | curricular commission | staff_operative |                    |
      | scoped@fi.uba.ar     | student affairs       | staff_operative | credit recognition |

  Scenario: anyone who signs in is an applicant by default
    When "new@fi.uba.ar" starts a "subject equivalence" procedure
    Then the procedure is started

  Scenario: staff cannot start a procedure
    When "operative@fi.uba.ar" starts a "subject equivalence" procedure
    Then the procedure is not started
    And a denied attempt is recorded with who tried, what they tried and when

  Scenario: a membership turns an applicant into staff
    When "admin@fi.uba.ar" grants "new@fi.uba.ar" a membership in "student affairs" with the "staff_operative" profile
    Then the membership is granted by "admin@fi.uba.ar" from now on
    And "new@fi.uba.ar" cannot start a "subject equivalence" procedure

  Scenario: only an administrator grants memberships
    When "approver@fi.uba.ar" grants "new@fi.uba.ar" a membership in "student affairs" with the "staff_operative" profile
    Then the membership is denied
    And a denied attempt is recorded with who tried, what they tried and when

  Scenario Outline: a membership needs a staff area, a known profile and published procedure types
    When "admin@fi.uba.ar" grants "new@fi.uba.ar" a membership in "<area>" with the "<profile>" profile on "<scope>"
    Then the membership is not granted

    Examples:
      | area            | profile         | scope              |
      | unknown area    | staff_operative |                    |
      | applicant       | staff_operative |                    |
      | student affairs | supervisor      |                    |
      | student affairs | staff_operative | unknown procedure  |

  Scenario: a person has one membership per area and scope
    When "admin@fi.uba.ar" grants "operative@fi.uba.ar" a membership in "student affairs" with the "staff_approver" profile
    Then the membership is not granted

  Scenario: a membership with a narrower scope raises the profile on those procedure types
    Given a "subject equivalence" procedure whose open step is "approve"
    When "admin@fi.uba.ar" grants "operative@fi.uba.ar" a membership in "student affairs" with the "staff_approver" profile on "subject equivalence"
    And "operative@fi.uba.ar" acts on that step
    Then the action is allowed

  Scenario Outline: who can act on each step
    Given a "<procedure type>" procedure whose open step is "<step>"
    When "<user>" acts on that step
    Then the action is <result>

    Examples:
      | procedure type      | step     | user                 | result  |
      | subject equivalence | request  | applicant@fi.uba.ar  | allowed |
      | subject equivalence | request  | other@fi.uba.ar      | denied  |
      | subject equivalence | request  | operative@fi.uba.ar  | denied  |
      | subject equivalence | verify   | operative@fi.uba.ar  | allowed |
      | subject equivalence | verify   | approver@fi.uba.ar   | allowed |
      | subject equivalence | approve  | approver@fi.uba.ar   | allowed |
      | subject equivalence | approve  | operative@fi.uba.ar  | denied  |
      | subject equivalence | evaluate | commission@fi.uba.ar | allowed |
      | subject equivalence | evaluate | operative@fi.uba.ar  | denied  |
      | subject equivalence | verify   | commission@fi.uba.ar | denied  |
      | credit recognition  | verify   | scoped@fi.uba.ar     | allowed |
      | subject equivalence | verify   | scoped@fi.uba.ar     | denied  |
      | subject equivalence | verify   | applicant@fi.uba.ar  | denied  |
      | subject equivalence | verify   | admin@fi.uba.ar      | denied  |

  Scenario: an ended membership no longer lets the person act
    Given a "subject equivalence" procedure whose open step is "verify"
    When "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    And "operative@fi.uba.ar" acts on that step
    Then the action is denied
    And the membership records that "admin@fi.uba.ar" ended it

  Scenario: a person whose memberships ended is an applicant again
    When "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    And "operative@fi.uba.ar" starts a "subject equivalence" procedure
    Then the procedure is started

  Scenario Outline: a membership with an end date holds until then
    Given "admin@fi.uba.ar" grants "substitute@fi.uba.ar" a membership in "student affairs" with the "staff_operative" profile until "2026-11-10"
    And a "subject equivalence" procedure whose open step is "verify"
    When it is "<date>"
    And "substitute@fi.uba.ar" acts on that step
    Then the action is <result>

    Examples:
      | date       | result  |
      | 2026-11-09 | allowed |
      | 2026-11-10 | denied  |

  Scenario: staff see the procedures that go through their areas
    Given a "subject equivalence" procedure
    And a "credit recognition" procedure
    When "commission@fi.uba.ar" lists the procedures
    Then they see 1 procedure
    And it is a "subject equivalence" procedure

  Scenario: the scope narrows the procedures staff see
    Given a "subject equivalence" procedure
    And a "credit recognition" procedure
    When "scoped@fi.uba.ar" lists the procedures
    Then they see 1 procedure
    And it is a "credit recognition" procedure

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

  Scenario: an administrator without a membership does not see procedures of others
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure
    When "admin@fi.uba.ar" lists the procedures
    Then they see 0 procedures

  Scenario: authorization is enforced by the server, not only by the interface
    Given a "subject equivalence" procedure whose open step is "approve"
    When "operative@fi.uba.ar" sends the approval directly to the API
    Then the action is denied
    And the procedure stays on the "approve" step

  Scenario: a membership keeps working when a new version is published
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
