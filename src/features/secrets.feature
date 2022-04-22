Feature: Secrets

    In order to get a visual on secrets
    As a user
    I need to be able to see the visual

    Scenario: View secrets
        Given there are 3 secrets attached to my user
        And 4 secrets not attached to my user
        When I load the secrets page
        Then I should see 7 secrets
        And I should see 3 owned by me
        And I should be 4 not owned by me