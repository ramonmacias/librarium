package app

import (
	"librarium/internal/postgres"
)

func (a *Application) setupInfra() error {
	var err error

	a.db, err = postgres.OpenConnection(a.databaseSource)
	if err != nil {
		return err
	}
	return nil
}
