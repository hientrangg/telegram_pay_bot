package manage

import (
	"errors"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/database"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TELEGRAM_APITOKEN = "6219020061:AAEHiiMLOsQ86xhnyEDBEY7wFrUIwNZ6vvQ"
)

func Deposit(bot *tgbotapi.BotAPI) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			value, err := strconv.ParseInt(update.Message.Text, 10, 64)
			if err != nil {
				return err
			}
			err = DoDeposit(update.Message.From.ID, value)
			if err != nil {
				return err
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Deposit Successful")
			bot.StopReceivingUpdates()
			if _, err := bot.Send(msg); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func Withdraw(bot *tgbotapi.BotAPI) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			value, err := strconv.ParseInt(update.Message.Text, 10, 64)
			if err != nil {
				return err
			}
			err = DoWithdraw(update.Message.From.ID, value)
			if err != nil {
				return err
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Withdraw Successful")
			bot.StopReceivingUpdates()
			if _, err := bot.Send(msg); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func DoDeposit(Uid int64, amount int64) error {
	value, err := database.QueryValue(Uid)
	if err != nil {
		return err
	}
	newValue := value + amount
	err = database.UpdateValue(Uid, newValue)
	if err != nil {
		return err
	}
	return nil
}

func DoWithdraw(Uid int64, amount int64) error {
	value, err := database.QueryValue(Uid)
	if err != nil {
		return err
	}
	if amount > value {
		return errors.New("withdraw value can not greater than the balance")
	}
	newValue := value - amount
	err = database.UpdateValue(Uid, newValue)
	if err != nil {
		return err
	}
	return nil
}

func DoGetStatus(Uid int64) (int64, error) {
	value, err := database.QueryValue(Uid)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func DoTranfer(sender, receiver, amount int64) error {
	senderValue, err := database.QueryValue(sender)
	if err != nil {
		return err
	}

	if amount > senderValue {
		return errors.New("tranfer value can not greater than the balance")
	}

	//do tranfer
	//sender
	newValue := senderValue - amount
	err = database.UpdateValue(sender, newValue)
	if err != nil {
		return err
	}

	//receiver
	receiverValue, err := database.QueryValue(receiver)
	if err != nil {
		return err
	}
	newValue = receiverValue + amount
	err = database.UpdateValue(receiver, newValue)
	if err != nil {
		return err
	}

	return nil
}

func DoRegister(uid int64) error {
	err := database.Add(uid)
	if err != nil {
		return err
	}
	return nil
}
