package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/initbot"
	"github.com/hientrangg/telegram_pay_bot/manage"
	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

var (
	cache         = sync.Map{}
	userDb        *sql.DB
	historyDb     *sql.DB
	inputTranfer  = make(chan manage.TranferParam)
	outputTranfer = make(chan string)
	inputCotpay   = make(chan manage.CotpayParam)
	outputCotpay  = make(chan string)

	inputDeposit  = make(chan manage.DepositParam)
	outputDeposit = make(chan string)
)

func init() {
	userDb, _ = database.InitUserDB("./db/userData.sqlite")
	historyDb, _ = database.InitHistodyDB("./db/history.sqlite")
}
func main() {
	//init bot to get update and send message
	bot := initbot.Init(TELEGRAM_APITOKEN)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go manage.Tranfer(userDb, historyDb, inputTranfer, outputTranfer)
	go manage.RequestCotpay(bot, userDb, historyDb, inputCotpay, outputCotpay)
	go manage.DoDeposit(userDb, historyDb, inputDeposit, outputDeposit)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					homePage(bot, update)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know this command, please try again")
					bot.Send(msg)
				}

			} else if update.Message.ReplyToMessage != nil {
				switch update.Message.ReplyToMessage.Text {
				case "Deposit value":
					if !manage.IsNumeric(update.Message.Text) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
						bot.Send(msg)
						homePage(bot, update)
						continue
					}
					valueInt, _ := strconv.Atoi(update.Message.Text)
					if valueInt < 0 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit value must > 0")
						bot.Send(msg)
						homePage(bot, update)
						continue
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit value is "+update.Message.Text)
					bot.Send(msg)
					depositParam := manage.DepositParam{Uid: int(update.Message.From.ID), Value: valueInt}
					inputDeposit <- depositParam
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Doing deposit")
					bot.Send(msg)
					status := <-outputDeposit

					if status == "error" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while add tx, please try again")
						bot.Send(msg)
					} else if status == "error!" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while update value, please try again")
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit successful, txId: "+status)
						bot.Send(msg)
					}
					homePage(bot, update)
				case "Withdraw value":
					if !manage.IsNumeric(update.Message.Text) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
						bot.Send(msg)
						homePage(bot, update)
						continue
					}
					valueInt, _ := strconv.Atoi(update.Message.Text)
					if valueInt < 0 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw value must > 0")
						bot.Send(msg)
						homePage(bot, update)
						continue
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw value is "+update.Message.Text)
					bot.Send(msg)
					depositParam := manage.DepositParam{Uid: int(update.Message.From.ID), Value: -valueInt}
					inputDeposit <- depositParam
					status := <-outputDeposit

					if status == "error" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while withdraw, please try again")
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw successful")
						bot.Send(msg)
					}
					homePage(bot, update)
				case "Tranfer receiver UID":
					getTranferReceiverUID(bot, update.Message)

				case "Tranfer receiver username":
					getTranferReceiverUsername(bot, update.Message)

				case "Tranfer value":
					err := getTranferValue(bot, update)
					if err != nil {
						continue
					}
					homePage(bot, update)
				case "Cotpay receiver UID":
					getCotpayReceiver(bot, update.Message)

				case "Cotpay value":
					err := getCotpayValue(bot, update.Message)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
						bot.Send(msg)
						continue
					}

				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "defaut !!!!!!!!!!!!")
					bot.Send(msg)
				}
			}
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}

			var msg tgbotapi.MessageConfig
			switch update.CallbackQuery.Data {
			case "username":
				updateUsername(bot, update.CallbackQuery)
				homePage(bot, update)

			case "deposit":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Deposit value")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				bot.Send(msg)

			case "withdraw":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Withdraw value")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "status":
				getBalance(bot, update.CallbackQuery)
				homePage(bot, update)

			case "register":

				register(bot, update.CallbackQuery)
				homePage(bot, update)

			case "tranfer":
				text := "CHOOSE TRANFER TYPE"
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				msg.ReplyMarkup = util.TranferKeyboard
				bot.Send(msg)

			case "tranferUsername":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Tranfer receiver username")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				bot.Send(msg)

			case "tranferUid":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Tranfer receiver UID")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				bot.Send(msg)

			case "cotpay":

				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Cotpay receiver UID")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "lockvalue":
				getLockValue(bot, update.CallbackQuery)
				homePage(bot, update)

			case "allowvalue":
				getAllowValue(bot, update.CallbackQuery)
				homePage(bot, update)

			case "uid":
				getUID(bot, update.CallbackQuery)
				homePage(bot, update)

			case "confirm-cotpay-receiver":
				confirm_cotpay_receiver(bot, update.CallbackQuery.Message)
				homePage(bot, update)

			case "cancel-cotpay-receiver":
				cancel_cotpay_receiver(bot, update.CallbackQuery.Message)
				homePage(bot, update)

			case "confirm-cotpay-sender":
				confirm_cotpay_sender(bot, update.CallbackQuery.Message)
				homePage(bot, update)

			}
		}
	}
}

