package util

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var StartKeyBoard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("REGISTER", "register"),
		tgbotapi.NewInlineKeyboardButtonData("UID", "uid"),
		tgbotapi.NewInlineKeyboardButtonData("GET BALANCE", "status"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("DEPOSIT", "deposit"),
		tgbotapi.NewInlineKeyboardButtonData("TRANFER", "tranfer"),
		tgbotapi.NewInlineKeyboardButtonData("WITHDRAW", "withdraw"),
	),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("COT PAY", "cotpay"),
        tgbotapi.NewInlineKeyboardButtonData("LOCK VALUE", "lockvalue"),
        tgbotapi.NewInlineKeyboardButtonData("ALLOW VALUE", "allowvalue"),
    ),
)

var PincodeKeyboard =  tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("1", "1"),
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("7", "7"),
        tgbotapi.NewInlineKeyboardButtonData("8", "8"),
        tgbotapi.NewInlineKeyboardButtonData("9", "9"),
    ),
	tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("<", "<"),
        tgbotapi.NewInlineKeyboardButtonData("0", "0"),
        tgbotapi.NewInlineKeyboardButtonData("ok", "ok"),
    ),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("text", "da"),
	),
)

var ReceiverConfirmKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("CONFIRM", "confirm-cotpay-receiver"),
		tgbotapi.NewInlineKeyboardButtonData("CANCEL", "cancel-cotpay-receiver"),
	),
)

var SenderConfirmKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("CONFIRM", "confirm-cotpay-sender"),
	),
)

func InitStartKeyboard(uid int, balance int) tgbotapi.InlineKeyboardMarkup {
	var keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("REGISTER", "register"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("UID: " + strconv.Itoa(uid), "uid"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Balance: " + strconv.Itoa(balance), "status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("DEPOSIT", "deposit"),
			tgbotapi.NewInlineKeyboardButtonData("TRANFER", "tranfer"),
			tgbotapi.NewInlineKeyboardButtonData("WITHDRAW", "withdraw"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("COT PAY", "cotpay"),
			tgbotapi.NewInlineKeyboardButtonData("LOCK VALUE", "lockvalue"),
			tgbotapi.NewInlineKeyboardButtonData("ALLOW VALUE", "allowvalue"),
		),
	)

	return keyboard
}
func InitPincodeKeyboard(data string) tgbotapi.InlineKeyboardMarkup {
	var pincodeKeyboard =  tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1", "1"),
			tgbotapi.NewInlineKeyboardButtonData("2", "2"),
			tgbotapi.NewInlineKeyboardButtonData("3", "3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("4", "4"),
			tgbotapi.NewInlineKeyboardButtonData("5", "5"),
			tgbotapi.NewInlineKeyboardButtonData("6", "6"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("7", "7"),
			tgbotapi.NewInlineKeyboardButtonData("8", "8"),
			tgbotapi.NewInlineKeyboardButtonData("9", "9"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<", "<"),
			tgbotapi.NewInlineKeyboardButtonData("0", "0"),
			tgbotapi.NewInlineKeyboardButtonData("ok", data),
		),
	)
	
	return pincodeKeyboard
}