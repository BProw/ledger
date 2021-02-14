Feature: User Management

    Users who want to use the product need to sign up!
    The product requires a contact email address, name, and password
    for identification.

  Scenario: New User
    Given a user does not have an account and wants to sign up
    When the user provides their email address as "ledger.user@example.com"
    And the user provides their name as "Ledger User"
    And the user provides their password as "supersecret"
    When the user asks to sign up
    Then the user has a valid session for "18h"

  Scenario: User Login
    Given the user does not have a valid session
    When the user provides their email address as "ledger.user@example.com"
    And the user provides their password as "supersecret"
    When the user attempts to login
    Then the user has a valid session for "18h"
  
  Scenario: User Logout
    Given the user has a valid session
    When the user decides to logout
    Then the user does not have a valid session
