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

  @SIG-61
  Scenario: an institutional identity new to SIGTA is registered as an applicant
    When "newcomer@fi.uba.ar" signs in through the identity provider
    Then a session is opened
    And their profile is "applicant"

  Scenario: a sign-in the provider rejects does not blame the account
    When the identity provider rejects a sign-in
    Then no session is opened
    And they are asked to try again

  @SIG-61
  Scenario: a sign-in cancelled at the provider does not blame the account
    When "applicant@fi.uba.ar" cancels the sign-in at the identity provider
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
