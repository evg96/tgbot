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
	duration    = 30
	getOrders   = "/get/orders?user=%s&token=%s"
	cancleOrder = "/delete/order"
	newClient   = "/client/new"
	checkClient = "/client/check"
)

type TgService interface {
	CancleOrder(orderID int) error
	GetOrders(userID int, firstName string, tokenMd5 string) ([]string, []int, error)
	AddNewClient(infoInfo model.Client) error
	CheckClient(info model.Client) (bool, error)
}

type Services struct {
	TgService
}

type HandleOrder struct {
	BackURL string
}

func NewHandleOrder(back string) *HandleOrder {
	return &HandleOrder{BackURL: back}
}

func NewTgService(back string) *Services {
	return &Services{
		NewHandleOrder(back),
	}
}

func (s *HandleOrder) CancleOrder(orderID int) error {
	client := http.Client{}
	url := s.BackURL + cancleOrder
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

func (s *HandleOrder) GetOrders(userID int, firstName string, tokenMd5 string) ([]string, []int, error) {
	report, err := s.getOrders(userID, tokenMd5)
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

func (s *HandleOrder) getOrders(userID int, tokenMd5 string) ([]model.Order, error) {
	client := http.Client{}
	path := fmt.Sprintf(getOrders, strconv.Itoa(userID), tokenMd5)
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
