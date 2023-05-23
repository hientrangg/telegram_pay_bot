package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

func InitUserDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			uid TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			value TEXT NOT NULL,
			lockvalue TEXT NOT NULL,
			allowvalue TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AddUser(db *sql.DB, uid string, value string, username string) error {
	stmt, err := db.Prepare(`
		INSERT INTO users (uid, username, value, lockvalue, allowvalue)
		VALUES (?, ?, ?, 0, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	allowvalue := value
	_, err = stmt.Exec(uid, username, value, allowvalue)
	if err != nil {
		return err
	}

	return nil
}

func UpdateValue(db *sql.DB, uid string, value string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue string
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	valueInt, _ := util.String2BigInt(value)
	currentValueInt, _ := util.String2BigInt(currentValue)
	lockValueInt, _ := util.String2BigInt(lockValue)

	newValue := new(big.Int)
	newValue = newValue.Add(currentValueInt, valueInt)
	newAllowValue := new(big.Int)
	newAllowValue = newAllowValue.Sub(newValue, lockValueInt)
	if newAllowValue.Cmp(zero) == -1 {
		tx.Rollback()
		return errors.New("not enough available funds")
	}

	_, err = tx.Exec(`UPDATE users SET value = ?, allowvalue = ? WHERE uid = ?`, newValue.String(), newAllowValue.String(), uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}


func LockValue(db *sql.DB, uid string, value string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue string
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	valueInt, _ := util.String2BigInt(value)
	currentValueInt, _ := util.String2BigInt(currentValue)
	lockValueInt, _ := util.String2BigInt(lockValue)

	newLockValue := new(big.Int)
	newLockValue.Add(lockValueInt, valueInt)
	newAllowValue := new(big.Int)
	newAllowValue.Sub(currentValueInt, newLockValue)
	if newAllowValue.Cmp(zero) == -1 {
		tx.Rollback()
		return errors.New("not enough available funds")
	}

	_, err = tx.Exec(`UPDATE users SET lockvalue = ?, allowvalue = ? WHERE uid = ?`, newLockValue.String(), newAllowValue.String(), uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func TranferLockValue(db *sql.DB, uid string, value string ) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT value, lockvalue, allowvalue FROM users WHERE uid = ?`, uid)
	var currentValue, lockValue, allowValue string
	err = row.Scan(&currentValue, &lockValue, &allowValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	valueInt, _ := util.String2BigInt(value)
	currentValueInt, _ := util.String2BigInt(currentValue)
	lockValueInt, _ := util.String2BigInt(lockValue)

	newLockValue := new(big.Int)
	newLockValue.Sub(lockValueInt, valueInt)
	newValue := new(big.Int)
	newValue.Sub(currentValueInt, valueInt)
	newAllowValue := new(big.Int)
	newAllowValue.Sub(newValue, newLockValue)
	if newLockValue.Cmp(zero) == -1 {
		tx.Rollback()
		return errors.New("not enough locked funds")
	}

	_, err = tx.Exec(`UPDATE users SET value = ?, lockvalue = ?, allowvalue = ? WHERE uid = ?`, newValue.String(), newLockValue.String(), newAllowValue.String(), uid)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func QueryUserValue(db *sql.DB, uid string) (string, string, string, error) {
	var value, lockValue, allowValue string
	err := db.QueryRow("SELECT value, lockvalue, allowvalue FROM users WHERE uid=?", uid).Scan(&value, &lockValue, &allowValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return "0", "0", "0", fmt.Errorf("user not found")
		}
		log.Fatalf("error querying user: %v", err)
	}
	return value, lockValue, allowValue, nil
}

func QueryUid(db *sql.DB, username string) (string, error) {
	var uid string
	err := db.QueryRow("SELECT uid FROM users WHERE username=?", username).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return "0", fmt.Errorf("username not found")
		}
		log.Fatalf("error querying username: %v", err)
	}

	return uid, nil
}

func UpdateUid(db *sql.DB, uid string, username string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE users SET uid = ? WHERE username = ?`, uid, username)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func UpdateUsername(db *sql.DB, uid string, username string) error {
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
