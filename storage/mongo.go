package storage

import (
	"context"
	"file-storage/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	collection *mongo.Collection
}

func NewMongoStorage(config *models.Mongo) (*MongoStorage, error) {
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(uint64(config.MaxPoolSize)).
		SetMinPoolSize(uint64(config.MinPoolSize)).
		SetMaxConnIdleTime(time.Duration(config.MaxConnIdleTimeSec) * time.Second)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	collection := client.Database(config.Name).Collection(config.Collection)
	return &MongoStorage{collection: collection}, nil
}

func (ms *MongoStorage) SaveFile(ctx context.Context, data []byte, metadata map[string]interface{}, fileID string) (string, error) {
	// Генерируем уникальный идентификатор файла, если он не указан
	if fileID == "" {
		fileID = uuid.New().String()
	}

	// Конвертация метаданных в BSON
	document, err := toBsonM(metadata)
	if err != nil {
		return "", err
	}

	// Добавляем данные файла и идентификатор в документ
	document["filedata"] = data

	// Условие для поиска документа по идентификатору
	filter := bson.M{"_id": fileID}

	// Опции для операции обновления с upsert (если не найден документ, будет вставлен новый)
	updateOptions := options.Update().SetUpsert(true)

	// Операция обновления
	update := bson.M{
		"$set": document,
	}

	updateResult, err := ms.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return "", err
	}

	if updateResult.UpsertedID != nil {
		upsertedId := updateResult.UpsertedID.(string)
		return upsertedId, nil
	}

	return fileID, nil
}

func (ms *MongoStorage) GetFile(ctx context.Context, fileID string) ([]byte, map[string]interface{}, error) {
	var result bson.M

	var metadata map[string]interface{}

	filter := bson.M{"_id": fileID}
	err := ms.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, metadata, err
	}

	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, metadata, err
	}

	err = bson.Unmarshal(bsonBytes, &metadata)
	if err != nil {
		return nil, metadata, err
	}

	// Приведение к []byte через тип []uint8
	filedata, ok := result["filedata"].(primitive.Binary)
	if !ok {
		return nil, metadata, fmt.Errorf("failed to assert filedata to primitive.Binary")
	}

	return filedata.Data, metadata, nil
}

func (ms *MongoStorage) DeleteFile(ctx context.Context, fileID string) error {
	filter := bson.M{"_id": fileID}
	_, err := ms.collection.DeleteOne(ctx, filter)
	return err
}

func toBsonM(v interface{}) (bson.M, error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result bson.M
	err = bson.Unmarshal(data, &result)
	return result, err
}
