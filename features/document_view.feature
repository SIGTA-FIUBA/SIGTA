@US16 @SIG-26
Feature: View and download the documents of a procedure
  As a person who takes part in a procedure,
  I want to open and download its documents,
  so that I can review them without leaving the platform.

  Background:
    Given a procedure type whose steps ask for these documents:
      | step      | done by               | document   | types                      | max size |
      | documents | the applicant         | transcript | application/pdf, image/png | 1024     |
      | review    | secretariat, "review" | ruling     | application/pdf            | 1024     |
    And these staff members:
      | staff                | area                  | permissions |
      | secretary@fi.uba.ar  | secretariat           | review      |
      | commission@fi.uba.ar | curricular commission | evaluate    |
    And "applicant@fi.uba.ar" has a procedure waiting on the "documents" step

  Scenario: each document of the procedure is downloaded on its own
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia.pdf" as the "transcript"
    And the procedure moved on to the "review" step
    And "secretary@fi.uba.ar" attached a PDF of 300 bytes named "dictamen.pdf" as the "ruling"
    When "secretary@fi.uba.ar" downloads "historia.pdf"
    Then they get "historia.pdf" with the content that was attached

  Scenario: a file that another one replaced can still be downloaded
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia-vieja.pdf" as the "transcript"
    And "applicant@fi.uba.ar" attached a PDF of 300 bytes named "historia.pdf" as the "transcript"
    When "secretary@fi.uba.ar" downloads "historia-vieja.pdf"
    Then they get "historia-vieja.pdf" with the content that was attached

  Scenario: the file comes with the type read from its content, so that the browser can show it
    Given "applicant@fi.uba.ar" attached a PNG image named "historia.pdf" as the "transcript"
    When "secretary@fi.uba.ar" downloads "historia.png"
    Then the file comes as "historia.png" with the type "image/png"

  Scenario Outline: only whoever can see the procedure opens and downloads its documents
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia.pdf" as the "transcript"
    And "scoped@fi.uba.ar" is in the "secretariat" with "review" only for another procedure type
    When "<person>" downloads "historia.pdf"
    Then <result>

    Examples:
      | person               | result                                                     |
      | applicant@fi.uba.ar  | they get "historia.pdf" with the content that was attached |
      | secretary@fi.uba.ar  | they get "historia.pdf" with the content that was attached |
      | commission@fi.uba.ar | the procedure is not found                                 |
      | scoped@fi.uba.ar     | the procedure is not found                                 |
      | other@fi.uba.ar      | the procedure is not found                                 |
