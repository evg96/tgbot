package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"tgbot/internal/model"
	"time"
)

const (
	duration        = 30
	getOrders       = "/get/orders?user=%s&token=%s"
	cancleOrder     = "/delete/order?token=%s"
	getCancleOrder  = "/get/cancle/orders?token=%s"
	getCreatedOrder = "/get/create/orders?token=%s"
	newClient       = "/client/new"
	checkClient     = "/client/check"
	groupChat       = "/get/group/chat?token=%s"
)

type TgService interface {
	CancleOrder(orderID int) error
	GetOrders(userID int, firstName string) ([]string, []int, error)
	AddNewClient(infoInfo model.Client) error
	CheckClient(info model.Client) (bool, error)
	GetCanceledOrders() (string, error)
	GetCreatedOrders() (string, error)
	GetGroupChat() int64
}

type Services struct {
	TgService
}

type HandleOrder struct {
	BackURL  string
	TokenMD5 string
}

func NewHandleOrder(back, tokenMd5 string) *HandleOrder {
	return &HandleOrder{
		BackURL:  back,
		TokenMD5: tokenMd5,
	}
}

func NewTgService(back, tokenMd5 string) *Services {
	return &Services{
		NewHandleOrder(back, tokenMd5),
	}
}

func (s *HandleOrder) GetGroupChat() int64 {
	client := http.Client{}
	path := fmt.Sprintf(groupChat, s.TokenMD5)
	url := s.BackURL + path
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0
	}
	resp, err := client.Do(request)
	if err != nil {
		return 0
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode != http.StatusOK {
		return 0
	}
	type response struct {
		ID int64 `json:"id"`
	}
	body := resp.Body
	b, _ := io.ReadAll(body)
	defer resp.Body.Close()
	r := response{}
	json.Unmarshal(b, &r)
	return r.ID
}

func (s *HandleOrder) GetCanceledOrders() (string, error) {
	client := http.Client{}
	path := fmt.Sprintf(getCancleOrder, s.TokenMD5)
	url := s.BackURL + path

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode != http.StatusOK {
		return "", nil
	}
	bodyR := resp.Body
	defer resp.Body.Close()
	body, err := io.ReadAll(bodyR)
	if err != nil {
		return "", err
	}
	var data []model.CancledOrder
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}
	message := genCancleMes(data)
	return message, nil
}

func genCancleMes(data []model.CancledOrder) string {
	message := fmt.Sprintf("Отмененные записи:\n\n")
	var info []string
	for _, i := range data {
		info = append(info, fmt.Sprintf("Время: %s\nСотрудник: %s\n\n", i.Time, i.EmployeeName))
	}
	for _, i := range info {
		message += i
	}
	return message
}

func (s *HandleOrder) CancleOrder(orderID int) error {
	client := http.Client{}
	path := fmt.Sprintf(cancleOrder, s.TokenMD5)
	url := s.BackURL + path
	id := model.ID{ID: orderID}
	body, err := json.Marshal(id)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return nil
}

func (s *HandleOrder) GetCreatedOrders() (string, error) {
	client := http.Client{}
	path := fmt.Sprintf(getCreatedOrder, s.TokenMD5)
	url := s.BackURL + path

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode != http.StatusOK {
		return "", nil
	}
	bodyR := resp.Body
	defer resp.Body.Close()
	body, err := io.ReadAll(bodyR)
	if err != nil {
		return "", err
	}
	var data []model.CreatedOrder
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}
	message := genCreateMes(data)
	return message, nil
}

func genCreateMes(data []model.CreatedOrder) string {
	message := fmt.Sprintf("Новые записи:\n\n")
	var info []string
	for _, i := range data {
		services := fmt.Sprintf("Услуги:\n")
		for _, d := range i.Title {
			services += fmt.Sprintf(" - %s\n", d)
		}
		duration := calcDur(i.Duration)
		client := fmt.Sprintf("@%s", i.ClientTgUsername)
		info = append(info, fmt.Sprintf("%sВремя: %s(%s)\nСотрудник: %s\nКлиент: %s \n\n", services, i.Time.Format("2006-01-02 15:04:05"), duration,
			i.EmployeeName, client))
	}
	for _, i := range info {
		message += i
	}
	return message
}

func calcDur(duration int) string {
	var h, m int
	var mes string
	h = duration / 3600
	if h != 0 {
		mes = fmt.Sprintf("%d ч ", h)
	}
	m = (duration - (h * 3600)) / 60
	if m != 0 {
		mes += fmt.Sprintf("%d мин", m)
	}
	return mes
}

func (s *HandleOrder) GetOrders(userID int, firstName string) ([]string, []int, error) {
	report, err := s.getOrders(userID)
	if err != nil {
		return nil, nil, err
	}
	var message []string
	var ids []int
	if len(report) > 0 {
		message, ids = genMessage(report)
	}
	return message, ids, nil
}

func (s *HandleOrder) getOrders(userID int) ([]model.Order, error) {
	client := http.Client{}
	path := fmt.Sprintf(getOrders, strconv.Itoa(userID), s.TokenMD5)
	// path := getOrders + strconv.Itoa(chatID)
	url := s.BackURL + path
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	bodyR := resp.Body
	defer resp.Body.Close()
	body, err := io.ReadAll(bodyR)
	if err != nil {
		return nil, err
	}
	var report []model.Order
	err = json.Unmarshal(body, &report)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func genMessage(orders []model.Order) ([]string, []int) {
	var messages []string
	var ids []int
	for i, order := range orders {
		d, s, t := getDateAndTime(order.Time, order.TimeSlots)
		services := getServices(order.Services)
		employee := getEmployee(order.Employee)
		r := fmt.Sprintf("%d. %s c %s до %s\n%sМастер: %s\n", i+1, d, s, t, services, employee)
		messages = append(messages, r)
		ids = append(ids, order.ID)
	}
	return messages, ids
}

func getDateAndTime(t time.Time, timeSlots int) (string, string, string) {
	date := t.Format("02.01.2006")
	sinceTime := t.Format("15:04")
	toTime := t.Add(time.Duration(duration*timeSlots) * time.Minute).Format("15:04")
	return date, sinceTime, toTime
}

func getServices(services []model.Service) string {
	var message string
	for _, service := range services {
		s := fmt.Sprintf("%s: %d₽\n", service.Title, service.Price)
		message += s
	}
	return message
}

func getEmployee(employee model.Employee) string {
	return fmt.Sprintf("%s", employee.Name)
}

func (s *HandleOrder) AddNewClient(info model.Client) error {
	url := s.BackURL + newClient
	body, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("service->AddNewUser: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("service->AddNewUser: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AddNewUser->status code: %d", resp.StatusCode)
	}
	return nil
}

func (s *HandleOrder) CheckClient(info model.Client) (bool, error) {
	url := s.BackURL + checkClient
	body, err := json.Marshal(info)
	if err != nil {
		return false, fmt.Errorf("service->CheckClient: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return false, fmt.Errorf("service->CheckClient: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, fmt.Errorf("CheckClient->status code: %d", resp.StatusCode)
}
