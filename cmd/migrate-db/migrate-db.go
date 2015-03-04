package main

import (
	"fmt"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/docopt/docopt-go"
	"github.com/rtfb/bark"
)

const (
	usage = `migrate-db. A tool to migrate production and staging DBs for rtfblog.

Usage:
  migrate-db [--db=<db>] [--env=<env>]
  migrate-db -h | --help
  migrate-db --version

Options:
  --db=<dir>    Path to DB specifier dir for goose [default: db].
  --env=<env>   Environment to migrate [default: staging].
  -h --help     Show this screen.
  --version     Show version.`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, "0.1", false)
	if err != nil {
		panic("Can't docopt.Parse!")
	}
	fmt.Println(args)
	fmt.Printf("db=%q, env=%q\n", args["--db"], args["--env"])
	db := args["--db"].(string)
	env := args["--env"].(string)
	/*
		TODO: if env=staging, do the following:
		1. migrate staging DB *down* to match prod DB
		2. copy over all data from prod DB to staging DB
		3. migrate staging DB up to latest
	*/
	logger := bark.Create()
	dbconf, err := goose.NewDBConf(db, env, "")
	logger.LogIf(err)
	target, err := goose.GetMostRecentDBVersion(dbconf.MigrationsDir)
	logger.LogIf(err)
	err = goose.RunMigrations(dbconf, dbconf.MigrationsDir, target)
	logger.LogIf(err)
}
