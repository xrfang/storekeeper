package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp"
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

func UpdateOTPKey(login, key string) {
	u, err := findUser(login)
	assert(err)
	res, err := db.Exec(`UPDATE "user" SET "otpkey"=? WHERE "id"=?`, key, u.ID)
	assert(err)
	ra, _ := res.RowsAffected()
	if ra == 0 {
		panic(fmt.Errorf("set otpkey for %s(%d) failed", login, u.ID))
	}
}

func CheckLogin(login, code string) (int, error) {
	u, err := findUser(login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidOTP
		}
		return 0, err
	}
	ok, err := totp.ValidateCustom(code, u.OTPKey, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: otp.AlgorithmSHA256,
	})
	if !ok {
		if err == nil {
			err = ErrInvalidOTP
		}
		return 0, err
	}
	return u.ID, nil
}

func GetUser(id interface{}) *User {
	var u User
	assert(db.Get(&u, `SELECT * FROM user WHERE id=?`, id))
	u.OTPKey = ""
	return &u
}

func ListUsers(account int) (users []User) {
	qry := `SELECT * FROM user`
	if account != 1 {
		qry += fmt.Sprintf(` WHERE client=%d OR id=%d`, account, account)
	}
	assert(db.Select(&users, qry))
	for i := range users {
		users[i].OTPKey = ""
	}
	return
}

func GetPrimaryUsers() []User {
	var us []User
	assert(db.Select(&us, `SELECT id,name from user WHERE id>1 AND client=0`))
	for i := range us {
		us[i].OTPKey = ""
	}
	return us
}

func UpdateUser(u *User) bool {
	if u.ID == 1 {
		return false //不允许更改1号用户（admin）的信息
	}
	if u.ID == 0 {
		res := db.MustExec(`INSERT INTO user (name,login,client,memo) VALUES
			(?,?,?,?)`, u.Name, u.Login, u.Client, u.Memo)
		id, _ := res.LastInsertId()
		u.ID = int(id)
	} else {
		cmd := `UPDATE user SET name=?,login=?,memo=?`
		args := []interface{}{u.Name, u.Login, u.Memo}
		cmd += ` WHERE id=?`
		args = append(args, u.ID)
		db.MustExec(cmd, args...)
	}
	return true
}

func DeleteUser(id int) {
	//TODO: 添加外部制约，凡是有交易记录的不可删除
	_, err := db.Exec(`DELETE FROM user WHERE id=?`, id)
	assert(err)
}
