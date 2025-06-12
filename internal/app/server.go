package app

import "librarium/internal/http"

func (a *Application) setupServer() (err error) {
	a.server, err = http.NewServer(a.serverAddress, a.authController, a.catalogController)
	if err != nil {
		return err
	}

	return nil
}
