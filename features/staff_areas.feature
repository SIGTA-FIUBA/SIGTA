@US07 @SIG-14
Feature: Staff areas
  As an administrator,
  I want to create and delete the areas of staff,
  so that SIGTA follows how the faculty is organized without changing code.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And "admin@fi.uba.ar" is an administrator

  Scenario: an administrator creates an area
    When "admin@fi.uba.ar" creates the area "degree office" named "Degree Office"
    Then the areas are "student affairs", "curricular commission" and "Degree Office"
    And "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "degree office"

  Scenario: a procedure type can use an area once it is created
    When "admin@fi.uba.ar" creates the area "degree office" named "Degree Office"
    Then a procedure type with a step in "degree office" can be published

  Scenario: only an administrator creates areas
    Given SIGTA has these memberships:
      | staff               | area            | profile        | scope |
      | approver@fi.uba.ar  | student affairs | staff_approver |       |
    When "approver@fi.uba.ar" creates the area "degree office" named "Degree Office"
    Then the area is denied

  Scenario Outline: an area needs a valid and unused identifier and name
    When "admin@fi.uba.ar" creates the area "<id>" named "<name>"
    Then the area is not created

    Examples:
      | id              | name                  |
      |                 | Degree Office         |
      | Degree Office   | Degree Office         |
      | 1st_office      | Degree Office         |
      | student_affairs | Degree Office         |
      | degree_office   |                       |
      | degree_office   | Curricular Commission |

  Scenario: an administrator deletes an area without memberships
    When "admin@fi.uba.ar" deletes the area "curricular commission"
    Then the areas are "student affairs"
    And "admin@fi.uba.ar" cannot grant "new@fi.uba.ar" a membership in "curricular commission"

  Scenario: an area with memberships that have not ended cannot be deleted
    Given SIGTA has these memberships:
      | staff               | area            | profile         | scope |
      | operative@fi.uba.ar | student affairs | staff_operative |       |
    When "admin@fi.uba.ar" deletes the area "student affairs"
    Then the area is not deleted
    And the areas are "student affairs" and "curricular commission"

  Scenario: an area can be deleted once its memberships ended
    Given SIGTA has these memberships:
      | staff               | area            | profile         | scope |
      | operative@fi.uba.ar | student affairs | staff_operative |       |
    When "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    And "admin@fi.uba.ar" deletes the area "student affairs"
    Then the areas are "curricular commission"
    And the area "student affairs" records that "admin@fi.uba.ar" deleted it

  Scenario: the identifier of a deleted area is not reused
    When "admin@fi.uba.ar" deletes the area "curricular commission"
    And "admin@fi.uba.ar" creates the area "curricular commission" named "Another Commission"
    Then the area is not created

  Scenario: only an administrator deletes areas
    When "operative@fi.uba.ar" deletes the area "curricular commission"
    Then the area is denied
    And the areas are "student affairs" and "curricular commission"
