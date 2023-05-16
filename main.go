package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq" // https://www.calhoun.io/why-we-import-sql-drivers-with-the-blank-identifier/
	phonedb "github.com/rafalmp/phone/db"
)

type phone struct {
	id     int
	number string
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbName   = "phone"
)

var numbers []string = []string{
	"1234567890",
	"123 456 7891",
	"(123) 456 7892",
	"(123) 456-7893",
	"123-456-7894",
	"123-456-7890",
	"1234567892",
	"(123)456-7892",
}

func main() {
	doResetDB := flag.Bool("r", false, "(Re)create the database on startup")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	if *doResetDB {
		checkErr(phonedb.Reset("postgres", psqlInfo, dbName))
	}
	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbName)
	db, err := sql.Open("postgres", psqlInfo)
	checkErr(err)
	defer db.Close()

	var id int
	checkErr(createPhoneNumbersTable(db))
	for _, n := range numbers {
		id, err = insertPhone(db, n)
		checkErr(err)
	}

	num, err := getPhone(db, id)
	checkErr(err)
	fmt.Println("Last inserted number is", num)

	phones, err := allPhones(db)
	checkErr(err)
	for _, p := range phones {
		fmt.Printf("Processing %+v\n", p)
		number := normalize(p.number)
		if number != p.number {
			fmt.Println("Updating or removing ", number)
			existing, err := findPhone(db, number)
			checkErr(err)
			if existing != nil {
				checkErr(deletePhone(db, p.id))
			} else {
				p.number = number
				checkErr(updatePhone(db, p))
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

func insertPhone(db *sql.DB, phone string) (int, error) {
	statement := `INSERT INTO phone_numbers(value) VALUES($1) RETURNING id`
	var id int
	err := db.QueryRow(statement, phone).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func getPhone(db *sql.DB, id int) (string, error) {
	var number string
	err := db.QueryRow("SELECT value FROM phone_numbers WHERE id=$1", id).Scan(&number)
	if err != nil {
		return "", err
	}
	return number, nil
}

func findPhone(db *sql.DB, number string) (*phone, error) {
	var p phone
	row := db.QueryRow("SELECT * FROM phone_numbers WHERE value=$1", number)
	err := row.Scan(&p.id, &p.number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &p, nil
}

func updatePhone(db *sql.DB, p phone) error {
	statement := `UPDATE phone_numbers SET value=$2 WHERE id=$1`
	_, err := db.Exec(statement, p.id, p.number)
	return err
}

func deletePhone(db *sql.DB, id int) error {
	statement := `DELETE FROM phone_numbers WHERE id=$1`
	_, err := db.Exec(statement, id)
	return err
}

func allPhones(db *sql.DB) ([]phone, error) {
	rows, err := db.Query("SELECT id, value FROM phone_numbers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []phone

	for rows.Next() {
		var p phone
		if err := rows.Scan(&p.id, &p.number); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

func normalize(phone string) string {
	var buf bytes.Buffer
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}
