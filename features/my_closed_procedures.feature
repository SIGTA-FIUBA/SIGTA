@US24 @SIG-35
Feature: My closed procedures
  As an applicant,
  I want to see my closed procedures,
  so that I recover a resolution without looking for an old email.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | asks for                     | next    |
      | request  | applicant             |             | enter_data            | reason                       | check   |
      | check    | student affairs       | verify      | confirm_external_fact |                              | resolve |
      | resolve  | student affairs       | verify      | upload_document       | dean resolution (resolution) | record  |
      | record   | student affairs       | approve     | confirm_external_fact |                              |         |
    And SIGTA has these memberships:
      | staff               | area            | permissions     | scope |
      | operative@fi.uba.ar | student affairs | verify          |       |
      | approver@fi.uba.ar  | student affairs | verify, approve |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: a closed procedure is listed with the date it closed and its resolution
    Given "operative@fi.uba.ar" confirmed the "check" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And it is "2026-11-25"
    And "approver@fi.uba.ar" confirmed the "record" step
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see these closed procedures:
      | procedure type      | closed on  | resolution      | file           |
      | subject equivalence | 2026-11-25 | dean resolution | resolucion.pdf |

  Scenario: only the last file attached as the resolution is listed
    Given "operative@fi.uba.ar" confirmed the "check" step
    And "operative@fi.uba.ar" attached "borrador.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And it is "2026-11-25"
    And "approver@fi.uba.ar" confirmed the "record" step
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see these closed procedures:
      | procedure type      | closed on  | resolution      | file           |
      | subject equivalence | 2026-11-25 | dean resolution | resolucion.pdf |

  Scenario: a procedure in progress is not listed as closed, even with its resolution attached
    Given "operative@fi.uba.ar" confirmed the "check" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see no closed procedures

  Scenario: the applicant downloads the resolution listed with their closed procedure
    Given "operative@fi.uba.ar" confirmed the "check" step
    And "operative@fi.uba.ar" attached "resolucion.pdf" as the "dean resolution"
    And "operative@fi.uba.ar" confirmed the "resolve" step
    And "approver@fi.uba.ar" confirmed the "record" step
    When "applicant@fi.uba.ar" looks at their procedures
    Then "applicant@fi.uba.ar" can download the resolution of each closed procedure
    And "applicant@fi.uba.ar" can open the detail of each one
