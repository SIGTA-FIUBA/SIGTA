@US19 @SIG-23
Feature: Detail of a procedure
  As a person who takes part in a procedure,
  I want to see its steps, who does each one and the documents they ask for,
  so that I know where it stands and what is expected now.

  Background:
    Given a procedure type whose steps ask for these documents:
      | step      | done by                    | document   | types                      | max size |
      | documents | the applicant              | transcript | application/pdf, image/png | 1024     |
      | ruling    | student affairs, "resolve" | ruling     | application/pdf            | 1024     |
    And these staff members:
      | staff                | area                  | permissions |
      | resolver@fi.uba.ar   | student affairs       | resolve     |
      | reader@fi.uba.ar     | student affairs       | review      |
      | commission@fi.uba.ar | curricular commission | evaluate    |
    And "applicant@fi.uba.ar" has a procedure waiting on the "documents" step

  Scenario: the detail lists every step in order with who does it and what it asks for
    When "applicant@fi.uba.ar" opens the procedure
    Then they see these steps:
      | step      | done by         | documents  |
      | documents | the applicant   | transcript |
      | ruling    | student affairs | ruling     |

  Scenario Outline: the detail says which step is open, who does it and whether it is up to them
    Given the procedure moved on to the "<step>" step
    When "<person>" opens the procedure
    Then the open step is "<step>", done by <done by>
    And the procedure is <state>
    And <turn>

    Examples:
      | step      | person              | done by         | state              | turn                 |
      | documents | applicant@fi.uba.ar | the applicant   | with the applicant | it is up to them     |
      | documents | reader@fi.uba.ar    | the applicant   | with the applicant | it is not up to them |
      | ruling    | resolver@fi.uba.ar  | student affairs | in an office       | it is up to them     |
      | ruling    | reader@fi.uba.ar    | student affairs | in an office       | it is not up to them |
      | ruling    | applicant@fi.uba.ar | student affairs | in an office       | it is not up to them |

  Scenario: a procedure with no open step is closed and up to nobody
    Given the procedure has no open step
    When "applicant@fi.uba.ar" opens the procedure
    Then the procedure is closed
    And it is not up to them

  Scenario: the attached documents show next to the requirement they cover
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia.pdf" as the "transcript"
    When "reader@fi.uba.ar" opens the procedure
    Then the "transcript" shows the file "historia.pdf"
    And the "ruling" shows no file

  Scenario Outline: whoever has no part in the procedure does not see its detail
    When "<person>" opens the procedure
    Then the procedure is not found

    Examples:
      | person               |
      | other@fi.uba.ar      |
      | commission@fi.uba.ar |
