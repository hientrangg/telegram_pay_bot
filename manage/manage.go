package manage

import (
	"database/sql"
	"strconv"
	"strings"
	"unicode"

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

func RequestCotpay(bot *tgbotapi.BotAPI, userdb *sql.DB, historyDb *sql.DB, inputSender chan int, inputSenderUsername chan string, inputReceiver chan int, inputValue chan int, inputStatus, output chan string) {
	for {
		sender := <-inputSender
		senderUsername := <-inputSenderUsername
		receiver := <-inputReceiver
		value := <-inputValue
		status := <- inputStatus

		if status != "ok" {
			continue
		}
		
		txIDInt, err := DoCotPay(bot, userdb, historyDb, sender, senderUsername, receiver, value)
		if err != nil {
			status := "error"
			output <- status
		}

		txID := strconv.Itoa(txIDInt)

		output <- txID
	}
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

func Tranfer(userDb *sql.DB, historyDb *sql.DB, inputSender chan int, inputReceiver chan int, inputValue chan int, inputStatus chan string, output chan string) {
	for {
		sender := <-inputSender
		receiver := <-inputReceiver
		value := <-inputValue
		status := <- inputStatus

		if status != "ok" {
			continue
		}

		txIDInt, err := DoTranfer(userDb, historyDb, sender, receiver, value)
		if err != nil {
			status := "error"
			output <- status
		}

		txID := strconv.Itoa(txIDInt)

		output <- txID
	}
}
func DoDeposit(userdb *sql.DB, historyDb *sql.DB, inputChan chan string, outputChan chan string) {
	for {
		Uid := <-inputChan
		amount := <-inputChan
		status := <-inputChan
		if status != "ok" {
			continue
		}
		UidInt, _ := strconv.Atoi(Uid)
		amountInt, _ := strconv.Atoi(amount)
		err := database.UpdateValue(userdb, UidInt, amountInt)
		if err != nil {
			outputChan <- "error"
			continue
		}

		t := database.Transaction{Type: "deposit", Sender: UidInt, Receiver: UidInt, Amount: amountInt, Status: "Done"}

		txID, _ := database.AddTransaction(historyDb, &t)

		txIDInt := strconv.Itoa(txID)

		outputChan <- txIDInt
	}
}

func DoGetStatus(db *sql.DB, Uid int) (UserData, error) {
	value, lockValue, allowValue, err := database.QueryUser(db, Uid)
	if err != nil {
		return UserData{Uid: 0, Value: 0, Lock_value: 0, Allow_value: 0}, err
	}

	User := UserData{Uid: Uid, Value: value, Lock_value: lockValue, Allow_value: allowValue}
	return User, nil
}

func DoTranfer(userDb *sql.DB, historyDb *sql.DB, sender, receiver, amount int) (int, error) {
	err := database.UpdateValue(userDb, sender, -amount)
	if err != nil {
		return 0, err
	}

	err = database.UpdateValue(userDb, receiver, amount)
	if err != nil {
		database.UpdateValue(userDb, sender, amount)
		return 0, err
	}

	t := database.Transaction{Type: "tranfer", Sender: sender, Receiver: receiver, Amount: amount, Status: "Done"}
	txID, _ := database.AddTransaction(historyDb, &t)
	return txID, nil
}

func DoRegister(db *sql.DB, uid int, pincode string) error {
	err := database.AddUser(db, uid, 0, pincode)
	if err != nil {
		return err
	}
	return nil
}

func DoCotPay(bot *tgbotapi.BotAPI, userdb *sql.DB, historyDb *sql.DB, sender int, senderUsername string, receiver int, amount int) (int, error) {
	err := database.LockValue(userdb, sender, amount)
	if err != nil {
		return 0, err
	}
	msg := tgbotapi.NewMessage(int64(receiver), "You have a Cotpay from "+senderUsername+" with UID "+strconv.Itoa(sender)+". BALANCE IS "+strconv.Itoa(amount))
	bot.Send(msg)
	transaction := database.Transaction{Type: "cotpay", Sender: sender, Receiver: receiver, Amount: amount, Status: "pending"}
	txID, _ := database.AddTransaction(historyDb, &transaction)
	msg.Text = strconv.Itoa(txID)
	msg.ReplyMarkup = util.ReceiverConfirmKeyboard
	bot.Send(msg)
	return txID, nil
}

func GetPincode(inputchan chan string, output chan string) {
	var pincode string
	for {
		num := <-inputchan

		switch num {
		case "0":
			pincode = pincode + num
		case "1":
			pincode = pincode + num
		case "2":
			pincode = pincode + num
		case "3":
			pincode = pincode + num
		case "4":
			pincode = pincode + num
		case "5":
			pincode = pincode + num
		case "6":
			pincode = pincode + num
		case "7":
			pincode = pincode + num
		case "8":
			pincode = pincode + num
		case "9":
			pincode = pincode + num
		case "<":
			pincode = strings.TrimSuffix(pincode, string(pincode[len(pincode)-1]))
		default:
			output <- pincode
			pincode = ""
		}
	}
}

func IsNumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

