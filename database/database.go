package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDb() *sql.DB {
	db, err := sql.Open("sqlite3", "./mydb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create users table if it doesn't exist
	sqlStmt := `
        CREATE TABLE IF NOT EXISTS users (
            uid INTEGER PRIMARY KEY,
    		value INTEGER
        );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func QueryValue(uid int64) (int64, error) {
    // Open the database file
    db, err := sql.Open("sqlite3", "./mydb.sqlite")
    if err != nil {
        return 0, err
    }
    defer db.Close()

    // Query the value by UID
    var value int64
    err = db.QueryRow("SELECT value FROM users WHERE uid = ?", uid).Scan(&value)
    if err != nil {
        return 0, err
    }
    return value, nil
}

func UpdateValue(uid int64, value int64) error {
    // Open the database file
    db, err := sql.Open("sqlite3", "./mydb.sqlite")
    if err != nil {
        return err
    }
    defer db.Close()

    // Update the value by UID
    _, err = db.Exec("UPDATE users SET value = ? WHERE uid = ?", value, uid)
    if err != nil {
        return err
    }
    return nil
}

func Add(uid int64) error {
    // Open the database file
    db, err := sql.Open("sqlite3", "./mydb.sqlite")
    if err != nil {
        return err
    }
    defer db.Close()

    // Insert the new UID and value
    _, err = db.Exec("INSERT INTO users (uid, value) VALUES (?, ?)", uid, 0)
    if err != nil {
        return err
    }
    return nil
}
