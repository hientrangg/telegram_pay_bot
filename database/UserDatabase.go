package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitUserDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			uid INTEGER PRIMARY KEY,
			username TEXT NOT NULL,
			passwd TEXT NOT NULL,
			value INTEGER NOT NULL,
			lockvalue INTEGER NOT NULL,
			allowvalue INTEGER NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AddUser(db *sql.DB, uid, value int, username, pincode string) error {
	stmt, err := db.Prepare(`
		INSERT INTO users (uid, username, passwd, value, lockvalue, allowvalue)
		VALUES (?, ?, ?, ?, 0, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	allowvalue := value
	_, err = stmt.Exec(uid, username, pincode, value, allowvalue)
	if err != nil {
		return err
	}

	return nil
}

func UpdateValue(db *sql.DB, uid, value int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue int
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	newValue := currentValue + value
	newAllowValue := newValue - lockValue
	if newAllowValue < 0 {
		tx.Rollback()
		return errors.New("not enough available funds")
	}

	_, err = tx.Exec(`UPDATE users SET value = ?, allowvalue = ? WHERE uid = ?`, newValue, newAllowValue, uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func LockValue(db *sql.DB, uid, value int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue int
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	newLockValue := lockValue + value
	newAllowValue := currentValue - newLockValue
	if newAllowValue < 0 {
		tx.Rollback()
		return errors.New("not enough available funds")
	}

	_, err = tx.Exec(`UPDATE users SET lockvalue = ?, allowvalue = ? WHERE uid = ?`, newLockValue, newAllowValue, uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func TranferLockValue(db *sql.DB, uid, value int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue int
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	newLockValue := lockValue - value
	newValue := currentValue - value
	newAllowValue := newValue - newLockValue
	if newLockValue < 0 {
		tx.Rollback()
		return errors.New("not enough locked funds")
	}

	_, err = tx.Exec(`UPDATE users SET value = ?, lockvalue = ?, allowvalue = ? WHERE uid = ?`, newValue, newLockValue, newAllowValue, uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func QueryUserValue(db *sql.DB, uid int) (int, int, int, error) {
	var value, lockValue, allowValue int
	err := db.QueryRow("SELECT value, lockvalue, allowvalue FROM users WHERE uid=?", uid).Scan(&value, &lockValue, &allowValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, 0, fmt.Errorf("user not found")
		}
		log.Fatalf("error querying user: %v", err)
	}
	return value, lockValue, allowValue, nil
}

func QueryPasswd(db *sql.DB, uid int) (string, error) {
	var passwd string 
	err := db.QueryRow("SELECT passwd FROM users WHERE uid=?", uid).Scan(&passwd)
	if err != nil {
		if err == sql.ErrNoRows {
			return "nil" ,fmt.Errorf("user not found")
		}
		log.Fatalf("error querying user: %v", err)
	}

	return passwd, nil
}

func QueryUid(db *sql.DB, username string) (int, error) {
	var uid int 
	err := db.QueryRow("SELECT uid FROM users WHERE username=?", username).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0 ,fmt.Errorf("username not found")
		}
		log.Fatalf("error querying username: %v", err)
	}

	return uid, nil
}

func UpdateUid(db *sql.DB, uid int, pincode, username string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	_, err = tx.Exec(`UPDATE users SET uid = ?, passwd = ? WHERE username = ?`, uid, pincode, username)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func UpdateUsername(db *sql.DB, uid int, username string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	_, err = tx.Exec(`UPDATE users SET username = ? WHERE uid = ?`, username, uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}