package tests

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"urleater/internal/handlers"
	"urleater/internal/service"
)

type BaseSuite struct {
	suite.Suite

	Handlers handlers.Handlers
}

type Handler = func(c echo.Context) error

func (s *BaseSuite) BaseSetupTest() {

}

func (s *BaseSuite) TearDownSuite() {

}

func (s *BaseSuite) MakeRequestWithBody(method string, f Handler, jsonString string) ([]byte, int) {
	e := echo.New()

	req := httptest.NewRequest(method, "http://localhost", strings.NewReader(jsonString))
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

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.PostRegister, string(res))
}

func (s *BaseSuite) LoginUser(data *handlers.LoginRequest) ([]byte, int) {
	res, err := json.Marshal(data)
	s.NoError(err)

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.PostLogin, string(res))
}

func (s *BaseSuite) CreateShortLink(data *handlers.CreateShortLinkRequest) ([]byte, int) {
	res, err := json.Marshal(data)
	s.NoError(err)

	return s.MakeRequestWithBody(http.MethodPost, s.Handlers.CreateShortLink, string(res))
}

func (s *BaseSuite) FinishSetupTest(
	postgresStorage service.PostgresStorage,
	redisStorage service.RedisStorage,
	elasticSearcher service.ElasticSearcher,
	consumers []service.Consumer,
	producer service.Producer,
	mockSessionStore handlers.SessionStore) {
	httpSegSvc := service.New(postgresStorage, redisStorage, producer, consumers, elasticSearcher, "")

	hndls := handlers.Handlers{
		Service: httpSegSvc,
		Store:   mockSessionStore,
	}

	s.Handlers = hndls
}
