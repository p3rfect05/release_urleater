package create_short_link

import (
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"urleater/internal/repository/postgresDB"
	base "urleater/tests"
	"urleater/tests/mocks"
)

type createShortLinkSuite struct {
	base.BaseSuite
}

func (s *createShortLinkSuite) SetupTest() {
	s.BaseSetupTest()

	storage := mocks.NewStorage(s.T())
	sessionStore := mocks.NewSessionStore(s.T())

	storage.On("GetUser", mock.Anything, mock.Anything).Return(&postgresDB.User{
		UrlsLeft: 1,
	}, nil).Maybe()

	storage.On("UpdateUserLinks", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	sessionStore.On("RetrieveEmailFromSession", mock.Anything).Return("any_email", nil)

	storage.On("GetShortLink", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)

	// 1
	longUrl1 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	createdNewLink1 := postgresDB.Link{
		ShortUrl: "new_short_link",
		LongUrl:  longUrl1,
	}

	storage.On("CreateShortLink", mock.Anything, mock.Anything, longUrl1, mock.Anything).Return(&createdNewLink1, nil).Once()

	// 4
	longUrl4 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias4 := "myAlias1"

	createdNewLink4 := postgresDB.Link{
		ShortUrl: alias4,
		LongUrl:  longUrl4,
	}

	storage.On("CreateShortLink", mock.Anything, alias4, longUrl4, mock.Anything).Return(&createdNewLink4, nil).Once()

	// 8
	longUrl8 := "https://www.gismeteo.ru/weather-moscow-4368/weekend/#dataset"
	alias8 := "myCustomAlias1234567"

	createdNewLink8 := postgresDB.Link{
		ShortUrl: alias8,
		LongUrl:  longUrl8,
	}

	storage.On("CreateShortLink", mock.Anything, alias8, longUrl8, mock.Anything).Return(&createdNewLink8, nil).Once()

	s.FinishSetupTest(storage, sessionStore)
}
