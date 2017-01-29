package main

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docopt/docopt-go"
	_ "github.com/lib/pq"
	"github.com/rtfb/bark"
	"github.com/steinbacher/goose"
)

const (
	usage = `migrate-db. A tool to migrate production and staging DBs for rtfblog.

Usage:
  migrate-db [--db=<db>] [--env=<env>] [--srcenv=<env>]
  migrate-db -h | --help
  migrate-db --version

Options:
  --db=<dir>      Path to DB specifier dir for goose [default: db].
  --env=<env>     Environment to migrate [default: staging].
  --srcenv=<env>  Environment to copy data from [default: production].
  -h --help       Show this screen.
  --version       Show version.`
)

var (
	logger *bark.Logger
)

func L10n(x string, y int) string {
	return ""
}

func Capitalize(s string) string {
	return ""
}

func migrateToLatest(db, env string) {
	dbconf, err := goose.NewDBConf(db, env)
	logger.LogIf(err)
	target, err := goose.GetMostRecentDBVersion(dbconf.MigrationsDir)
	logger.LogIf(err)
	err = goose.RunMigrations(dbconf, dbconf.MigrationsDir, target)
	logger.LogIf(err)
}

func migrateToTarget(db, env string, target int64) {
	dbconf, err := goose.NewDBConf(db, env)
	logger.LogIf(err)
	err = goose.RunMigrations(dbconf, dbconf.MigrationsDir, target)
	logger.LogIf(err)
}

func clearTable(db *sql.DB, table string) error {
	fmt.Printf("Clearing table %v...\n", table)
	_, err := db.Exec("delete from " + table)
	return err
}

func makePlaceholders(num int) string {
	items := make([]string, 0)
	for i := 0; i < num; i++ {
		items = append(items, fmt.Sprintf("$%d", i+1))
	}
	return strings.Join(items, ", ")
}

func copyTable(src, dst *sql.DB, table string) error {
	fmt.Printf("Copying table %v...\n", table)
	rows, err := src.Query("select * from " + table)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		ss := make([]interface{}, 0)
		for i := 0; i < len(columns); i++ {
			var x interface{}
			ss = append(ss, &x)
		}
		rows.Scan(ss...)
		sql := fmt.Sprintf("insert into %s (%s) values (%s)", table,
			strings.Join(columns, ", "), makePlaceholders(len(columns)))
		_, err = dst.Exec(sql, ss...)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyData(source *goose.DBConf, db, env string) {
	targetConf, err := goose.NewDBConf(db, env)
	logger.LogIf(err)
	tDB, err := sql.Open(targetConf.Driver.Name, targetConf.Driver.DSN)
	logger.LogIf(err)
	logger.LogIf(clearTable(tDB, "comment"))
	logger.LogIf(clearTable(tDB, "commenter"))
	logger.LogIf(clearTable(tDB, "post"))
	logger.LogIf(clearTable(tDB, "tag"))
	logger.LogIf(clearTable(tDB, "tagmap"))
	logger.LogIf(clearTable(tDB, "author"))
	sDB, err := sql.Open(source.Driver.Name, source.Driver.DSN)
	logger.LogIf(err)
	logger.LogIf(copyTable(sDB, tDB, "author"))
	logger.LogIf(copyTable(sDB, tDB, "post"))
	logger.LogIf(copyTable(sDB, tDB, "commenter"))
	logger.LogIf(copyTable(sDB, tDB, "comment"))
	logger.LogIf(copyTable(sDB, tDB, "tag"))
	logger.LogIf(copyTable(sDB, tDB, "tagmap"))
	tDB.Close()
}

func main() {
	args, err := docopt.Parse(usage, nil, true, "0.1", false)
	if err != nil {
		panic("Can't docopt.Parse!")
	}
	fmt.Println(args)
	db := args["--db"].(string)
	env := args["--env"].(string)
	srcenv := args["--srcenv"].(string)
	dbConf := path.Join(db, "dbconf.yml")
	if _, err := os.Stat(dbConf); os.IsNotExist(err) {
		fmt.Printf("Can't find Goose dbconf: %q, exiting.\n", dbConf)
		return
	}
	logger = bark.Create()
	if env == "production" {
		migrateToLatest(db, env)
		return
	} else {
		prodConf, err := goose.NewDBConf(db, srcenv)
		logger.LogIf(err)
		target, err := goose.GetDBVersion(prodConf)
		logger.LogIf(err)
		migrateToTarget(db, env, target)
		copyData(prodConf, db, env)
		migrateToLatest(db, env)
	}
}
