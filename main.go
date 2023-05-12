package main

import (
	"database/sql"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/initbot"
	"github.com/hientrangg/telegram_pay_bot/manage"
	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6236521521:AAHLlRtOQHvns5DXlU28hCiIU0dYch2ByzU"
)

var (
	userDb                    *sql.DB
	historyDb                 *sql.DB
	inputSender               = make(chan int)
	inputReceiver             = make(chan int)
	inputValue                = make(chan int)
	tranferOutput             = make(chan string)
	inputCotpaySender         = make(chan int)
	inputCotpaySenderUsername = make(chan string)
	inputCotpayReceiver       = make(chan int)
	inputCotpayValue          = make(chan int)
	cotpayOutput              = make(chan string)
)

func init() {
	userDb, _ = database.InitDB("./userData.sqlite")
	historyDb, _ = database.InitHistodyDB("./history.sqlite")
}
func main() {
	//init bot to get update and send message
	bot := initbot.Init(TELEGRAM_APITOKEN)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go manage.Tranfer(userDb, historyDb, inputSender, inputReceiver, inputValue, tranferOutput)
	go manage.RequestCotpay(bot, userDb, historyDb, inputCotpaySender, inputCotpaySenderUsername, inputCotpayReceiver, inputCotpayValue, cotpayOutput)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					homePage(bot, update.Message)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know this command, please try again")

					bot.Send(msg)
				}

			} else if update.Message.ReplyToMessage != nil {
				switch update.Message.ReplyToMessage.Text {
				case "Deposit value":
					deposit(bot, update.Message)
					homePage(bot, update.Message)

				case "Withdraw value":
					withdraw(bot, update.Message)
					homePage(bot, update.Message)

				case "Tranfer receiver UID":
					getTranferReceiver(bot, update.Message)

				case "Tranfer value":
					getTranferValue(bot, update.Message)
					homePage(bot, update.Message)

				case "Cotpay receiver UID":
					getCotpayReceiver(bot, update.Message)

				case "Cotpay value":
					getCotpayValue(bot, update.Message)
					homePage(bot, update.Message)

				}
			}
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
			
			var msg tgbotapi.MessageConfig
			switch update.CallbackQuery.Data {
			case "deposit":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Deposit value")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "tranfer":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Tranfer receiver UID")
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

			case "register":
				register(bot, update.CallbackQuery)

			case "cotpay":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Cotpay receiver UID")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "lockvalue":
				getLockValue(bot, update.CallbackQuery)
				homePage(bot, update.CallbackQuery.Message)

			case "allowvalue":
				getAllowValue(bot, update.CallbackQuery)
				homePage(bot, update.CallbackQuery.Message)

			case "uid":
				getUID(bot, update.CallbackQuery)
				homePage(bot, update.CallbackQuery.Message)

			case "confirm-cotpay-receiver":
				confirm_cotpay_receiver(bot, update.CallbackQuery.Message)
				homePage(bot, update.CallbackQuery.Message)

			case "cancel-cotpay-receiver":
				cancel_cotpay_receiver(bot, update.CallbackQuery.Message)
				homePage(bot, update.CallbackQuery.Message)

			case "confirm-cotpay-sender":
				confirm_cotpay_sender(bot, update.CallbackQuery.Message)
                homePage(bot, update.CallbackQuery.Message)

			}
		}
	}
}

func homePage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "*** E-WALLET")
	user := int(message.From.ID)
	userData, _ := manage.DoGetStatus(userDb, user)
	StartKeyBoard := util.InitStartKeyboard(user, userData.Value)
	msg.ReplyMarkup = StartKeyBoard
	bot.Send(msg)
}

func deposit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	value := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Deposit value is "+value)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	valueInt, _ := strconv.Atoi(value)
	user := int(message.From.ID)
	txId, err := manage.DoDeposit(userDb, historyDb, user, valueInt)
	if err != nil {
		msg.Text = "Error while deposit, please try again"
		bot.Send(msg)
	}

	msg.Text = "Deposit Sucessful, txId is " + strconv.Itoa(txId)
	bot.Send(msg)
}

func withdraw(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	value := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Withdraw value is "+value)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	valueInt, _ := strconv.Atoi(value)
	user := int(message.From.ID)
	txID, err := manage.DoWithdraw(userDb, historyDb, user, valueInt)
	if err != nil {
		msg.Text = "Error while withdraw, please try again"
		bot.Send(msg)
	}
	msg.Text = "Withdraw Successful, txID is: " + strconv.Itoa(txID)
	bot.Send(msg)
}

func getTranferReceiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	receiver := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	receiverInt, _ := strconv.Atoi(receiver)
	msg = tgbotapi.NewMessage(message.Chat.ID, "Tranfer value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
	sender := int(message.From.ID)
	inputSender <- sender
	inputReceiver <- receiverInt
}

func getTranferValue(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	value := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Tranfer value "+value)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	valueInt, _ := strconv.Atoi(value)
	inputValue <- valueInt

	status := <-tranferOutput
	if status == "error" {
		msg.Text = "Error while tranfer, please try again"
		bot.Send(msg)
	} else {
		msg.Text = "Tranfer successful, txID is: " + status
		bot.Send(msg)
	}
}

func getCotpayReceiver(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	receiver := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Receiver is "+receiver)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	receiverInt, _ := strconv.Atoi(receiver)
	sender := int(message.From.ID)
	inputCotpaySender <- sender
	inputCotpaySenderUsername <- message.From.UserName
	inputCotpayReceiver <- receiverInt

	msg = tgbotapi.NewMessage(message.Chat.ID, "Cotpay value")
	msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
}

func getCotpayValue(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	value := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "Cotpay value "+value)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	valueInt, _ := strconv.Atoi(value)
	inputCotpayValue <- valueInt

	status := <-cotpayOutput
	if status == "error" {
		msg.Text = "Error while cotpay, please try again"
		bot.Send(msg)
	} else {
		msg.Text = "Cotpay request successful, TxID: " + status
		bot.Send(msg)
	}
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
	err := manage.DoRegister(userDb, user)
	if err != nil {
		text := "your UID is " + strconv.Itoa(user) + " error while register, please try again: " + err.Error()
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		bot.Send(msg)
	} else {

		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Register Successful")

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
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
