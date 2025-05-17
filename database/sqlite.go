package database

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

const sqliteDialect = "sqlite3"

var conn *sqlstore.Container

func NewSqlite() (*sqlstore.Container, error) {
	container, err := sqlstore.New(context.Background(), sqliteDialect, "file:examplestore.db?_foreign_keys=on", nil)
	if err != nil {
		log.Errorf("err sqlstore.New : %v ", err)
		return nil, err
	}

	SetConnection(container)

	return container, nil
}

// GetConnection : Get Available Connection
func GetConnection() *sqlstore.Container {
	return conn
}

// SetConnection : Set Available Connection
func SetConnection(connection *sqlstore.Container) {
	conn = connection
}
