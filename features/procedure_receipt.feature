@US38 @SIG-53
Feature: Receipt of my procedure
  As an applicant,
  I want to download a receipt of my procedure,
  so that I have proof of it outside the platform.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | asks for                     | next                                         |
      | request  | applicant             |             | enter_data            | reason                       | check                                        |
      | check    | student affairs       | verify      | decide                |                              | accept: evaluate, send back: back to request |
      | evaluate | curricular commission | evaluate    | upload_document       | minutes                      | resolve                                      |
      | resolve  | student affairs       | verify      | upload_document       | dean resolution (resolution) | record                                       |
      | record   | student affairs       | approve     | confirm_external_fact |                              |                                              |
    And SIGTA has these memberships:
      | staff                | area                  | permissions     | scope |
      | operative@fi.uba.ar  | student affairs       | verify          |       |
      | approver@fi.uba.ar   | student affairs       | verify, approve |       |
      | commission@fi.uba.ar | curricular commission | evaluate        |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: the receipt of a procedure in progress has its identifier, type, dates, state and open step
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And it is "2026-11-18"
    When "applicant@fi.uba.ar" downloads the receipt of the procedure
    Then the receipt names the procedure by its identifier
    And the receipt shows a "subject equivalence" procedure, "referred"
    And the receipt shows it started on "2026-11-16", last moved on "2026-11-17" and was issued on "2026-11-18"
    And the receipt shows the open step "evaluate", in "curricular commission"

  Scenario: the receipt of a closed procedure has its state and no open step
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And it is "2026-11-20"
    And "approver@fi.uba.ar" confirmed the "record" step
    When "applicant@fi.uba.ar" downloads the receipt of the procedure
    Then the receipt shows a "subject equivalence" procedure, "closed"
    And the receipt shows it started on "2026-11-16", last moved on "2026-11-20" and was issued on "2026-11-20"
    And the receipt shows no open step

  Scenario: the receipt has the route of the procedure with the date of each step, as the applicant sees it
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept" with the observation "falta el programa" and the external file number "EX-2026-001"
    And it is "2026-11-20"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    When "applicant@fi.uba.ar" downloads the receipt of the procedure
    Then the receipt shows the route, in this order:
      | step     | done by               | on         |
      | request  | applicant@fi.uba.ar   | 2026-11-16 |
      | check    | student affairs       | 2026-11-17 |
      | evaluate | curricular commission | 2026-11-20 |
    And the receipt does not show "falta el programa", "EX-2026-001", "acta.pdf" nor "operative@fi.uba.ar"

  Scenario: the route keeps what a correction request asked for and why, and the reason of a send back
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" requested a correction of "reason" because "el motivo no es claro"
    And "applicant@fi.uba.ar" entered "Cursé la materia en otra facultad" as the "reason"
    And "applicant@fi.uba.ar" sent the answer to the correction
    And it is "2026-11-18"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "send back" with the observation "falta el programa"
    When "applicant@fi.uba.ar" downloads the receipt of the procedure
    Then the receipt shows the route, in this order:
      | step    | done by             | on         | observation           | correction |
      | request | applicant@fi.uba.ar | 2026-11-16 |                       |            |
      | check   | student affairs     | 2026-11-17 | el motivo no es claro | reason     |
      | request | applicant@fi.uba.ar | 2026-11-17 |                       |            |
      | check   | applicant@fi.uba.ar | 2026-11-17 |                       |            |
      | check   | student affairs     | 2026-11-18 | falta el programa     |            |
    And the receipt does not show "operative@fi.uba.ar"

  Scenario Outline: only the applicant downloads the receipt of their procedure
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    When "<person>" downloads the receipt of the procedure
    Then there is no receipt for them

    Examples:
      | person               |
      | operative@fi.uba.ar  |
      | commission@fi.uba.ar |
      | other@fi.uba.ar      |
