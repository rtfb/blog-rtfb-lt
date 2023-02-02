package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rtfb/bark"
	"gopkg.in/yaml.v2"
)

const (
	usage = `migrate-db. A tool to migrate production and staging DBs for rtfblog.

Usage:
  migrate-db [--db=<db>] [--env=<env>] [--srcenv=<env>]
  migrate-db -h | --help
  migrate-db --version

Options:
  --db=<dir>      Path to a dir where dbconf.yml lives [default: db].
  --env=<env>     Environment to migrate [default: staging].
  --srcenv=<env>  Environment to copy data from [default: production].
  -h --help       Show this screen.
  --version       Show version.`
)

var (
	logger *bark.Logger
)

type dbConfTopLevel struct {
	path string
	envs map[string]EnvDBConf
}

type EnvDBConf struct {
	Driver        string `yaml:"driver"`
	DSN           string `yaml:"dsn"`
	MigrationsDir string `yaml:"migrationsDir"`
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

func copyData(source, target EnvDBConf) {
	tDB, err := sql.Open(target.Driver, target.DSN)
	logger.LogIf(err)
	logger.LogIf(clearTable(tDB, "comment"))
	logger.LogIf(clearTable(tDB, "commenter"))
	logger.LogIf(clearTable(tDB, "post"))
	logger.LogIf(clearTable(tDB, "tag"))
	logger.LogIf(clearTable(tDB, "tagmap"))
	logger.LogIf(clearTable(tDB, "author"))
	sDB, err := sql.Open(source.Driver, source.DSN)
	logger.LogIf(err)
	logger.LogIf(copyTable(sDB, tDB, "author"))
	logger.LogIf(copyTable(sDB, tDB, "post"))
	logger.LogIf(copyTable(sDB, tDB, "commenter"))
	logger.LogIf(copyTable(sDB, tDB, "comment"))
	logger.LogIf(copyTable(sDB, tDB, "tag"))
	logger.LogIf(copyTable(sDB, tDB, "tagmap"))
	tDB.Close()
}

func gormConnStringToPostgresURL(conn string) (string, error) {
	params := strings.Split(conn, " ")
	if len(params) != 3 {
		return "", fmt.Errorf("bad conn string, expect 'user=u dbname=n password=p'")
	}
	kvs := map[string]string{}
	for i, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			return "", fmt.Errorf("bad %dth kv in conn string, expect key=value", i)
		}
		kvs[kv[0]] = kv[1]
	}
	return fmt.Sprintf("postgres://%s:%s@localhost:5432/%s",
		kvs["user"],
		kvs["password"],
		kvs["dbname"],
	), nil
}

func readDBConfYAML(location string) (dbConfTopLevel, error) {
	conf := dbConfTopLevel{
		path: location,
	}
	dbConf := path.Join(location, "dbconf.yml")
	if _, err := os.Stat(dbConf); os.IsNotExist(err) {
		return conf, fmt.Errorf("find Goose dbconf: %q", dbConf)
	}
	confBytes, err := ioutil.ReadFile(dbConf)
	if err != nil {
		return conf, fmt.Errorf("read %s: %v", dbConf, err)
	}
	err = yaml.Unmarshal(confBytes, &conf.envs)
	if err != nil {
		return conf, fmt.Errorf("parse %s: %v", dbConf, err)
	}
	return conf, nil
}

func createForEnv(conf dbConfTopLevel, env string) (*migrate.Migrate, error) {
	envConfig, ok := conf.envs[env]
	if !ok {
		return nil, fmt.Errorf("config does not contain environment %s", env)
	}
	if envConfig.Driver != "postgres" {
		return nil, fmt.Errorf("unsupported driver '%s'. Only 'postgres' is supported", envConfig.Driver)
	}
	source := "file://" + path.Join(conf.path, envConfig.MigrationsDir, "migrations")
	pgURL, err := gormConnStringToPostgresURL(envConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse conn string: %v", err)
	}
	return migrate.New(source, pgURL)
}

func doDataMigration(
	conf dbConfTopLevel,
	srcenv string,
	tgtenv string,
	targetM *migrate.Migrate,
) error {
	srcM, err := createForEnv(conf, srcenv)
	if err != nil {
		return fmt.Errorf("create migration object: %v", err)
	}
	srcVersion, dirty, err := srcM.Version()
	if err != nil {
		if err != migrate.ErrNilVersion { // allow source to be of unknown version
			return fmt.Errorf("get source env version: %v", err)
		}
	}
	if dirty {
		return fmt.Errorf("source env dirty")
	}
	targetM.Migrate(srcVersion)
	copyData(conf.envs[srcenv], conf.envs[tgtenv])
	targetM.Up()
	return nil
}

func main() {
	args, err := docopt.Parse(usage, nil, true, "0.2", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't docopt.Parse!")
		return
	}
	fmt.Println(args)
	db := args["--db"].(string)
	env := args["--env"].(string)
	srcenv := args["--srcenv"].(string)
	logger = bark.Create()
	conf, err := readDBConfYAML(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		return
	}
	m, err := createForEnv(conf, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create target migration object: %v\n", err)
		return
	}
	if env == "production" {
		m.Up()
	} else {
		err := doDataMigration(conf, srcenv, env, m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to do data migration %s->%s: %v\n", srcenv, env, err)
		}
	}
}
