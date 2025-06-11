package app

import "librarium/internal/postgres"

// Option handles the functional options for setup the application
type Option func(a *Application)

// WithDatabaseSource adds the provided data source as the database
// source for this application
func WithDatabaseSource(ds *postgres.DataSource) Option {
	return func(a *Application) {
		a.databaseSource = ds
	}
}

// WithServerAddress adds the provided address as the http server
// address for this application
func WithServerAddress(addr string) Option {
	return func(a *Application) {
		a.serverAddress = addr
	}
}
