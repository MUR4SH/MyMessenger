package databaseInterface

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MUR4SH/MyMessenger/structures"
)

const LIMIT = 20

type DatabaseInterface struct {
	clientOptions          options.ClientOptions
	client                 mongo.Client
	database               mongo.Database
	collectionMessages     mongo.Collection
	collectionUsers        mongo.Collection
	collectionChats        mongo.Collection
	collectionFiles        mongo.Collection
	collectionChatsArray   mongo.Collection
	collectionChatSettings mongo.Collection
	collectionUserSettings mongo.Collection
}

func Encrypt(s string, key *rsa.PublicKey) string {
	crypt, _ := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(s))

	return string(crypt)
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
	collectionUserSettings := db.Collection(coll_personal_settings)
	log.Print("Connected to database\n")
	log.Print(coll_chats)

	return DatabaseInterface{
		*clientOptions,
		*client,
		*db,
		*collectionMessages,
		*collectionUsers,
		*collectionChats,
		*collectionFiles,
		*collectionChatsArray,
		*collectionChatSettings,
		*collectionUserSettings,
	}
}

//Получаем конкретный чат пользователя
func (d DatabaseInterface) GetUsersChat(user_id string, chat_id string) (structures.Chats_array, error) {
	var res []structures.Chats_array_agregate
	userId, _ := primitive.ObjectIDFromHex(user_id)
	chatId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}

	result, err := (d.collectionUsers.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userId}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chats_array"},
			{Key: "localField", Value: "chats_array"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "chats_array"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$match", Value: bson.D{{Key: "_id", Value: chatId}},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_array", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))

	for result.Next(context.TODO()) {
		var elem structures.Chats_array_agregate
		err := result.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res[0].Chats_array[0], err
}

//Получаем список чатов пользователя
func (d DatabaseInterface) GetUsersChats(user_id string, limit int, offset int) ([]structures.Chats_array_agregate, error) {
	var res []structures.Chats_array_agregate
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
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userId}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chats_array"},
			{Key: "localField", Value: "chats_array"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "chats_array"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$skip", Value: offset,
				}},
				{{
					Key: "$limit", Value: limit,
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_array", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))

	for result.Next(context.TODO()) {
		var elem structures.Chats_array_agregate
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

//Получить ключ пользователя
func (d DatabaseInterface) GetUsersKey(user_id string, chat_id string) (*rsa.PrivateKey, error) {
	chats, err := d.GetUsersChat(user_id, chat_id)

	if err != nil {
		log.Fatal("Error getting chats")
		return nil, err
	}

	return chats.Key, err
}

//Получить пользователей чата
func (d DatabaseInterface) GetUsersOfChat(user_id string, chat_id string, limit int, offset int) ([]structures.User_lite, error) {
	var res []structures.User_lite

	var cur *mongo.Cursor
	var err error
	objectId, err := primitive.ObjectIDFromHex(chat_id)

	if !d.UserInChat(user_id, chat_id) {
		return nil, err
	}

	if limit <= 0 {
		limit = LIMIT
	}

	if offset < limit || offset < 0 {
		offset = 0
	}

	cur, err = (d.collectionChats.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: objectId}}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "users_array", Value: bson.D{
				{Key: "$slice", Value: []string{"$users_array", strconv.Itoa(offset), strconv.Itoa(limit)}},
			}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "users_array"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "users_array"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$project", Value: bson.D{
						{Key: "password", Value: 0},
						{Key: "chats_array", Value: 0},
						{Key: "email", Value: 0},
						{Key: "phone", Value: 0},
						{Key: "personal_settings", Value: 0},
					},
				}},
			}},
		}}},
	}))

	for cur.Next(context.TODO()) {
		var elem structures.User_lite
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res, err
}

//Получаем данные чата, убирая ненужные данные
func (d DatabaseInterface) GetChat(chat_id string) (structures.Chat_lite, error) {
	var res []structures.Chat_lite
	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	cur, err := (d.collectionChats.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: objectId}}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "users_count", Value: bson.D{
				{Key: "$size", Value: "$users_array"},
			}},
			{Key: "_id", Value: 1},
			{Key: "chat_name", Value: 1},
			{Key: "chat_logo", Value: 1},
			{Key: "options", Value: 1},
		}}},
	}))
	for cur.Next(context.TODO()) {
		var elem structures.Chat_lite
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res[0], err
}

