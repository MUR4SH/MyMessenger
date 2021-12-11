package structures

type Chat struct {
	Id             string `bson:"_id"`
	Users_array    []string
	Messages_array []string
	Files_array    []string
	Options        struct {
		Chat_name     string
		Chat_logo     string
		Visability    string
		Hide_users    bool
		Invites_only  bool
		Security_keys []struct {
			User_id     string
			Public_key  string
			Private_key string
		}
	}
	Admins_array  []string
	Invited_array []string
	Banned_array  []string
}

type Files struct {
	Id         string `bson:"_id"`
	Name       string
	Type       string
	Gtm_date   string
	Message_id string
	Url        string
}

type User struct {
	Id          string `bson:"_id"`
	Login       string
	Password    *string
	Email       *string
	Phone       *string
	Chats_array []*struct {
		Chat_id      string
		Muted        bool
		Notification string
	}
	Photos_array []string
	Status       string
	About        string
	Keys_array   []*struct {
		Chat_id     string
		Public_ley  string
		Private_key string
	}
	Devices_array []*string
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
	Hidden_login   bool
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
	Hidden_login   bool
}

type ID struct {
	Id string `bson:"_id"`
}

type TokenJson struct {
	Token string `json:"token"`
}

type ChatJSON struct {
	Token          string   `json:"token"`
	Id             string   `json:"id"`
	Users_array    []string `json:"users_array"`
	Messages_array []string `json:"messages_array"`
	Files_array    []string `json:"files_array"`
	Options        struct {
		Chat_name    string `json:"chat_name"`
		Chat_logo    string `json:"chat_logo"`
		Muted        bool   `json:"muted"`
		Hide_users   bool   `json:"hide_users"`
		Invites_only bool   `json:"invites_only"`
	} `json:"options"`
	Admins_array  []string `json:"admins_array"`
	Invited_array []string `json:"invited_array"`
	Banned_array  []string `json:"banned_array"`
}

type FilesJSON struct {
	Token string `json:"token"`
	Id    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Url   string `json:"url"`
}

type UserAuthorise struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserJSON struct {
	Token       string `json:"token"`
	Id          string `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Chats_array []struct {
		Id           string `json:"string"`
		Muted        bool   `json:"muted"`
		Notification string `json:"notification"`
	} `json:"chats_array"`
	Photos_array []string `json:"photos_array"`
	Status       string   `json:"status"`
	About        string   `json:"about"`
	Keys_array   []struct {
		Chat_id     string `json:"chat_id"`
		Private_key string `json:"private_key"`
		Public_key  string `json:"public_key"`
	} `json:"keys_array"`
	Devices_array []string `json:"devices_array"`
}

type MessageJSON struct {
	Token          string   `json:"token"`
	Id             string   `json:"id"`
	Gtm_date       string   `json:"gtm_date"`
	User_id        string   `json:"user_id"`
	Text           string   `json:"text"`
	Files_array    []string `json:"files_array"`
	Resend_array   []string `json:"resend_array"`
	Replied_id     string   `json:"replied_id"`
	Comments_array []string `json:"comments_array"`
	Chat_id        string   `json:"chat_id"`
	Hidden_login   string   `json:"hidden_login"`
}
