@US15 @SIG-18
Feature: Upload and store documents
  As a person who takes part in a procedure,
  I want to attach files to it,
  so that they stay with the request.

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

  Scenario: the applicant attaches a document the open step asks for
    When "applicant@fi.uba.ar" attaches a PDF of 200 bytes named "historia.pdf" as the "transcript"
    Then the "transcript" of the procedure is "historia.pdf", uploaded by "applicant@fi.uba.ar" on the "documents" step

  Scenario Outline: the type and size come from the configuration of the document
    When "applicant@fi.uba.ar" attaches <file> named "<name>" as the "transcript"
    Then the file is rejected because <reason>
    And the procedure has no attached documents

    Examples:
      | file                | name         | reason                   |
      | a text file         | historia.pdf | its type is not accepted |
      | a ZIP file          | historia.pdf | its type is not accepted |
      | a PDF of 1025 bytes | historia.pdf | it is too large          |
      | an empty file       | historia.pdf | it is empty              |

  Scenario: the type is read from the content, not from the name
    When "applicant@fi.uba.ar" attaches a PNG image named "historia.pdf" as the "transcript"
    Then the "transcript" of the procedure has the type "image/png"

  Scenario: a document that the open step does not ask for is rejected
    When "applicant@fi.uba.ar" attaches a PDF of 200 bytes named "resolucion.pdf" as the "ruling"
    Then the file is rejected because the open step does not ask for it

  Scenario: the staff attach the ruling on their step
    Given the procedure moved on to the "ruling" step
    When "resolver@fi.uba.ar" attaches a PDF of 200 bytes named "resolucion.pdf" as the "ruling"
    Then the "ruling" of the procedure is "resolucion.pdf", uploaded by "resolver@fi.uba.ar" on the "ruling" step

  Scenario Outline: only whoever acts on the open step attaches documents to it
    Given the procedure moved on to the "<step>" step
    When "<person>" attaches a PDF of 200 bytes named "archivo.pdf" as the "<document>"
    Then the file is rejected because <reason>

    Examples:
      | step      | person              | document   | reason                       |
      | ruling    | reader@fi.uba.ar    | ruling     | they cannot act on that step |
      | ruling    | applicant@fi.uba.ar | ruling     | they cannot act on that step |
      | documents | resolver@fi.uba.ar  | transcript | they cannot act on that step |
      | documents | other@fi.uba.ar     | transcript | the procedure is not found   |

  Scenario Outline: only whoever can see the procedure downloads its documents
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia.pdf" as the "transcript"
    When "<person>" downloads the "transcript"
    Then <result>

    Examples:
      | person               | result                                      |
      | applicant@fi.uba.ar  | they get the same content that was attached |
      | reader@fi.uba.ar     | they get the same content that was attached |
      | other@fi.uba.ar      | the procedure is not found                  |
      | commission@fi.uba.ar | the procedure is not found                  |

  Scenario: the file is kept outside the database, named by its content
    When "applicant@fi.uba.ar" attaches a PDF of 200 bytes named "historia.pdf" as the "transcript"
    Then the document store keeps the file under the SHA-256 digest of its content
    And the record of the document names that digest and its size

  Scenario: a new file replaces the current one and the previous one stays recorded
    Given "applicant@fi.uba.ar" attached a PDF of 200 bytes named "historia.pdf" as the "transcript"
    When "applicant@fi.uba.ar" attaches a PDF of 300 bytes named "historia-corregida.pdf" as the "transcript"
    Then the "transcript" of the procedure is "historia-corregida.pdf", uploaded by "applicant@fi.uba.ar" on the "documents" step
    And the procedure has 2 attached documents
