package models

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	DB      *mongo.Database
	content *mongo.Collection
	ctx     = context.TODO()
)

func init() {
	godotenv.Load(".env")
	Setup()
}

func Setup() {
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")))
	if err != nil {
		panic(err.Error())
	}

	DB = client.Database(os.Getenv("MONGO_DBNAME"))
	content = DB.Collection("content")
	fmt.Println("mongo db is initialized")
}
