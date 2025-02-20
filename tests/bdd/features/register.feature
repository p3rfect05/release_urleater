Feature: URLEater Register

  Scenario Outline: Registration
    Given I have navigated to main page and not logged in
    When I type email <email>
    And I type password <password>
    And I type confirm_password <password_confirm>
    And I clicked Register
    Then I should see the main page and be logged in

    Examples:
      | email             | password   | password_confirm   |
      | test@example.com  | 12341234   | 12341234           |

