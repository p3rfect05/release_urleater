package login_user

import (
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	base "urleater/tests"
	"urleater/tests/mocks"
)

type loginUserSuite struct {
	base.BaseSuite
}

func (s *loginUserSuite) SetupTest() {
	s.BaseSetupTest()

	storage := mocks.NewStorage(s.T())
	sessionStore := mocks.NewSessionStore(s.T())

	sessionStore.On("RetrieveEmailFromSession", mock.Anything).Return("", nil)
	sessionStore.On("Get", mock.Anything, mock.Anything).Return(nil, nil)
	sessionStore.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 1
	storage.On("VerifyUserPassword", mock.Anything, "test_name1@mail.ru", mock.Anything).
		Return(nil).Once()

	// 5
	storage.On("VerifyUserPassword", mock.Anything, "test_name10@mail.com", mock.Anything).
		Return(pgx.ErrNoRows).Once()

	// 6
	storage.On("VerifyUserPassword", mock.Anything, "test_name1@mail.ru", mock.Anything).
		Return(errors.New("invalid password")).Once()

	s.FinishSetupTest(storage, sessionStore)
}
