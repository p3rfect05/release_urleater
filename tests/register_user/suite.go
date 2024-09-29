package register_user

import (
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"urleater/internal/repository/postgresDB"
	base "urleater/tests"
	"urleater/tests/mocks"
)

type registerUserSuite struct {
	base.BaseSuite
}

func (s *registerUserSuite) SetupTest() {
	s.BaseSetupTest()

	storage := mocks.NewStorage(s.T())
	sessionStore := mocks.NewSessionStore(s.T())

	sessionStore.On("RetrieveEmailFromSession", mock.Anything).Return("", nil)
	sessionStore.On("Get", mock.Anything, mock.Anything).Return(nil, nil)
	sessionStore.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	storage.On("GetUser", mock.Anything, "test_name1@mail.ru").
		Return(nil, pgx.ErrNoRows).Once()
	storage.On("CreateUser", mock.Anything, "test_name1@mail.ru", "qwertyui").Return(nil).Once()

	// 2
	user2 := postgresDB.User{
		Email:        "test_name1@mail.ru",
		PasswordHash: "some_hash",
	}

	storage.On("GetUser", mock.Anything, "test_name1@mail.ru").
		Return(&user2, nil).Once()

	// 3
	user3 := postgresDB.User{
		Email:        "test_name1@mail.ru",
		PasswordHash: "some_hash",
	}

	storage.On("GetUser", mock.Anything, "test_name1@mail.ru").
		Return(&user3, nil).Once()

	s.FinishSetupTest(storage, sessionStore)

}
