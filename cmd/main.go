package main

import (
	"log"
	// "net/http"
	"os"

	"websockettwo/chaty.com/authenticate"

	"github.com/gin-gonic/gin"
)

func main(){
	route := gin.Default()
	
	route.Use(authenticate.Cors)
	route.StaticFS("/static", gin.Dir("./static", true))
	route.POST("/create", authenticate.Create)
	route.POST("/signin", authenticate.Signin)
	route.GET("/ws", authenticate.WebSocket)
	// http.HandleFunc("/create",authenticate.Create)
	// http.HandleFunc("/signin",authenticate.Signin)
	// http.HandleFunc("/ws",authenticate.WebSocket)
	PORT := os.Getenv("PORT")
	if PORT == ""{
		PORT = "4500"
	}
	log.Printf("Server Listenning on PORT: %v", PORT)
	// http.ListenAndServe(":"+PORT,nil)
	route.Run(":"+PORT)

}