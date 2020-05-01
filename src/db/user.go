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

func ListUsers(account int) (users []User, err error) {
	qry := `SELECT * FROM user`
	if account != 1 {
		qry += fmt.Sprintf(` WHERE client=%d`, account)
	}
	err = db.Select(&users, qry)
	return
}
