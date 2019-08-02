package postgres

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type client struct {
	host     string
	port     string
	user     string
	dbname   string
	password string
}

type Connection struct {
	conn *gorm.DB
}

var (
	connInstance *Connection
)

func NewClient(host, port, user, dbname, password string) *client {
	return &client{
		host:     host,
		port:     port,
		user:     user,
		dbname:   dbname,
		password: password,
	}
}

func (c *client) Connect() *Connection {
	if connInstance == nil {
		db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", c.host, c.port, c.user, c.dbname, c.password))
		if err != nil {
			log.Panicf("Error trying to connect: %v", err)
		}
		connInstance = &Connection{
			conn: db,
		}
	}
	return connInstance
}

func (c *Connection) DB() *gorm.DB {
	return c.conn
}
