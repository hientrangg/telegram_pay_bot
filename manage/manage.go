package manage

import (
	"math/rand"
	"database/sql"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

type UserData struct {
	Uid         int
	Value       int
	Allow_value int
	Lock_value  int
}

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

func RequestCotpay(bot *tgbotapi.BotAPI ,userdb *sql.DB, historyDb *sql.DB, inputSender chan int, inputSenderUsername chan string, inputReceiver chan int, inputValue chan int, output chan string) {
	sender := <- inputSender
	senderUsername := <-inputSenderUsername
	receiver := <- inputReceiver
	value := <- inputValue

	err := DoCotPay(bot, userdb, historyDb, sender, senderUsername, receiver, value) 
	if err != nil {
		status := "error"
		output <- status
	}

	status := "ok"
	
	output <- status
}

func TranferCotpay(userDb *sql.DB, t database.Transaction) error {
	err := database.TranferLockValue(userDb, t.Sender, t.Amount)
	if err != nil {
		return err
	}

	err = database.UpdateValue(userDb, t.Receiver, t.Amount)
	if err != nil {
		return err
	}

	return nil
}

func Tranfer(userDb *sql.DB, historyDb *sql.DB, inputSender chan int, inputReceiver chan int, inputValue chan int, output chan string) {
	for {
		sender := <-inputSender
		receiver := <-inputReceiver
		value := <-inputValue

		err := DoTranfer(userDb, historyDb ,sender, receiver, value)
		if err != nil {
			status := "error"
			output <- status
		}

		status := "ok"
		output <- status
	}
}
func DoDeposit(userdb *sql.DB, historyDb *sql.DB, Uid int, amount int) error {
	err := database.UpdateValue(userdb, Uid, amount)
	if err != nil {
		return err
	}
	txId := rand.Intn(99999999)
	t := database.Transaction{ID: txId ,Type: "deposit", Sender: Uid, Receiver: Uid, Amount: amount, Status: "Done" }

	database.AddTransaction(historyDb, &t)

	return nil
}

func DoWithdraw(userDb *sql.DB, historyDb *sql.DB,Uid int, amount int) error {
	err := database.UpdateValue(userDb, Uid, -amount)
	if err != nil {
		return err
	}
	txId := rand.Intn(99999999)
	t := database.Transaction{ID: txId ,Type: "withdraw", Sender: Uid, Receiver: Uid, Amount: amount, Status: "Done" }
	database.AddTransaction(historyDb, &t)

	return nil
}

func DoGetStatus(db *sql.DB, Uid int) (UserData, error) {
	value, lockValue, allowValue, err := database.QueryUser(db, Uid)
	if err != nil {
		return UserData{Uid: 0, Value: 0, Lock_value: 0, Allow_value: 0}, err
	}

	User := UserData{Uid: Uid, Value: value, Lock_value: lockValue, Allow_value: allowValue}
	return User, nil
}

func DoTranfer(userDb *sql.DB, historyDb *sql.DB, sender, receiver, amount int) error {
	err := database.UpdateValue(userDb, sender, -amount)
	if err != nil {
		return err
	}

	err = database.UpdateValue(userDb, receiver, amount)
	if err != nil {
		return err
	}
	txId := rand.Intn(99999999)
	t := database.Transaction{ID: txId ,Type: "tranfer", Sender: sender, Receiver: receiver, Amount: amount, Status: "Done" }
	database.AddTransaction(historyDb, &t)
	return nil
}

func DoRegister(db *sql.DB, uid int) error {
	err := database.AddUser(db, uid, 0)
	if err != nil {
		return err
	}
	return nil
}

func DoCotPay(bot *tgbotapi.BotAPI ,userdb *sql.DB, historyDb *sql.DB, sender int, senderUsername string, receiver int, amount int) error {
	err := database.LockValue(userdb, sender, amount)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(int64(receiver), "You have a Cotpay from " + senderUsername + " with UID " + strconv.Itoa(sender) + ". BALANCE IS " + strconv.Itoa(amount)) 
	msg.ReplyMarkup = util.ReceiverConfirmKeyboard
	bot.Send(msg)
	txID := rand.Intn(999999)
	transaction := database.Transaction{ID: txID ,Type: "cotpay", Sender: sender, Receiver: receiver, Amount: amount, Status: "pending"}
	database.AddTransaction(historyDb, &transaction)
	return nil
}
