package databaseInterface

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MUR4SH/MyMessenger/structures"
)

type DatabaseInterface struct {
	clientOptions      options.ClientOptions
	client             mongo.Client
	collectionMessages mongo.Collection
	collectionChats    mongo.Collection
	collectionUsers    mongo.Collection
	collectionFiles    mongo.Collection
}

func New(address string, database string, coll_messages string, coll_users string, coll_files string, coll_chats string) DatabaseInterface {
	// Set client options
	clientOptions := options.Client().ApplyURI(address)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	// get collection as ref
	collectionMessages := client.Database(database).Collection(coll_messages)
	collectionUsers := client.Database(database).Collection(coll_users)
	collectionFiles := client.Database(database).Collection(coll_files)
	collectionChats := client.Database(database).Collection(coll_chats)
	log.Print(" Connected to database\n")

	return DatabaseInterface{*clientOptions, *client, *collectionMessages, *collectionChats, *collectionUsers, *collectionFiles}
}

func (d DatabaseInterface) GetUsersChats(user_id string, limit int, offset int) (structures.User, error) {
	var res structures.User
	objectId, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		log.Println("Invalid id")
	}
	err = d.collectionUsers.FindOne(context.TODO(), bson.D{{"_id", objectId}}).Decode(&res)
	return res, err
}

func (d DatabaseInterface) GetUsers(login string, limit int, offset int) ([]structures.User, error) {
	var res []structures.User
	filter := bson.D{primitive.E{Key: "login", Value: login}}
	err := d.collectionUsers.FindOne(context.TODO(), filter).Decode(&res)
	return res, err
}

func (d DatabaseInterface) GetUsersOfChat(chat_id string, limit int, offset int) (structures.Chat, error) {
	var res structures.Chat
	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	err = d.collectionChats.FindOne(context.TODO(), bson.D{{"_id", objectId}}).Decode(&res)
	return res, err
}

func (d DatabaseInterface) GetChat(chat_id string) (structures.Chat, error) {
	var res structures.Chat
	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	err = d.collectionChats.FindOne(context.TODO(), bson.D{{"_id", objectId}}).Decode(&res)
	return res, err
}

func (d DatabaseInterface) GetMessages(chat_id string, limit int, offset int) ([]structures.Message, error) {
	var res []structures.Message
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	cur, err := d.collectionMessages.Find(context.TODO(), bson.D{{"chat_id", chat_id}}, findOptions)
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem structures.Message
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}
	return res, err
}

func (d DatabaseInterface) Authorise(login string, password string) (string, error) {
	var res structures.User
	err := d.collectionUsers.FindOne(context.TODO(), bson.D{{"login", login}, {"password", password}}).Decode(&res)
	return res.Id, err
}

func (d DatabaseInterface) SendMessage(chat_id string, user_id string, text string) (bool, error) {
	time := time.Now().UTC()
	var msg structures.MessageInsert
	msg.Chat_id = chat_id
	msg.Gtm_date = time.Format("2006-01-02T15:04:05")
	msg.Text = text
	msg.User_id = user_id

	ins, err := d.collectionMessages.InsertOne(context.TODO(), msg)
	if err != nil {
		return false, err
	}
	oid, _ := ins.InsertedID.(primitive.ObjectID)
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "messages_array", Value: oid.Hex()},
		}},
	}

	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	_, err = d.collectionChats.UpdateOne(context.TODO(), bson.D{{"_id", objectId}}, update)
	return err == nil, err
}
