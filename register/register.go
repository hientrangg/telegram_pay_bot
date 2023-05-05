package register

import(

	_ "github.com/mattn/go-sqlite3"
	"github.com/hientrangg/telegram_pay_bot/database"
)

func register(uid int64) error {
	err := database.Add(uid)
	if err != nil {
		return err
	}
	return nil
}