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
	"github.com/gorilla/websocket"
)

//Карта авторизованных пользователей строка - токен, значение - id
var users map[string]structures.TokenStore

//Карта времени удаления пользователей, где ключ - время создания, значение - массив токенов
var delete_users map[string][]string

var dbInterface *databaseInterface.DatabaseInterface

const COOKIE_NAME = "token"
const NOT_DONE = 501
const NOT_AUTHORISED = 200
const NOT_FOUND = 400
const OK = 200

const DATE_FORMAT = "2006-01-02 15"

//Карта чат - пользователи
var chatUsers map[string][]*websocket.Conn
var userChats map[*websocket.Conn][]string

//Карта id - connection
var usersId map[string]*websocket.Conn

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

//Удаляем токены пользователей раз в час
func timeoutTokens() {
	log.Print("Initiate deleting timeout tokens\n")
	for {
		log.Print("Deleting timeout tokens\n")
		//Получаем дату создания записи методом текущая дата минус 23 часа
		//Чтобы не удалить только что созданные записи
		pastDate := (time.Now().UTC().Add(-23 * time.Hour)).Format(DATE_FORMAT)
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

func enableCors(w *http.ResponseWriter, r string) {
	(*w).Header().Set("Access-Control-Allow-Origin", r)
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Max-Age", "1000")
	(*w).Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
}

//Функция шифрования пароля
func GetSHA256Hash(text string) string {
	sha := sha256.New()
	sha.Write([]byte(text))
	return hex.EncodeToString(sha.Sum(nil))
}

//Обновляет токен
func updateToken(token string) {
	date := (time.Now().UTC()).Format(DATE_FORMAT)

	user_date := users[token].Date
	array := delete_users[user_date]
	var new_array []string

	//Переопределяем объект в карте
	users[token] = structures.TokenStore{Id: users[token].Id, Date: date}

	for i := 0; i < len(array); i++ {
		if array[i] != users[token].Id {
			new_array = append(new_array, array[i])
		}
	}

	//Удаляем из карты отслеживания времени
	delete_users[user_date] = new_array

	//Записываем в карту под новым ключом
	if _, ok := delete_users[date]; ok {
		delete_users[date] = append(delete_users[date], token)
	} else {
		delete_users[date] = []string{token}
	}
}

//Создает запись в картах и возвращает токен, обёрнутый в json
func createUser(id string) structures.TokenJson {
	var t structures.TokenJson
	t.Token = generateString()
	date := (time.Now().UTC()).Format(DATE_FORMAT)

	users[t.Token] = structures.TokenStore{Id: id, Date: date}

	if _, ok := delete_users[date]; ok {
		delete_users[date] = append(delete_users[date], t.Token)
	} else {
		delete_users[date] = []string{t.Token}
	}

	return t
}

//Функция проверки токена из кук
func verifyTokenCookie(c *http.Cookie, e error) bool {
	if c == nil {
		return false
	}
	token := c.Value

	if token != "" {
		if _, ok := users[token]; ok {

			a, _ := time.Parse(DATE_FORMAT, users[token].Date)
			d := (time.Since(a)).Hours()
			if d >= 12 {
				updateToken(token)
			}

			return true
		}
	}
	return false
}

//Функция проверки токена из кук
func verifyToken(c string) bool {
	if c == "" {
		return false
	}
	token := c

	if token != "" {
		if _, ok := users[token]; ok {

			a, _ := time.Parse(DATE_FORMAT, users[token].Date)
			d := (time.Since(a)).Hours()
			if d >= 12 {
				updateToken(token)
			}

			return true
		}
	}
	return false
}

func deleteUser(token string) {
	date := users[token].Date
	array := delete_users[date]
	var new_array []string

	for i := 0; i < len(array); i++ {
		if array[i] != users[token].Id {
			new_array = append(new_array, array[i])
		}
	}

	//Удаляем из карты отслеживания времени
	delete_users[date] = new_array
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
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	if !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") || !r.URL.Query().Has("chat_id") {
		answ.Text = "NOT_FOUND"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, string(b))
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(b))
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		answ.Text = "offset error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		answ.Text = "limit error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
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
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	if !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") {
		answ.Text = "NOT_FOUND"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, string(b))
		return
	}
	log.Println(verifyTokenCookie(r.Cookie(COOKIE_NAME)))
	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(b))
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		answ.Text = "offset error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		answ.Text = "limit error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	arr, err := dbInterface.GetUsersChats(
		users[c.Value].Id,
		limit,
		offset,
	)

	if err != nil {
		answ.Text = "Error getting messages"
		b, _ := json.Marshal(answ)
		w.WriteHeader(501)
		fmt.Fprintf(w, string(b))
		return
	}

	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получить чат
func getChatLite(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting chat info\n")
	var answ structures.Answer
	enableCors(&w, r.Header.Get("Origin"))

	c, _ := r.Cookie(COOKIE_NAME)

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(b))
		return
	}

	if !r.URL.Query().Has("chat_id") {
		answ.Text = "NO CHAT_ID"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(b))
		return
	}

	arr, err := dbInterface.GetChat(users[c.Value].Id, r.URL.Query().Get("chat_id"))
	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
		return
	}
	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получаем сообщения чата
