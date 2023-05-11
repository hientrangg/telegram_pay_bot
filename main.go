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
	"github.com/hientrangg/telegram_pay_bot/util"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

var (
	userDb *sql.DB
    historyDb *sql.DB
    inputSender = make(chan int)
    inputReceiver = make(chan int)
    inputValue = make(chan int)
    tranferOutput = make(chan string)
    inputCotpaySender = make(chan int)
    inputCotpaySenderUsername = make(chan string)
    inputCotpayReceiver = make(chan int)
    inputCotpayValue = make(chan int)
    cotpayOutput = make(chan string)
)

func init() {
	userDb, _ = database.InitDB("./userData.sqlite")
    historyDb, _ =database.InitHistodyDB("./history.sqlite")
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
			    msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			    switch update.Message.Command() {
			    case "start":
                    msg.Text = "WELCOME TO *** E-WALLET"
                    user := int(update.Message.From.ID)
				    userData, _ := manage.DoGetStatus(userDb, user)
                    StartKeyBoard := util.InitStartKeyboard(user, userData.Value)
				    msg.ReplyMarkup = StartKeyBoard
                case "testacc": //register test account
                    uid := rand.Intn(999999999)
                    err := manage.DoRegister(userDb, uid)
                    if err != nil {
                    msg.Text = "error while register, please try again"
                    continue
                    }   

                    msg.Text = "Register Successful"
                case "cid":
                    msg.Text = "chatID is " + strconv.Itoa(int(update.Message.Chat.ID))
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

                    valueInt, _ := strconv.Atoi(value)
                    user := int(update.Message.From.ID)
                    err := manage.DoDeposit(userDb, historyDb, user, valueInt)
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

                    valueInt, _ := strconv.Atoi(value)
                    user := int(update.Message.From.ID)
                    err := manage.DoWithdraw(userDb, historyDb, user, valueInt)
                    if err != nil {
                        msg.Text = "Error while withdraw, please try again"
                        bot.Send(msg)
                    }
                    msg.Text = "Withdraw Successful"
                    bot.Send(msg)
                case "Tranfer receiver UID":
                    receiver := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Receiver is " + receiver)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    receiverInt, _ := strconv.Atoi(receiver)
                    msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Tranfer value")
                    msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
                    if _, err := bot.Send(msg); err != nil {
                        panic(err)
                    }
                    sender := int(update.Message.From.ID)
                    inputSender <- sender
                    inputReceiver <- receiverInt
                case "Tranfer value":
                    value := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tranfer value " + value)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    valueInt, _ := strconv.Atoi(value)
                    inputValue <- valueInt
                    
                    status := <- tranferOutput
                    if status == "error" {
                        msg.Text = "Error while tranfer, please try again"
                        bot.Send(msg)
                    } else if status == "ok" {
                        msg.Text = "Tranfer successful"
                        bot.Send(msg)
                    }
                case "Cotpay receiver UID":
                    receiver := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Receiver is " + receiver)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    receiverInt, _ := strconv.Atoi(receiver)
                    msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Cotpay value")
                    msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
                    if _, err := bot.Send(msg); err != nil {
                        panic(err)
                    }
                    sender := int(update.Message.From.ID)
                    inputCotpaySender <- sender
                    inputCotpaySenderUsername <- update.Message.From.UserName
                    inputCotpayReceiver <- receiverInt
                case "Cotpay value":
                    value := update.Message.Text
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Cotpay value " + value)
                    if _, err := bot.Send(msg); err != nil {
					    log.Panic(err)
				    }

                    valueInt, _ := strconv.Atoi(value)
                    inputCotpayValue <- valueInt
                    
                    status := <- cotpayOutput
                    if status == "error" {
                        msg.Text = "Error while cotpay, please try again"
                        bot.Send(msg)
                    } else if status == "ok" {
                        msg.Text = "Cotpay request successful"
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
                user := int(update.CallbackQuery.From.ID)
				userData, err := manage.DoGetStatus(userDb, user)
				if err != nil {
					text := "error while get account value, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

				text := "Your value is " + strconv.Itoa(userData.Value)
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

            case "register":
                user := int(update.CallbackQuery.From.ID)
                err := manage.DoRegister(userDb, user)
                if err != nil {
                    text:= "your UID is " + strconv.Itoa(user) + " error while register, please try again: " + err.Error()
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Register Successful")

                if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
            case "cotpay":
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Cotpay receiver UID")
                msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
                if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
            case "lockvalue":
                user := int(update.CallbackQuery.From.ID)
				userData, err := manage.DoGetStatus(userDb, user)
				if err != nil {
					text := "error while get account value, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

				text := "Your lock value is " + strconv.Itoa(userData.Lock_value)
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
            case "allowvalue":
                user := int(update.CallbackQuery.From.ID)
				userData, err := manage.DoGetStatus(userDb, user)
				if err != nil {
					text := "error while get account value, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

				text := "Your allow value is " + strconv.Itoa(userData.Allow_value)
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
            case "uid":
                userUID := int(update.CallbackQuery.From.ID)
                text := "Your UID is " + strconv.Itoa(userUID)
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                bot.Send(msg)
            case "confirm-cotpay-receiver":
                receiver := int(update.CallbackQuery.From.ID)
                transaction, err := database.FilterTransactionsReceiver(historyDb, "pending", receiver)
                if err != nil {
                    // text := "error while cotpay, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error())
                    bot.Send(msg)
                    continue
                }

                if len(transaction) > 1 {
                    text := "error while cotpay, please try again !!!!"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                msg = tgbotapi.NewMessage(int64(transaction[0].Sender), "Recerver confirm, you need confirm to do tranfer ")
                msg.ReplyMarkup = util.SenderConfirmKeyboard
                bot.Send(msg)
            case "cancel-cotpay-receiver":
                receiver := int(update.CallbackQuery.From.ID)
                transaction, err := database.FilterTransactionsReceiver(historyDb, "pending", receiver)
                if err != nil {
                    text := "error while cotpay, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                if len(transaction) > 1 {
                    text := "error while cotpay, please try again !!!!"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                err = database.LockValue(userDb, transaction[0].Sender, -transaction[0].Amount)
                if err != nil {
                    text := "error while cotpay, please try again !!"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                transaction[0].Status = "cancel"
                database.UpdateStatus(historyDb, transaction[0])
                text := "receiver cancel cotpay, you will return lock value"
                msg = tgbotapi.NewMessage(int64(transaction[0].Sender), text)
                bot.Send(msg)
            case "confirm-cotpay-sender":
                sender := int(update.CallbackQuery.From.ID)
                transaction, err := database.FilterTransactionsSender(historyDb, "pending", sender)
                if err != nil {
                    text := "error while cotpay, please try again"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                if len(transaction) > 1 {
                    text := "error while cotpay, please try again !!!!"
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }
                err = manage.TranferCotpay(userDb,transaction[0])
                if err != nil {
                    text := "error while cotpay, please try again !! "
                    msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                    bot.Send(msg)
                    continue
                }

                transaction[0].Status = "done"
                database.UpdateStatus(historyDb, transaction[0])
                text := "Cotpay sucessful"
                msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
                bot.Send(msg)
			}
        }
	}
}
