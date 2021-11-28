package databaseInterface

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type DatabaseInterface struct {
	db pg.DB
}

type Chat struct {
	Id             string
	Users_array    []int
	Messages_array []int
	Files_array    []int
	Options        struct {
		Chat_name    string
		Chat_logo    int
		Hide_users   bool
		Invites_only bool
	}
	Admins_array  []int
	Invited_array []int
	Banned_array  []int
}

type Files struct {
	Id   string
	Name string
	Type string
	Url  string
}

type User struct {
	Id           string
	Login        string
	Password     string
	Key          string
	Email        string
	Phone        int
	Chats_array  []int
	Photos_array []int
	Status       string
	About        string
	Keys_array   []struct {
		Chat_id string
		Key     string
	}
	Devices_array []string
}

type Message struct {
	Id             string
	Gtm_date       string
	User_id        string
	Text           string
	Files_array    []string
	Resend_array   []string
	Replied_id     string
	Comments_array []string
	Chat_id        string
	Hidden_login   string
}

func New(address string, user string, password string, database string) DatabaseInterface {
	db := pg.Connect(&pg.Options{
		Addr:     address,
		User:     user,
		Password: password,
		Database: database,
	})

	return DatabaseInterface{*db}
}

func (d DatabaseInterface) GetUser(limit int, offset int) []User {
	var users []User
	_, err := d.db.Query(&users, `SELECT id, first_name, last_name, patronym, username, role FROM public.user ORDER BY id LIMIT ? OFFSET ?;`, limit, offset)

	if err != nil {
		panic(err)
	}

	return users
}

func (d DatabaseInterface) VerifyLoginPass(login string, password string) int {
	var id int
	_, err := d.db.Query(&id, `SELECT id FROM public.user WHERE username=? AND password_hash=? AND role>0;`, login, password)

	if err != nil {
		panic(err)
	}

	return id
}

func (d DatabaseInterface) AddUser(id int, login string, first_name string, last_name string, patronym string, password string, code string, role int) int {
	var err error
	if (d.GetRole(id) == 1 && role < 2) || (d.GetRole(id) == 2) {
		_, err = d.db.Exec(`INSERT INTO public.user (username, first_name, last_name, patronym, password_hash, code_hash, role) VALUES (?, ?, ?, ?, ?, ?, ?);`, login, first_name, last_name, patronym, password, code, role)
	}

	if err != nil {
		return 0
	}

	return 1
}

func (d DatabaseInterface) AddCodelock(id int, name string, description string) int {
	var err error
	if d.GetRole(id) >= 1 && name != "" && description != "" {
		_, err = d.db.Exec(`INSERT INTO public.codelock (name, description) VALUES (?, ?);`, name, description)
	}
	if err != nil {
		return 0
	}

	return 1
}

func (d DatabaseInterface) GetRole(id int) int {
	var role int
	_, err := d.db.Query(&role, `SELECT role FROM public.user WHERE id=?;`, id)

	if err != nil {
		panic(err)
	}

	return role
}

func (d DatabaseInterface) DeleteUser(id int, user_id int) int {
	var err error
	var count int
	var res orm.Result
	if (d.GetRole(id) == 2) || (d.GetRole(id) == 1 && d.GetRole(user_id) <= 1) {
		res, err = d.db.Exec(`DELETE FROM public.user WHERE id=?;`, user_id)
	} else {
		return 0
	}

	if err != nil {
		panic(err)
	} else {
		count = res.RowsAffected()
		if count < 1 {
			return 0
		}
	}

	return 1
}

func (d DatabaseInterface) DeleteCodelock(id int, codelock_id int) int {
	var err error
	var count int
	var res orm.Result
	if d.GetRole(id) >= 1 {
		res, err = d.db.Exec(`DELETE FROM public.codelock WHERE id=?;`, codelock_id)
	} else {
		return 0
	}

	if err != nil {
		panic(err)
	} else {
		count = res.RowsAffected()
		if count < 1 {
			return 0
		}
	}

	return 1
}

func (d DatabaseInterface) EditCodelock(id int, codelock_id int, name string, describe string) int {
	var err error
	if d.GetRole(id) >= 1 {
		if name != "" {
			_, err = d.db.Exec(`UPDATE public.codelock SET name=? WHERE id=?;`, name, codelock_id)
		}
		if describe != "" {
			_, err = d.db.Exec(`UPDATE public.codelock SET description=? WHERE id=?;`, describe, codelock_id)
		}
	} else {
		return 0
	}

	if err != nil {
		panic(err)
	}

	return 1
}

func (d DatabaseInterface) EditUser(id int, user_id int, login string, first_name string, last_name string, patronym string, password string, code string, role int) int {
	var err error
	if (d.GetRole(id) == 2) || (d.GetRole(id) == 1 && d.GetRole(user_id) <= 1) {
		if login != "" {
			_, err = d.db.Exec(`UPDATE public.user SET username=? WHERE id=?;`, login, user_id)
		}
		if first_name != "" {
			_, err = d.db.Exec(`UPDATE public.user SET first_name=? WHERE id=?;`, first_name, user_id)
		}
		if last_name != "" {
			_, err = d.db.Exec(`UPDATE public.user SET last_name=? WHERE id=?;`, last_name, user_id)
		}
		if patronym != "" {
			_, err = d.db.Exec(`UPDATE public.user SET patronym=? WHERE id=?;`, patronym, user_id)
		}
		if password != "" {
			_, err = d.db.Exec(`UPDATE public.user SET password_hash=? WHERE id=?;`, password, user_id)
		}
		if code != "" {
			_, err = d.db.Exec(`UPDATE public.user SET code_hash=? WHERE id=?;`, code, user_id)
		}
		if role >= 0 {
			_, err = d.db.Exec(`UPDATE public.user SET role=? WHERE id=?;`, role, user_id)
		}
	} else {
		return 0
	}

	if err != nil {
		panic(err)
	}

	return 1
}
