package http_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"librarium/internal/http"
	"librarium/internal/mocks"
	"librarium/internal/user"
)

func TestNewAuthController(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockUserRepository(ctrl)

	testCases := map[string]struct {
		userRepo         user.Repository
		expectedErr      error
		assertController func(controller *http.AuthController)
	}{
		"it should return an error if the user repository is missing": {
			expectedErr: errors.New("user repository is mandatory for auth controller"),
			assertController: func(controller *http.AuthController) {
				assert.Nil(t, controller)
			},
		},
		"it should return the controller created and no error": {
			userRepo: userRepo,
			assertController: func(controller *http.AuthController) {
				assert.NotNil(t, controller)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller, err := http.NewAuthController(tc.userRepo)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertController(controller)
		})
	}
}
