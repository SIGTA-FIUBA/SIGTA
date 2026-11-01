@US06 @SIG-11
Feature: Sign in through an abstracted identity provider
  As a user,
  I want to sign in,
  so that I can access only what belongs to me.

  Background:
    Given only accounts from the "fi.uba.ar" domain are allowed
    And SIGTA has "applicant@fi.uba.ar" registered as an applicant
    And SIGTA has "operator@fi.uba.ar" registered as staff with the operator profile

  Scenario: the profile comes from SIGTA, not from the provider
    When "operator@fi.uba.ar" signs in through the identity provider
    Then a session is opened
    And their profile is staff operator

  Scenario: an identity unknown to SIGTA gets no session
    When "nobody@fi.uba.ar" signs in through the identity provider
    Then no session is opened

  Scenario Outline: only institutional accounts sign in
    When "<email>" signs in through the identity provider
    Then no session is opened
    And they are asked to sign in with their institutional account

    Examples:
      | email                       |
      | applicant@example.com       |
      | operator@evilfi.uba.ar      |
      | operator@fi.uba.ar.evil.com |

  Scenario: a sign-in the browser did not start is rejected
    When a provider response arrives with a state different from the one the browser started
    Then no session is opened

  Scenario: the session expires after inactivity
    Given "applicant@fi.uba.ar" has an open session
    When 30 minutes pass without requests
    Then the session expires
    And they have to sign in again

  Scenario: activity keeps the session open
    Given "applicant@fi.uba.ar" has an open session
    When they make a request every 29 minutes
    Then the session stays open

  Scenario: an applicant signs in with a link sent by email
    When "applicant@fi.uba.ar" asks for a sign-in link
    Then they receive a single-use link by email
    And opening it opens a session

  Scenario: the link cannot be reused
    Given "applicant@fi.uba.ar" already used their sign-in link
    When they open it again
    Then no session is opened

  Scenario: the link expires
    Given "applicant@fi.uba.ar" asked for a sign-in link
    When 15 minutes pass
    Then the link no longer opens a session

  Scenario: the link is only sent to institutional emails
    When "applicant@example.com" asks for a sign-in link
    Then they are asked to use their institutional email
    And no email is sent

  Scenario: asking for a link does not reveal who has an account
    When "nobody@fi.uba.ar" asks for a sign-in link
    Then the response is the same as for an existing account
    And no email is sent

  Scenario: staff sign in only through the identity provider
    When "operator@fi.uba.ar" asks for a sign-in link
    Then no email is sent
