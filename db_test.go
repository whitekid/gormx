package gormx

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/fx"
)

func forEachSQLDriver(t *testing.T, testfn func(t *testing.T, dbURL string, reset func())) {
	fx.Each(fx.Of("sqlite", "mysql", "pgsql"), func(_ int, driver string) {
		if os.Getenv("GX_SKIP_SQL_"+strings.ToUpper(driver)) == "true" {
			return
		}

		forOneSQLDriver(t, driver, testfn)
	})
}

func forOneSQLDriver(t *testing.T, driver string, testfn func(t *testing.T, dbURL string, reset func())) {
	t.Run(driver, func(t *testing.T) {
		dbname := dbName(t.Name())
		dburl := ""
		var db *sql.DB
		var err error
		var reset = func() {}
		switch driver {
		case "sqlite":
			reset = func() { os.Remove(dbname + ".db") }
			dburl = fmt.Sprintf("sqlite://%s.db", dbname)

		case "mysql":
			db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/mysql")
			require.NoError(t, err)

			reset = func() {
				_, err := db.Exec("DROP DATABASE IF EXISTS " + dbname)
				require.NoError(t, err)

				_, err = db.Exec("CREATE DATABASE " + dbname)
				require.NoError(t, err)
			}

			dburl = fmt.Sprintf("mysql://root:@127.0.0.1:3306/%s?parseTime=true", dbname)

		case "pgsql":
			db, err := sql.Open("pgx", "dbname=postgres")
			require.NoError(t, err)

			reset = func() {
				_, err = db.Exec("DROP DATABASE IF EXISTS " + dbname)
				require.NoError(t, err)
				_, err = db.Exec("CREATE DATABASE " + dbname)
				require.NoError(t, err)
			}

			dburl = fmt.Sprintf("postgresql:///%s", dbname)

		default:
			require.Failf(t, "not supported scheme", driver)
		}

		reset()
		testfn(t, dburl, reset)
	})
}

func dbName(name string) string {
	return strings.ToLower(strings.NewReplacer(
		"/", "_",
		":", "_",
		"#", "_",
	).Replace(name))
}
