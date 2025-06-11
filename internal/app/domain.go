package app

import (
	"librarium/internal/http"
	"librarium/internal/postgres"
)

func (a *Application) setupDomain() (err error) {
	a.userRepo, err = postgres.NewUserRepository(a.db)
	if err != nil {
		return err
	}

	a.authController, err = http.NewAuthController(a.userRepo)
	if err != nil {
		return err
	}
	return nil
}
