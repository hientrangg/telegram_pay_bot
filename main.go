package main

import (
	"database/sql"
	"log"
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
	//init bot to get update
	bot := initbot.Init(TELEGRAM_APITOKEN)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "start":
				msg.ReplyMarkup = numericKeyboard
			default:
				msg.Text = "I don't know this command, please try again"
			}
			// Send the message.
			if _, err := bot.Send(msg); err != nil {
				panic(err)
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

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

                bot2 := initbot.Init(TELEGRAM_APITOKEN)
				if err := manage.Deposit(bot2); err != nil {
					msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Error while deposit, please try again")
					if _, err := bot.Send(msg); err != nil {
						panic(err)
					}
				}
			case "tranfer":

			case "withdraw":
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Withdraw value")

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
                
                bot2 := initbot.Init(TELEGRAM_APITOKEN)
				if err := manage.Withdraw(bot2); err != nil {
					msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Error while withdraw, please try again")
					if _, err := bot.Send(msg); err != nil {
						panic(err)
					}
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
