@SIG-84
Feature: Offer only the actions each person can take on a procedure type
  As the team,
  I want the screen that edits procedure types to offer each person only what they can do,
  so that nobody is offered to approve a change they made, or to edit without the permission.

  Background:
    Given the staff areas for the drafts are "student affairs" and "curricular commission"
    And "ana" can propose procedure types, "beto" can approve them and "eva" can do both

  Scenario Outline: a draft being edited is offered to edit only to whoever can, and to approve to nobody
    Given "<creator>" created the draft of a new procedure type "subject equivalence"
    When "<person>" looks at the draft
    Then they are offered to edit it: <edit>, and to approve or reject it: no

    Examples:
      | creator | person | edit |
      | ana     | ana    | yes  |
      | ana     | beto   | no   |
      | ana     | eva    | yes  |

  Scenario Outline: a proposed draft is not offered to edit, and is offered to approve only to another person with the permission
    Given "<proposer>" proposed the draft of a new procedure type "subject equivalence"
    When "<person>" looks at the draft
    Then they are offered to edit it: no, and to approve or reject it: <review>

    Examples:
      | proposer | person | review |
      | ana      | ana    | no     |
      | ana      | beto   | yes    |
      | ana      | eva    | yes    |
      | eva      | eva    | no     |

  Scenario Outline: only whoever can edit creates a procedure type or opens a draft
    Given version 1 of "subject equivalence" is published without a draft
    When "<person>" looks at the procedure types they can edit
    Then they <create> create a procedure type and <open> open a draft of "subject equivalence"

    Examples:
      | person | create | open   |
      | ana    | can    | can    |
      | beto   | cannot | cannot |
