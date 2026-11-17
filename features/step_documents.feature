@SIG-82
Feature: A step is not confirmed without its documents
  As the Secretariat,
  I want a step that asks for a document not to be confirmed without it,
  so that every closed procedure has its documentation.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step       | done by               | permissions | action                | asks for                                           | next       |
      | request    | applicant             |             | enter_data            |                                                    | evaluate   |
      | evaluate   | curricular commission | evaluate    | upload_document       | commission record                                  | resolution |
      | resolution | student affairs       | verify      | upload_document       | draft resolution, signed note, cover note optional | file       |
      | file       | student affairs       | verify      | confirm_external_fact |                                                    |            |
    And SIGTA has these memberships:
      | staff                | area                  | permissions | scope |
      | operative@fi.uba.ar  | student affairs       | verify      |       |
      | commission@fi.uba.ar | curricular commission | evaluate    |       |

  Scenario: a step cannot be confirmed while its mandatory document is missing
    Given a "subject equivalence" procedure whose open step is "evaluate"
    When "commission@fi.uba.ar" confirms the "evaluate" step
    Then the confirmation is rejected because "commission record" is missing
    And the procedure stays on the "evaluate" step

  Scenario: with its document attached, the step is confirmed
    Given a "subject equivalence" procedure whose open step is "evaluate"
    And the "commission record" document of the "evaluate" step is attached
    When "commission@fi.uba.ar" confirms the "evaluate" step
    Then the confirmation is recorded
    And the procedure moves to the "resolution" step

  Scenario Outline: the rejection names every mandatory document that is missing
    Given a "subject equivalence" procedure whose open step is "resolution"
    And the "<attached>" documents of the "resolution" step are attached
    When "operative@fi.uba.ar" confirms the "resolution" step
    Then the confirmation is rejected because "<missing>" is missing
    And the procedure stays on the "resolution" step

    Examples:
      | attached         | missing                        |
      |                  | draft resolution, signed note  |
      | draft resolution | signed note                    |
      | cover note       | draft resolution, signed note  |

  Scenario: an optional document does not hold back the confirmation
    Given a "subject equivalence" procedure whose open step is "resolution"
    And the "draft resolution, signed note" documents of the "resolution" step are attached
    When "operative@fi.uba.ar" confirms the "resolution" step
    Then the confirmation is recorded
    And the procedure moves to the "file" step

  Scenario: a document attached to another step does not cover the open one
    Given a "subject equivalence" procedure whose open step is "resolution"
    And the "draft resolution, signed note" documents of the "evaluate" step are attached
    When "operative@fi.uba.ar" confirms the "resolution" step
    Then the confirmation is rejected because "draft resolution, signed note" is missing
