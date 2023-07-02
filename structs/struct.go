package structs

import "github.com/gorilla/websocket"

type Create struct {
	UserId   string `bson:"_id" json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Surname  string `json:"surname"`
	LastName string `json:"lastname"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Status   string `json:"status"`
	Imgurl   string `json:"imgurl"`
}
type WebHandler struct {
	Type       string          `json:"type"`
	Connection *websocket.Conn `json:"connection"`
	Pasword    string          `json:"password"`
	Uid        string          `json:"uid"`

	MessageId string `bson:"_id"`
	UserId    string `bson:"userid"`
	ReciverId string `bson:"reciverid"`
}
type ChatWith struct {
	Type string `json:"type"`
	Uid  string `json:"uid"`
	Rid  string `json:"rid"`
}
type Message struct {
	MessageId string `bson:"_id"`
	UserId    string `bson:"userid"`
	ReciverId string `bson:"reciverid"`
	Text      string `bson:"text"`
	ImgUrl    string `bson:"imgurl"`
	Type      string `json:"type"`
}
type Notification struct {
	NotificationId string `bson:"_id"`
	UserId         string `bson:"userid"`
	ReciverId      string `bson:"reciverid"`
	Type           string `json:"type"`
}
type NotificationStruct struct {
	UserId string `json:"userid"`
	Amount int    `json:"amount"`
	Type   string `json:"type"`
}
