package database

import (
	"crypto/rand"
	"database/sql"
	"log"
	"math/big"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	ID       int
	Type     string
	Sender   int
	Receiver int
	Amount   int
	Status   string
}

func InitHistodyDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions
	                  (id INT,
						type TEXT,
						 sender INT,
						  receiver INT,
						  amount INT,
						   status TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

func AddTransaction(db *sql.DB, t *Transaction) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("INSERT INTO transactions(id, type, sender, receiver, amount, status) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	t.ID, _ = RandUID()
	_, err = stmt.Exec(t.ID, t.Type, t.Sender, t.Receiver, t.Amount, t.Status)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return t.ID, tx.Commit()
}

func QueryTransactionsBySender(db *sql.DB, sender int) ([]Transaction, error) {
	rows, err := db.Query("SELECT id, param, sender, receiver, status, amount FROM transactions WHERE sender=?", sender)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.Type, &t.Sender, &t.Receiver, &t.Status, &t.Amount)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func QueryTransactionsByReceiver(db *sql.DB, receiver int) ([]Transaction, error) {
	rows, err := db.Query("SELECT id, param, sender, receiver, status, amount FROM transactions WHERE receiver=?", receiver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.Type, &t.Sender, &t.Receiver, &t.Status, &t.Amount)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func QueryTransactionByID(db *sql.DB, id int) (*Transaction, error) {
	var t Transaction
	err := db.QueryRow("SELECT id, type, sender, receiver, amount, status FROM transactions WHERE id=?", id).Scan(&t.ID, &t.Type, &t.Sender, &t.Receiver, &t.Amount, &t.Status)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func UpdateStatus(historyDB *sql.DB, transaction Transaction) error {
	stmt, err := historyDB.Prepare("UPDATE transactions SET status=? WHERE id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(transaction.Status, transaction.ID)
	if err != nil {
		return err
	}
	return nil
}

func FilterTransactionsReceiver(db *sql.DB, status string, receiver int) ([]Transaction, error) {
	rows, err := db.Query("SELECT * FROM transactions WHERE status=? AND receiver=?", status, receiver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(&txn.ID, &txn.Type, &txn.Sender, &txn.Receiver, &txn.Amount, &txn.Status); err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return txns, nil
}

func FilterTransactionsSender(db *sql.DB, status string, sender int) ([]Transaction, error) {
	rows, err := db.Query("SELECT * FROM transactions WHERE status=? AND sender=?", status, sender)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(&txn.ID, &txn.Type, &txn.Sender, &txn.Receiver, &txn.Amount, &txn.Status); err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return txns, nil
}

func RandUID() (int, error) {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return 0, err
	}
	return int(randomInt.Int64()), nil
}
