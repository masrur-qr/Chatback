package mongoconn

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Ctx context.Context
	Client *mongo.Client
	DB_Url = os.Getenv("DBURL")
)
func Connection(){
	if(DB_Url ==  ""){
		DB_Url = "mongodb://127.0.0.1:27017"
		// DB_Url = "mongodb://root:PCxwZMLwKA@34.30.133.36:27017"
	}
	// log.Printf("DB URl%v\n",DB_Url)
	// log.Printf("DB URl%v\n",os.Getenv("DBURL"))
	clientOptions := options.Client().ApplyURI(DB_Url)
	// clientOptions := options.Client().ApplyURI("mongodb://mas:mas@mongo:27017")
	// clientOptions := options.Client().ApplyURI(os.Getenv("DBURL"))
	clientG, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Mongo.connect() ERROR: ", err)
	}
	ctxG, _ := context.WithTimeout(context.Background(), 15*time.Minute)
	// collection := client.Database("MedCard").Collection("users")
	Ctx = ctxG
	Client = clientG
}