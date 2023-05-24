package main

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
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
	//check if string are in telegram username format
	isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9_-]*$`).MatchString

	//cache
	Txcache = sync.Map{}

	//database
	userDb        *sql.DB
	historyDb     *sql.DB

	//channel
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
					register(bot, update.Message)
					homePage(bot, update)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know this command, please try again")
					if _, err := bot.Send(msg); err != nil {
						panic(err)
					}
				}

			} else if update.Message.ReplyToMessage != nil {
				switch update.Message.ReplyToMessage.Text {
				case "Deposit value":
					err := deposit(bot, update)
					if err != nil {
						homePage(bot, update)
						continue
					}
					homePage(bot, update)
				case "Withdraw value":
					err := withdraw(bot, update)
					if err != nil {
						homePage(bot, update)
						continue
					}
					homePage(bot, update)
				case "Tranfer receiver UID":
					err := getTranferReceiverUID(bot, update.Message)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
						if _, err := bot.Send(msg); err != nil {
							panic(err)
						}
						homePage(bot, update)
					}
				case "Tranfer receiver username":
					err := getTranferReceiverUsername(bot, update.Message)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
						if _, err := bot.Send(msg); err != nil {
							panic(err)
						}
						homePage(bot, update)
					}

				case "Tranfer value":
					err := getTranferValue(bot, update)
					if err != nil {
						homePage(bot, update)
						continue
					}
					homePage(bot, update)
				case "Cotpay receiver UID":
					getCotpayReceiver(bot, update.Message)

				case "Cotpay value":
					err := getCotpayValue(bot, update.Message)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
						if _, err := bot.Send(msg); err != nil {
							panic(err)
						}
						homePage(bot, update)
						continue
					}
					homePage(bot, update)
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
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

			case "withdraw":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Withdraw value")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "status":
				getBalance(bot, update.CallbackQuery)
				homePage(bot, update)

			case "tranfer":
				text := "CHOOSE TRANFER TYPE"
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				msg.ReplyMarkup = util.TranferKeyboard
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

			case "tranferUsername":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Tranfer receiver username")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

			case "tranferUid":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Tranfer receiver UID")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

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
		msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "*** E-WALLET")
		userUID := int(update.CallbackQuery.From.ID)
		userName := update.CallbackQuery.From.UserName
		userData, _ := manage.DoGetStatus(userDb, strconv.Itoa(userUID))
		StartKeyBoard := util.InitStartKeyboard(userUID, userName, userData.Value)
		msg.ReplyMarkup = StartKeyBoard
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.From.ID, "*** E-WALLET")
		userUID := int(update.Message.From.ID)
		userName := update.Message.From.UserName
		userData, _ := manage.DoGetStatus(userDb, strconv.Itoa(userUID))
		StartKeyBoard := util.InitStartKeyboard(userUID, userName, userData.Value)
		msg.ReplyMarkup = StartKeyBoard
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	}
}

func deposit(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if !util.IsNumeric(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
		return errors.New("invalid value")
	}
	value := update.Message.Text
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit value is "+value)
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	depositParam := manage.DepositParam{Uid: int(update.Message.From.ID), Value: value}
	inputDeposit <- depositParam
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Doing deposit")
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	status := <-outputDeposit

	if status == "error" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while add tx, please try again")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit successful, txId: "+status)
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	}
	return nil
}

func withdraw(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if !util.IsNumeric(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
		return errors.New("invalid value")
	}

	value := update.Message.Text
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw value is "+update.Message.Text)
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	depositParam := manage.DepositParam{Uid: int(update.Message.From.ID), Value: "-" + value}
	inputDeposit <- depositParam
	status := <-outputDeposit

	if status == "error" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while withdraw, please try again")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw successful")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	}
	return nil
}
func getTranferReceiverUID(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	receiver := message.Text
	if !util.IsNumeric(message.Text) {
		return errors.New("invalid uid type")
	}

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
	tx := database.Transaction{
		ID:       sender,
		Type:     "tranfer",
		Sender:   strconv.Itoa(sender),
		Receiver: receiver,
		Amount:   "0",
		Status:   "",
	}
	Txcache.Store(sender, tx)
	return nil
}

func getTranferReceiverUsername(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	receiver := message.Text
	if !isStringAlphabetic(receiver) {
		return errors.New("invalid username type")
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}

	uid, err := database.QueryUid(userDb, receiver)

	if err != nil {
		uidInt, _ := database.RandUID()
		uid = "0x" + strconv.Itoa(uidInt)
		manage.DoRegister(userDb, uid, receiver)
	}

	sender := int(message.From.ID)
	tx := database.Transaction{
		ID:       sender,
		Type:     "tranfer",
		Sender:   strconv.Itoa(sender),
		Receiver: uid,
		Amount:   "0",
		Status:   "",
	}
	Txcache.Store(sender, tx)
	msg = tgbotapi.NewMessage(message.Chat.ID, "Tranfer value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	
	return nil
}

func getTranferValue(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if !util.IsNumeric(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid value, value must be number")
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
		sender := int(update.Message.From.ID)
		Txcache.Delete(sender)

		return errors.New("invalid value")
	} else {
		sender := int(update.Message.From.ID)

		value := update.Message.Text
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tranfer value "+value)
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}

		cacheValue, ok := Txcache.Load(sender)
		if !ok {
			return errors.New("pending tx not found")
		}
		tx, _ := cacheValue.(database.Transaction)
		tx.Amount = value
		tranferParam := manage.TranferParam{
			Sender:   tx.Sender,
			Receiver: tx.Receiver,
			Value:    tx.Amount,
		}

		inputTranfer <- tranferParam
		Txcache.Delete(sender)

		msg.Text = "Doing Tranfer"
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}

		txID := <-outputTranfer
		if txID == "error" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "error while do tranfer, please try again !!!")
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tranfer success, txID is: "+txID)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
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

	tx := manage.CotpayParam{
		Sender:   strconv.Itoa(sender),
		Username: message.From.UserName,
		Receiver: receiver,
		Value:    "0",
	}

	Txcache.Store(sender, tx)

	msg = tgbotapi.NewMessage(message.Chat.ID, "Cotpay value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
}

func getCotpayValue(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if !util.IsNumeric(message.Text) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid value, value must be number")
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return errors.New("invalid value")
	} else {
		value := message.Text
		valueInt, _ := strconv.Atoi(value)
		if valueInt < 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "cotpay value must > 0")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		
			return errors.New("invalid value")
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Cotpay value "+value)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			sender := int(message.From.ID)
			cacheValue, ok := Txcache.Load(sender)
			if !ok {
				return errors.New("pending tx not found")
			}
			tx, _ := cacheValue.(manage.CotpayParam)
			tx.Value = value
			cotpayParam := manage.CotpayParam{
				Sender:   tx.Sender,
				Username: tx.Username,
				Receiver: tx.Receiver,
				Value:    tx.Value,
			}
			inputCotpay <- cotpayParam
			msg = tgbotapi.NewMessage(message.Chat.ID, "Sending cotpay request")
			bot.Send(msg)
			Txcache.Delete(sender)
			txID := <-outputCotpay
			if txID == "error" {
				msg := tgbotapi.NewMessage(message.Chat.ID, "error while send cotpay request, please try again")
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Send cotpay request success, txID is: "+txID)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			
			}

		}
	}
	return nil
}

func getBalance(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, strconv.Itoa(user))
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	

	} else {
		text := "Your value is " + userData.Value
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func register(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user := int(message.From.ID)
	username := message.From.UserName
	uid, err := database.QueryUid(userDb, username)
	if err != nil {
		err := manage.DoRegister(userDb, strconv.Itoa(user), username)
		if err != nil {
			text := "error while register, please try again: "
			msg := tgbotapi.NewMessage(message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Register Successful")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		
		}
	} else {
		if uid != strconv.Itoa(user) {
			database.UpdateUid(userDb, strconv.Itoa(user), username)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Register Successful !!!")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		
		}
	}
}

func getLockValue(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, strconv.Itoa(user))
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
	} else {

		text := "Your lock value is " + userData.Lock_value
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func getAllowValue(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	user := int(callback.From.ID)
	userData, err := manage.DoGetStatus(userDb, strconv.Itoa(user))
	if err != nil {
		text := "error while get account value, please try again"
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
	} else {
		text := "Your allow value is " + userData.Allow_value
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
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

}

func confirm_cotpay_receiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		text := "error while cotpay, please try again"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
	} else {
		sender, _ := strconv.Atoi(transaction.Sender)
		msg := tgbotapi.NewMessage(int64(sender), "Recerver confirm, you need confirm to do tranfer ")
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		msg.Text = strconv.Itoa(txID)
		msg.ReplyMarkup = util.SenderConfirmKeyboard
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
	}
}

func cancel_cotpay_receiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		text := "error while cotpay, please try again, txID is: " + strconv.Itoa(txID)
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return err
	}
	err = database.LockValue(userDb, transaction.Sender, "-" + transaction.Amount)
	if err != nil {
		text := "error while cotpay, please try again !!"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return err
	}
	transaction.Status = "cancel"
	database.UpdateStatus(historyDb, *transaction)
	text := "receiver cancel cotpay, you will return lock value"
	sender, _ := strconv.Atoi(transaction.Sender)
	msg := tgbotapi.NewMessage(int64(sender), text)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	return nil
}

func confirm_cotpay_sender(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	txID, _ := strconv.Atoi(message.Text)
	transaction, err := database.QueryTransactionByID(historyDb, txID)
	if err != nil {
		text := "error while cotpay, please try again"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return err
	}

	err = manage.TranferCotpay(userDb, *transaction)
	if err != nil {
		text := "error while cotpay, please try again !! "
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return err
	}

	transaction.Status = "Done"
	database.UpdateStatus(historyDb, *transaction)
	text := "Cotpay sucessful"
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	return nil
}

func updateUsername(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	username := callback.From.UserName
	uid := callback.From.ID
	err := database.UpdateUsername(userDb, strconv.Itoa(int(uid)), username)
	if err != nil {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, err.Error())
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	
		return err
	}
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Update username success")
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}


	return nil
}
