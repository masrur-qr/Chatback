package authenticate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"websockettwo/chaty.com/mongoconn"
	"websockettwo/chaty.com/structs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/websocket"
)

type WebHandler struct {
	Type       string          `json:"type"`
	Connection *websocket.Conn `json:"Connection"`
	Pasword    string          `json:"password"`
	Uid        string          `json:"uid"`
	Rid        string          `json:"rid"`
}

var (
	handler         structs.WebHandler
	ConnectionArr   []WebHandler
	ChatWithArr     []structs.ChatWith
	MessageFromUser []byte
)

func (cl *WebHandler) ReadMessage() {
	for {
		_, message, err := cl.Connection.ReadMessage()
		if err != nil {
			log.Printf("Error Read Message %v", err)
			return
		}
		MessageFromUser = message
		json.Unmarshal(message, &handler)

		fmt.Printf("handler: %v\n", handler)
		curentuser := &WebHandler{
			Connection: cl.Connection,
			Type:       handler.Type,
			Uid:        cl.Uid,
			Rid:        handler.ReciverId,
		}
		fmt.Printf("curentuser: %v\n", curentuser)
		fmt.Printf("ConnectionArr: %v\n", ConnectionArr)

		curentuser.WriteMessage()
	}
	// fmt.Printf("handler: %v\n", string(message))

}

func (cl *WebHandler) WriteMessage() {
	if cl.Type == "message" {
		cl.message()
	} else if cl.Type == "list" {
		cl.list()
	} else if cl.Type == "chatwith" {
		cl.chatwith()
	}
}

