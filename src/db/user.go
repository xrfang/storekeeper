package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
)

type User struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Login   string    `json:"login"`
	OTPKey  string    `json:"-"`
	Client  int       `json:"client"`
	Memo    string    `json:"memo"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

var ErrInvalidOTP = errors.New("invalid otp password")

func findUser(login string) (*User, error) {
	var u User
	err := db.Get(&u, `SELECT * FROM "user" WHERE "login"=?`, login)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateOTPKey(login, key string) error {
	u, err := findUser(login)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE "user" SET "otpkey"=? WHERE "id"=?`, key, u.ID)
	return err
}

func CheckLogin(login, otp string) (int, error) {
	u, err := findUser(login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidOTP
		}
		return 0, err
	}
	if totp.Validate(otp, u.OTPKey) {
		return u.ID, nil
	}
	return 0, ErrInvalidOTP
}

func GetUser(id int) (*User, error) {
	var u User
	err := db.Get(&u, `SELECT * FROM user WHERE id=?`, id)
	return &u, err
}

func ListUsers(account int) (users []User, err error) {
	qry := `SELECT * FROM user`
	if account != 1 {
		qry += fmt.Sprintf(` WHERE client=%d`, account)
	}
	err = db.Select(&users, qry)
	return
}

func GetPrimaryUsers() ([]User, error) {
	var us []User
	err := db.Select(&us, `SELECT id,name from user WHERE id>1 AND client=0`)
	return us, err
}

func UpdateUser(u *User) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if u.ID == 1 {
		return //不允许更改1号用户（admin）的信息
	}
	if u.ID == 0 {
		db.MustExec(`INSERT INTO user (name,login,client,memo) VALUES
		    (?,?,?,?)`, u.Name, u.Login, u.Client, u.Memo)
	} else {
		cmd := `UPDATE user SET name=?,login=?,memo=?`
		args := []interface{}{u.Name, u.Login, u.Memo}
		if u.Client >= 0 {
			cmd += `,client=?`
			args = append(args, u.Client)
		}
		cmd += ` WHERE id=?`
		args = append(args, u.ID)
		db.MustExec(cmd, args...)
	}
	return
}