func homePage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "*** E-WALLET")
		userUID := int(update.CallbackQuery.From.ID)
		userName := update.CallbackQuery.From.UserName
		userData, _ := manage.DoGetStatus(userDb, userUID)
		StartKeyBoard := util.InitStartKeyboard(userUID, userName, userData.Value)
		msg.ReplyMarkup = StartKeyBoard
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "*** E-WALLET")
		userUID := int(update.Message.From.ID)
		userName := update.Message.From.UserName
		userData, _ := manage.DoGetStatus(userDb, userUID)
		StartKeyBoard := util.InitStartKeyboard(userUID, userName, userData.Value)
		msg.ReplyMarkup = StartKeyBoard
		bot.Send(msg)
	}
}

func getTranferReceiverUID(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	receiver := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	msg = tgbotapi.NewMessage(message.Chat.ID, "Tranfer value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	sender := int(message.From.ID)
	receiverInt, _ := strconv.Atoi(receiver)
	tx := database.Transaction{
		ID:       sender,
		Type:     "tranfer",
		Sender:   sender,
		Receiver: receiverInt,
		Amount:   0,
		Status:   "",
	}
	cache.Store(sender, tx)
}

func getTranferReceiverUsername(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	receiver := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	bot.Send(msg)

	uid, err := database.QueryUid(userDb, receiver)

	if err != nil {
		uid, _ = database.RandUID()
		uidstr := "11111" + strconv.Itoa(uid)
		uid, _ = strconv.Atoi(uidstr)
		manage.DoRegister(userDb, uid, receiver)
	}

	msg = tgbotapi.NewMessage(message.Chat.ID, "Tranfer value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	bot.Send(msg)
	sender := int(message.From.ID)
	tx := database.Transaction{
		ID:       sender,
		Type:     "tranfer",
		Sender:   sender,
		Receiver: uid,
		Amount:   0,
		Status:   "",
	}
	cache.Store(sender, tx)
}

func getTranferValue(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if !manage.IsNumeric(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
		bot.Send(msg)
		homePage(bot, update)
		return errors.New("invalid value")
	} else {
		sender := int(update.Message.From.ID)
		valueInt, _ := strconv.Atoi(update.Message.Text)
		if valueInt < 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tranfer value must > 0")
			bot.Send(msg)
			return errors.New("invalid value")
		} else {
			value := update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tranfer value "+value)
			bot.Send(msg)

			cacheValue, ok := cache.Load(sender)
			if !ok {
				return errors.New("pending tx not found")
			}
			tx, _ := cacheValue.(database.Transaction)
			tx.Amount = valueInt
			tranferParam := manage.TranferParam{
				Sender:   tx.Sender,
				Receiver: tx.Receiver,
				Value:    tx.Amount,
			}
			inputTranfer <- tranferParam
			cache.Delete(sender)
			txID := <-outputTranfer
			if txID == "error" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "error while do tranfer, please try again")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tranfer success, txID is: "+txID)
				bot.Send(msg)
			}
		}
	}
	return nil
}

func getCotpayReceiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	receiver := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	sender := int(message.From.ID)
	receiverInt, _ := strconv.Atoi(receiver)

	tx := manage.CotpayParam{
		Sender:   sender,
		Username: message.From.UserName,
		Receiver: receiverInt,
		Value:    0,
	}

	cache.Store(sender, tx)

	msg = tgbotapi.NewMessage(message.Chat.ID, "Cotpay value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
}

