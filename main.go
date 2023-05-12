package main

import (
	"bytes"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // https://www.calhoun.io/why-we-import-sql-drivers-with-the-blank-identifier/
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbName   = "phone"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	db, err := sql.Open("postgres", psqlInfo)
	checkErr(err)
	err = resetDB(db, dbName)
	checkErr(err)
	db.Close()

	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbName)
	db, err = sql.Open("postgres", psqlInfo)
	checkErr(err)
	defer db.Close()

	checkErr(createPhoneNumbersTable(db))
	id, err := insertPhone(db, "1234567890")
	checkErr(err)
	fmt.Println("New record ID:", id)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func insertPhone(db *sql.DB, phone string) (int, error) {
	statement := `INSERT INTO phone_numbers(value) VALUES($1) RETURNING id`
	var id int
	err := db.QueryRow(statement, phone).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func createPhoneNumbersTable(db *sql.DB) error {
	statement := `
	CREATE TABLE IF NOT EXISTS phone_numbers (
		id SERIAL,
		value VARCHAR(255)
	)`
	_, err := db.Exec(statement)
	return err
}

func resetDB(db *sql.DB, name string) error {
	_, err := db.Exec("DROP DATABASE IF EXISTS " + name)
	if err != nil {
		return err
	}
	return createDB(db, name)
}

func createDB(db *sql.DB, name string) error {
	_, err := db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return err
	}
	return nil
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
