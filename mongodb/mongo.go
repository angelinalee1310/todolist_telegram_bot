package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func Connect(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}

	// Проверка подключения
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Не удалось подключиться к MongoDB:", err)
	}

	log.Println("Успешное подключение к MongoDB")
	Client = client
	return client
}
