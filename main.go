package main

import (
	"crypto/md5"
	"encoding/hex"
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

const (
	period = 3
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

	hash := md5.Sum([]byte(Token))
	tokenMd5 := hex.EncodeToString(hash[:])
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

	service := service.NewTgService(backURl, tokenMd5)
	tg := handler.NewTg(service, tokenMd5)

	go func() {
		groupChat := service.GetGroupChat()
		if groupChat != 0 {
			for {
				time.Sleep(period * time.Minute)
				cancleOrders, err := service.GetCanceledOrders()
				newOrders, err := service.GetCreatedOrders()
				if err == nil && newOrders != "" {
					b.Send(&tele.Chat{ID: groupChat}, newOrders)
				}
				if err == nil && cancleOrders != "" {
					b.Send(&tele.Chat{ID: groupChat}, cancleOrders)
				}
			}
		}
	}()

	b.Handle(&btnOrders, tg.GetOrders)

	b.Handle(&btnCancel, tg.CanclOrder)

	b.Handle(&btnConf, tg.ConfCancle)

	b.Handle("/start", tg.HandleStartBtn)

	b.Handle("/test", func(c tele.Context) error {
		// fmt.Println(menu)
		fmt.Println(c.Message().Sender.ID)
		// return c.Send("message", menu)
		return nil
	})

	b.Start()
}
