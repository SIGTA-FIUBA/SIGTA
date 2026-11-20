@US22 @SIG-30
Feature: My procedures in progress
  As an applicant,
  I want to see my active procedures when I sign in,
  so that I know where they are without asking.

  Background:
    Given the staff areas "student affairs" and "curricular commission"
    And the area "student affairs" has the permission "verify"
    And the area "curricular commission" has the permission "evaluate"
    And a published "subject equivalence" procedure type with this flow:
      | step     | done by               | permissions | action                | next     |
      | request  | applicant             |             | enter_data            | check    |
      | check    | student affairs       | verify      | confirm_external_fact | evaluate |
      | evaluate | curricular commission | evaluate    | confirm_external_fact | file     |
      | file     | student affairs       | verify      | confirm_external_fact |          |
    And SIGTA has these memberships:
      | staff                | area                  | permissions | scope |
      | operative@fi.uba.ar  | student affairs       | verify      |       |
      | commission@fi.uba.ar | curricular commission | evaluate    |       |

  Scenario: each active procedure shows its open step and its state
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is "request"
    And "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is "check"
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see these procedures:
      | procedure type      | open step | state               |
      | subject equivalence | check     | in an office        |
      | subject equivalence | request   | with the applicant  |

  Scenario: a referred procedure shows as referred
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is "check"
    And "operative@fi.uba.ar" confirmed the "check" step
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see these procedures:
      | procedure type      | open step | state    |
      | subject equivalence | evaluate  | referred |

  # A closed procedure is listed too, so that its applicant reaches its
  # resolution (SIG-31).
  Scenario: a closed procedure is listed as closed, without an open step
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is ""
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see these procedures:
      | procedure type      | open step | state  |
      | subject equivalence |           | closed |

  Scenario: the procedures of other applicants are not listed
    Given "other@fi.uba.ar" has a "subject equivalence" procedure whose open step is "check"
    When "applicant@fi.uba.ar" looks at their procedures
    Then they see no procedures

  Scenario: staff see the procedures they started, not the ones of their areas
    Given "operative@fi.uba.ar" has a "subject equivalence" procedure whose open step is "request"
    And "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is "check"
    When "operative@fi.uba.ar" looks at their procedures
    Then they see these procedures:
      | procedure type      | open step | state              |
      | subject equivalence | request   | with the applicant |

  Scenario: each listed procedure opens its detail
    Given "applicant@fi.uba.ar" has a "subject equivalence" procedure whose open step is "check"
    When "applicant@fi.uba.ar" looks at their procedures
    Then "applicant@fi.uba.ar" can open the detail of each one
