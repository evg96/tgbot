package model

import "time"

type Order struct {
	ID        int       `json:"id"`
	Time      time.Time `json:"time"`
	TimeSlots int       `json:"time_slots"`
	Employee  Employee  `json:"employee"`
	Services  []Service `json:"services"`
}

type Service struct {
	Title string `json:"title"`
	Price int    `json:"price"`
}

type Employee struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ID struct {
	ID int `json:"id"`
}

type Client struct {
	TgUserID   int64  `json:"user_id"`
	TgUserName string `json:"username"`
	TgFirsName string `json:"firstname"`
	ChatID     int64  `json:"chat_id"`
	TgToken    string `json:"tg_token"`
}
