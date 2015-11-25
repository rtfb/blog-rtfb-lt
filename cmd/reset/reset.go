package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/howeyc/gopass"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	usage = `passwd-reset. A helper to reset password for rtfblog.

Usage:
  passwd-reset <env var>
  passwd-reset -h | --help
  passwd-reset --version

Options:
  It takes a single argument -- a name of environment variable to look up a
  connection string.
  -h --help     Show this screen.
  --version     Show version.`
	defaultCookieSecret = "dont-forget-to-change-me"
)

func EncryptBcrypt(passwd string) (hash string, err error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	hash = string(hashBytes)
	return
}

func updateAuthorRow(connString, uname, passwd, fullname, email, www string) {
	db, err := sql.Open("postgres", connString)
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

func main() {
	args, err := docopt.Parse(usage, nil, true, "1.0", false)
	if err != nil {
		panic("Can't docopt.Parse!")
	}
	envVar := args["<env var>"].(string)
	fmt.Printf("Looking up connstr in $%s...\n", envVar)
	dbFile := os.Getenv(envVar)
	uname := "rtfb"
	fmt.Printf("New password: ")
	passwd := gopass.GetPasswd()
	fmt.Printf("Confirm: ")
	passwd2 := gopass.GetPasswd()
	if string(passwd2) != string(passwd) {
		panic("Passwords do not match")
	}
	fullname := "Vytautas Šaltenis"
	email := "vytas@rtfb.lt"
	www := "http://rtfb.lt/"
	updateAuthorRow(dbFile, uname, string(passwd), fullname, email, www)
}
