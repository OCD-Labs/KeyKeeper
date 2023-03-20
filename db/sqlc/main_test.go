package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/OCD-Labs/KeyKeeper/internal/util"
	_ "github.com/lib/pq"
)

var testQuerier Querier

func TestMain(m *testing.M) {
	config, err := util.ParseConfigs("../../configs/")
	if err != nil {
		log.Fatalf("cannot parse configs: %v", err)
	}

	testdb, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot open db connection: %v", err)
	}

	testQuerier = New(testdb)

	os.Exit(m.Run())
}
