package main

import (
	"fmt"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/docopt/docopt-go"
	"github.com/jinzhu/gorm"
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
	dbconf, err := goose.NewDBConf(db, env, "")
	logger.LogIf(err)
	target, err := goose.GetMostRecentDBVersion(dbconf.MigrationsDir)
	logger.LogIf(err)
	err = goose.RunMigrations(dbconf, dbconf.MigrationsDir, target)
	logger.LogIf(err)
}

func migrateToTarget(db, env string, target int64) {
	dbconf, err := goose.NewDBConf(db, env, "")
	logger.LogIf(err)
	err = goose.RunMigrations(dbconf, dbconf.MigrationsDir, target)
	logger.LogIf(err)
}

func copyData(source *goose.DBConf, db, env string) {
	targetConf, err := goose.NewDBConf(db, env, "")
	logger.LogIf(err)
	tDB, err := gorm.Open(targetConf.Driver.Name, targetConf.Driver.OpenStr)
	logger.LogIf(err)
	tDB.SingularTable(true)
	logger.LogIf(tDB.Delete(EntryTable{}).Error)
	logger.LogIf(tDB.Delete(TagMap{}).Error)
	logger.LogIf(tDB.Delete(Tag{}).Error)
	logger.LogIf(tDB.Delete(CommentTable{}).Error)
	logger.LogIf(tDB.Delete(CommenterTable{}).Error)
	logger.LogIf(tDB.Delete(Author{}).Error)
	sDB, err := gorm.Open(source.Driver.Name, source.Driver.OpenStr)
	sDB.SingularTable(true)
	logger.LogIf(err)
	// author
	var authors []Author
	logger.LogIf(sDB.Find(&authors).Error)
	for _, a := range authors {
		logger.LogIf(tDB.Create(&a).Error)
	}
	// post
	var posts []EntryTable
	logger.LogIf(sDB.Find(&posts).Error)
	for _, p := range posts {
		logger.LogIf(tDB.Create(&p).Error)
	}
	// commenter
	var commenters []CommenterTable
	logger.LogIf(sDB.Find(&commenters).Error)
	for _, c := range commenters {
		logger.LogIf(tDB.Create(&c).Error)
	}
	// comment
	var comments []CommentTable
	logger.LogIf(sDB.Find(&comments).Error)
	for _, c := range comments {
		logger.LogIf(tDB.Create(&c).Error)
	}
	// tag
	var tags []Tag
	logger.LogIf(sDB.Find(&tags).Error)
	for _, t := range tags {
		logger.LogIf(tDB.Create(&t).Error)
	}
	// tagmap
	var tagmaps []TagMap
	logger.LogIf(sDB.Find(&tagmaps).Error)
	for _, tm := range tagmaps {
		logger.LogIf(tDB.Create(&tm).Error)
	}
	tDB.Close()
}

func main() {
	args, err := docopt.Parse(usage, nil, true, "0.1", false)
	if err != nil {
		panic("Can't docopt.Parse!")
	}
	fmt.Println(args)
	fmt.Printf("db=%q, env=%q\n", args["--db"], args["--env"])
	db := args["--db"].(string)
	env := args["--env"].(string)
	logger = bark.Create()
	if env == "production" {
		migrateToLatest(db, env)
		return
	} else {
		prodConf, err := goose.NewDBConf(db, "production", "")
		logger.LogIf(err)
		target, err := goose.GetDBVersion(prodConf)
		logger.LogIf(err)
		migrateToTarget(db, env, target)
		copyData(prodConf, db, env)
		migrateToLatest(db, env)
	}
}
