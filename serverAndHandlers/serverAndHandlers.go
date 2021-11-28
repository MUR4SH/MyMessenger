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

	"github.com/WeWillRenameIT/iot_server/databaseInterface"
)

//Карта авторизованных пользователей строка - токен, значение - id
var users map[string]int
var dbInterface *databaseInterface.DatabaseInterface

type TokenJson struct {
	Token string `json:"token"`
}

type Chat struct {
	Id             string `json: "id"`
	Users_array    []int  `json: "users_array"`
	Messages_array []int  `json: "messages_array"`
	Files_array    []int  `json: "files_array"`
	Options        struct {
		Chat_name    string `json: "chat_name"`
		Chat_logo    int    `json: "chat_logo"`
		Hide_users   bool   `json: "hide_users"`
		Invites_only bool   `json: "invites_only"`
	} `json: "options"`
	Admins_array  []int `json: "admins_array"`
	Invited_array []int `json: "invited_array"`
	Banned_array  []int `json: "banned_array"`
}

type Files struct {
	Id   string `json: "id"`
	Name string `json: "name"`
	Type string `json: "type"`
	Url  string `json: "url"`
}

type UserAuthorise struct {
	Login    string `json: "login"`
	Password string `json: "password"`
}

type User struct {
	Id           string `json: "id"`
	Login        string `json: "login"`
	Password     string `json: "password"`
	Key          string `json: "key"`
	Email        string `json: "email"`
	Phone        int    `json: "phone"`
	Chats_array  []int  `json: "chats_array"`
	Photos_array []int  `json: "photos_array"`
	Status       string `json: "status"`
	About        string `json: "about"`
	Keys_array   []struct {
		Chat_id string `json: "chat_id"`
		Key     string `json: "key"`
	} `json: "keys_array"`
	Devices_array []string `json: "devices_array"`
}

type Message struct {
	Id             string   `json: "id"`
	Gtm_date       string   `json: "gtm_date"`
	User_id        string   `json: "user_id"`
	Text           string   `json: "text"`
	Files_array    []string `json: "files_array"`
	Resend_array   []string `json: "resend_array"`
	Replied_id     string   `json: "replied_id"`
	Comments_array []string `json: "comments_array"`
	Chat_id        string   `json: "chat_id"`
	Hidden_login   string   `json: "hidden_login"`
}

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

//Функция проверки гет-токена
func verifyGetToken(r *http.Request) bool {
	r.ParseForm()
	if r.URL.Query().Has("token") {
		token := r.URL.Query().Get("token")
		if _, ok := users[token]; ok {
			return true
		}
	}
	return false
}

//Функция проверки текстового-токена
func verifyPostToken(m string) bool {
	if m != "" {
		token := m
		if _, ok := users[token]; ok {
			return true
		}
	}
	return false
}

//Получаем пользователей
func getUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if !verifyGetToken(r) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	arr := dbInterface.GetUser(limit, offset)
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
	enableCors(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	var m UserAuthorise
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
		w.WriteHeader(400)
		http.Error(w, err.Error(), 404)
		return
	}
	id := dbInterface.VerifyLoginPass(m.Login, GetSHA256Hash(m.Password))

	for k, e := range users {
		if e == id {
			delete(users, k)
			break
		}
	}

	if id > 0 {
		token := generateString()
		var t TokenJson
		t.Token = token
		b, _ := json.Marshal(t)
		users[token] = id
		w.WriteHeader(200)
		fmt.Fprintf(w, string(b))
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "")
	}
}

//Проверка токена
func verifyToken(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	var m TokenJson
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

	if !verifyPostToken(m.Token) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "500")
	} else {
		w.WriteHeader(200)
		fmt.Fprintf(w, "200")
	}
}

//Получаем порт и интерфейс для работы с бд
func InitServer(port string, db *databaseInterface.DatabaseInterface) {
	rand.Seed(time.Now().Unix())
	dbInterface = db
	users = make(map[string]int)

	//GET Ручки
	http.HandleFunc("/user", getUser) //Получить пользователей

	//POST Ручки
	http.HandleFunc("/authorise", authoriseUser)  //Авторизовать
	http.HandleFunc("/token_verify", verifyToken) //Перепроверить токен

	log.Fatal(http.ListenAndServe(":"+port, nil)) //Запускаем сервер, оборачиваем в логирование чтоб видеть результат
}
