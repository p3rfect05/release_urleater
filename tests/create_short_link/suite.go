package create_short_link

import (
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"urleater/dto"
	base "urleater/tests"
	"urleater/tests/mocks"
)

type createShortLinkSuite struct {
	base.BaseSuite
}

func (s *createShortLinkSuite) SetupTest() {
	s.BaseSetupTest()

	storage := mocks.NewPostgresStorage(s.T())
	sessionStore := mocks.NewSessionStore(s.T())
	redisStorage := mocks.NewRedisStorage(s.T())
	searcherStorage := mocks.NewElasticSearcher(s.T())

	storage.On("GetShortLink", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows).Maybe()

	redisStorage.On("SaveShortLinkToLongLink", mock.Anything, mock.Anything).Return(nil).Maybe()

	searcherStorage.On("AddShortLink", mock.Anything, mock.Anything).Return(nil).Maybe()

	storage.On("GetUser", mock.Anything, mock.Anything).Return(&dto.User{
		UrlsLeft: 1,
	}, nil).Maybe()

	storage.On("UpdateUserLinks", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	sessionStore.On("RetrieveEmailFromSession", mock.Anything).Return("any_email", nil)

	// 1
	longUrl1 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	createdNewLink1 := dto.Link{
		ShortUrl: "new_short_link",
		LongUrl:  longUrl1,
	}

	storage.On("CreateShortLink", mock.Anything, mock.Anything, longUrl1, mock.Anything).Return(&createdNewLink1, nil).Once()

	// 4
	longUrl4 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias4 := "myAlias1"

	createdNewLink4 := dto.Link{
		ShortUrl: alias4,
		LongUrl:  longUrl4,
	}

	storage.On("CreateShortLink", mock.Anything, alias4, longUrl4, mock.Anything).Return(&createdNewLink4, nil).Once()

	// 8
	longUrl8 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias8 := "myCustomAlias1234567"

	createdNewLink8 := dto.Link{
		ShortUrl: alias8,
		LongUrl:  longUrl8,
	}

	storage.On("CreateShortLink", mock.Anything, alias8, longUrl8, mock.Anything).Return(&createdNewLink8, nil).Once()

	s.FinishSetupTest(storage, redisStorage, searcherStorage, nil, nil, sessionStore)
}
