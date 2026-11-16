@US10 @SIG-20
Feature: Step confirmation
  As staff of an area,
  I want to confirm that a step was done,
  so that the procedure moves forward and the confirmation is recorded.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | next                                 |
      | request  | applicant             |             | enter_data            | check                                |
      | check    | student affairs       | verify      | decide                | accept: evaluate, send back: request |
      | evaluate | curricular commission | evaluate    | confirm_external_fact | file                                 |
      | file     | student affairs       | verify      | confirm_external_fact | approve                              |
      | approve  | student affairs       | approve     | confirm_external_fact |                                      |
    And SIGTA has these memberships:
      | staff                | area                  | permissions     | scope |
      | operative@fi.uba.ar  | student affairs       | verify          |       |
      | approver@fi.uba.ar   | student affairs       | verify, approve |       |
      | commission@fi.uba.ar | curricular commission | evaluate        |       |

  Scenario: a confirmation records who, when and with what observation
    Given a "subject equivalence" procedure whose open step is "file"
    And it is "2026-11-16"
    When "operative@fi.uba.ar" confirms the "file" step with the observation "case file created"
    Then the confirmation is recorded
    And the "file" step is confirmed by "operative@fi.uba.ar" on "2026-11-16" with the observation "case file created"

  Scenario: a confirmation needs an observation
    Given a "subject equivalence" procedure whose open step is "file"
    When "operative@fi.uba.ar" confirms the "file" step with the observation ""
    Then the confirmation is rejected
    And the procedure stays on the "file" step

  Scenario: a confirmed step opens the next one
    Given a "subject equivalence" procedure whose open step is "file"
    When "operative@fi.uba.ar" confirms the "file" step
    Then the procedure moves to the "approve" step

  Scenario Outline: a decision opens the step of the chosen outcome
    Given a "subject equivalence" procedure whose open step is "check"
    When "operative@fi.uba.ar" confirms the "check" step choosing "<outcome>"
    Then the procedure moves to the "<next>" step

    Examples:
      | outcome   | next     |
      | accept    | evaluate |
      | send back | request  |

  Scenario Outline: a decision needs one of its outcomes
    Given a "subject equivalence" procedure whose open step is "check"
    When "operative@fi.uba.ar" confirms the "check" step choosing "<outcome>"
    Then the confirmation is rejected
    And the procedure stays on the "check" step

    Examples:
      | outcome    |
      |            |
      | fast track |

  Scenario: confirming the last step closes the procedure
    Given a "subject equivalence" procedure whose open step is "approve"
    When "approver@fi.uba.ar" confirms the "approve" step
    Then no step is left open

  Scenario: a confirmed step cannot be confirmed again
    Given a "subject equivalence" procedure whose open step is "file"
    And "operative@fi.uba.ar" confirmed the "file" step
    When "approver@fi.uba.ar" confirms the "file" step
    Then the confirmation is rejected because the step is not open
    And the procedure stays on the "approve" step
    And the "file" step has 1 confirmation

  Scenario: only the open step can be confirmed
    Given a "subject equivalence" procedure whose open step is "file"
    When "approver@fi.uba.ar" confirms the "approve" step
    Then the confirmation is rejected because the step is not open
    And the procedure stays on the "file" step

  Scenario: two confirmations of the same step at the same time move the procedure once
    Given a "subject equivalence" procedure whose open step is "file"
    When "operative@fi.uba.ar" and "approver@fi.uba.ar" confirm the "file" step at the same time
    Then the "file" step has 1 confirmation
    And the procedure moves to the "approve" step

  Scenario Outline: only who holds the permissions of the step in its area confirms it
    Given a "subject equivalence" procedure whose open step is "<step>"
    When "<user>" confirms the "<step>" step
    Then the confirmation is <result>

    Examples:
      | step     | user                 | result   |
      | file     | operative@fi.uba.ar  | recorded |
      | approve  | approver@fi.uba.ar   | recorded |
      | evaluate | commission@fi.uba.ar | recorded |
      | approve  | operative@fi.uba.ar  | denied   |
      | file     | commission@fi.uba.ar | denied   |
      | evaluate | approver@fi.uba.ar   | denied   |
      | file     | applicant@fi.uba.ar  | denied   |

  Scenario: a denied confirmation changes nothing
    Given a "subject equivalence" procedure whose open step is "approve"
    When "operative@fi.uba.ar" confirms the "approve" step
    Then the confirmation is denied
    And the procedure stays on the "approve" step
    And the "approve" step has 0 confirmations

  Scenario: a step done outside SIGTA is confirmed as declared by who confirms it
    Given a "subject equivalence" procedure whose open step is "file"
    When "operative@fi.uba.ar" confirms the "file" step with the observation "the case file was created in the electronic file system"
    Then the confirmation is recorded
    And the "file" step is recorded as declared by "operative@fi.uba.ar", not verified by SIGTA

  Scenario: a confirmation that fails leaves the procedure as it was
    Given a "subject equivalence" procedure whose open step is "file"
    And recording confirmations fails
    When "operative@fi.uba.ar" confirms the "file" step
    Then the confirmation is not recorded
    And the procedure stays on the "file" step
    And the "file" step has 0 confirmations

  Scenario: moving to a step of another area refers the procedure
    Given a "subject equivalence" procedure whose open step is "check"
    And it is "2026-11-16"
    When "operative@fi.uba.ar" confirms the "check" step choosing "accept" with the external file number "EX-2026-00001"
    Then the procedure is referred to "curricular commission"
    And its state is referred
    And the referral is recorded by "operative@fi.uba.ar" on "2026-11-16" to "curricular commission" with the external file number "EX-2026-00001"

  Scenario: a referral does not need an external file number
    Given a "subject equivalence" procedure whose open step is "check"
    When "operative@fi.uba.ar" confirms the "check" step choosing "accept"
    Then the procedure is referred to "curricular commission"
    And the referral is recorded without an external file number

  Scenario: a referred procedure counts the days without return
    Given a "subject equivalence" procedure whose open step is "check"
    And it is "2026-11-16"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    When it is "2026-11-20"
    Then the procedure is referred to "curricular commission" for 4 days without return
    And its detail shows it in "curricular commission", from "student affairs", for 4 days without return

  Scenario: moving to a step of the same area is not a referral
    Given a "subject equivalence" procedure whose open step is "file"
    When "operative@fi.uba.ar" confirms the "file" step
    Then the procedure moves to the "approve" step
    And the procedure is not referred

  Scenario: going back to the area it was referred from returns the procedure
    Given a "subject equivalence" procedure whose open step is "check"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    When "commission@fi.uba.ar" confirms the "evaluate" step
    Then the procedure moves to the "file" step
    And the procedure is not referred
    And its state is in an office

  Scenario: closing a referred procedure ends the referral
    Given a published "credit recognition" procedure type with this flow:
      | step     | done by               | permissions | action                | next                                 |
      | request  | applicant             |             | enter_data            | check                                |
      | check    | student affairs       | verify      | decide                | accept: evaluate, send back: request |
      | evaluate | curricular commission | evaluate    | confirm_external_fact |                                      |
    And a "credit recognition" procedure whose open step is "check"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    When "commission@fi.uba.ar" confirms the "evaluate" step
    Then no step is left open
    And the procedure is not referred
    And its state is closed
