@US23 @SIG-31
Feature: Detail of my procedure
  As an applicant,
  I want to see which step my procedure is on and what is expected now,
  so that I do not depend on an email to know it.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | asks for                     | next                                 |
      | request  | applicant             |             | enter_data            | reason                       | check                                |
      | check    | student affairs       | verify      | decide                |                              | accept: evaluate, send back: request |
      | evaluate | curricular commission | evaluate    | upload_document       | minutes                      | resolve                              |
      | resolve  | student affairs       | verify      | upload_document       | dean resolution (resolution) | record                               |
      | record   | student affairs       | approve     | confirm_external_fact |                              |                                      |
    And SIGTA has these memberships:
      | staff                | area                  | permissions     | scope |
      | operative@fi.uba.ar  | student affairs       | verify          |       |
      | approver@fi.uba.ar   | student affairs       | verify, approve |       |
      | commission@fi.uba.ar | curricular commission | evaluate        |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: the applicant sees which step is open, which area has it and that it is out for evaluation
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    When "applicant@fi.uba.ar" looks at the procedure
    Then they see the open step "evaluate", in "curricular commission"
    And they see the procedure is "referred"
    And they see it is referred to "curricular commission" from "student affairs"

  Scenario: the applicant sees the whole route with the date of each step
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And it is "2026-11-20"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    When "applicant@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step     | on         | document |
      | request  | 2026-11-16 |          |
      | check    | 2026-11-17 |          |
      | evaluate | 2026-11-20 |          |

  Scenario Outline: who did each step of staff, the outcome, the observations and the external file number stay with staff
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept" with the observation "falta el programa" and the external file number "EX-2026-001"
    When "<person>" reads the history of the procedure
    Then the history shows, in this order:
      | step    | done by             | outcome   | observation   | external file number |
      | request | applicant@fi.uba.ar |           |               |                      |
      | check   | <done by>           | <outcome> | <observation> | <file number>        |

    Examples:
      | person              | done by             | outcome | observation       | file number |
      | applicant@fi.uba.ar |                     |         |                   |             |
      | operative@fi.uba.ar | operative@fi.uba.ar | accept  | falta el programa | EX-2026-001 |

  Scenario: a member of staff sees their own procedure as its applicant
    Given "operative@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"
    And "approver@fi.uba.ar" confirmed the "check" step choosing "accept" with the observation "falta el programa" and the external file number "EX-2026-001"
    When "operative@fi.uba.ar" reads the history of the procedure
    Then the history shows, in this order:
      | step    | done by             | observation | external file number |
      | request | operative@fi.uba.ar |             |                      |
      | check   |                     |             |                      |

  Scenario: the applicant does not see the documents of staff while the procedure is open
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    When "applicant@fi.uba.ar" looks at the procedure
    And "applicant@fi.uba.ar" reads the history of the procedure
    Then they see no files
    And the history shows, in this order:
      | step     | document |
      | request  |          |
      | check    |          |
      | evaluate |          |

  Scenario: the applicant sees the final state and the resolution once the procedure closes
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And "approver@fi.uba.ar" confirmed the "record" step
    When "applicant@fi.uba.ar" looks at the procedure
    And "applicant@fi.uba.ar" reads the history of the procedure
    Then they see the procedure is "closed"
    And they see these files:
      | step    | document        | file           |
      | resolve | dean resolution | resolucion.pdf |
    And the history shows, in this order:
      | step     | document        |
      | request  |                 |
      | check    |                 |
      | evaluate |                 |
      | resolve  | dean resolution |
      | resolve  |                 |
      | record   |                 |

  Scenario Outline: once the procedure closes, the applicant downloads the resolution and nothing else of staff
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And "approver@fi.uba.ar" confirmed the "record" step
    When "<person>" downloads "<file>" from the procedure
    Then <result>

    Examples:
      | person              | file           | result                 |
      | applicant@fi.uba.ar | resolucion.pdf | they get the file      |
      | applicant@fi.uba.ar | acta.pdf       | the file is not found  |
      | operative@fi.uba.ar | acta.pdf       | they get the file      |

  Scenario: before the procedure closes, the applicant cannot download the resolution
    Given "operative@fi.uba.ar" confirmed the "check" step choosing "accept"
    And "commission@fi.uba.ar" attached "acta.pdf" as the "minutes"
    And "commission@fi.uba.ar" confirmed the "evaluate" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    When "applicant@fi.uba.ar" downloads "resolucion.pdf" from the procedure
    Then the file is not found
