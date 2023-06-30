package authenticate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"websockettwo/chaty.com/mongoconn"
	"websockettwo/chaty.com/readwrite"
	"websockettwo/chaty.com/structs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebHandler struct {
	Type       string          `json:"type"`
	Connection *websocket.Conn `json:"Connection"`
	Pasword    string          `json:"password"`
	Uid        string          `json:"uid"`
	Rid        string          `json:"rid"`
	contextGin *gin.Context
}

var (
	handler               structs.WebHandler
	ConnectionArr         []WebHandler
	ChatWithArr           []structs.ChatWith
	MessageFromUser       []byte
	testArr               []structs.Notification
	userSendMEssage       []string
	NotificationStructArr []structs.NotificationStruct
	randId int
)

func (cl *WebHandler) ReadMessage() {
	// ! Read message can be handled with code bellow
	for {
		_, message, err := cl.Connection.ReadMessage()
		if err != nil {
			log.Printf("Error Read Message %v", err)
			return
		}
		MessageFromUser = message
		// userType := strings.Split(strings.Split(strings.Split(string(MessageFromUser), ":")[1], ",")[0], `"`)[1]
		userType := strings.Split(string(MessageFromUser), ",")
		fmt.Printf("userType: %v\n", len(userType))
		if len(userType) > 5 {
			// !handling img
			randId = rand.Intn(1000) * int(time.Second)
			fmt.Printf("randId: %v\n", randId)
			// "./staticupload-"+
			err = ioutil.WriteFile("./static/"+fmt.Sprint(randId)+".png", MessageFromUser, 0644)
			if err != nil {
				log.Println("Error saving file:", err)
				return
			}
			fmt.Println("File received and saved successfully.")
			// !handling img
			cl.Type = ""
			cl.message()
		}

		json.Unmarshal(message, &handler)

		// fmt.Printf("handler: %v\n", handler)
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
}

func (cl *WebHandler) WriteMessage() {
	// ? Write message can be handled by code bellow
	if cl.Type == "message" {
		fmt.Println("mes")
		cl.message()
	} else if cl.Type == "list" {
		fmt.Println("lis")
		cl.list()
	} else if cl.Type == "chatwith" {
		fmt.Println("cht")
		cl.chatwith()
	}
}

func NotificationSend(ReciverId string, connection *websocket.Conn) {
	// log.Printf("id%v", ReciverId)
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
	// fmt.Printf("user: %v\n", user)

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
	for _, notItem := range notificationArr {
		if user.UserId == notItem.ReciverId {
			notificationArrSend = append(notificationArrSend, notItem)

			// ! created new slice and add type of user you have to make filter
			if len(userSendMEssage) != 0 {
				var isFind bool = false
				for _, item := range userSendMEssage {
					if item == notItem.UserId {
						isFind = true
					}
				}
				if !isFind {
					userSendMEssage = append(userSendMEssage, notItem.UserId)
				}
			} else {
				userSendMEssage = append(userSendMEssage, notItem.UserId)
			}
		}
	}
	// fmt.Printf("userSendMEssage: %v\n", userSendMEssage)
	// ! create slice of user id and amount of message
	for _, itemUser := range userSendMEssage {
		testArr = make([]structs.Notification, 0)

		for _, itemNote := range notificationArrSend {
			if itemUser == itemNote.UserId {
				testArr = append(testArr, itemNote)
			}
		}

		if len(testArr) != 0 {
			NotificationStructArr = append(NotificationStructArr, structs.NotificationStruct{
				UserId: itemUser,
				Amount: len(testArr),
				Type:   "notification",
			})
		}
	}
	// fmt.Printf("testArr: %v\n", testArr)
	if len(notificationArrSend) != 0 {
		for _, item := range ConnectionArr {
			if item.Uid == user.UserId {
				jsnote, _ := json.Marshal(NotificationStructArr)
				// fmt.Printf("NotificationStructArr: %v\n", NotificationStructArr)
				NotificationStructArr = make([]structs.NotificationStruct, 0)
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
	// fmt.Printf("cl: %v\n", cl)
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
	_, err = connectionNot.DeleteMany(mongoconn.Ctx, bson.M{"userid": cl.Rid, "reciverid": cl.Uid})
	// connectionNot.DeleteMany(mongoconn.Ctx, bson.M{"userid": cl.Uid,"reciverid": handler.ReciverId})
	if err != nil {
		log.Printf("Error delete the notification %v", err)
	}
	// fmt.Printf("result: %v\n", result)
	// ! ===================================== Add users to chat withh arrey =========================================
	// loop though the chatwith and delete the user if it exist there
	cl.ClearCatwith()
	// Add it again to chatithh arrey
	ChatWithArr = append(ChatWithArr, structs.ChatWith{
		Type: "catwith",
		Uid:  cl.Uid,
		Rid:  cl.Rid,
	})
}
func (cl *WebHandler) message() {
	var (
		notification structs.Notification
		// OnlineUsersList []structs.Create
		Users          structs.Create
		Messages       structs.Message
		MessagesDecode structs.Message
		MessagesArr    []structs.Message
	)
	json.Unmarshal(MessageFromUser, &MessagesDecode)
	// fmt.Printf("MessageFromUser: %v\n", MessagesDecode)
	mongoconn.Connection()
	connection := mongoconn.Client.Database("Chat").Collection("messages")
	// ? ===================================== Insert Message into DB ====================================
	messageid := primitive.NewObjectID().Hex()
	MessagesDecode.MessageId = messageid
	MessagesDecode.UserId = cl.Uid
	MessagesDecode.Type = "message"
	MessagesDecode.ReciverId = handler.ReciverId
	if cl.Type == ""{
		MessagesDecode.Type = "file"
		fmt.Sprint(randId)
		fmt.Println(randId)
		MessagesDecode.Text = fmt.Sprint(randId)+".png"
	}
	if MessagesDecode.Text != "" {
		connection.InsertOne(mongoconn.Ctx, MessagesDecode)
	}
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
	for cur.Next(mongoconn.Ctx) {
		cur.Decode(&Messages)
		if Messages.ReciverId == cl.Uid && Messages.UserId == cl.Rid || Messages.ReciverId == cl.Rid && Messages.UserId == cl.Uid {
			MessagesArr = append(MessagesArr, Messages)
		}
	}
	var isFind = false
	jsmessages, _ := json.Marshal(MessagesArr)
	for _, chatWithItem := range ChatWithArr {
		//!----------------------------- Marshel the message -----------------------------
		if chatWithItem.Uid == cl.Rid && chatWithItem.Rid == cl.Uid {
			for _, item := range ConnectionArr {
				println("salom")
				item.Connection.WriteMessage(websocket.TextMessage, jsmessages)
				cl.Connection.WriteMessage(websocket.TextMessage, jsmessages)
				isFind = true
			}
		}
	}

	if isFind == false {
		println("walek")
		var ()
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
		NotificationSend(cl.Rid, cl.Connection)
		cl.Connection.WriteMessage(websocket.TextMessage, jsmessages)
	}
}
func (cl *WebHandler) handleFile() {
	cl.message()
	imguid := readwrite.ParseFile(cl.contextGin, "./", 3)
	fmt.Printf("imguid: %v\n", imguid)
}

func (cl *WebHandler) ClearCatwith() {
	for index, item := range ChatWithArr {
		if item.Uid == cl.Uid {
			if index == len(ChatWithArr) {
				ChatWithArr = append(ChatWithArr[:index-1], ChatWithArr[index:]...)
			} else {
				ChatWithArr = append(ChatWithArr[:index], ChatWithArr[index+1:]...)
			}
		}
	}
}