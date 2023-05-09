package main

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/initbot"
	"github.com/hientrangg/telegram_pay_bot/manage"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

var (
	ValueDb *sql.DB
    inputSender = make(chan int64)
    inputReceiver = make(chan int64)
    inputValue = make(chan int64)
    output = make(chan string)
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("REGISTER", "register"),
		tgbotapi.NewInlineKeyboardButtonData("GET BALANCE", "status"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("DEPOSIT", "deposit"),
		tgbotapi.NewInlineKeyboardButtonData("TRANFER", "tranfer"),
		tgbotapi.NewInlineKeyboardButtonData("WITHDRAW", "withdraw"),
	),
)

func init() {
	ValueDb = database.InitDb()
}

func main() {
    go manage.Tranfer(inputSender, inputReceiver, inputValue, output)
	//init bot to get update
	bot := initbot.Init(TELEGRAM_APITOKEN)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
            if update.Message.IsCommand() {
			    msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			    switch update.Message.Command() {
			    case "start":
				    msg.ReplyMarkup = numericKeyboard
                case "testacc": //register test account
                    uid := rand.Int63n(999999999)
                    err := manage.DoRegister(uid)
                    if err != nil {
                    msg.Text = "error while register, please try again"
                    continue
                    }   

                    msg.Text = "Register Successful"
			    default:
				    msg.Text = "I don't know this command, please try again"
			    }
			    // Send the message.
			    if _, err := bot.Send(msg); err != nil {
				    panic(err)
			    }
            } else if update.Message.ReplyToMessage != nil {
                switch update.Message.ReplyToMessage.Text{
                case "Deposit value":
                    value := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit value is " + value)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    valueInt, _ := strconv.ParseInt(value, 10, 64)
                    err := manage.DoDeposit(update.Message.From.ID, valueInt)
                    if err != nil {
                        msg.Text = "Error while deposit, please try again"
                        bot.Send(msg)
                    }

                    msg.Text = "Deposit Sucessful"
                    bot.Send(msg)
                case "Withdraw value":
                    value := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw value is " + value)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    valueInt, _ := strconv.ParseInt(value, 10, 64)
                    err := manage.DoWithdraw(update.Message.From.ID, valueInt)
                    if err != nil {
                        msg.Text = "Error while withdraw, please try again"
                        bot.Send(msg)
                    }
                    msg.Text = "Withdraw Successful"
                    bot.Send(msg)
                case "Receiver UID":
                    receiver := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Receiver is " + receiver)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    receiverInt, _ := strconv.ParseInt(receiver, 10, 64)
                    msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Value")
                    msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
                    if _, err := bot.Send(msg); err != nil {
                        panic(err)
                    }
                    inputSender <- update.Message.From.ID
                    inputReceiver <- receiverInt
                case "Value":
                    value := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tranfer value " + value)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    valueInt, _ := strconv.ParseInt(value, 10, 64)
                    inputValue <- valueInt
                    
                    status := <- output
                    if status == "error" {
                        msg.Text = "Error while tranfer, please try again"
                        bot.Send(msg)
                    } else if status == "ok" {
                        msg.Text = "Tranfer successful"
                        bot.Send(msg)
                    }
                }
        }
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
			// And finally, send a message containing the data received.
			var msg tgbotapi.MessageConfig
			switch update.CallbackQuery.Data {
			case "deposit":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Deposit value")
                msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			case "tranfer":
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Receiver UID")
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
				msg.Text = "Get wallet status"
				value, err := manage.DoGetStatus(int64(update.CallbackQuery.From.ID))
				if err != nil {
					msg.Text = "error while get account value, please try again"
                    continue
                }

				text := "Your value is " + strconv.Itoa(int(value))
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

            case "register":
                err := manage.DoRegister(update.CallbackQuery.From.ID)
                if err != nil {
                    msg.Text = "error while register, please try again"
                    continue
                }

                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Register Successful")

                if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
        }
	}
}
