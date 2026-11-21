@SIG-84
Feature: Offer only the actions each person can take on a procedure type
  As the team,
  I want the screen that edits procedure types to offer each person only what they can do,
  so that nobody is offered to approve a change they made, or to edit without the permission.

  Background:
    Given the staff areas for the drafts are "student affairs" and "curricular commission"
    And "ana" can propose procedure types, "beto" can approve them and "eva" can do both

  Scenario Outline: a proposed draft offers to approve it only to another person with the permission
    Given "<proposer>" proposed the draft of a new procedure type "subject equivalence"
    When "<person>" looks at the draft
    Then they are offered to edit it: <edit>, and to approve or reject it: <review>

    Examples:
      | proposer | person | edit | review |
      | ana      | ana    | yes  | no     |
      | ana      | beto   | no   | yes    |
      | ana      | eva    | yes  | yes    |
      | eva      | eva    | yes  | no     |

  Scenario Outline: only whoever can edit creates a procedure type or opens a draft
    Given version 1 of "subject equivalence" is published without a draft
    When "<person>" looks at the procedure types they can edit
    Then they <offer> create a procedure type and <offer> open a draft of "subject equivalence"

    Examples:
      | person | offer  |
      | ana    | can    |
      | beto   | cannot |
