@US33 @SIG-49
Feature: Request a correction from the applicant
  As staff of the Secretariat,
  I want to ask the applicant for something specific without closing or restarting the procedure,
  so that the documentation can be completed.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | next     | asks for |
      | request  | applicant             |             | enter_data            | check    | subjects |
      | check    | student affairs       | verify      | confirm_external_fact | evaluate |          |
      | evaluate | curricular commission | evaluate    | confirm_external_fact |          |          |
    And SIGTA has these memberships:
      | staff                | area                  | permissions | scope |
      | operative@fi.uba.ar  | student affairs       | verify      |       |
      | commission@fi.uba.ar | curricular commission | evaluate    |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: a correction request says what is missing and why
    When "operative@fi.uba.ar" requests a correction of "subjects" because "the subject codes are missing"
    Then the correction request is recorded
    And the applicant is asked to correct "subjects" because "the subject codes are missing"

  Scenario Outline: a correction request needs what is missing and why
    When "operative@fi.uba.ar" requests a correction of "<requirements>" because "<reason>"
    Then the correction request is rejected
    And its state is in an office

    Examples:
      | requirements | reason                        |
      | subjects     |                               |
      |              | the subject codes are missing |
      | grades       | not a requirement of the type |

  Scenario: the procedure waits for the applicant without leaving its step
    When "operative@fi.uba.ar" requests a correction of "subjects" because "the subject codes are missing"
    Then the procedure moves to the "check" step
    And its state is with the applicant

  Scenario: the step is not confirmed while the applicant has a correction to make
    Given "operative@fi.uba.ar" requested a correction of "subjects" because "the subject codes are missing"
    When "operative@fi.uba.ar" confirms the "check" step
    Then the confirmation is rejected
    And the procedure moves to the "check" step

  Scenario: the procedure leaves the inbox of whoever acts on the step until the applicant answers
    When "operative@fi.uba.ar" requests a correction of "subjects" because "the subject codes are missing"
    Then the procedure leaves the inbox of "operative@fi.uba.ar"

  Scenario: only whoever can act on the open step requests a correction
    When "commission@fi.uba.ar" requests a correction of "subjects" because "the subject codes are missing"
    Then the correction request is denied

  Scenario: a request of the external evaluator reaches the applicant as a correction
    Given "operative@fi.uba.ar" confirmed the "check" step
    When "commission@fi.uba.ar" requests a correction of "subjects" because "the commission asks for the syllabus of each subject"
    Then the correction request is recorded
    And the applicant is asked to correct "subjects" because "the commission asks for the syllabus of each subject"
    And the procedure is referred to "curricular commission"
    And its state is with the applicant

  Scenario: more than one correction can be requested on the same procedure
    Given "operative@fi.uba.ar" requested a correction of "subjects" because "the subject codes are missing"
    When "operative@fi.uba.ar" requests a correction of "subjects" because "the grades are missing too"
    Then the correction request is recorded
    And the applicant is asked to correct "subjects" because "the subject codes are missing"
    And the applicant is asked to correct "subjects" because "the grades are missing too"

  Scenario Outline: the history records each correction request with who, when, what and why
    Given it is "2026-11-17"
    And "operative@fi.uba.ar" requested a correction of "subjects" because "the subject codes are missing"
    When "<person>" reads the history of the procedure
    Then the history shows, in this order:
      | step    | done by             | on         | observation                   | correction |
      | request | applicant@fi.uba.ar | 2026-11-16 |                               |            |
      | check   | <requested by>      | 2026-11-17 | the subject codes are missing | subjects   |

    # The applicant sees what is asked and why, but not who asked (SIG-31).
    Examples:
      | person              | requested by        |
      | applicant@fi.uba.ar |                     |
      | operative@fi.uba.ar | operative@fi.uba.ar |
