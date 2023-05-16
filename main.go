package main

import (
	"bytes"
	"flag"
	"fmt"

	_ "github.com/lib/pq" // https://www.calhoun.io/why-we-import-sql-drivers-with-the-blank-identifier/
	phonedb "github.com/rafalmp/phone/db"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbName   = "phone"
)

func main() {
	doResetDB := flag.Bool("r", false, "(Re)create the database on startup")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	if *doResetDB {
		checkErr(phonedb.Reset("postgres", psqlInfo, dbName))
	}

	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbName)
	checkErr(phonedb.Migrate("postgres", psqlInfo))
	db, err := phonedb.Open("postgres", psqlInfo)
	checkErr(err)
	defer db.Close()

	checkErr(db.Seed())

	phones, err := db.AllPhones()
	checkErr(err)
	for _, p := range phones {
		fmt.Printf("Processing %+v\n", p)
		number := normalize(p.Number)
		if number != p.Number {
			fmt.Println("Updating or removing ", number)
			existing, err := db.FindPhone(number)
			checkErr(err)
			if existing != nil {
				checkErr(db.DeletePhone(p.ID))
			} else {
				p.Number = number
				checkErr(db.UpdatePhone(&p))
			}
		} else {
			fmt.Println("No changes required")
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func normalize(phone string) string {
	var buf bytes.Buffer
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}
