@US12 @SIG-44
Feature: Send back a step
  As staff of an area,
  I want to send a procedure back to an earlier step stating the reason,
  so that something can be fixed without closing the procedure.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | next                                         |
      | request  | applicant             |             | enter_data            | check                                        |
      | check    | student affairs       | verify      | decide                | accept: evaluate, send back: back to request |
      | evaluate | curricular commission | evaluate    | confirm_external_fact |                                              |
    And SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: a send back needs a reason
    When "operative@fi.uba.ar" confirms the "check" step choosing "send back" with the observation ""
    Then the confirmation is rejected
    And the procedure stays on the "check" step

  Scenario: a send back returns the procedure to the step of its configuration without closing it
    When "operative@fi.uba.ar" confirms the "check" step choosing "send back" with the observation "missing signature"
    Then the procedure moves to the "request" step
    And its state is with the applicant

  Scenario Outline: the history records the send back with who, when and the reason, for whoever can see it
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "send back" with the observation "missing signature"
    When "<person>" reads the history of the procedure
    Then the history shows, in this order:
      | step    | done by             | on         | observation       | outcome   | send back |
      | request | applicant@fi.uba.ar | 2026-11-16 |                   |           |           |
      | check   | <sent back by>      | 2026-11-17 | missing signature | send back | yes       |

    # The applicant sees the area of a staff step, not the person (SIG-31).
    Examples:
      | person              | sent back by        |
      | applicant@fi.uba.ar |                     |
      | operative@fi.uba.ar | operative@fi.uba.ar |

  Scenario: a decision that moves the procedure forward is not a send back
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept" with the observation "complete request"
    When "operative@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step    | observation      | outcome | send back |
      | request |                  |         |           |
      | check   | complete request | accept  |           |
