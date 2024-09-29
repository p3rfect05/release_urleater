package login_user

import (
	"net/http"
	"urleater/internal/handlers"
)

func (s *loginUserSuite) TestLoginUser() {

	// 1
	_, code := s.LoginUser(&handlers.LoginRequest{
		Email:    "test_name1@mail.ru",
		Password: "qwertyui",
	})

	s.Equal(http.StatusOK, code)

	// 2
	_, code = s.LoginUser(&handlers.LoginRequest{
		Email:    "",
		Password: "",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 3
	_, code = s.LoginUser(&handlers.LoginRequest{
		Email:    "        ",
		Password: "        ",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 4
	_, code = s.LoginUser(&handlers.LoginRequest{
		Email:    "test_name1",
		Password: "qwertyui",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 5
	_, code = s.LoginUser(&handlers.LoginRequest{
		Email:    "test_name10@mail.com",
		Password: "12345678",
	})

	s.Equal(http.StatusInternalServerError, code)

	// 6
	_, code = s.LoginUser(&handlers.LoginRequest{
		Email:    "test_name1@mail.ru",
		Password: "qwertyui5",
	})

	s.Equal(http.StatusInternalServerError, code)

}
