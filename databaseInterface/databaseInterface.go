package databaseInterface

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MUR4SH/MyMessenger/structures"
)

const LIMIT = 20

type DatabaseInterface struct {
	clientOptions              options.ClientOptions
	client                     mongo.Client
	collectionMessages         mongo.Collection
	collectionUsers            mongo.Collection
	collectionChats            mongo.Collection
	collectionFiles            mongo.Collection
	collectionChatsArray       mongo.Collection
	collectionChatSettings     mongo.Collection
	collectionPersonalSettings mongo.Collection
}

func New(
	address string,
	database string,
	coll_messages string,
	coll_users string,
	coll_chats string,
	coll_files string,
	coll_chat_settings string,
	coll_chats_array string,
	coll_personal_settings string,
) DatabaseInterface {
	clientOptions := options.Client().ApplyURI(address)
	//Коннект к бд
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	//Проверяем подключение
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	db := client.Database(database)

	collectionMessages := db.Collection(coll_messages)
	collectionUsers := db.Collection(coll_users)
	collectionFiles := db.Collection(coll_files)
	collectionChats := db.Collection(coll_chats)
	collectionChatsArray := db.Collection(coll_chats_array)
	collectionChatSettings := db.Collection(coll_chat_settings)
	collectionPersonalSettings := db.Collection(coll_personal_settings)
	log.Print("Connected to database\n")
	log.Print(coll_chats)

	return DatabaseInterface{
		*clientOptions,
		*client,
		*collectionMessages,
		*collectionUsers,
		*collectionChats,
		*collectionFiles,
		*collectionChatsArray,
		*collectionChatSettings,
		*collectionPersonalSettings,
	}
}

//Получаем список чатов пользователя
func (d DatabaseInterface) GetUsersChats(user_id string, limit int, offset int) ([]structures.Chats_array, error) {
	var res []structures.Chats_array
	userId, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		log.Println("Invalid id")
	}

	if limit <= 0 {
		limit = LIMIT
	}

	if offset < limit || offset < 0 {
		offset = 0
	}

	result, err := (d.collectionUsers.Aggregate(context.TODO(), mongo.Pipeline{
		{{"$match", bson.D{{"_id", userId}}}},
		{{"$lookup", bson.M{
			"from":         "Chats_array",
			"localField":   "chats_array",
			"foreignField": "_id",
			"as":           "chats_array",
			"pipeline": []bson.M{
				{
					"$skip": offset,
				},
				{
					"$limit": limit,
				},
			},
		}}},
		{{"$project", bson.M{
			"chats_array": 1,
			"_id":         0,
		}}},
	}))

	for result.Next(context.TODO()) {
		var elem structures.Chats_array
		err := result.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res, err
}

//Получаем данные пользователя
//Агрегируем настройки и список чатов
func (d DatabaseInterface) GetUser(login string, limit int, offset int) ([]structures.User, error) {
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
	var res structures.Chat
	err := d.collectionUsers.FindOne(context.TODO(), bson.D{{"login", login}, {"password", password}}).Decode(&res)

	return res.Id, err
}

func (d DatabaseInterface) SendMessage(chat_id string, user_id string, text string) (bool, error) {
	time := time.Now().UTC()
	var msg structures.MessageInsert
	msg.Chat_id = chat_id
	msg.Gtm_date = time.Format("2006-01-02 15:04:05")
	msg.Text = text
	msg.User_id = user_id

	ins, err := d.collectionMessages.InsertOne(context.TODO(), msg)
	if err != nil {
		return false, err
	}
	oid := ins.InsertedID
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "messages_array", Value: oid},
		}},
	}

	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	_, err = d.collectionChats.UpdateOne(context.TODO(), bson.D{{"_id", objectId}}, update)
	return err == nil, err
}

//Метод создания настроек чата
func (d DatabaseInterface) insertChatSettings(
	user_id string,
	chat_id string,
	secured bool,
	search_visible bool,
	resend bool,
	users_write_permission bool,
	personal bool,
) (*mongo.InsertOneResult, error) {
	var f structures.Chat_settings
	f.Secured = secured || personal //Если чат персональный, то автоматически защищенный
	f.Search_visible = search_visible
	f.Resend = resend && !secured //Если чат защищен, то запрещаем пересылку
	f.Users_write_permission = users_write_permission
	f.Personal = personal

	res, err := d.collectionChatSettings.InsertOne(context.TODO(), f)
	if err != nil {
		log.Println(err)
		return res, err
	}

	return res, err
}