func getMessages(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting messages of chat\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	if !r.URL.Query().Has("chat_id") {
		answ.Text = "No chat_id"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(b))
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(b))
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		answ.Text = "offset error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(b))
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		answ.Text = "limit error"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(b))
		return
	}
	c, _ := r.Cookie(COOKIE_NAME)
	arr, err := dbInterface.GetMessages(users[c.Value].Id, r.URL.Query().Get("chat_id"), limit, offset)
	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
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
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	var m structures.UserAuthorise
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		http.Error(w, string(b), NOT_FOUND)
		return
	}

	// Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		log.Println(answ.Text)
		w.WriteHeader(NOT_FOUND)
		http.Error(w, string(b), NOT_FOUND)
		return
	}
	id, err := dbInterface.Authorise(m.Login, GetSHA256Hash(m.Password))

	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		w.WriteHeader(501)
		http.Error(w, string(b), 501)
		return
	}

	token := createUser(id)

	b, _ = json.Marshal(createUser(id)) //Делаем json ответ с токеном

	ck := http.Cookie{
		Name:   "token",
		Value:  token.Token,
		Domain: "",
		Path:   "/",
	}
	http.SetCookie(w, &ck)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

func exit(w http.ResponseWriter, r *http.Request) {
	log.Print(" Exiting\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	var m structures.TokenJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		http.Error(w, string(bs), NOT_FOUND)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "token not found"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		fmt.Fprintf(w, string(bs))
		return
	}

	deleteUserByCookie(r.Cookie(COOKIE_NAME))

	answ.Text = "success"
	bs, _ := json.Marshal(answ)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(bs))
}

//Проверка токена
func verifyTokenReq(w http.ResponseWriter, r *http.Request) {
	log.Print(" Verifying token\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	var m structures.TokenJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	cookie, _ := r.Cookie(COOKIE_NAME)
	var verify bool
	if cookie == nil {
		verify = verifyToken(m.Token)
	} else {
		verify = verifyTokenCookie(r.Cookie(COOKIE_NAME))
	}

	if verify {
		answ.Text = users[cookie.Value].Id
		log.Printf("id")
		log.Printf(answ.Text)
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
	} else {
		answ.Text = ""
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
	}
}

//Функция отправки сообщений
func sendMessage(w http.ResponseWriter, r *http.Request) {
	log.Print(" Sending message\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	var m structures.MessageJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil || len(m.Text) == 0 {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	res, err := dbInterface.SendMessage(m.Chat_id, users[c.Value].Id, m.Text)
	if err != nil && !res {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
		return
	}
	answ.Text = "success"
	bs, _ := json.Marshal(answ)

	w.WriteHeader(OK)
	fmt.Fprintf(w, string(bs))

	for i := 0; i < len(chatUsers[m.Chat_id]); i++ {
		chatUsers[m.Chat_id][i].WriteMessage(websocket.TextMessage, []byte(m.Chat_id))
	}
}

//Ручка создания чата
func createChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Creating chat\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	var m structures.ChatCreationJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		http.Error(w, string(bs), NOT_FOUND)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	//Verivying token
	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
		return
	}

	//Getting token
	c, _ := r.Cookie(COOKIE_NAME)
	var logo_id string

	if m.Logo != nil {
		//TODO процесс преобразования файла и его сохранение в директорию files
		logo_id, err = dbInterface.CreateFile(users[c.Value].Id, m.Logo, m.Logo_url)
	}

	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	//Создаем чат (файл)
	res, err := dbInterface.CreateChat(
		users[c.Value].Id,
		m.Name,
		logo_id,
		m.Users,
		*key,          //Приватный ключ
		key.PublicKey, //Публичный ключ
		m.Secured,
		m.Search_visible,
		m.Resend,
		m.Users_write_permission,
		m.Personal,
	)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
		return
	}

	answ.Text = res
	bs, _ := json.Marshal(answ)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(bs))
}

func getChatKey(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting user's key\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	if !r.URL.Query().Has("chat_id") {
		answ.Text = "NO CHAT_ID"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(bs))
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	res, err := dbInterface.GetUsersKey(users[c.Value].Id, r.URL.Query().Get("chat_id"))
	if err != nil {
		answ.Text = "Error getting key"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
		return
	}

	b, _ := json.Marshal(res)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

func registration(w http.ResponseWriter, r *http.Request) {
	log.Print(" Registration\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	var m structures.CreateUserJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_FOUND)
		http.Error(w, string(bs), NOT_FOUND)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		http.Error(w, string(bs), NOT_DONE)
		return
	}

	//Создаем чат (файл)
	res, err := dbInterface.Registration(&m)
	if err != nil {
		answ.Text = err.Error()
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
		return
	}

	answ.Text = res
	bs, _ := json.Marshal(answ)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(bs))
}