//Получаем настройки чата
func (d DatabaseInterface) getChatsOptions(chat_id string) (structures.Chat_settings, error) {
	var res []structures.Chat_settings_agregate

	objectId, err := primitive.ObjectIDFromHex(chat_id)
	if err != nil {
		log.Println("Invalid id")
	}
	cur, err := (d.collectionChats.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: objectId}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chat_settings"},
			{Key: "localField", Value: "options"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "options"},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "options", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))
	for cur.Next(context.TODO()) {
		var elem structures.Chat_settings_agregate
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(elem)
		res = append(res, elem)
	}

	return res[0].Chat_settings_array[0], err
}

//Получаем параметр защищенности чата
func (d DatabaseInterface) ChatIsSecured(chat_id string) bool {
	res, _ := d.getChatsOptions(chat_id)
	return res.Secured
}

//Получаем значение состоит ли пользователь в чате
func (d DatabaseInterface) UserInChat(user_id string, chat_id string) bool {
	res := false
	userId, _ := primitive.ObjectIDFromHex(user_id)
	chatId, _ := primitive.ObjectIDFromHex(chat_id)

	cur, err := (d.collectionUsers.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userId}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chats_array"},
			{Key: "localField", Value: "chats_array"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "chats_array"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$match", Value: bson.D{
						{Key: "chat_id", Value: chatId},
					},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_array", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))

	if err != nil {
		return false
	}

	for cur.Next(context.TODO()) {
		var elem structures.Chats_array_agregate
		err := cur.Decode(&elem)
		if err != nil {
			log.Println(err)
		} else {
			res = (len(elem.Chats_array) == 1)
		}
	}

	return res
}

//Получить сообщения
func (d DatabaseInterface) GetMessages(user_id string, chat_id string, limit int, offset int) ([]structures.MessageToUser, error) {
	var res []structures.MessageToUser
	objectId, err := primitive.ObjectIDFromHex(chat_id)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//Если пользователь не состоит в чате
	if !d.UserInChat(user_id, chat_id) {
		var er error
		return nil, er
	}

	if limit <= 0 {
		limit = LIMIT
	}

	if offset < limit || offset < 0 {
		offset = 0
	}

	cur, err := (d.collectionMessages.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "chat_id", Value: objectId}}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "gtm_date", Value: 1},
		}}},
		bson.D{{Key: "$skip", Value: offset}},
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "expiredAt", Value: 0},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "user_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "user"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$project", Value: bson.D{
						{Key: "password", Value: 0},
						{Key: "chats_array", Value: 0},
						{Key: "email", Value: 0},
						{Key: "phone", Value: 0},
						{Key: "personal_settings", Value: 0},
					},
				}},
			}},
		}}},
	}))
	for cur.Next(context.TODO()) {
		var elem structures.MessageToUser
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
	err := d.collectionUsers.FindOne(context.TODO(), bson.D{{Key: "login", Value: login}, {Key: "password", Value: password}}).Decode(&res)

	return res.Id.Hex(), err
}

func (d DatabaseInterface) SendMessage(chat_id string, user_id string, text string) (bool, error) {
	time := time.Now().UTC()
	var msg structures.Message_noid
	objectId, err := primitive.ObjectIDFromHex(chat_id)
	msg.Chat_id = objectId
	msg.Gtm_date = time.Format("2006-01-02 15:04:05")
	if d.ChatIsSecured(chat_id) {
		key, _ := d.GetUsersKey(user_id, chat_id)
		text = Encrypt(text, &key.PublicKey)
	}
	msg.Text = text
	userId, err := primitive.ObjectIDFromHex(user_id)
	msg.User_id = userId

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

	if err != nil {
		log.Println("Invalid id")
	}
	_, err = d.collectionChats.UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: objectId}}, update)
	return err == nil, err
}

