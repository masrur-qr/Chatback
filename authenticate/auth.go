package authenticate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"websockettwo/chaty.com/mongoconn"
	"websockettwo/chaty.com/readwrite"
	"websockettwo/chaty.com/structs"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	urlcors = os.Getenv("URL")
)

func Create(c *gin.Context) {
	jsonFM := c.Request.FormValue("json")
	files, _, errIMG := c.Request.FormFile("img")
	if errIMG != nil {
		c.JSON(409, gin.H{
			"sttus": "NOIMGFILEEXIST",
		})
	}
	fmt.Printf("jsonFM: %v\n", string(jsonFM))	
	fmt.Printf("files: %v\n", files)
	imguid := readwrite.ParseFile(c,"./",10)
	// log.Println("9i")
	var (
		user structs.Create
		// decode structs.Create
	)
	err := json.Unmarshal([]byte(jsonFM), &user)
	if err != nil{
		log.Printf("err %v", err)
	}
	// fmt.Printf("user: %v\n", user)

	mongoconn.Connection()
	connection := mongoconn.Client.Database("Chat").Collection("users")
	usercount , err := connection.CountDocuments(mongoconn.Ctx, bson.M{"name": user.Name, "lastname": user.LastName, "surname": user.Surname, "login": user.Login})
	fmt.Printf("userCount: %v\n", usercount)
	if err != nil{
		log.Printf("Error count %v ",err)
	}


	if usercount == 0 && user.Name != ""{
		userID := primitive.NewObjectID().Hex()
		user.UserId = userID
		user.Imgurl = "-" + imguid
		// fmt.Printf("imguid: %v\n", imguid)
		connection.InsertOne(mongoconn.Ctx, user)

		user.Password = ""
		jsstring, _ := json.Marshal(user)

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "token",
			Value:    string(jsstring),
			Expires:  time.Now().Add(30 * time.Hour),
			HttpOnly: false,
			Secure:   false,
			Path:     "/",
			Domain:   "",
		})

		c.JSON(200, gin.H{
			"Code":"User",
		})
	} else {
		// w.Write([]byte("User Already Exist"))
		c.JSON(309, gin.H{
			"Code":"user Exist",
		})
	}
}
func Signin(c *gin.Context) {
	var user structs.Create
	c.ShouldBindJSON(&user)
	fmt.Printf("user: %v\n", user)
	// err := json.NewDecoder(r.Body).Decode(&user)
	// if err != nil {
		// log.Printf("Err Decode%v\n", err)
	// }

	var DecodeUser structs.Create
	mongoconn.Connection()
	connection := mongoconn.Client.Database("Chat").Collection("users")
	connection.FindOne(mongoconn.Ctx, bson.M{"login": user.Login}).Decode(&DecodeUser)

	jsstring , _ := json.Marshal(DecodeUser)
	if DecodeUser.Password == user.Password {
		log.Println(string(jsstring))
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "token",
			Value:    string(jsstring),
			Expires:  time.Now().Add(30 * time.Hour),
			HttpOnly: false,
			Secure:   false,
			Path:     "/",
			Domain:   "",
		})
		c.JSON(200, gin.H{
			"Code":"Authorize",
		})
	} else {
		// w.Write([]byte("Cannot Authorize"))
		c.JSON(309, gin.H{
			"Code":"Cannot Authorize",
		})
	}
}
func WebSocket(c *gin.Context) {
	var (
		// user structs.Create
	)
	cookie , err := c.Request.Cookie("token")
	if err != nil{
		log.Printf("Cookie err %v",err)
	}
	cookiedata := strings.Split(strings.Join(strings.Split(strings.Join(strings.Split(cookie.Value, "{"), " "), "}"), " "), ",")
	log.Printf("cookie value%v",strings.Split(cookiedata[0], ":")[1])
	
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	// ========================= *websocket.ConnwebSock is the connection that user makes with server ============================
	webSock, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error Wile Upgrading web socket %v", err)
	}
	curentuser := &WebHandler{
		Connection: webSock,
		Uid: strings.Split(cookiedata[0], ":")[1],
	}
	// !========================= Add connection into Db ============================
	mongoconn.Connection()
	userconnection := mongoconn.Client.Database("Chat").Collection("users")
	var Decodedata structs.Create
	userconnection.FindOne(mongoconn.Ctx,bson.M{"_id": strings.Split(cookiedata[0], ":")[1]}).Decode(&Decodedata)
	if Decodedata.LastName != ""{
		// var LocalPaste bool = false
		for index , item := range ConnectionArr {
			if item.Uid == strings.Split(cookiedata[0], ":")[1]{
				// LocalPaste = true
				ConnectionArr = ConnectionArr[:index]
				fmt.Printf("ConnectionArr[:index]: %v\n", ConnectionArr[index:index+1])
			}
		}
		ConnectionArr = append(ConnectionArr, WebHandler{
			Connection: webSock,
			Uid: Decodedata.UserId,
		})
	}
	
	// =================== CaLl read Massage function =================================
	go curentuser.ReadMessage()
}

func Cors(c *gin.Context) {
	fmt.Printf("urlcors: %v\n", urlcors)
	if urlcors == "" {
		urlcors = "http://127.0.0.1:3000"
	}
	fmt.Printf("urlcors: %v\n", urlcors)

	c.Writer.Header().Set("Access-Control-Allow-Origin", urlcors)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, ResponseType, accept, origin, Cache-Control, X-Requested-With")
}
