package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rtfb/gopass"
	"golang.org/x/crypto/bcrypt"
)

const (
	usage = `passwd-reset. A helper to reset password for rtfblog.

Usage:
  passwd-reset <env var or sqlite db file>
  passwd-reset -h | --help
  passwd-reset --version

Options:
  It takes a single argument -- a name of environment variable to look up a
  connection string.
  -h --help     Show this screen.
  --version     Show version.`
)

func EncryptBcrypt(passwd string) (hash string, err error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	hash = string(hashBytes)
	return
}

func updateAuthorRow(dialect, connString, uname, passwd, fullname, email, www string) {
	db, err := sql.Open(dialect, connString)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()
	stmt, _ := db.Prepare(`update author set
	disp_name=$1, passwd=$2, full_name=$3, email=$4, www=$5
	where id=1`)
	defer stmt.Close()
	passwdHash, err := EncryptBcrypt(passwd)
	if err != nil {
		fmt.Printf("Error in Encrypt(): %s\n", err)
		return
	}
	fmt.Printf("Updating user ID=1...\n")
	fmt.Printf("dbstr: %q\nuname: %q\npasswd: %q\nhash: %q\nfullname: %q\nemail: %q\nwww: %q\n",
		connString, uname, "***", passwdHash, fullname, email, www)
	stmt.Exec(uname, passwdHash, fullname, email, www)
}

func dbConn(param string) (dialect, conn string) {
	envVar := os.Getenv(param)
	if envVar != "" {
		dialect = "postgres"
		fmt.Printf("%s seems to be an env var, assuming %s\n", param, dialect)
		return dialect, envVar
	}
	dialect = "sqlite3"
	fmt.Printf("%s seems to be a file, assuming %s\n", param, dialect)
	return dialect, param
}

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		panic("Can't docopt.Parse!")
	}
	param := args["<env var or sqlite db file>"].(string)
	dialect, conn := dbConn(param)
	fmt.Printf("Connecting to %s@%s...\n", dialect, conn)
	uname := "rtfb"
	fmt.Printf("New password: ")
	passwd, err := gopass.GetPasswd()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Confirm: ")
	passwd2, err := gopass.GetPasswd()
	if err != nil {
		panic(err.Error())
	}
	if string(passwd2) != string(passwd) {
		panic("Passwords do not match")
	}
	fullname := "Vytautas Šaltenis"
	email := "vytas@rtfb.lt"
	www := "http://rtfb.lt/"
	updateAuthorRow(dialect, conn, uname, string(passwd), fullname, email, www)
}
