package serverAndHandlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/MUR4SH/MyMessenger/databaseInterface"
	"github.com/MUR4SH/MyMessenger/structures"
)

//Карта авторизованных пользователей строка - токен, значение - id
var users map[string]structures.TokenStore

//Карта времени удаления пользователей, где ключ - время создания, значение - массив токенов
var delete_users map[string][]string

var dbInterface *databaseInterface.DatabaseInterface

const COOKIE_NAME = "token"
const NOT_DONE = 501
const NOT_AUTHORISED = 500
const NOT_FOUND = 400
const OK = 200

//Удаляем токены пользователей раз в час
func timeoutTokens() {
	log.Print("Initiate deleting timeout tokens\n")
	for {
		log.Print("Deleting timeout tokens\n")
		//Получаем дату создания записи методом текущая дата минус 23 часа
		//Чтобы не удалить только что созданные записи
		pastDate := (time.Now().UTC().Add(-23 * time.Hour)).Format("OK6-01-02 15")
		counter := 0

		if arr, ok := delete_users[pastDate]; ok && len(arr) > 0 {
			for i := 0; i < len(arr); i++ {
				deleteUser(arr[i])
				counter++
			}
			if len(delete_users[pastDate]) == 0 {
				delete(delete_users, pastDate)
			}
		}

		log.Print(counter, " token(-s) has(-ve) been deleted\n")
		//Ждем один час для повтора
		time.Sleep(time.Hour)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET")
	(*w).Header().Set("Access-Control-Max-Age", "1000")
	(*w).Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
}

//Функция шифрования пароля
func GetSHA256Hash(text string) string {
	sha := sha256.New()
	sha.Write([]byte(text))
	return hex.EncodeToString(sha.Sum(nil))
}

//Обновляет и возвращает токен, обёрнутый в json
func updateToken(token string) structures.TokenJson {
	id := users[token].Id
	deleteUser(token)
	return createUser(id)
}

//Создает запись в картах и возвращает токен, обёрнутый в json
func createUser(id string) structures.TokenJson {
	var t structures.TokenJson
	t.Token = generateString()
	date := (time.Now().UTC()).Format("OK6-01-02 15")

	users[t.Token] = structures.TokenStore{Id: id, Date: date}

	if _, ok := delete_users[date]; ok {
		delete_users[date] = append(delete_users[date], t.Token)
	} else {
		delete_users[date] = []string{t.Token}
	}

	return t
}

//Функция проверки токена из кук
func verifyToken(c *http.Cookie, e error) bool {
	//TODO: добавить проверку что если дата создания > 12 часов назад, то обновить куку
	//curDate := (time.Now().UTC().Add(-23 * time.Hour)).Format("OK6-01-02 15")
	token := c.Value
	if token != "" {
		if _, ok := users[token]; ok {
			return true
		}
	}
	return false
}

func deleteUser(token string) {
	array := delete_users[token]
	var new_array []string

	for i := 0; i < len(array); i++ {
		if array[i] != users[token].Id {
			new_array = append(new_array, array[i])
		}
	}

	//Удаляем из карты отслеживания времени
	delete_users[token] = new_array
	//Удаляем из карты пользователей
	delete(users, token)
}

//Получаем токен из куки и удаляем пользователя
func deleteUserByCookie(c *http.Cookie, e error) {
	deleteUser(c.Value)
}

//Получаем пользователей чата
func getUsersOfChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting users of chat\n")

	enableCors(&w)
	if !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") || !r.URL.Query().Has("chat_id") {
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, "NOT_FOUND")
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "limit error")
		return
	}
	c, _ := r.Cookie(COOKIE_NAME)
	arr, _ := dbInterface.GetUsersOfChat(users[c.Value].Id, r.URL.Query().Get("chat_id"), limit, offset)
	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получаем чаты пользователя
func getUsersChats(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting chats of user\n")

	enableCors(&w)
	if !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") {
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, "NOT_FOUND")
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "limit error")
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	arr, err := dbInterface.GetUsersChats(
		users[c.Value].Id,
		limit,
		offset,
	)

	if err != nil {
		w.WriteHeader(503)
		fmt.Fprintf(w, "Error getting messages")
		return
	}

	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получить чат
func getChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting chat info\n")

	enableCors(&w)
	if !r.URL.Query().Has("chat_id") {
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, "NOT_FOUND")
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	arr, err := dbInterface.GetChat(r.URL.Query().Get("chat_id"))
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, err.Error())
		return
	}
	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получаем сообщения чата
func getMessages(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting messages of chat\n")

	enableCors(&w)
	if !r.URL.Query().Has("chat_id") {
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, "No chat_id")
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, "limit error")
		return
	}
	c, _ := r.Cookie(COOKIE_NAME)
	arr, err := dbInterface.GetMessages(users[c.Value].Id, r.URL.Query().Get("chat_id"), limit, offset)
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, err.Error())
		return
	}
	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Генерация токена