func getCotpayValue(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if !manage.IsNumeric(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid value, value must be number")
		bot.Send(msg)
		return errors.New("invalid value")
	} else {
		value := message.Text
		valueInt, _ := strconv.Atoi(value)
		if valueInt < 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "cotpay value must > 0")
			bot.Send(msg)
			return errors.New("invalid value")
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Cotpay value "+value)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			sender := int(message.From.ID)
			cacheValue, ok := cache.Load(sender)
			if !ok {
				return errors.New("pending tx not found")
			}
			tx, _ := cacheValue.(manage.CotpayParam)
			tx.Value = valueInt
			cotpayParam := manage.CotpayParam{
				Sender:   tx.Sender,
				Username: tx.Username,
				Receiver: tx.Receiver,
				Value:    tx.Value,
			}
			inputCotpay <- cotpayParam
			msg = tgbotapi.NewMessage(message.Chat.ID, "Sending cotpay request")
			bot.Send(msg)
			cache.Delete(sender)
			txID := <-outputCotpay
			if txID == "error" {
				msg := tgbotapi.NewMessage(message.Chat.ID, "error while send cotpay request, please try again")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Send cotpay request success, txID is: "+txID)
				bot.Send(msg)
			}

		}
	}
	return nil
}

func getBalance(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, user)
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		bot.Send(msg)

	} else {
		text := "Your value is " + strconv.Itoa(userData.Value)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func register(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	username := callback.From.UserName
	uid, err := database.QueryUid(userDb, username)
	if err != nil {
		err := manage.DoRegister(userDb, user, username)
		if err != nil {
			text := " error while register, please try again: " + err.Error()
			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Register Successful")
			bot.Send(msg)
		}
	} else {

		if uid != user {
			database.UpdateUid(userDb, user, username)
			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Register Successful !!!")
			bot.Send(msg)
		} else {
			err := manage.DoRegister(userDb, user, username)
			if err != nil {
				text := " error while register, please try again: " + err.Error()
				msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Register Successful")
				bot.Send(msg)
			}
		}
	}
}

func getLockValue(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, user)
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		bot.Send(msg)
	} else {

		text := "Your lock value is " + strconv.Itoa(userData.Lock_value)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func getAllowValue(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, user)
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		bot.Send(msg)
	} else {
		text := "Your allow value is " + strconv.Itoa(userData.Allow_value)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func getUID(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	userUID := int(callback.From.ID)
	text := "Your UID is " + strconv.Itoa(userUID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
	bot.Send(msg)
}

func confirm_cotpay_receiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		// text := "error while cotpay, please try again"
		msg := tgbotapi.NewMessage(message.Chat.ID, err.Error())
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(int64(transaction.Sender), "Recerver confirm, you need confirm to do tranfer ")
		bot.Send(msg)
		msg.Text = strconv.Itoa(txID)
		msg.ReplyMarkup = util.SenderConfirmKeyboard
		bot.Send(msg)
	}
}

func cancel_cotpay_receiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		text := "error while cotpay, please try again, txID is: " + strconv.Itoa(txID)
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)
		return err
	}
	err = database.LockValue(userDb, transaction.Sender, -transaction.Amount)
	if err != nil {
		text := "error while cotpay, please try again !!"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)
		return err
	}
	transaction.Status = "cancel"
	database.UpdateStatus(historyDb, *transaction)
	text := "receiver cancel cotpay, you will return lock value"
	msg := tgbotapi.NewMessage(int64(transaction.Sender), text)
	bot.Send(msg)
	return nil
}

func confirm_cotpay_sender(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		text := "error while cotpay, please try again"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)
		return err
	}

	err = manage.TranferCotpay(userDb, *transaction)
	if err != nil {
		text := "error while cotpay, please try again !! "
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)
		return err
	}

	transaction.Status = "Done"
	database.UpdateStatus(historyDb, *transaction)
	text := "Cotpay sucessful"
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	bot.Send(msg)
	return nil
}

// func openPincode(bot *tgbotapi.BotAPI, message *tgbotapi.Message, data string) {
// 	msg := tgbotapi.NewMessage(message.Chat.ID, "Enter pincode")
// 	msg.ReplyMarkup = util.InitPincodeKeyboard(data)
// 	bot.Send(msg)
// }

func updateUsername(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	username := callback.From.UserName
	uid := callback.From.ID
	err := database.UpdateUsername(userDb, int(uid), username)
	if err != nil {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, err.Error())
		bot.Send(msg)

		return err
	}
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Update username success")
	bot.Send(msg)

	return nil
}
