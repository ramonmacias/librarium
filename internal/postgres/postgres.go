package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbConnectionTimeout = 10 * time.Second
)

// DataSource defines all the required data that enables the auth and connection
// between the application and the postgres
type DataSource struct {
	UserName string // Name of the user to connect
	Password string // Password to be used for connect the user to the db
	Host     string // Host where the db is placed
	Port     string // Number of the port where the db is exposed
	DBName   string // Name of the database
	SSLMode  string // Use the ssl mode or node, should be enable or disable
}

// URL builds the url to be used while starting the connection to the db
func (ds *DataSource) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", ds.UserName, ds.Password, ds.Host, ds.Port, ds.DBName, ds.SSLMode)
}

// OpenConnection starts a new connection between the application and the postgres
// database, it checks the connection for that db before we return the *sql.DB, so we
// are sure the connection is alive before we try to reach the db.
// It returns specific errors while openning and pinging.
func OpenConnection(ds *DataSource) (*sql.DB, error) {
	db, err := sql.Open("postgres", ds.URL())
	if err != nil {
		return nil, fmt.Errorf("error while opening db connection %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging the db while opening db connection %w", err)
	}

	return db, nil
}

// CloseConnection finishes the connection between the application and the postgres database.
// It returns an error if we can't close it.
func CloseConnection(conn *sql.DB) error {
	return conn.Close()
}