func getUser(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting user\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	if !r.URL.Query().Has("user_id") {
		answ.Text = "NO USER_ID"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(bs))
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)

	res, err := dbInterface.GetUserId(users[c.Value].Id, r.URL.Query().Get("user_id"))
	if err != nil {
		answ.Text = "Error getting user"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(bs))
		return
	}

	b, _ := json.Marshal(res)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

func webSocket(w http.ResponseWriter, r *http.Request) {
	log.Print(" Connecting\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))

	if !r.URL.Query().Has("user_id") {
		answ.Text = "NOT_AUTHORISED"
		bs, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(bs))
		return
	}

	c := r.URL.Query().Get("user_id")

	connection, _ := upgrader.Upgrade(w, r, nil)
	re, _ := dbInterface.GetUsersChatsId(c)

	updateChats(connection, re, c)
}

func updateChats(connection *websocket.Conn, array []structures.Chat_Id, user_id string) {
	log.Println(array)
	for i := 0; i < len(array); i++ {
		chatUsers[array[i].Chat_id.Hex()] = append(chatUsers[array[i].Chat_id.Hex()], connection)
		userChats[connection] = append(userChats[connection], array[i].Chat_id.Hex())
		usersId[user_id] = connection
	}
}

func deletChatUser(connection *websocket.Conn) {
	if array, ok := userChats[connection]; ok {
		for i := 0; i < len(array); i++ {
			var new_array []*websocket.Conn
			for j := 0; j < len(chatUsers[array[i]]); j++ {
				if chatUsers[array[i]][j] != connection {
					new_array = append(new_array, chatUsers[array[i]][j])
				}
			}
			chatUsers[array[i]] = new_array
		}
		delete(userChats, connection)
	}
}

//Получаем новые сообщения из чата
func getNewMessages(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting new messages of chat\n")
	var answ structures.Answer

	enableCors(&w, r.Header.Get("Origin"))
	if !r.URL.Query().Has("chat_id") {
		answ.Text = "No chat_id"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_DONE)
		fmt.Fprintf(w, string(b))
		return
	}

	if !verifyTokenCookie(r.Cookie(COOKIE_NAME)) {
		answ.Text = "NOT_AUTHORISED"
		b, _ := json.Marshal(answ)
		w.WriteHeader(NOT_AUTHORISED)
		fmt.Fprintf(w, string(b))
		return
	}

	c, _ := r.Cookie(COOKIE_NAME)
	log.Println(c.Value)
	arr, err := dbInterface.GetNewMessages(users[c.Value].Id, r.URL.Query().Get("chat_id"))
	if err != nil {
		answ.Text = err.Error()
		b, _ := json.Marshal(answ)
		w.WriteHeader(OK)
		fmt.Fprintf(w, string(b))
		return
	}
	b, _ := json.Marshal(arr)
	w.WriteHeader(OK)
	fmt.Fprintf(w, string(b))
}

//Получаем порт и интерфейс для работы с бд
func InitServer(port string, db *databaseInterface.DatabaseInterface) {
	mrand.Seed(time.Now().Unix())
	dbInterface = db
	users = make(map[string]structures.TokenStore)
	delete_users = make(map[string][]string)
	chatUsers = make(map[string][]*websocket.Conn)
	userChats = make(map[*websocket.Conn][]string)
	usersId = make(map[string]*websocket.Conn)

	go timeoutTokens() //Запускаем функцию на проверку актуальности токенов в отдельном потоке

	//GET Ручки
	http.HandleFunc("/usersChats", getUsersChats)   //Получить чаты пользователя
	http.HandleFunc("/chatUsers", getUsersOfChat)   //Получить пользователей чата
	http.HandleFunc("/chat", getChatLite)           //Получить информацию чата
	http.HandleFunc("/messages", getMessages)       //Получить сообщения чата
	http.HandleFunc("/newMessages", getNewMessages) //Получить новые сообщения чата
	http.HandleFunc("/chatKey", getChatKey)         //Получить ключ чата
	http.HandleFunc("/user", getUser)               //Получить пользователя
	http.HandleFunc("/ws", webSocket)               //Подключиться по вебсокету
	//TODO: гет-ручка обновления токена

	//POST Ручки
	http.HandleFunc("/authorise", authoriseUser)    //Авторизовать
	http.HandleFunc("/exit", exit)                  //Выйти
	http.HandleFunc("/registration", registration)  //Выйти
	http.HandleFunc("/verifyToken", verifyTokenReq) //Перепроверить токен
	http.HandleFunc("/sendMessage", sendMessage)    //Отправить сообщение
	http.HandleFunc("/createChat", createChat)      //Создать чат

	log.Print(" Starting server\n")
	log.Print(" Server started\n")
	log.Fatal(http.ListenAndServe(":"+port, nil)) //Запускаем сервер, оборачиваем в логирование чтоб видеть результат
	log.Print(" Server finished\n")
}
