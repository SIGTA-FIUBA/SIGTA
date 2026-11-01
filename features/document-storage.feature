@US15 @SIG-18
Feature: Upload and store documents
  As a user, I want to attach files to a procedure,
  so that they stay associated with the request.

  Background:
    Given a "subject equivalence" procedure started by an applicant
    And its procedure type requires a "transcript" document in PDF of up to 10 MB

  Rule: file type and size are checked against the requirement configuration

    Scenario: a document that meets the requirement is accepted
      When the applicant uploads a 2 MB PDF as the transcript
      Then the document is stored

    Scenario Outline: a document the requirement does not accept is rejected
      When the applicant uploads <file> as the transcript
      Then the document is not stored
      And the applicant is told <reason>

      Examples:
        | file               | reason                                 |
        | a PNG image        | the transcript must be a PDF           |
        | an 11 MB PDF       | the transcript can be up to 10 MB      |
        | an empty file      | the file is empty                      |

    Scenario: the type is checked on the content, not on the file name
      When the applicant uploads an image renamed as "transcript.pdf" as the transcript
      Then the document is not stored
      And the applicant is told the transcript must be a PDF

  Rule: each file is associated with the procedure and the requirement it covers

    Scenario: a stored document records its procedure and requirement
      When the applicant uploads a PDF as the transcript
      Then the document belongs to the procedure
      And it covers the "transcript" requirement
      And it records who uploaded it and when

    Scenario: an applicant cannot upload a document that covers no requirement
      When the applicant uploads a PDF that covers no requirement of the procedure
      Then the document is not stored
      And the applicant is told the requirement does not exist

  Rule: the Secretaría can attach documents to the procedure

    Scenario Outline: staff attach a document that covers no requirement
      Given a staff member with the <profile> profile
      When they upload a resolution document to the procedure
      Then the document is stored
      And it belongs to the procedure without covering a requirement

      Examples:
        | profile   |
        | operator  |
        | approver  |

  Rule: only someone with permission on the procedure can access its files

    Scenario: an applicant cannot upload to another applicant's procedure
      Given another applicant
      When the other applicant uploads a PDF as the transcript
      Then the document is not stored
      And they are told they have no access to the procedure

    Scenario Outline: access to a stored file follows the permission on the procedure
      Given the applicant uploaded a PDF as the transcript
      When <someone> asks for the file
      Then they <outcome>

      Examples:
        | someone                  | outcome                             |
        | the applicant            | get the file                        |
        | a staff member           | get the file                        |
        | another applicant        | are told they have no access to it  |
        | someone without session  | are asked to sign in                |

    Scenario: a file is only reachable through the procedure it belongs to
      Given the applicant uploaded a PDF as the transcript
      And another applicant has their own procedure
      When the other applicant asks for that file through their own procedure
      Then they are told the document does not exist

  Rule: files are stored outside the database

    Scenario: the file goes to the document storage and the database keeps only its record
      When the applicant uploads a PDF as the transcript
      Then the file content is kept in the document storage
      And the document record keeps its name, type, size and requirement
