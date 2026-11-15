@US07 @SIG-14
Feature: Staff areas
  As an administrator,
  I want to create and delete the areas of staff, their permissions and roles,
  so that SIGTA follows how the faculty is organized without changing code.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And "admin@fi.uba.ar" is an administrator

  Scenario: an administrator creates an area
    When "admin@fi.uba.ar" creates the area "degree office" named "Degree Office"
    And "admin@fi.uba.ar" creates the permission "sign" named "Sign" in "degree office"
    Then the areas besides the Secretariat are "student affairs", "curricular commission" and "Degree Office"
    And "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "degree office" with the permission "sign"

  Scenario: a procedure type can use an area once it is created
    When "admin@fi.uba.ar" creates the area "degree office" named "Degree Office"
    And "admin@fi.uba.ar" creates the permission "sign" named "Sign" in "degree office"
    Then a procedure type with a step in "degree office" that asks for "sign" can be published

  Scenario: only an administrator creates areas
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | approver@fi.uba.ar  | student affairs | verify      |       |
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
    Then the areas besides the Secretariat are "student affairs"
    And "admin@fi.uba.ar" cannot grant "new@fi.uba.ar" a membership in "curricular commission" with the permission "evaluate"

  Scenario: an area with memberships that have not ended cannot be deleted
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" deletes the area "student affairs"
    Then the area is not deleted
    And the areas besides the Secretariat are "student affairs" and "curricular commission"

  Scenario: an area can be deleted once its memberships ended
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    And "admin@fi.uba.ar" deletes the area "student affairs"
    Then the areas besides the Secretariat are "curricular commission"
    And the area "student affairs" records that "admin@fi.uba.ar" deleted it

  Scenario: the identifier of a deleted area is not reused
    When "admin@fi.uba.ar" deletes the area "curricular commission"
    And "admin@fi.uba.ar" creates the area "curricular commission" named "Another Commission"
    Then the area is not created

  Scenario: only an administrator deletes areas
    When "operative@fi.uba.ar" deletes the area "curricular commission"
    Then the area is denied
    And the areas besides the Secretariat are "student affairs" and "curricular commission"

  Scenario: an administrator creates a permission in an area
    When "admin@fi.uba.ar" creates the permission "sign" named "Sign" in "student affairs"
    Then "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "student affairs" with the permission "sign"
    And "admin@fi.uba.ar" cannot grant "other@fi.uba.ar" a membership in "curricular commission" with the permission "sign"

  Scenario: only an administrator creates permissions
    When "operative@fi.uba.ar" creates the permission "sign" named "Sign" in "student affairs"
    Then the permission is denied

  Scenario Outline: a permission needs a valid identifier and a name unused in its area
    When "admin@fi.uba.ar" creates the permission "<id>" named "<name>" in "student affairs"
    Then the permission is not created

    Examples:
      | id     | name   |
      |        | Sign   |
      | Sign   | Sign   |
      | verify | Review |
      | sign   | verify |
      | sign   |        |

  Scenario: two areas can have a permission with the same identifier
    When "admin@fi.uba.ar" creates the permission "verify" named "verify" in "curricular commission"
    Then "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "curricular commission" with the permission "verify"

  Scenario: a permission held by a membership that has not ended cannot be deleted
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" deletes the permission "verify" of "student affairs"
    Then the permission is not deleted

  Scenario: a deleted permission can no longer be granted
    When "admin@fi.uba.ar" deletes the permission "verify" of "student affairs"
    Then "admin@fi.uba.ar" cannot grant "new@fi.uba.ar" a membership in "student affairs" with the permission "verify"

  Scenario: the identifier of a deleted permission is not reused
    When "admin@fi.uba.ar" deletes the permission "verify" of "student affairs"
    And "admin@fi.uba.ar" creates the permission "verify" named "Verify again" in "student affairs"
    Then the permission is not created

  Scenario: SIGTA brings the Secretariat with the permissions to edit and publish procedure types
    Then the area "secretariat" exists without anyone creating it
    And "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "secretariat" with the permissions "edit_procedure_types, publish_procedure_types"

  Scenario: the Secretariat cannot be deleted
    When "admin@fi.uba.ar" deletes the area "secretariat"
    Then the area is not deleted

  Scenario: another area cannot take the identifier of the Secretariat
    When "admin@fi.uba.ar" creates the area "secretariat" named "Another Secretariat"
    Then the area is not created

  Scenario Outline: the permissions to edit and publish procedure types cannot be deleted
    When "admin@fi.uba.ar" deletes the permission "<permission>" of "secretariat"
    Then the permission is not deleted

    Examples:
      | permission              |
      | edit_procedure_types    |
      | publish_procedure_types |

  Scenario: the Secretariat takes permissions for its steps like any other area
    When "admin@fi.uba.ar" creates the permission "register" named "Register" in "secretariat"
    Then "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "secretariat" with the permission "register"
    And a procedure type with a step in "secretariat" that asks for "register" can be published

  Scenario: an administrator creates a role
    When "admin@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    Then "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "student affairs" with the role "clerk"
    And "admin@fi.uba.ar" cannot grant "other@fi.uba.ar" a membership in "curricular commission" with the role "clerk"

  Scenario: a role in the Secretariat can group the permissions to edit and publish procedure types
    When "admin@fi.uba.ar" creates the role "setup" named "Setup" with "edit_procedure_types, publish_procedure_types" in "secretariat"
    Then "admin@fi.uba.ar" can grant "new@fi.uba.ar" a membership in "secretariat" with the role "setup"

  Scenario: only an administrator creates roles
    When "operative@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    Then the role is denied

  Scenario Outline: a role needs a valid and unused identifier, a name and permissions of its area
    Given "admin@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    When "admin@fi.uba.ar" creates the role "<id>" named "<name>" with "<permissions>" in "student affairs"
    Then the role is not created

    Examples:
      | id       | name     | permissions    |
      |          | Reviewer | verify         |
      | Reviewer | Reviewer | verify         |
      | clerk    | Reviewer | verify         |
      | reviewer | Clerk    | verify         |
      | reviewer |          | verify         |
      | reviewer | Reviewer |                |
      | reviewer | Reviewer | evaluate       |
      | reviewer | Reviewer | verify, verify |

  Scenario: a role held by a membership that has not ended cannot be deleted
    Given "admin@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    And "admin@fi.uba.ar" grants "operative@fi.uba.ar" a membership in "student affairs" with the role "clerk"
    When "admin@fi.uba.ar" deletes the role "clerk" of "student affairs"
    Then the role is not deleted

  Scenario: a deleted role can no longer be granted, and its identifier is not reused
    Given "admin@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    When "admin@fi.uba.ar" deletes the role "clerk" of "student affairs"
    Then "admin@fi.uba.ar" cannot grant "new@fi.uba.ar" a membership in "student affairs" with the role "clerk"
    And "admin@fi.uba.ar" creates the role "clerk" named "Clerk again" with "verify" in "student affairs"
    And the role is not created

  Scenario: a permission that is part of a role cannot be deleted
    Given "admin@fi.uba.ar" creates the role "clerk" named "Clerk" with "verify" in "student affairs"
    When "admin@fi.uba.ar" deletes the permission "verify" of "student affairs"
    Then the permission is not deleted
