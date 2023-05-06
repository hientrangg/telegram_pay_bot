package withdraw

import(
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hientrangg/telegram_pay_bot/manage"
)

func Withdraw(bot *tgbotapi.BotAPI) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	
	for update := range updates {
		if update.Message != nil {
			value, err := strconv.ParseInt(update.Message.Text, 10 , 64)
			if err != nil {
				return err
			}
			err = manage.DoWithdraw(update.Message.From.ID, value)
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