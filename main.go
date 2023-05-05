package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hientrangg/telegram_pay_bot/database"
	"github.com/hientrangg/telegram_pay_bot/manage"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

var (
    ValueDb *sql.DB
)

func init() {
    ValueDb = database.InitDb()
}

func main() {
    bot, err := tgbotapi.NewBotAPI(TELEGRAM_APITOKEN)
    if err != nil {
        log.Panic(err)
    }
    
    bot.Debug = true
    log.Printf("Authorized on account %s", bot.Self.UserName)
    
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    
    updates, err := bot.GetUpdatesChan(u)
    if err != nil {
        panic(err)
    }

    for update := range updates {
        if update.Message == nil { // ignore any non-Message updates
            continue
        }

        if !update.Message.IsCommand() { // ignore any non-command Messages
            continue
        }

        // Create a new MessageConfig. We don't have text yet,
        // so we leave it empty.
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

        // Extract the command from the Message.
        switch update.Message.Command() {
        case "help":
            msg.Text = "use /deposit to deposit; /tranfer to tranfer; /withdraw to withdraw; /status to get wallet status; /register to register."
        
        case "register":
            msg.Text = "REGISTER"
            err := manage.Register(int64(update.Message.From.ID))
            if err != nil {
                msg.Text = "error while register, please try again"
            }
        case "deposit":
            msg.Text = "DEPOSIT"
            value, err := SplitValue(update.Message.CommandArguments())
            if err != nil {
                msg.Text = "Please provide value"
                continue
            }
            
            //Process value
            valueInt, err := strconv.ParseInt(value, 10, 64)
            if err != nil {
                msg.Text = "Error white deposit, please try again"
                continue
            }

            err = manage.Deposit(int64(update.Message.From.ID), valueInt)
            if err != nil {
                msg.Text = "Error white deposit, please try again"
            }

        case "tranfer":
            msg.Text = "TRANFER"

        case "withdraw":
            msg.Text = "WITHDRAW"
            value, err := SplitValue(update.Message.CommandArguments())
            if err != nil {
                msg.Text = "Please provide value"
                continue
            }

            //Process value
            valueInt, err := strconv.ParseInt(value, 10, 64)
            if err != nil {
                msg.Text = "Error white withdraw, please try again"
                continue
            }
            err = manage.Withdraw(int64(update.Message.From.ID), valueInt)
            if err != nil {
                msg.Text = "Error white withdraw, please try again"
            }

        case "status":
            msg.Text = "Get wallet status"
            value, err := manage.GetStatus(int64(update.Message.From.ID))
            if err != nil {
                msg.Text = "error while get account value, please try again"
            }

            msg.Text = strconv.Itoa(int(value))
        default:
            msg.Text = "I don't know that command"
        }

        if _, err := bot.Send(msg); err != nil {
            log.Panic(err)
        }
    }
}
    
func SplitValue(command string) (string, error) {
    parts := strings.SplitN(command, "", 2)
	if len(parts) < 2 {
		// The command was not followed by a value
		return "", errors.New("no have value after command")
	}
	commandValue := parts[1]

    return commandValue, nil
}
