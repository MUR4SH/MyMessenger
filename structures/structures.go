package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EditedPublicKey struct {
	N string
	E int
}

type EditedPrivateKey struct {
	PublicKey EditedPublicKey
	D         string
	Primes    []string

	Precomputed EditedPrecomputedValues
}

type EditedPrecomputedValues struct {
	Dp, Dq    string
	Qinv      string
	CRTValues []EditedCRTValue
}

type EditedCRTValue struct {
	Exp   string
	Coeff string
	R     string
}

type Chat struct {
	Id             primitive.ObjectID `bson:"_id"`
	Chat_name      string
	Chat_logo      primitive.ObjectID
	Users_array    []primitive.ObjectID
	Messages_array []primitive.ObjectID
	Files_array    []primitive.ObjectID
	Options        primitive.ObjectID
	Admins_array   []primitive.ObjectID
	Invited_array  []primitive.ObjectID
	Banned_array   []primitive.ObjectID
	Key            []byte
}

type Chat_noid struct {
	Chat_name      string
	Chat_logo      primitive.ObjectID
	Users_array    []primitive.ObjectID
	Messages_array []primitive.ObjectID
	Files_array    []primitive.ObjectID
	Options        primitive.ObjectID
	Admins_array   []primitive.ObjectID
	Invited_array  []primitive.ObjectID
	Banned_array   []primitive.ObjectID
	Key            []byte
}

type Chat_lite struct {
	Id          primitive.ObjectID `bson:"_id"`
	Chat_name   string
	Chat_logo   string
	Users_count int64
	Options     string
}

type Chat_settings struct {
	Id                     primitive.ObjectID `bson:"_id"`
	Chat_id                primitive.ObjectID
	Secured                bool
	Search_visible         bool
	Resend                 bool
	Users_write_permission bool
	Personal               bool
}

type Chat_settings_noid struct {
	Chat_id                primitive.ObjectID
	Secured                bool
	Search_visible         bool
	Resend                 bool
	Users_write_permission bool
	Personal               bool
}

type Files struct {
	Id         primitive.ObjectID `bson:"_id"`
	Name       string
	Type       string
	Gtm_date   string
	ExpiredAt  string
	Message_id *string
	Url        string
}

type Files_noid struct {
	Name       string
	Type       string
	Gtm_date   string
	ExpiredAt  string
	Message_id *string
	Url        string
}

type UserId struct {
	Id primitive.ObjectID `bson:"_id"`
}

type User struct {
	Id                primitive.ObjectID `bson:"_id"`
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

type Chat_User_aggregate_lite struct {
	Id          primitive.ObjectID `bson:"_id"`
	Users_array []User_lite
}

type User_lite struct {
	Id           primitive.ObjectID `bson:"_id"`
	Login        string
	Email        *string
	Phone        *string
	Photos_array []*string
	Status       string
	About        string
}

type Chats_array struct {
	Id                   primitive.ObjectID `bson:"_id"`
	Chat_id              primitive.ObjectID
	Notifications        bool
	Key                  []byte
	Personal             bool
	Last_messages_number int
}

type Chats_array_noid struct {
	Chat_id              primitive.ObjectID
	Notifications        bool
	Personal             bool
	Key                  []byte
	Last_messages_number int
}

type Chats_array_agregate struct {
	Chats_arrays []Chats_array
}

type Chat_settings_agregate struct {
	Options []Chat_settings
}

type Messages_array_agregate struct {
	Messages_array []Message
}

type Answer struct {
	Text string `json:"text"`
}

type Personal_settings struct {
	Id            primitive.ObjectID `bson:"_id"`
	User_id       string
	Phone_visible bool
	Email_visible bool
}

type Message struct {
	Id             primitive.ObjectID `bson:"_id"`
	Gtm_date       string
	User_id        primitive.ObjectID
	Text           []byte
	Files_array    []primitive.ObjectID
	Resend_array   []primitive.ObjectID
	Replied_id     primitive.ObjectID
	Comments_array []primitive.ObjectID
	Chat_id        primitive.ObjectID
	ExpiredAt      string
}

type MessageToUser struct {
	Id             primitive.ObjectID `bson:"_id"`
	Gtm_date       string
	User_id        string
	Text           []byte
	Files_array    []string
	Resend_array   []string
	Replied_id     string
	Comments_array []string
	Chat_id        string
	User           []User_lite
}

type Message_noid struct {
	Gtm_date       string
	User_id        primitive.ObjectID
	Text           []byte
	Files_array    []primitive.ObjectID
	Resend_array   []primitive.ObjectID
	Replied_id     primitive.ObjectID
	Comments_array []primitive.ObjectID
	Chat_id        primitive.ObjectID
}

type ID struct {
	Id primitive.ObjectID `bson:"_id"`
}

type TokenJson struct {
	Token string `json:"token"`
}

type ChatJSON struct {
	Id             string   `json:"id"`
	Chat_name      string   `json:"chat_name"`
	Chat_logo      string   `json:"chat_logo"`
	Users_array    []string `json:"users_array"`
	Messages_array []string `json:"messages_array"`
	Files_array    []string `json:"files_array"`
	Options        string   `json:"options"`
	Admins_array   []string `json:"admins_array"`
	Invited_array  []string `json:"invited_array"`
	Banned_array   []string `json:"banned_array"`
	Key            []byte   `json:"key"`
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

type CreateUserJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
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

type ChatIdJSON struct {
	Id string `json:"chat_id`
}

type ChatCreationJSON struct {
	Name                   string   `json:"name"`
	Logo                   []byte   `json:"logo"`
	Logo_url               *string  `json:"logo_url"`
	Users                  []string `json:"users"`
	Secured                bool     `json:"secured"`
	Search_visible         bool     `json:"search_visible"`
	Resend                 bool     `json:"resend"`
	Users_write_permission bool     `json:"users_write_permission"`
	Personal               bool     `json:"personal"`
}
