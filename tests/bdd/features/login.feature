Feature: URLEater Login

  Scenario Outline: Login
    Given I have navigated to main page and not logged in
    When I type email <email>
    When I type password <password>
    When I clicked Login
    Then I should see the main page and be logged in

    Examples:
      | email             | password   |
      | test@example.com  | 12341234   |