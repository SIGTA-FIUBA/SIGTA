@US72 @SIG-74
Feature: Edit and publish a procedure type
  As the team,
  I want to edit a procedure type in SIGTA and publish it with the approval of another person,
  so that procedure types are loaded without code and no single person changes one.

  Background:
    Given the staff areas for the drafts are "student affairs" and "curricular commission"
    And "ana" can propose procedure types, "beto" can approve them and "eva" can do both

  Scenario: a new procedure type is published when another person approves it
    Given "ana" created the draft of a new procedure type "subject equivalence"
    When "ana" proposes the draft
    And "beto" approves the draft
    Then version 1 of the new procedure type is published, proposed by "ana" and approved by "beto"
    And the procedure type has no draft left

  Scenario: the draft of a new version starts from the latest version
    Given version 1 of "subject equivalence" is published without a draft
    When "ana" opens a draft of "subject equivalence"
    Then the draft has the definition of version 1

  Scenario: a new version leaves the published one as it was
    Given version 1 of "subject equivalence" is published without a draft
    And "ana" proposed a draft of "subject equivalence" with another document name
    When "beto" approves the draft
    Then version 2 of "subject equivalence" is published from the draft
    And version 1 keeps its content

  Scenario: whoever edited the draft does not approve it, even if another person proposed it
    Given "eva" created the draft of a new procedure type "subject equivalence"
    When "ana" proposes the draft
    And "eva" approves the draft
    Then no version is published from the draft

  Scenario: a proposed draft is not edited
    Given "ana" proposed the draft of a new procedure type "subject equivalence"
    When "ana" edits the draft
    Then the change is refused because the draft is proposed

  Scenario: a rejected proposal goes back to editing with its reason
    Given "ana" proposed the draft of a new procedure type "subject equivalence"
    When "beto" rejects the draft because "Falta el plazo de presentación"
    Then the draft can be edited again
    And the draft shows that "beto" rejected it because "Falta el plazo de presentación"

  Scenario: an invalid definition is not proposed
    Given "ana" created the draft of a new procedure type with a step in an unknown area
    When "ana" proposes the draft
    Then the draft is not proposed
    And the problem points at the step in the unknown area

  Scenario: a draft opened from an outdated version is not published
    Given version 1 of "subject equivalence" is published without a draft
    And "ana" proposed a draft of "subject equivalence" with another document name
    And version 2 of "subject equivalence" is published by other means
    When "beto" approves the draft
    Then no version is published from the draft
    And the approval is refused because another version was published since

  Scenario Outline: proposing a draft asks for the permission to propose
    Given "ana" created the draft of a new procedure type "subject equivalence"
    When "<person>" proposes the draft
    Then the draft is not proposed

    Examples:
      | person |
      | beto   |
      | dani   |

  Scenario Outline: only another person with the permission approves
    Given "<proposer>" proposed the draft of a new procedure type "subject equivalence"
    When "<approver>" approves the draft
    Then no version is published from the draft

    Examples:
      | proposer | approver |
      | ana      | ana      |
      | ana      | dani     |
      | eva      | eva      |
      | eva      | Eva      |
