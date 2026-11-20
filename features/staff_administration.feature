@US74 @SIG-76
Feature: Staff administration
  As an administrator,
  I want to add and remove areas, permissions, memberships and administrators from SIGTA,
  so that the organization of staff is kept up to date without the console and with a record of who changed it.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permissions "verify" and "approve"
    And "admin@fi.uba.ar" is an administrator

  Scenario: an administrator sees the areas with their permissions and roles
    Given the area "student affairs" has the role "clerk" with the permission "verify"
    When "admin@fi.uba.ar" lists the areas to administer
    Then the area "student affairs" is listed with the permissions "verify, approve" and the roles "clerk"
    And the area "secretariat" is listed with the permissions "edit_procedure_types, publish_procedure_types" and the roles ""

  Scenario: an administrator finds by email someone who signed in
    Given "new@fi.uba.ar" signed in to SIGTA
    When "admin@fi.uba.ar" looks for "New@FI.uba.ar"
    Then they find "new@fi.uba.ar"

  Scenario: someone who never signed in is not found
    When "admin@fi.uba.ar" looks for "nobody@fi.uba.ar"
    Then nobody is found

  Scenario Outline: only an administrator sees what is administered
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "operative@fi.uba.ar" <looks>
    Then it is denied

    Examples:
      | looks                                    |
      | lists the areas to administer            |
      | lists the memberships of "student affairs" |
      | looks for "admin@fi.uba.ar"              |
      | lists the administrators                 |

  Scenario: the memberships of an area show who granted them and their validity
    Given it is "2026-11-04"
    And "admin@fi.uba.ar" grants "operative@fi.uba.ar" a membership in "student affairs" with the permission "verify" until "2027-03-01"
    Then the memberships of "student affairs" are:
      | staff               | roles | permissions | from       | until      | granted by      |
      | operative@fi.uba.ar |       | verify      | 2026-11-04 | 2027-03-01 | admin@fi.uba.ar |

  Scenario: ending a membership keeps it, ended, with who ended it
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    Then the membership records that "admin@fi.uba.ar" ended it
    And no membership of "student affairs" is listed

  Scenario: a membership that ended cannot be ended again
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    And "admin@fi.uba.ar" ends the membership of "operative@fi.uba.ar" in "student affairs"
    When "admin@fi.uba.ar" ends that membership again
    Then the membership is refused because it ended

  Scenario: changing the permissions of a membership ends it and opens another
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    And it is "2026-11-10"
    When "admin@fi.uba.ar" changes the membership of "operative@fi.uba.ar" in "student affairs" to the permissions "verify, approve"
    Then the membership records that "admin@fi.uba.ar" ended it
    And the memberships of "student affairs" are:
      | staff               | roles | permissions     | from       | until | granted by      |
      | operative@fi.uba.ar |       | verify, approve | 2026-11-10 |       | admin@fi.uba.ar |

  Scenario: changing a membership to the permissions it has is refused
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" changes the membership of "operative@fi.uba.ar" in "student affairs" to the permissions "verify"
    Then the membership is not granted

  Scenario: only an administrator changes a membership
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "operative@fi.uba.ar" changes the membership of "operative@fi.uba.ar" in "student affairs" to the permissions "verify, approve"
    Then it is denied

  Scenario: an area with memberships that have not ended says why it is not deleted
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" deletes the area "student affairs"
    Then it is refused because the area has memberships that have not ended

  Scenario: the Secretariat says why it is not deleted
    When "admin@fi.uba.ar" deletes the area "secretariat"
    Then it is refused because the area comes with SIGTA

  Scenario: a permission of SIGTA says why it is not deleted
    When "admin@fi.uba.ar" deletes the permission "publish_procedure_types" of "secretariat"
    Then it is refused because the permission comes with SIGTA

  Scenario: a permission held by a membership says why it is not deleted
    Given SIGTA has these memberships:
      | staff               | area            | permissions | scope |
      | operative@fi.uba.ar | student affairs | verify      |       |
    When "admin@fi.uba.ar" deletes the permission "verify" of "student affairs"
    Then it is refused because the permission is held by memberships that have not ended

  Scenario: a permission in a role says why it is not deleted
    Given the area "student affairs" has the role "clerk" with the permission "approve"
    When "admin@fi.uba.ar" deletes the permission "approve" of "student affairs"
    Then it is refused because the permission is part of a role

  Scenario: a role held by a membership says why it is not deleted
    Given the area "student affairs" has the role "clerk" with the permission "approve"
    And "admin@fi.uba.ar" grants "operative@fi.uba.ar" a membership in "student affairs" with the role "clerk"
    When "admin@fi.uba.ar" deletes the role "clerk" of "student affairs"
    Then it is refused because the role is held by memberships that have not ended

  Scenario: an area and its permissions record who created them
    Given it is "2026-11-04"
    When "admin@fi.uba.ar" creates the area "degree office" named "Degree Office"
    And "admin@fi.uba.ar" creates the permission "sign" named "Sign" in "degree office"
    Then the area "degree office" and its permission "sign" record that "admin@fi.uba.ar" created them on "2026-11-04"

  Scenario: an administrator grants the administration to someone who signed in
    Given "new@fi.uba.ar" signed in to SIGTA
    When "admin@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    Then the administrators are "admin@fi.uba.ar" and "new@fi.uba.ar"
    And the administration of "new@fi.uba.ar" records that "admin@fi.uba.ar" granted it

  Scenario: being an administrator gives no access to procedures
    Given "new@fi.uba.ar" signed in to SIGTA
    When "admin@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    Then "new@fi.uba.ar" is an administrator who is not staff

  Scenario: the administration is granted only to someone who signed in
    When "admin@fi.uba.ar" grants the administration to "nobody@fi.uba.ar"
    Then it is refused because nobody signed in with that account

  Scenario: the administration is not granted twice
    Given "new@fi.uba.ar" signed in to SIGTA
    And "admin@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    When "admin@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    Then it is refused because the person is an administrator already

  Scenario: only an administrator grants the administration
    Given "new@fi.uba.ar" signed in to SIGTA
    When "new@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    Then it is denied

  Scenario: an administrator ends the administration of another
    Given "new@fi.uba.ar" signed in to SIGTA
    And "admin@fi.uba.ar" grants the administration to "new@fi.uba.ar"
    When "new@fi.uba.ar" ends the administration of "admin@fi.uba.ar"
    Then the administrators are "new@fi.uba.ar"
    And "admin@fi.uba.ar" can no longer administer

  Scenario: SIGTA is never left without administrators
    When "admin@fi.uba.ar" ends the administration of "admin@fi.uba.ar"
    Then it is refused because SIGTA would be left without administrators
    And the administrators are "admin@fi.uba.ar"
