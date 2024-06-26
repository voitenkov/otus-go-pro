package initstorage

import (
	"fmt"
	"log"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	memorystorage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/sql"
)

func New(cfg *config.Config) (app.Storage, error) {
	switch cfg.DB.Type {
	case "memory":
		return memorystorage.New(), nil
	case "sql":
		sqlConf := cfg.DB.SQL
		if sqlConf.Driver != "pgx" {
			log.Fatal("unsupported db driver is selected, pgx (postgresql) driver to be used")
		}
		return sqlstorage.New(cfg, GetDsn(sqlConf)), nil
	default:
		return nil, fmt.Errorf("unknown database type: %q", cfg.DB.Type)
	}
}

func GetDsn(sql config.SQLConf) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", sql.User, sql.Password, sql.Host, sql.Port, sql.Name)
}