//Добавляем чат в список пользователя
func (d DatabaseInterface) pushUsersChats(
	user_id string,
	chat_element_id string,
) (bool, error) {
	objectId, err := primitive.ObjectIDFromHex(chat_element_id)

	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "chats_array", Value: objectId},
		}},
	}

	objectId, err = primitive.ObjectIDFromHex(user_id)
	if err != nil {
		log.Println("Invalid users id")
	}
	_, err = d.collectionUsers.UpdateOne(context.TODO(), bson.D{{"_id", objectId}}, update)

	return err != nil, err
}

//Метод создания элемента чата пользователя
func (d DatabaseInterface) insertUsersChatsArray(
	user_id string,
	chat_id string,
	privateKey *rsa.PrivateKey,
) (string, error) {
	var f structures.Chats_array
	f.Chat_id = chat_id
	f.Key = privateKey

	res, err := d.collectionChatsArray.InsertOne(context.TODO(), f)
	if err != nil {
		log.Println(err)
		return "Error inserting new chat element", err
	}

	res2, err2 := d.pushUsersChats(user_id, res.InsertedID.(string))
	if !res2 {
		log.Println(err2)
		return "Error pushing chat to user", err2
	}

	return res.InsertedID.(string), err
}

//Метод создания чата
func (d DatabaseInterface) CreateChat(
	user_id string,
	name string,
	logo string,
	users []string,
	privateKey *rsa.PrivateKey,
	publicKey rsa.PublicKey,
	secured bool,
	search_visible bool,
	resend bool,
	users_write_permission bool,
	personal bool,
) (string, error) {
	var f structures.Chat
	f.Chat_name = name
	f.Chat_logo = logo
	f.Users_array = append([]string{user_id}, users...)
	f.Admins_array = []string{user_id}

	//Если зашифрованный или персональный чат, то шифруем
	if secured || personal {
		f.Key = &publicKey
		r, _ := rsa.EncryptPKCS1v15(rand.Reader, &publicKey, []byte(name))
		f.Chat_name = string(r)
		r, _ = rsa.EncryptPKCS1v15(rand.Reader, &publicKey, []byte(logo))
		f.Chat_logo = string(r)

		var arr []string
		for i := 0; i < len(users); i++ {
			r, _ = rsa.EncryptPKCS1v15(rand.Reader, &publicKey, []byte(users[i]))
			arr = append(arr, string(r))
		}
		f.Users_array = arr

		r, _ = rsa.EncryptPKCS1v15(rand.Reader, &publicKey, []byte(user_id))
		f.Admins_array = []string{string(r)}
	}

	res, err := d.collectionChats.InsertOne(context.TODO(), f)

	if err != nil {
		log.Println(err)
		return "", err
	}

	//Создаем элемент настроек чата
	res_settings, err := d.insertChatSettings(
		user_id,
		res.InsertedID.(string),
		secured,
		search_visible,
		resend,
		users_write_permission,
		personal,
	)
	//Добавляем чат пользователzv
	for i := 0; i < len(users); i++ {
		_, err = d.insertUsersChatsArray(
			users[i],
			res.InsertedID.(string),
			privateKey,
		)
	}

	d.collectionChats.UpdateOne(
		context.TODO(),
		bson.M{"_id": res},
		bson.D{
			{"$set", bson.D{{"options", res_settings.InsertedID}}},
		},
	)

	return res.InsertedID.(string), err
}

//Метод сохранения файла и добавления записи в бд
func (d DatabaseInterface) CreateFile(user_id string, file []byte) (string, error) {
	var err error
	var f structures.Files
	tm := time.Now().UTC()
	f.Name = "name"
	f.Type = "type"
	f.Gtm_date = tm.Format("2006-01-02 15:04:05")
	f.ExpiredAt = tm.AddDate(0, 6, 0).Format("2006-01-02 15:04:05")
	f.Message_id = nil
	f.Url = "/files/*.type"

	res, err := d.collectionFiles.InsertOne(context.TODO(), f)

	return res.InsertedID.(string), err
}
