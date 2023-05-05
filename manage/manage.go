package manage

import (
	"errors"

	"github.com/hientrangg/telegram_pay_bot/database"
	_ "github.com/mattn/go-sqlite3"
)

func Deposit(Uid int64, amount int64) error {
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

func Withdraw(Uid int64, amount int64) error {
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

func GetStatus(Uid int64) (int64, error) {
	value, err := database.QueryValue(Uid)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func Tranfer(sender, receiver, amount int64) error {
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

func Register(uid int64) error {
	err := database.Add(uid)
	if err != nil {
		return err
	}
	return nil
}