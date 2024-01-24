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
	TokenMd5   string `json:"token_md5"`
}

type CancledOrder struct {
	Time         string `json:"time"`
	EmployeeName string `json:"employee"`
}

type CreatedOrder struct {
	Title            []string  `json:"title"`
	Time             time.Time `json:"time"`
	Duration         int       `json:"duration"`
	EmployeeName     string    `json:"empl_name"`
	ClientTgUsername string    `json:"tg_username"`
}
