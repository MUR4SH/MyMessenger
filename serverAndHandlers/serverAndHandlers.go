package serverAndHandlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/MUR4SH/MyMessenger/databaseInterface"
	"github.com/MUR4SH/MyMessenger/structures"
)

//Карта авторизованных пользователей строка - токен, значение - id
var users map[string]string
var dbInterface *databaseInterface.DatabaseInterface

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Max-Age", "1000")
	(*w).Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
}

//Функция шифрования пароля
func GetSHA256Hash(text string) string {
	sha := sha256.New()
	sha.Write([]byte(text))
	return hex.EncodeToString(sha.Sum(nil))
}

//Функция проверки токена
func verifyToken(m string) bool {
	if m != "" {
		token := m
		if _, ok := users[token]; ok {
			return true
		}
	}
	return false
}

//Получаем пользователей чата
func getUsersOfChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting users of chat\n")

	enableCors(&w)
	if !r.URL.Query().Has("token") || !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") || !r.URL.Query().Has("chat_id") {
		w.WriteHeader(400)
		fmt.Fprintf(w, "400")
		return
	}

	if !verifyToken(r.URL.Query().Get("token")) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "limit error")
		return
	}

	arr, err := dbInterface.GetUsersOfChat(r.URL.Query().Get("chat_id"), limit, offset)
	b, _ := json.Marshal(arr.Users_array)
	w.WriteHeader(200)
	fmt.Fprintf(w, string(b))
}

//Получаем чаты пользователя
func getUsersChats(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting chats of user\n")

	enableCors(&w)
	if !r.URL.Query().Has("token") || !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") {
		w.WriteHeader(400)
		fmt.Fprintf(w, "400")
		return
	}

	if !verifyToken(r.URL.Query().Get("token")) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "limit error")
		return
	}

	arr, err := dbInterface.GetUsersChats(users[r.URL.Query().Get("token")], limit, offset)
	b, _ := json.Marshal(arr.Chats_array)
	w.WriteHeader(200)
	fmt.Fprintf(w, string(b))
}

//Получить чат
func getChat(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting chat info\n")

	enableCors(&w)
	if !r.URL.Query().Has("token") || !r.URL.Query().Has("chat_id") {
		w.WriteHeader(400)
		fmt.Fprintf(w, "400")
		return
	}

	if !verifyToken(r.URL.Query().Get("token")) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	arr, err := dbInterface.GetChat(r.URL.Query().Get("chat_id"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, err.Error())
		return
	}
	arr.Options.Security_keys = nil
	arr.Banned_array = nil
	arr.Invited_array = nil
	b, _ := json.Marshal(arr)
	w.WriteHeader(200)
	fmt.Fprintf(w, string(b))
}

//Получаем сообщения чата
func getMessages(w http.ResponseWriter, r *http.Request) {
	log.Print(" Getting messages of chat\n")

	enableCors(&w)
	if !r.URL.Query().Has("token") || !r.URL.Query().Has("limit") || !r.URL.Query().Has("offset") || !r.URL.Query().Has("chat_id") {
		w.WriteHeader(400)
		fmt.Fprintf(w, "400")
		return
	}

	if !verifyToken(r.URL.Query().Get("token")) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "offset error")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, "limit error")
		return
	}

	arr, err := dbInterface.GetMessages(r.URL.Query().Get("chat_id"), limit, offset)
	if err != nil {
		w.WriteHeader(200)
		fmt.Fprintf(w, err.Error())
		return
	}
	b, _ := json.Marshal(arr)
	w.WriteHeader(200)
	fmt.Fprintf(w, string(b))
}

//Генерация токена
func generateString() string {
	str := ""
	bl := true
	for bl {
		charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP1234567890"
		length := rand.Intn(20) + 20
		for i := 0; i < length; i++ {
			random := rand.Intn(len(charSet))
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
	if (*r).Method == "OPTIONS" {
		return
	}
	var m structures.UserAuthorise
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	// Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}
	id, err := dbInterface.Authorise(m.Login, GetSHA256Hash(m.Password))

	if err != nil {
		w.WriteHeader(400)
		http.Error(w, err.Error(), 400)
		return
	}

	/*
		for k, e := range users {
			if e == id {
				delete(users, k)
				break
			}
		}
	*/

	token := generateString()
	var t structures.TokenJson
	t.Token = token
	b, _ = json.Marshal(t) //Делаем json ответ с токеном
	users[token] = id
	w.WriteHeader(200)
	fmt.Fprintf(w, string(b))
}

func exit(w http.ResponseWriter, r *http.Request) {
	log.Print(" Exiting\n")

	enableCors(&w)

	var m structures.TokenJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	if !verifyToken(m.Token) {
		w.WriteHeader(200)
		fmt.Fprintf(w, "token not found")
		return
	}

	for k, e := range users {
		if e == m.Token {
			delete(users, k)
			break
		}
	}

	w.WriteHeader(200)
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
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	if !verifyToken(m.Token) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
	} else {
		w.WriteHeader(200)
		fmt.Fprintf(w, "200")
	}
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	log.Print(" Sending message\n")

	enableCors(&w)

	var m structures.MessageJSON
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	//Unmarshal
	err = json.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, err.Error(), 404)
		return
	}

	if !verifyToken(m.Token) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	res, err := dbInterface.SendMessage(m.Chat_id, users[m.Token], m.Text)
	if err != nil && !res {
		w.WriteHeader(200)
		fmt.Fprintf(w, err.Error())
		return
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "success")
}

//Получаем порт и интерфейс для работы с бд
func InitServer(port string, db *databaseInterface.DatabaseInterface) {
	rand.Seed(time.Now().Unix())
	dbInterface = db
	users = make(map[string]string)

	//GET Ручки
	http.HandleFunc("/usersChats", getUsersChats)   //Получить чаты пользователя
	http.HandleFunc("/usersOfChat", getUsersOfChat) //Получить пользователей чата
	http.HandleFunc("/сhat", getChat)               //Получить чат
	http.HandleFunc("/chatsMessages", getMessages)  //Получить сообщения чата
	//http.HandleFunc("/users", getUsers)         //Получить пользователей

	//POST Ручки
	http.HandleFunc("/authorise", authoriseUser)     //Авторизовать
	http.HandleFunc("/exit", exit)                   //Выйти
	http.HandleFunc("/tokenVerify", verifyTokenFunc) //Перепроверить токен
	http.HandleFunc("/sendMessage", sendMessage)     //Перепроверить токен

	log.Print(" Starting server\n")
	log.Print(" Server started\n")
	log.Fatal(http.ListenAndServe(":"+port, nil)) //Запускаем сервер, оборачиваем в логирование чтоб видеть результат
	log.Print(" Server finished\n")
}
