@US34 @SIG-50
Feature: The applicant answers a correction request
  As an applicant,
  I want to complete what I am asked for on the same procedure,
  so that I do not have to start it again.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | next     | asks for                               |
      | request  | applicant             |             | enter_data            | files    | subjects                               |
      | files    | applicant             |             | upload_document       | check    | transcript optional, syllabus optional |
      | check    | student affairs       | verify      | confirm_external_fact | evaluate |                                        |
      | evaluate | curricular commission | evaluate    | confirm_external_fact |          |                                        |
    And SIGTA has these memberships:
      | staff                | area                  | permissions | scope |
      | operative@fi.uba.ar  | student affairs       | verify      |       |
      | commission@fi.uba.ar | curricular commission | evaluate    |       |
    And "applicant@fi.uba.ar" submitted a "subject equivalence" procedure on "2026-11-16"

  Scenario: the applicant sees what they are asked for and why, without who asked
    Given "operative@fi.uba.ar" requested a correction of "subjects, syllabus" because "the subject codes and the syllabus are missing"
    When "applicant@fi.uba.ar" looks at the procedure
    Then they are asked for "subjects, syllabus" because "the subject codes and the syllabus are missing", by "student affairs"

  Scenario: what the applicant loads lands on the step of theirs that asks for it
    Given "operative@fi.uba.ar" requested a correction of "subjects, syllabus" because "the subject codes and the syllabus are missing"
    When "applicant@fi.uba.ar" enters "61.03, 62.01" as the "subjects"
    And "applicant@fi.uba.ar" attaches "programa.pdf" as the "syllabus"
    Then the "subjects" of the "request" step is "61.03, 62.01"
    And the "syllabus" of the "files" step is "programa.pdf"
    And its state is with the applicant

  Scenario: the applicant loads only what they are asked for
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    When "applicant@fi.uba.ar" tries to attach "analitico.pdf" as the "transcript"
    Then the attachment is rejected

  Scenario: the applicant loads nothing while no correction is pending
    When "applicant@fi.uba.ar" tries to attach "programa.pdf" as the "syllabus"
    Then the attachment is rejected

  Scenario: the answer is not sent until everything asked for is loaded again
    Given "operative@fi.uba.ar" requested a correction of "subjects, syllabus" because "the subject codes and the syllabus are missing"
    And "applicant@fi.uba.ar" attached "programa.pdf" as the "syllabus"
    When "applicant@fi.uba.ar" sends the answer to the correction
    Then the answer is rejected because "subjects" is missing
    And its state is with the applicant

  Scenario: sending the answer takes the procedure back to staff, on the same step
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    And "applicant@fi.uba.ar" attached "programa.pdf" as the "syllabus"
    When "applicant@fi.uba.ar" sends the answer to the correction
    Then the procedure stays on the "check" step
    And its state is in an office
    And "operative@fi.uba.ar" has the procedure in their inbox

  Scenario: an answer during an evaluation goes back to the evaluator
    Given "operative@fi.uba.ar" confirmed the "check" step
    And "commission@fi.uba.ar" requested a correction of "syllabus" because "the commission asks for the syllabus"
    And "applicant@fi.uba.ar" attached "programa.pdf" as the "syllabus"
    When "applicant@fi.uba.ar" sends the answer to the correction
    Then the procedure stays on the "evaluate" step
    And the procedure is referred to "curricular commission"

  Scenario: one answer covers every pending request
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    And "operative@fi.uba.ar" requested a correction of "subjects" because "the subject codes are missing"
    And "applicant@fi.uba.ar" attached "programa.pdf" as the "syllabus"
    And "applicant@fi.uba.ar" entered "61.03" as the "subjects"
    When "applicant@fi.uba.ar" sends the answer to the correction
    Then the answer is sent
    And its state is in an office

  Scenario: only the applicant sends the answer
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    When "operative@fi.uba.ar" sends the answer to the correction
    Then the answer is denied

  Scenario: the history records the answer, with who and when
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    And it is "2026-11-18"
    And "applicant@fi.uba.ar" attached "programa.pdf" as the "syllabus"
    When "applicant@fi.uba.ar" sends the answer to the correction
    Then "operative@fi.uba.ar" sees in the history the answer of "applicant@fi.uba.ar" on "2026-11-18", on the "check" step

  Scenario: my procedures say when the applicant has a correction to make
    Given "operative@fi.uba.ar" requested a correction of "syllabus" because "the syllabus is missing"
    When "applicant@fi.uba.ar" looks at their procedures
    Then the procedure shows a pending correction
