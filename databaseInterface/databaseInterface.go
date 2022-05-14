package databaseInterface

import (
	"context"
	"crypto/rsa"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	security "github.com/MUR4SH/MyMessenger/security"
	"github.com/MUR4SH/MyMessenger/structures"
)

const LIMIT = 20
const SECURED_MESSAGE_LIMIT = 245
const DATE_FORMAT = "2006-01-02 15:04:05"

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

//Функция инициальзации подключения к бд и создания интерфейса взаимодействия
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
			{Key: "as", Value: "chats_arrays"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$match", Value: bson.D{{Key: "chat_id", Value: chatId}},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_arrays", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))

	if err != nil {
		log.Println("err")
		log.Println(err)
	}

	for result.Next(context.TODO()) {
		var elem structures.Chats_array_agregate
		err := result.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res[0].Chats_arrays[0], err
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
			{Key: "as", Value: "chats_arrays"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$skip", Value: offset,
				}},
				{{
					Key: "$limit", Value: limit,
				}},
				{{Key: "$project", Value: bson.D{
					{Key: "key", Value: 0},
				}}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_arrays", Value: 1},
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
func (d DatabaseInterface) GetUsersKey(user_id string, chat_id string) ([]byte, error) {
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
				{Key: "$slice", Value: []interface{}{"$users_array", offset, limit}},
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

	if err != nil {
		log.Println("err")
		log.Println(err)
	}

	for cur.Next(context.TODO()) {
		var elem structures.Chat_User_aggregate_lite
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = elem.Users_array
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

	if err != nil {
		log.Println("here")
		log.Println(err)
	}

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

	if err != nil {
		log.Println("error")
		log.Println(err)
	}

	for cur.Next(context.TODO()) {
		var elem structures.Chat_settings_agregate
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, elem)
	}

	return res[0].Options[0], err
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
			{Key: "as", Value: "chats_arrays"},
			{Key: "pipeline", Value: []bson.D{
				{{
					Key: "$match", Value: bson.D{
						{Key: "chat_id", Value: chatId},
					},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "chats_arrays", Value: 1},
			{Key: "_id", Value: 0},
		}}},
	}))

	if err != nil {
		return false
	}

	for cur.Next(context.TODO()) {
		var elem structures.Chats_array_agregate
		err := cur.Decode(&elem)
		log.Println(elem)
		if err != nil {
			log.Println(err)
		} else {
			res = (len(elem.Chats_arrays) == 1)
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
		log.Println("User not in chat")
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
		key, _ := d.GetUsersKey(user_id, chat_id)
		decrypted_key := security.PrivateKeyFromPEM(key)
		log.Println(string(security.Decrypt(elem.Text, decrypted_key)))
		res = append(res, elem)
	}
	return res, err
}

//Метод получения расшифрованных сообщений
func (d DatabaseInterface) GetDecryptedMessages(user_id string, chat_id string, limit int, offset int) ([]structures.MessageToUser, error) {
	key, _ := d.GetUsersKey(user_id, chat_id)
	decrypted_key := security.PrivateKeyFromPEM(key)

	messages, err := d.GetMessages(user_id, chat_id, limit, offset)

	for i := 0; i < len(messages); i++ {
		messages[i].Text = security.Decrypt(messages[i].Text, decrypted_key)
	}

	return messages, err
}

//Метод авторизации, проверяет пользователя по логину и паролю, возвращая id
func (d DatabaseInterface) Authorise(login string, password string) (string, error) {
	var res structures.Chat
	err := d.collectionUsers.FindOne(context.TODO(), bson.D{{Key: "login", Value: login}, {Key: "password", Value: password}}).Decode(&res)

	return res.Id.Hex(), err
}

//Метод отправки уже зашифрованных сообщений
func (d DatabaseInterface) SendEncryptedMessage(chat_id string, user_id string, text []byte) (bool, error) {
	var msg structures.Message_noid

	time := time.Now().UTC()
	objectId, _ := primitive.ObjectIDFromHex(chat_id)
	msg.Chat_id = objectId
	msg.Gtm_date = time.Format(DATE_FORMAT)

	if !d.ChatIsSecured(chat_id) {
		return false, errors.New("chat is not secured")
	}

	msg.Text = text
	userId, _ := primitive.ObjectIDFromHex(user_id)
	msg.User_id = userId

	ins, err := d.collectionMessages.InsertOne(context.TODO(), msg)
	if err != nil {
		log.Println(err)
		return false, err
	}
	oid := ins.InsertedID
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "messages_array", Value: oid},
		}},
	}

	_, err = d.collectionChats.UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: objectId}}, update)
	return err == nil, err
}

