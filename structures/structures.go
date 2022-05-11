package structures

import "crypto/rsa"

type Chat struct {
	Id             string `bson:"_id"`
	Chat_name      string
	Chat_logo      string
	Users_array    []string
	Messages_array []string
	Files_array    []string
	Options        string
	Admins_array   []string
	Invited_array  []string
	Banned_array   []string
	Key            *rsa.PublicKey
}

type Chat_settings struct {
	Id                     string `bson:"_id"`
	Secured                bool
	Search_visible         bool
	Resend                 bool
	Users_write_permission bool
	Personal               bool
}

type Files struct {
	Id         string `bson:"_id"`
	Name       string
	Type       string
	Gtm_date   string
	ExpiredAt  string
	Message_id *string
	Url        string
}

type UserId struct {
	Id string `bson:"_id"`
}

type User struct {
	Id                string `bson:"_id"`
	Login             string
	Password          *string
	Email             *string
	Phone             *string
	Chats_array       []*string
	Photos_array      []*string
	Status            string
	About             string
	Personal_settings string
}

type Chats_array struct {
	Id                   string `bson:"_id"`
	Chat_id              string
	Notifications        bool
	Key                  *rsa.PrivateKey
	Last_messages_number int
}

type Personal_settings struct {
	Id            string `bson:"_id"`
	User_id       string
	Phone_visible bool
	Email_visible bool
}

type Message struct {
	Id             string `bson:"_id"`
	Gtm_date       string
	User_id        string
	Text           string
	Files_array    []string
	Resend_array   []string
	Replied_id     string
	Comments_array []string
	Chat_id        string
	ExpiredAt      string
}

type MessageInsert struct {
	Gtm_date       string
	User_id        string
	Text           string
	Files_array    []string
	Resend_array   []string
	Replied_id     string
	Comments_array []string
	Chat_id        string
}

type ID struct {
	Id string `bson:"_id"`
}

type TokenJson struct {
	Token string `json:"token"`
}

type ChatJSON struct {
	Id             string        `json:"id"`
	Chat_name      string        `json:"chat_name"`
	Chat_logo      string        `json:"chat_logo"`
	Users_array    []string      `json:"users_array"`
	Messages_array []string      `json:"messages_array"`
	Files_array    []string      `json:"files_array"`
	Options        string        `json:"options"`
	Admins_array   []string      `json:"admins_array"`
	Invited_array  []string      `json:"invited_array"`
	Banned_array   []string      `json:"banned_array"`
	Key            rsa.PublicKey `json:"key"`
}

type FilesJSON struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

type UserAuthorise struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenStore struct {
	Id   string
	Date string
}

type UserJSON struct {
	Id                string   `json:"id"`
	Login             string   `json:"login"`
	Password          string   `json:"password"`
	Email             string   `json:"email"`
	Phone             string   `json:"phone"`
	Chats_array       []string `json:"chats_array"`
	Photos_array      []string `json:"photos_array"`
	Status            string   `json:"status"`
	About             string   `json:"about"`
	Personal_settings string   `json:"personal_settings"`
}

type MessageJSON struct {
	Id             string   `json:"id"`
	Gtm_date       string   `json:"gtm_date"`
	User_id        string   `json:"user_id"`
	Text           string   `json:"text"`
	Files_array    []string `json:"files_array"`
	Resend_array   []string `json:"resend_array"`
	Replied_id     string   `json:"replied_id"`
	Comments_array []string `json:"comments_array"`
	Chat_id        string   `json:"chat_id"`
	ExpiredAt      string   `json:"expired_at"`
}

type ChatCreationJSON struct {
	User_id                string   `json:"user_id"`
	Name                   string   `json:"name"`
	Logo                   []byte   `json:"logo"`
	Users                  []string `json:"users"`
	Secured                bool     `json:"secured"`
	Search_visible         bool     `json:"search_visible"`
	Resend                 bool     `json:"resend"`
	Users_write_permission bool     `json:"users_write_permission"`
	Personal               bool     `json:"personal"`
}
