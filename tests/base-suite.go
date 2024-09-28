package tests

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"urleater/internal/handlers"
	"urleater/internal/service"
	"urleater/tests/mocks"
)

type BaseSuite struct {
	suite.Suite

	Handlers handlers.Handlers
}

type Handler = func(c echo.Context) error

func (s *BaseSuite) SetupTest() {

}

func (s *BaseSuite) TearDownSuite() {

}

func (s *BaseSuite) MakeRequestWithBody(method string, f Handler, url string, jsonString string) ([]byte, int) {
	e := echo.New()

	req := httptest.NewRequest(method, url, strings.NewReader(jsonString))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := f(c)

	s.NoError(err)

	return rec.Body.Bytes(), rec.Code
}

func (s *BaseSuite) RegisterUser(data *handlers.RegisterRequest) ([]byte, int) {
	res, err := json.Marshal(data)
	s.NoError(err)

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.PostRegister, "", string(res))
}

func (s *BaseSuite) LoginUser(data *handlers.LoginRequest) ([]byte, int) {
	res, err := json.Marshal(data)
	s.NoError(err)

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.PostLogin, "", string(res))
}

func (s *BaseSuite) CreateShortLink(data *handlers.CreateShortLinkRequest) ([]byte, int) {
	res, err := json.Marshal(data)
	s.NoError(err)

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.CreateShortLink, "", string(res))
}

func (s *BaseSuite) FinishSetupTest(storage service.Storage) {

	httpSegSvc := service.New(storage)

	sessionStore := mocks.NewSessionStore(s.T())
	sessionStore.On("RetrieveEmailFromSession", mock.Anything).Return("valid_email", nil)

	hndls := handlers.Handlers{
		Service: httpSegSvc,
		Store:   sessionStore,
	}

	s.Handlers = hndls
}
