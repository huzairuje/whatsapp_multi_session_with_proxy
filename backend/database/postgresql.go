package database

import (
	"context"
	"database/sql"
	"fmt"

	"whatsapp_multi_session/config"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

const postgresDialect = "postgres"

func NewPostgresClient(conf *config.Config) (*sqlstore.Container, error) {
	db := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s fallback_application_name=whatsapp_multi_session TimeZone=Asia/Jakarta",
		conf.Postgres.Host,
		conf.Postgres.Port,
		conf.Postgres.User,
		conf.Postgres.Password,
		conf.Postgres.DBName,
		"disable")
	dbConn, err := sqlstore.New(context.Background(), postgresDialect, db, nil)
	if err != nil {
		return nil, err
	}
	sqlstore.PostgresArrayWrapper = pq.Array

	SetConnection(dbConn)

	return dbConn, nil
}

func GetRawPostgresDB(conf *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Postgres.Host,
		conf.Postgres.Port,
		conf.Postgres.User,
		conf.Postgres.Password,
		conf.Postgres.DBName)
	return sql.Open("postgres", connStr)
}