//Метод создания настроек чата
func (d DatabaseInterface) insertChatSettings(
	chat_id string,
	secured bool,
	search_visible bool,
	resend bool,
	users_write_permission bool,
	personal bool,
) (*mongo.InsertOneResult, error) {
	var f structures.Chat_settings_noid
	f.Secured = secured || personal                               //Если чат персональный, то автоматически защищенный
	f.Search_visible = search_visible && !personal                //Персональные не видны в поиске
	f.Resend = resend && !f.Secured                               //Если чат защищен, то запрещаем пересылку
	f.Users_write_permission = users_write_permission || personal //В персональном чате все могут писать
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
	objectId, _ := primitive.ObjectIDFromHex(chat_element_id)

	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "chats_array", Value: objectId},
		}},
	}

	objectId, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		log.Println("Invalid users id")
		return false, err
	}
	_, err = d.collectionUsers.UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: objectId}}, update)

	return err != nil, err
}

//Метод создания элемента чата пользователя
func (d DatabaseInterface) insertUsersChatsArray(
	user_id string,
	chat_id string,
	privateKey *rsa.PrivateKey,
) (string, error) {
	var f structures.Chats_array_noid
	objectId, _ := primitive.ObjectIDFromHex(chat_id)
	f.Chat_id = objectId
	f.Key = privateKey

	res, err := d.collectionChatsArray.InsertOne(context.TODO(), f)
	if err != nil {
		log.Println(err)
		return "Error inserting new chat element", err
	}
	oid, _ := res.InsertedID.(primitive.ObjectID)
	res2, err2 := d.pushUsersChats(user_id, oid.Hex())
	if !res2 {
		log.Println(err2)
		return "Error pushing chat to user", err2
	}

	return oid.Hex(), err
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
	var null_arr []primitive.ObjectID
	var f structures.Chat_noid
	f.Chat_name = name
	logoId, _ := primitive.ObjectIDFromHex(logo)
	f.Chat_logo = logoId
	//Если зашифрованный или персональный чат, то шифруем
	if secured || personal {
		f.Key = &publicKey
	}
	userId, _ := primitive.ObjectIDFromHex(user_id)
	var ar []primitive.ObjectID
	f.Admins_array = append(ar, userId)

	var arr []primitive.ObjectID
	arr = append(arr, userId)
	for i := 0; i < len(users); i++ {
		userId, _ = primitive.ObjectIDFromHex(users[i])
		arr = append(arr, userId)
	}
	f.Users_array = arr

	f.Messages_array = null_arr
	f.Files_array = null_arr
	f.Invited_array = null_arr
	f.Banned_array = null_arr

	res, err := d.collectionChats.InsertOne(context.TODO(), f)

	if err != nil {
		log.Println(err)
		return "", err
	}
	//Создаем элемент настроек чата
	oid, _ := res.InsertedID.(primitive.ObjectID)
	res_settings, err := d.insertChatSettings(
		oid.Hex(),
		secured,
		search_visible,
		resend,
		users_write_permission,
		personal,
	)
	//Добавляем чат пользователю
	for i := 0; i < len(f.Users_array); i++ {
		_, err = d.insertUsersChatsArray(
			f.Users_array[i].Hex(),
			oid.Hex(),
			privateKey,
		)
	}

	d.collectionChats.UpdateOne(
		context.TODO(),
		bson.M{"_id": res.InsertedID},
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "options", Value: res_settings.InsertedID}}},
		},
	)
	oids, _ := res.InsertedID.(primitive.ObjectID)
	return oids.Hex(), err
}

//Метод сохранения файла и добавления записи в бд
func (d DatabaseInterface) CreateFile(user_id string, file []byte, url *string) (string, error) {
	var err error
	var f structures.Files_noid
	tm := time.Now().UTC()
	f.Name = "name"
	f.Type = "type"
	f.Gtm_date = tm.Format("2006-01-02 15:04:05")
	f.ExpiredAt = tm.AddDate(0, 6, 0).Format("2006-01-02 15:04:05")
	f.Message_id = nil
	if url != nil {
		f.Url = *url
	} else {
		f.Url = "/files/*.type"
	}

	res, err := d.collectionFiles.InsertOne(context.TODO(), f)
	oid, _ := res.InsertedID.(primitive.ObjectID)
	return oid.Hex(), err
}
