package register_user

import (
	"net/http"
	"urleater/internal/handlers"
)

func (s *registerUserSuite) TestRegisterUser() {
	// 1
	_, code := s.RegisterUser(&handlers.RegisterRequest{
		Email:    "test_name1@mail.ru",
		Password: "qwertyui",
	})

	s.Equal(http.StatusOK, code)

	// 2
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "test_name1@mail.ru",
		Password: "qwertyui",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 3
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "test_name1@mail.ru",
		Password: "12345678",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 4
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "",
		Password: "",
	})

	s.Equal(http.StatusInternalServerError, code)

	//5
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "        ",
		Password: "        ",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 6

	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "testname",
		Password: "12345678",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 7
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "test_name2@ya.com",
		Password: "1234567",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 8
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "юзер@маил.ком",
		Password: "12345678",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 9
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "test_name3@ya.com",
		Password: "мойпароль",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 10
	_, code = s.RegisterUser(&handlers.RegisterRequest{
		Email:    "!?test_name4@ya.com",
		Password: "qwertyui",
	})

	s.Equal(http.StatusInternalServerError, code)
}
