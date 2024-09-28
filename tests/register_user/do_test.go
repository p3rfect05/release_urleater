package register_user

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestSuite(t *testing.T) {
	suite.Run(t, new(registerUserSuite))
}
