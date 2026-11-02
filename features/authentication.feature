@US06 @SIG-11
Feature: Sign in through an abstracted identity provider
  As a user,
  I want to sign in,
  so that I can access only what belongs to me.

  Background:
    Given only accounts from the "fi.uba.ar" domain are allowed
    And SIGTA has "applicant@fi.uba.ar" registered as an applicant
    And SIGTA has "operative@fi.uba.ar" registered as staff with the "staff_operative" profile

  Scenario: the profile comes from SIGTA, not from the provider
    When "operative@fi.uba.ar" signs in through the identity provider
    Then a session is opened
    And their profile is "staff_operative"

  Scenario: an identity unknown to SIGTA gets no session
    When "nobody@fi.uba.ar" signs in through the identity provider
    Then no session is opened
    And they are told SIGTA has no account for that identity

  Scenario: a sign-in the provider rejects does not blame the account
    When the identity provider rejects a sign-in
    Then no session is opened
    And they are asked to try again

  Scenario Outline: only institutional accounts sign in
    When "<email>" signs in through the identity provider
    Then no session is opened
    And they are asked to sign in with their institutional account

    Examples:
      | email                        |
      | applicant@example.com        |
      | operative@evilfi.uba.ar      |
      | operative@fi.uba.ar.evil.com |

  Scenario: a sign-in the browser did not start is rejected
    When a provider response arrives with a state different from the one the browser started
    Then no session is opened

  Scenario: a pending sign-in cannot be completed twice
    Given "applicant@fi.uba.ar" completed a sign-in
    When the same provider response arrives again in that browser
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
    When "operative@fi.uba.ar" asks for a sign-in link
    Then no email is sent

  Scenario: a link stops working when its holder becomes staff
    Given "applicant@fi.uba.ar" asked for a sign-in link
    And their profile changed to "staff_operative"
    When they open the link
    Then no session is opened
