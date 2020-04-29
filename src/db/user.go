package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/pquerna/otp/totp"
)

type User struct {
	ID      int
	Name    string
	Login   string
	OTPKey  string
	Client  int
	Memo    string
	Created time.Time
	Updated time.Time
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

func CheckLogin(login, otp string) error {
	u, err := findUser(login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidOTP
		}
		return err
	}
	if totp.Validate(otp, u.OTPKey) {
		return nil
	}
	return ErrInvalidOTP
}
