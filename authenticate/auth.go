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
	imguid := readwrite.ParseFile(c,"./static/upload")
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

		// jsstring, _ := json.Marshal(user.Name + ":" + user.Surname + ":" + user.UserId + ":" + user.Login + ":" + user.Imgurl)

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "token",
			Value:    user.Name + ":" + user.Surname + ":" + user.UserId + ":" + user.Login + ":" + user.Imgurl,
			Expires:  time.Now().Add(30 * time.Hour),
			HttpOnly: false,
			SameSite: http.SameSiteNoneMode,
			MaxAge: 0,
			Secure:   true,
			Path:     "/",
			Domain:   ".khorog.dev",
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

	jsstring, _ := json.Marshal(DecodeUser.Name+":"+DecodeUser.Surname+":"+DecodeUser.UserId + ":" + DecodeUser.Login + ":" + DecodeUser.Imgurl)

	if DecodeUser.Password == user.Password {
		log.Println(string(jsstring))
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "token",
			Value:    DecodeUser.Name+":"+DecodeUser.Surname+":"+DecodeUser.UserId + ":" + DecodeUser.Login + ":" + DecodeUser.Imgurl,
			Expires:  time.Now().Add(30 * time.Hour),
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path:     "/",
			Domain:   ".khorog.dev",
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
	cookiedata := strings.Split(cookie.Value, ":")
	log.Printf("cookie value%v",strings.Split(cookie.Value, ":")[2])
	
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
		Uid: cookiedata[2],
	}
	// !========================= Add connection into Db ============================
	mongoconn.Connection()
	userconnection := mongoconn.Client.Database("Chat").Collection("users")
	var Decodedata structs.Create
	userconnection.FindOne(mongoconn.Ctx,bson.M{"_id": cookiedata[2]}).Decode(&Decodedata)
	if Decodedata.LastName != ""{
		println("Hello")
		// var NewConnectionArr []WebHandler
		// var LocalPaste bool = false
		// ! remocve existing connection and appened new one
		for index , item := range ConnectionArr {
			fmt.Printf("item.Uid: %v\n", item.Uid)
			fmt.Printf("item.Rid: %v\n", item.Rid)
			if item.Uid == cookiedata[2]{
				// LocalPaste = true
				if index == len(ConnectionArr){
					ConnectionArr = append(ConnectionArr[:index-1],ConnectionArr[index:]...)
				}else{
					ConnectionArr = append(ConnectionArr[:index],ConnectionArr[index+1:]...)
				}
			}else{
				fmt.Println("sorry error")
			}
		}
		ConnectionArr = append(ConnectionArr, WebHandler{
			Connection: webSock,
			Uid: Decodedata.UserId,
		})
		fmt.Printf("ConnectionArr 3 final %v\n", ConnectionArr)
		//! olso remove clear chat with arrey
		curentuser.ClearCatwith()
	}
	// =================== CaLl read Massage function =================================
	go curentuser.ReadMessage()
}

func Cors(c *gin.Context) {
	if urlcors == "" {
		// urlcors = "http://127.0.0.1:3000"
		// urlcors = "http://192.168.0.108:3000"
		urlcors = "https://chat.khorog.dev"
		// urlcors = "https://chat.khorog.dev"
	}
// ssd

	c.Writer.Header().Set("Access-Control-Allow-Origin", urlcors)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, ResponseType, accept, origin, Cache-Control, X-Requested-With")
}

