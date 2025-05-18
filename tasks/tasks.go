package tasks

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Task struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` // добавляем поле ID
	ChatID int64              `bson:"chat_id"`
	Text   string             `bson:"text"`
	IsDone bool               `bson:"is_done"`
}

type TaskService struct {
	collection *mongo.Collection
}

func NewTaskService(collection *mongo.Collection) *TaskService {
	return &TaskService{collection: collection}
}

func (s *TaskService) AddTask(ctx context.Context, chatID int64, text string) error {
	task := Task{
		ChatID: chatID,
		Text:   text,
		IsDone: false,
	}
	_, err := s.collection.InsertOne(ctx, task)
	return err
}

func (s *TaskService) ListTasks(ctx context.Context, chatID int64) ([]Task, error) {
	filter := bson.M{"chat_id": chatID}
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	err = cursor.All(ctx, &tasks)
	return tasks, err
}

func (s *TaskService) MarkTaskDone(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_done": true}}
	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}