//Метод отправки сообщений
func (d DatabaseInterface) SendMessage(chat_id string, user_id string, text string) (bool, error) {
	time := time.Now().UTC()
	var msg structures.Message_noid
	var byte_text []byte
	objectId, _ := primitive.ObjectIDFromHex(chat_id)
	msg.Chat_id = objectId
	msg.Gtm_date = time.Format(DATE_FORMAT)
	if d.ChatIsSecured(chat_id) {

		if len(text) > SECURED_MESSAGE_LIMIT {
			return false, errors.New("message length more than limit")
		}

		key, e := d.GetUsersKey(user_id, chat_id)
		if e != nil {
			log.Println(e)
			return false, e
		}

		decodedKey := security.PrivateKeyFromPEM(key)

		byte_text = security.Encrypt(text, &decodedKey.PublicKey)
	} else {
		byte_text = []byte(text)
	}

	msg.Text = byte_text
	userId, _ := primitive.ObjectIDFromHex(user_id)
	msg.User_id = userId

	ins, err := d.collectionMessages.InsertOne(context.TODO(), msg)
	if err != nil {
		log.Println(err)
		return false, err
	}
	oid := ins.InsertedID
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{
			primitive.E{Key: "messages_array", Value: oid},
		}},
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
	chatId, _ := primitive.ObjectIDFromHex(chat_id)
	f.Chat_id = chatId
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
		return false, errors.New("invalid user's id")
	}
	_, err = d.collectionUsers.UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: objectId}}, update)

	return err != nil, err
}

//Метод создания элемента чата пользователя
func (d DatabaseInterface) insertUsersChatsArray(
	user_id string,
	chat_id string,
	privateKey *rsa.PrivateKey,
	personal bool,
) (string, error) {
	var f structures.Chats_array_noid
	objectId, _ := primitive.ObjectIDFromHex(chat_id)
	f.Chat_id = objectId
	f.Key = security.PrivateKeyPEM(privateKey)
	f.Personal = personal

	res, err := d.collectionChatsArray.InsertOne(context.TODO(), f)
	if err != nil {
		log.Println(err)
		return "", err
	}
	oid, _ := res.InsertedID.(primitive.ObjectID)
	res2, err2 := d.pushUsersChats(user_id, oid.Hex())
	if !res2 {
		log.Println(err2)
		return "", err2
	}

	return oid.Hex(), err
}

//Метод вычисляет есть ли персональный чат у двух пользователей
func (d DatabaseInterface) hasPersonalChat(first_id string, second_id string) (bool, string) {
	var res []structures.ID
	firstId, _ := primitive.ObjectIDFromHex(first_id)
	secondId, _ := primitive.ObjectIDFromHex(second_id)

	cur, err := (d.collectionChats.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "users_array", Value: firstId}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "users_array", Value: secondId}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Chat_settings"},
			{Key: "localField", Value: "options"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "options"},
			{Key: "pipeline", Value: []bson.D{{{
				Key: "$match", Value: bson.D{
					{Key: "personal", Value: true},
				}}},
			}},
		}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "options.0", Value: bson.D{{
			Key: "$exists", Value: true,
		}}}}}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 1}}}},
	}))

	if err != nil {
		return false, ""
	}

	for cur.Next(context.TODO()) {
		var elem structures.ID
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(elem)
		res = append(res, elem)
	}

	id := ""
	if len(res) > 0 {
		id = res[0].Id.Hex()
	}

	return len(res) > 0, id
}

//Метод создания чата
func (d DatabaseInterface) CreateChat(
	user_id string,
	name string,
	logo string,
	users []string,
	privateKey rsa.PrivateKey,
	publicKey rsa.PublicKey,
	secured bool,
	search_visible bool,
	resend bool,
	users_write_permission bool,
	personal bool,
) (string, error) {
	if personal {
		if len(users) != 1 {
			return "", errors.New("wrong users length. Must be 1")
		}
		ok, id := d.hasPersonalChat(user_id, users[0])
		if ok {
			return id, nil
		}
	}

	var f structures.Chat_noid
	f.Chat_name = name
	logoId, _ := primitive.ObjectIDFromHex(logo)
	f.Chat_logo = logoId
	userId, _ := primitive.ObjectIDFromHex(user_id)
	var ar []primitive.ObjectID
	f.Admins_array = append(ar, userId)
	//Если зашифрованный или персональный чат, то шифруем
	if secured || personal {
		f.Key = security.PublicKeyPEM(&publicKey)
	}

	var arr []primitive.ObjectID
	arr = append(arr, userId)
	for i := 0; i < len(users); i++ {
		userId, _ = primitive.ObjectIDFromHex(users[i])
		arr = append(arr, userId)
	}
	if personal {
		f.Admins_array = arr
	}
	f.Users_array = arr

	f.Messages_array = []primitive.ObjectID{}
	f.Files_array = []primitive.ObjectID{}
	f.Invited_array = []primitive.ObjectID{}
	f.Banned_array = []primitive.ObjectID{}

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
			&privateKey,
			personal,
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
	f.Gtm_date = tm.Format(DATE_FORMAT)
	f.ExpiredAt = tm.AddDate(0, 6, 0).Format(DATE_FORMAT)
	f.Message_id = nil
	if url != nil {
		f.Url = *url
	} else {
		f.Url = "/files/*.type"
	}

	res, err := d.collectionFiles.InsertOne(context.TODO(), f)
	if err != nil {
		log.Println(err)
	}
	oid, _ := res.InsertedID.(primitive.ObjectID)
	return oid.Hex(), err
}

//Метод регистрации
func (d DatabaseInterface) Registration(user *structures.CreateUserJSON) (string, error) {
	//TODO - доделать
	var res structures.Chat
	err := d.collectionUsers.FindOne(context.TODO(), bson.D{{Key: "login", Value: user.Login}, {Key: "email", Value: user.Email}}).Decode(&res)

	return res.Id.Hex(), err
}
