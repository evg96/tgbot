package main

import (
	"fmt"
	"log"
	"os"
	"tgbot/internal/handler"
	"tgbot/internal/service"
	"time"

	tele "gopkg.in/telebot.v3"
)

var (
	menu    = &tele.ReplyMarkup{ResizeKeyboard: true}
	cancel  = &tele.ReplyMarkup{}
	confirm = &tele.ReplyMarkup{}

	btnConf   = confirm.Data("", "confirm", "")
	btnCancel = cancel.Data("", "cancle", "")

	btnOrders = menu.Text("ℹ Мои записи")
)

func main() {
	backURl := os.Getenv("BACK")
	if backURl == "" {
		fmt.Println("Set env. BACK. Example: export BACK=http://127.0.0.1:8080")
		os.Exit(-1)
	}
	Token := os.Getenv("TOKEN")
	if Token == "" {
		fmt.Println("Set env. TOKEN.")
		os.Exit(-1)
	}
	pref := tele.Settings{
		Token:  Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	menu.Reply(
		menu.Row(btnOrders),
	)

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	service := service.NewTgService(backURl)
	tg := handler.NewTg(service) //http://127.0.0.1:8095

	b.Handle(&btnOrders, tg.GetOrders)

	b.Handle(&btnCancel, tg.CanclOrder)

	b.Handle(&btnConf, tg.ConfCancle)

	b.Handle("/start", tg.HandleStartBtn)

	b.Handle("/test", func(c tele.Context) error {
		fmt.Println(menu)
		return c.Send("message", menu)
	})

	b.Start()
}