func NotificationSend(ReciverId string, connection *websocket.Conn) {
	log.Printf("id%v", ReciverId)
	var (
		// message      structs.Message
		notification        structs.Notification
		notificationArr     []structs.Notification
		notificationArrSend []structs.Notification
		user                structs.Create
	)
	mongoconn.Connection()
	// ! ===================================== Get the note from DB ====================================
	userconnection := mongoconn.Client.Database("Chat").Collection("users")
	userconnection.FindOne(mongoconn.Ctx, bson.M{"_id": ReciverId, "status": "online"}).Decode(&user)
	// fmt.Printf("user: %v\n", user.Connection)
	fmt.Printf("user: %v\n", user)
	// fmt.Printf("user: %v\n", connection)

	// ! ===================================== Get the note from DB ====================================
	notificationConnection := mongoconn.Client.Database("Chat").Collection("notifications")
	cursNot, err := notificationConnection.Find(mongoconn.Ctx, bson.M{"reciverid": ReciverId})
	if err != nil {
		log.Printf("error Db get %v", err)
	}
	defer cursNot.Close(mongoconn.Ctx)
	for cursNot.Next(mongoconn.Ctx) {
		cursNot.Decode(&notification)
		notificationArr = append(notificationArr, notification)
	}
	// fmt.Printf("notificationArr: %v\n", notificationArr)
	// log.Printf("err %v", err)
	// notificationConnection.FindOne(mongoconn.Ctx,bson.M{"":})
	for _, notItem := range notificationArr {
		// fmt.Printf("user.UserId: %v\n", user.UserId)
		// fmt.Printf("notItem.ReciverId: %v\n", notItem.ReciverId)
		if user.UserId == notItem.ReciverId {
			notificationArrSend = append(notificationArrSend, notItem)
		}
	}
	// log.Printf("note%v", notificationArrSend)
	if len(notificationArrSend) != 0 {
		for _, item := range ConnectionArr {
			// fmt.Printf("item.Uid: %v\n", item.Uid)
			if item.Uid == user.UserId {
				jsnote, _ := json.Marshal(notificationArrSend)
				err := item.Connection.WriteMessage(websocket.TextMessage, jsnote)
				if err != nil {
					log.Printf("error %v", err)
				}
			}
		}
	}
}
func (cl *WebHandler) list() {
	userconnection := mongoconn.Client.Database("Chat").Collection("users")
	var (
		Decodedata      structs.Create
		Users           structs.Create
		OnlineUsersList []structs.Create
	)
	fmt.Printf("cl: %v\n", cl)
	//===================================== Make User Online ====================================
	err := userconnection.FindOneAndUpdate(context.Background(), bson.M{"_id": cl.Uid}, bson.M{"$set": bson.M{"status": "online"}}).Decode(&Decodedata)
	if err != nil {
		log.Printf("err%v", err)
	}
	// =========================== greb all online users ===================================
	cursNot, err := userconnection.Find(mongoconn.Ctx, bson.M{"status": "online"})
	if err != nil {
		log.Printf("Err Find DB %v", err)
	}
	
	// fmt.Printf("Decodedata.Connection: %v\n", Decodedata.Connection)
	// ? ===================================== Send Message To User =====================================
	defer cursNot.Close(mongoconn.Ctx)
	for cursNot.Next(mongoconn.Ctx) {
		cursNot.Decode(&Users)
		Users.Type = "list"
		OnlineUsersList = append(OnlineUsersList, Users)
	}
	byteonlineusers, _ := json.Marshal(OnlineUsersList)
	cl.Connection.WriteMessage(websocket.TextMessage, byteonlineusers)
	// ================================Update user list in all sessions ===========================
	for _, item := range ConnectionArr {
		item.Connection.WriteMessage(websocket.TextMessage, byteonlineusers)
	}
	NotificationSend(cl.Uid, cl.Connection)
}
func (cl *WebHandler) chatwith() {
	var (
		Messages    structs.Message
		MessagesArr []structs.Message
	)
	// NotificationSend(cl.Uid, cl.Connection)
	// ! =========================== Get and Send Messages To user with  specific client ========================
	mongoconn.Connection()
	connection := mongoconn.Client.Database("Chat").Collection("messages")
	cur, err := connection.Find(mongoconn.Ctx, bson.M{})
	if err != nil {
		log.Printf("Error get Messages %v", err)
	}
	defer cur.Close(mongoconn.Ctx)

	for cur.Next(mongoconn.Ctx) {
		cur.Decode(&Messages)
		if Messages.ReciverId == cl.Uid && Messages.UserId == cl.Rid || Messages.ReciverId == cl.Rid && Messages.UserId == cl.Uid {
			MessagesArr = append(MessagesArr, Messages)
		}
	}
	jsmessages, _ := json.Marshal(MessagesArr)
	cl.Connection.WriteMessage(websocket.TextMessage, jsmessages)
	// * ======================================== Remove Notification with that user ==================================
	connectionNot := mongoconn.Client.Database("Chat").Collection("notifications")
	result, err := connectionNot.DeleteMany(mongoconn.Ctx, bson.M{"userid": cl.Rid, "reciverid": cl.Uid})
	// connectionNot.DeleteMany(mongoconn.Ctx, bson.M{"userid": cl.Uid,"reciverid": handler.ReciverId})
	if err != nil {
		log.Printf("Error delete the notification %v", err)
	}
	fmt.Printf("result: %v\n", result)
	// ! ===================================== Add users to chat withh arrey =========================================
	// loop though the chatwith and delete the user if it exist there
	for index, item := range ChatWithArr {
		if item.Uid == cl.Uid {
			// fmt.Printf("ChatWithArr[index]: %v\n", ChatWithArr[index])
			ChatWithArr = ChatWithArr[:index]
			// fmt.Printf("ChatWithArr[index]: %v\n", ChatWithArr[index])
		}
	}
	// Add it again to chatithh arrey
	ChatWithArr = append(ChatWithArr, structs.ChatWith{
		Type: "catwith",
		Uid:  cl.Uid,
		Rid:  cl.Rid,
	})
	fmt.Printf("ChatWithArr: %v\n", ChatWithArr)
}
func (cl *WebHandler) message() {
	var (
		notification structs.Notification
		// OnlineUsersList []structs.Create
		Users          structs.Create
		Messages       structs.Message
		MessagesDecode structs.Message
		MessagesArr    []structs.Message
		// user         structs.Create
		// user         structs.Create
	)
	json.Unmarshal(MessageFromUser, &MessagesDecode)
	fmt.Printf("MessageFromUser: %v\n", MessagesDecode)
	mongoconn.Connection()
	connection := mongoconn.Client.Database("Chat").Collection("messages")
	// ? ===================================== Insert Message into DB ====================================
	messageid := primitive.NewObjectID().Hex()
	MessagesDecode.MessageId = messageid
	MessagesDecode.UserId = cl.Uid
	MessagesDecode.Type = "message"
	MessagesDecode.ReciverId = handler.ReciverId
	connection.InsertOne(mongoconn.Ctx, MessagesDecode)
	// ! ===================================== Get the note from DB ====================================
	// ===================================== Send Message To User =====================================
	userconnection := mongoconn.Client.Database("Chat").Collection("users")
	err := userconnection.FindOne(mongoconn.Ctx, bson.M{"_id": cl.Rid, "status": "online"}).Decode(&Users)
	if err != nil {
		log.Printf("Err Find DB %v", err)
	}
	// *+++++++++++++++++++++++++++++++++++ Check if the user online if true  Send the messages ++++++++++++++++++++++++++++
	cur, err := connection.Find(mongoconn.Ctx, bson.M{})
	if err != nil {
		log.Printf("Error get Messages %v", err)
	}
	defer cur.Close(mongoconn.Ctx)
	// log.Printf("test  %v", 2)
	for cur.Next(mongoconn.Ctx) {
		cur.Decode(&Messages)
		// fmt.Printf("handler.ReciverId: %v\n", handler.ReciverId)
		if Messages.ReciverId == cl.Uid && Messages.UserId == cl.Rid || Messages.ReciverId == cl.Rid && Messages.UserId == cl.Uid {
			MessagesArr = append(MessagesArr, Messages)
			// fmt.Printf("MessagesArr: %v\n", MessagesArr)
		}
		// log.Printf("test one %v", 1)
	}
	for _, chatWithItem := range ChatWithArr {

		//!----------------------------- Marshel the message -----------------------------
		jsmessages, _ := json.Marshal(MessagesArr)
		// fmt.Printf("string(jsmessages): %v\n", string(jsmessages))
		if chatWithItem.Uid == cl.Rid && chatWithItem.Rid == cl.Uid {
			for _, item := range ConnectionArr {
				println("salom")
				item.Connection.WriteMessage(websocket.TextMessage, jsmessages)
				cl.Connection.WriteMessage(websocket.TextMessage, jsmessages)
			}
		} else {
			println("walek")
			var (
				noteArr []structs.Notification
				note structs.Notification
			)
			// ++++++++++++++++++++++++ Else add notification ++++++++++++++++++++++++++++++++++++
			// ? ===================================== Insert Notification to user into DB ====================================
			notificationConnection := mongoconn.Client.Database("Chat").Collection("notifications")
			NotificationId := primitive.NewObjectID().Hex()
			notification.NotificationId = NotificationId
			notification.ReciverId = cl.Rid
			notification.UserId = cl.Uid
			notification.Type = "notification"
			notificationConnection.InsertOne(mongoconn.Ctx, notification)
			// fmt.Printf("string(jsmessages): %v\n", string(jsmessages))

			// ?============================= Send Notification to user =============================
			curNote , Err := notificationConnection.Find(mongoconn.Ctx,bson.M{"reciverid":cl.Rid})
			if Err != nil{
				log.Fatalf("Error find Any note %v", Err)
			}
			defer curNote.Close(mongoconn.Ctx)
			for curNote.Next(mongoconn.Ctx){
				curNote.Decode(&note)
				noteArr = append(noteArr, note)
			}
			for _, item := range ConnectionArr {
				println("send note")
				jsmessages , _:= json.Marshal(&noteArr)
				item.Connection.WriteMessage(websocket.TextMessage, jsmessages)
			}
			cl.Connection.WriteMessage(websocket.TextMessage, jsmessages)
		}
	}
}
