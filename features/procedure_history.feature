@US11 @SIG-21
Feature: History of a procedure
  As staff of an area,
  I want to see everything that happened in a procedure,
  so that I know where it is and who had it.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | asks for      | next                                 |
      | request  | applicant             |             | enter_data            | reason        | check                                |
      | check    | student affairs       | verify      | decide                |               | accept: evaluate, send back: request |
      | evaluate | curricular commission | evaluate    | confirm_external_fact |               | ruling                               |
      | ruling   | student affairs       | verify      | upload_document       | ruling        | record                               |
      | record   | student affairs       | approve     | enter_data            | record number |                                      |
    And SIGTA has these memberships:
      | staff                | area                  | permissions     | scope |
      | operative@fi.uba.ar  | student affairs       | verify          |       |
      | approver@fi.uba.ar   | student affairs       | verify, approve |       |
      | commission@fi.uba.ar | curricular commission | evaluate        |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: the history shows each step with who did it, when and the observation, in order
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept" with the observation "complete request"
    And it is "2026-11-20"
    And "commission@fi.uba.ar" confirmed the "evaluate" step with the observation "ruling received"
    When "operative@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step     | done by              | on         | observation      | outcome |
      | request  | applicant@fi.uba.ar  | 2026-11-16 |                  |         |
      | check    | operative@fi.uba.ar  | 2026-11-17 | complete request | accept  |
      | evaluate | commission@fi.uba.ar | 2026-11-20 | ruling received  |         |

  Scenario: every kind of step the engine records shows in the history
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" confirmed the "ruling" step
    And "approver@fi.uba.ar" confirmed the "record" step
    When "approver@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step     | action                |
      | request  | enter_data            |
      | check    | decide                |
      | evaluate | confirm_external_fact |
      | ruling   | upload_document       |
      | record   | enter_data            |

  @SIG-81
  Scenario: a document attached after the start shows in the history with who attached it, when and what it covers
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And it is "2026-11-20"
    When "operative@fi.uba.ar" uploads "dictamen.pdf" as the "ruling"
    And "approver@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step     | done by              | on         | document | file         |
      | request  | applicant@fi.uba.ar  | 2026-11-16 |          |              |
      | check    | operative@fi.uba.ar  | 2026-11-16 |          |              |
      | evaluate | commission@fi.uba.ar | 2026-11-16 |          |              |
      | ruling   | operative@fi.uba.ar  | 2026-11-20 | ruling   | dictamen.pdf |

  Scenario: what the history shows stays as it was when the procedure moves on
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "operative@fi.uba.ar" read the history of the procedure
    When "commission@fi.uba.ar" confirms the "evaluate" step
    And "operative@fi.uba.ar" reads the history of the procedure
    Then the history starts with what it showed before
    And then shows the "evaluate" step done by "commission@fi.uba.ar"

  Scenario: changing what was read of the history does not change it
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "operative@fi.uba.ar" read the history of the procedure
    When they change what they read of the history
    And "operative@fi.uba.ar" reads the history of the procedure
    Then the history starts with what it showed before

  Scenario Outline: the history is read in the detail, by whoever can see the procedure
    When "<person>" reads the history of the procedure
    Then <result>

    Examples:
      | person               | result                               |
      | applicant@fi.uba.ar  | the history shows the "request" step |
      | operative@fi.uba.ar  | the history shows the "request" step |
      | commission@fi.uba.ar | the history shows the "request" step |
      | other@fi.uba.ar      | they cannot read the history         |
