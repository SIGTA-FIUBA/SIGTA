@US08 @SIG-12
Feature: Define and publish a procedure type
  As the team,
  I want to define a procedure type as configuration,
  so that a new procedure type needs no code changes.

  Background:
    Given the staff areas are "student affairs" and "curricular commission"
    And "ana" proposes and "beto" approves every publication

  Scenario: publishing a valid definition creates its first version
    When a valid "subject equivalence" definition is published from version 0
    Then version 1 of "subject equivalence" is published
    And it records who proposed it, who approved it and when it was published

  Scenario: publishing the same content again creates no version
    Given version 1 of "subject equivalence" is published
    When the same definition is published from version 1
    Then the latest version is still 1

  Scenario: publishing the same content by other people records nothing new
    Given version 1 of "subject equivalence" was proposed by "ana" and approved by "beto"
    When "carla" proposes and "dario" approves the same definition
    Then the latest version is still 1
    And it still records "ana" and "beto"

  Scenario: any visible change creates a new version
    Given version 1 of "subject equivalence" is published
    When the name of a document is corrected and the definition is published from version 1
    Then version 2 of "subject equivalence" is published

  Scenario: a definition edited from an outdated version is rejected
    Given version 2 of "subject equivalence" is published
    When a different definition is published from version 1
    Then no version is published
    And whoever published it is asked to start again from version 2

  Scenario Outline: the same person cannot propose and approve a version
    When "ana" proposes and "<approver>" approves a "subject equivalence" definition
    Then no version is published

    Examples:
      | approver |
      | ana      |
      | Ana      |

  Scenario Outline: an invalid definition is not published
    When a "subject equivalence" definition with <problem> is published
    Then no version is published
    And the problem is reported

    Examples:
      | problem                                  |
      | a step with an action outside the four   |
      | a step in an unknown area                |
      | a reference to an undeclared document    |
      | a step unreachable from the first one    |
      | a step that never reaches a closing step |
      | a field no step uses                     |
      | a staff step without a profile           |
      | an applicant step in a staff area        |

  Scenario: a procedure in progress keeps its version
    Given a procedure that started with version 1 of "subject equivalence"
    When version 2 is published
    Then the procedure still follows version 1
    And new procedures follow version 2

  Scenario: renaming an area does not change a published version
    Given version 1 of "subject equivalence" is published
    When the "student affairs" area is renamed
    Then version 1 still shows the previous name of the area

  Scenario: an altered version is not used
    Given the stored content of version 1 no longer matches its digest
    When a procedure asks for version 1
    Then the version is not returned
    And nothing is published on top of it
