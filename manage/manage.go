package manage

import (
	"database/sql"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

type UserData struct {
	Uid         string
	Value       string
	Allow_value string
	Lock_value  string
}

type CotpayParam struct {
	Sender   string
	Username string
	Receiver string
	Value    string
}

type TranferParam struct {
	Sender   string
	Receiver string
	Value    string
}

type DepositParam struct {
	Uid   int
	Value string
}

func RequestCotpay(bot *tgbotapi.BotAPI, userdb *sql.DB, historyDb *sql.DB, input chan CotpayParam, output chan string) {
	for {
		cotpayParam := <-input
		txIDInt, err := DoCotPay(bot, userdb, historyDb, cotpayParam.Sender, cotpayParam.Username, cotpayParam.Receiver, cotpayParam.Value)
		if err != nil {
			status := "error"
			output <- status
			continue
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

func Tranfer(userDb *sql.DB, historyDb *sql.DB, input chan TranferParam, output chan string) {
	for {
		tranferParam := <-input

		txIDInt, err := DoTranfer(userDb, historyDb, tranferParam.Sender, tranferParam.Receiver, tranferParam.Value)
		if err != nil {
			status := "error"
			output <- status
			continue
		}

		txID := strconv.Itoa(txIDInt)

		output <- txID
	}
}
func DoDeposit(userdb *sql.DB, historyDb *sql.DB, inputChan chan DepositParam, outputChan chan string) {
	for {
		depositParam := <-inputChan

		err := database.UpdateValue(userdb, strconv.Itoa(depositParam.Uid), depositParam.Value)
		if err != nil {
			outputChan <- "error"
			continue
		}

		if depositParam.Value[0] != '-' {
			t := database.Transaction{Type: "deposit", Sender: strconv.Itoa(depositParam.Uid), Receiver: strconv.Itoa(depositParam.Uid), Amount: depositParam.Value, Status: "Done"}

			txID, err := database.AddTransaction(historyDb, &t)
			if err != nil {
				outputChan <- "error"
				continue
			}

			txIDInt := strconv.Itoa(txID)

			outputChan <- txIDInt
		} else {
			t := database.Transaction{Type: "withdraw", Sender: strconv.Itoa(depositParam.Uid), Receiver: strconv.Itoa(depositParam.Uid), Amount: depositParam.Value[1:], Status: "Done"}

			txID, err := database.AddTransaction(historyDb, &t)
			if err != nil {
				outputChan <- "error"
				continue
			}

			txIDInt := strconv.Itoa(txID)

			outputChan <- txIDInt
		}
	}
}

func DoGetStatus(db *sql.DB, Uid string) (UserData, error) {
	value, lockValue, allowValue, err := database.QueryUserValue(db, Uid)
	if err != nil {
		return UserData{Uid: "0", Value: "0", Lock_value: "0", Allow_value: "0"}, err
	}

	User := UserData{Uid: Uid, Value: value, Lock_value: lockValue, Allow_value: allowValue}
	return User, nil
}

func DoTranfer(userDb *sql.DB, historyDb *sql.DB, sender, receiver, amount string) (int, error) {

	err := database.UpdateValue(userDb, sender, "-"+amount)
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

func DoRegister(db *sql.DB, uid string, username string) error {
	err := database.AddUser(db, uid, "0", username)
	if err != nil {
		return err
	}
	return nil
}

func DoCotPay(bot *tgbotapi.BotAPI, userdb *sql.DB, historyDb *sql.DB, sender string, senderUsername string, receiver string, amount string) (int, error) {

	receiverUid, _ := strconv.Atoi(receiver)
	err := database.LockValue(userdb, sender, amount)
	if err != nil {
		return 0, err
	}
	msg := tgbotapi.NewMessage(int64(receiverUid), "You have a Cotpay from "+senderUsername+" with UID "+sender+". BALANCE IS "+amount)
	bot.Send(msg)
	transaction := database.Transaction{Type: "cotpay", Sender: sender, Receiver: receiver, Amount: amount, Status: "pending"}
	txID, _ := database.AddTransaction(historyDb, &transaction)
	msg.Text = strconv.Itoa(txID)
	msg.ReplyMarkup = util.ReceiverConfirmKeyboard
	bot.Send(msg)
	return txID, nil
}