func generateString() string {
	str := ""
	bl := true
	for bl {
		charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP1234567890"
		length := mrand.Intn(20) + 20
		for i := 0; i < length; i++ {
			random := mrand.Intn(len(charSet))
			str += string(charSet[random])
		}
		if _, ok := users[str]; ok {
			bl = true
		} else {
			bl = false
		}
	}
	return str
}

//Авторизация
func authoriseUser(w http.ResponseWriter, r *http.Request) {
	log.Print(" Authorising\n")

	enableCors(&w)

	var m structures.UserAuthorise
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(NOT_FOUND)
		http.Error(w, err.Error(), NOT_FOUND)
		return
	}

	// Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_FOUND)
		http.Error(w, err.Error(), NOT_FOUND)
		return
	}
	id, err := dbInterface.Authorise(m.Login, GetSHA256Hash(m.Password))

	if err != nil {
		w.WriteHeader(503)
		log.Print(err)
		http.Error(w, "Authorising error", 503)
		return
	}

	b, _ = json.Marshal(createUser(id)) //Делаем json ответ с токеном
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

func exit(w http.ResponseWriter, r *http.Request) {
	log.Print(" Exiting\n")

	enableCors(&w)

	var m structures.TokenJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(NOT_FOUND)
		http.Error(w, err.Error(), NOT_FOUND)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, "token not found")
		return
	}

	deleteUserByCookie(r.Cookie(COOKIE_NAME))

	w.WriteHeader(OK)
	fmt.Fprintf(w, "success")
}

//Проверка токена
func verifyTokenFunc(w http.ResponseWriter, r *http.Request) {
	log.Print(" Verifying token\n")

	enableCors(&w)

	var m structures.TokenJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
	} else {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "OK")
	}
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	log.Print(" Sending message\n")

	enableCors(&w)

	var m structures.MessageJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	res, err := dbInterface.SendMessage(m.Chat_id, users[c.Value].Id, m.Text)
	if err != nil && !res {
		w.WriteHeader(OK)
		fmt.Fprintf(w, err.Error())
		return
	}
	w.WriteHeader(OK)
	fmt.Fprintf(w, "success")
}

//Ручка создания чата
func createChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Creating chat\n")

	enableCors(&w)
	var m structures.ChatCreationJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(NOT_FOUND)
		http.Error(w, err.Error(), NOT_FOUND)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	//Verivying token
	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	//Getting token
	c, _ := r.Cookie(COOKIE_NAME)
	var logo_id string

	if m.Logo != nil {
		//TODO процесс преобразования файла и его сохранение в директорию files
		logo_id, err = dbInterface.CreateFile(users[c.Value].Id, m.Logo, m.Logo_url)
	}

	key, _ := rsa.GenerateKey(rand.Reader, 50)

	//Создаем чат (файл)
	res, err := dbInterface.CreateChat(
		users[c.Value].Id,
		m.Name,
		logo_id,
		m.Users,
		key,           //Приватный ключ
		key.PublicKey, //Публичный ключ
		m.Secured,
		m.Search_visible,
		m.Resend,
		m.Users_write_permission,
		m.Personal,
	)
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(OK)
	fmt.Fprintf(w, res)
}

func getChatKey(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting users key\n")

	enableCors(&w)

	var m structures.ChatJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(NOT_DONE)
		http.Error(w, err.Error(), NOT_DONE)
		return
	}

	if !verifyToken(r.Cookie(COOKIE_NAME)) {
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, "NOT_AUTHORISED")
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	res, err := dbInterface.GetUsersKey(users[c.Value].Id, m.Id)
	if err != nil {
		w.WriteHeader(OK)
		fmt.Fprintf(w, "Error getting key")
		return
	}

	b, _ = json.Marshal(res)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получаем порт и интерфейс для работы с бд
func InitServer(port string, db *databaseInterface.DatabaseInterface) {
	mrand.Seed(time.Now().Unix())
	dbInterface = db
	users = make(map[string]structures.TokenStore)
	delete_users = make(map[string][]string)

	go timeoutTokens() //Запускаем функцию на проверку актуальности токенов в отдельном потоке

	//GET Ручки
	http.HandleFunc("/usersChats", getUsersChats) //Получить чаты пользователя
	http.HandleFunc("/chatUsers", getUsersOfChat) //Получить пользователей чата
	http.HandleFunc("/сhat", getChat)             //Получить информацию чата
	http.HandleFunc("/messages", getMessages)     //Получить сообщения чата
	http.HandleFunc("/chatKey", getChatKey)       //Получить сообщения чата
	//TODO: гет-ручка обновления токена

	//POST Ручки
	http.HandleFunc("/authorise", authoriseUser)     //Авторизовать
	http.HandleFunc("/exit", exit)                   //Выйти
	http.HandleFunc("/verifyToken", verifyTokenFunc) //Перепроверить токен
	http.HandleFunc("/sendMessage", sendMessage)     //Отправить сообщение
	http.HandleFunc("/createChat", createChat)       //Отправить сообщение

	log.Print(" Starting server\n")
	log.Print(" Server started\n")
	log.Fatal(http.ListenAndServe(":"+port, nil)) //Запускаем сервер, оборачиваем в логирование чтоб видеть результат
	log.Print(" Server finished\n")
}
