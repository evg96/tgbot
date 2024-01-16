package handler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"tgbot/internal/model"
	"tgbot/internal/service"

	tele "gopkg.in/telebot.v3"
)

var (
	menu      = &tele.ReplyMarkup{ResizeKeyboard: true}
	btnOrders = menu.Text("ℹ Мои записи")
)

type TgHandler struct {
	// *tele.Bot
	TgService *service.Services
}

func NewTg(service *service.Services) *TgHandler {
	return &TgHandler{TgService: service}
}

func (tg *TgHandler) GetOrders(c tele.Context) error {
	firstName := c.Message().Chat.FirstName
	chatID := c.Chat().ID

	messages, ids, err := tg.TgService.GetOrders(int(chatID), firstName)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		message := fmt.Sprintf("%s, Ваши записи пусты", firstName)
		return c.Send(message, menu)
	}

	cancel := &tele.ReplyMarkup{}

	for i, m := range messages {
		data := strconv.Itoa(i+1) + "-" + strconv.Itoa(ids[i])
		btnCancel := cancel.Data("Отменить запись", "cancle", data)
		cancel.Inline(
			cancel.Row(btnCancel),
		)
		err := c.Send(m, cancel)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (tg *TgHandler) CanclOrder(c tele.Context) error {
	confirm := &tele.ReplyMarkup{}
	orderN := strings.Split(c.Data(), "-")[0]
	btnConf := confirm.Data("Подтвердить", "confirm", c.Data())
	confirm.Inline(
		confirm.Row(btnConf),
	)
	text := "Подтвердите отмену записи № " + orderN
	err := c.Send(text, confirm)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (tg *TgHandler) ConfCancle(c tele.Context) error {
	orderN := strings.Split(c.Data(), "-")[0]
	orderID := strings.Split(c.Data(), "-")[1]
	id, err := strconv.Atoi(orderID)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := tg.TgService.CancleOrder(id); err != nil {
		return err
	}

	text := "Запись № " + orderN + " отменена"
	err = c.Send(text, menu)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (tg *TgHandler) HandleStartBtn(c tele.Context) error {
	clientInfo := model.Client{
		TgUserID:   c.Message().Sender.ID,
		TgFirsName: c.Message().Sender.FirstName,
		TgUserName: c.Message().Sender.Username,
		ChatID:     c.Chat().ID,
		TgToken:    os.Getenv("TOKEN"),
	}

	if clientInfo.TgUserID == 0 || clientInfo.TgFirsName == "" ||
		clientInfo.TgUserName == "" || clientInfo.ChatID == 0 || clientInfo.TgToken == "" {
		return c.Send("Что-то пошло не так. Попробуйте еще раз. Введите /start.", &tele.ReplyMarkup{})
	}
	exist, err := tg.TgService.CheckClient(clientInfo)
	if err != nil {
		return c.Send("Что-то пошло не так. Попробуйте еще раз. Введите /start.", &tele.ReplyMarkup{})
	}
	if !exist {
		if err := tg.TgService.AddNewClient(clientInfo); err != nil {
			return c.Send("Что-то пошло не так. Попробуйте еще раз. Введите /start.", &tele.ReplyMarkup{})
		} else {
			menu.Reply(
				menu.Row(btnOrders),
			)
			message := welcomeString()
			return c.Send(message, menu)
		}
	} else {
		menu.Reply(
			menu.Row(btnOrders),
		)
		message := welcomeString()
		return c.Send(message, menu)
	}
}

func welcomeString() string {
	m1 := fmt.Sprintf("Добро пожаловать в наш салон красоты!\n")
	m2 := fmt.Sprintf("Нажмите синюю иконку слева внизу, чтобы запустить приложение и записаться к нам в салон.\n")
	m3 := fmt.Sprintf("Посмотреть Ваши текущие записи, а также отменить их можно, нажав на кнопку Мои записи.")
	return fmt.Sprintf("%s\n%s\n%s", m1, m2, m3)
}
