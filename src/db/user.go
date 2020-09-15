package db

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func uuid(L int) string {
	cs := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, L)
	rand.Read(buf)
	for i := 0; i < L; i++ {
		buf[i] = cs[buf[i]%62]
	}
	return string(buf)
}

type User struct {
	ID      int        `json:"id"`
	Name    string     `json:"name,omitempty"`
	Login   string     `json:"login,omitempty"`
	Markup  float64    `json:"markup"`
	OTPKey  string     `json:"-"`
	Client  int        `json:"client,omitempty"`
	Memo    string     `json:"memo,omitempty"`
	Paid    float64    `json:"paid"`
	Due     float64    `json:"due"`
	Created *time.Time `json:"created,omitempty"`
	Updated *time.Time `json:"updated,omitempty"`
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

func Login(login, code string) (string, error) {
	u, err := findUser(login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidOTP
		}
		return "", err
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
		return "", err
	}
	tok := uuid(16)
	_, err = db.Exec(`INSERT INTO access (tok,uid,upd) VALUES (?,?,?)`,
		tok, u.ID, time.Now())
	return tok, err
}

func Logout(token string) error {
	_, err := db.Exec(`DELETE FROM access WHERE tok=?`, token)
	return err
}

func CheckToken(token string) int {
	var uid int
	var upd time.Time
	r := db.QueryRow(`SELECT uid,upd FROM access WHERE tok=?`, token)
	err := r.Scan(&uid, &upd)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	if time.Since(upd).Hours() > 168 { //超过7天没有使用
		return 0
	}
	return uid
}

func GetUser(id interface{}) *User {
	var u User
	assert(db.Get(&u, `SELECT * FROM user WHERE id=?`, id))
	u.OTPKey = ""
	return &u
}

func ListUsers(account int, props ...string) (users []User) {
	qry := `SELECT %s FROM user`
	acc := len(props) == 0
	if acc {
		qry = fmt.Sprintf(qry, "*")
	} else {
		fields := strings.Join(props, ",")
		qry = fmt.Sprintf(qry, fields)
	}
	if account != 1 {
		qry += fmt.Sprintf(` WHERE client=%d OR id=%d`, account, account)
	}
	assert(db.Select(&users, qry))
	if !acc {
		for _, p := range props {
			acc = p == "paid" || p == "due"
			if acc {
				break
			}
		}
	}
	if acc {
		var account []struct {
			UserID int     `db:"user_id"`
			Status int     `db:"status"`
			Total  float64 `db:"total"`
		}
		type balance struct {
			paid float64
			due  float64
		}
		ab := make(map[int]balance)
		db.Select(&account, `SELECT user_id,status,sum(total) AS total FROM
			(SELECT user_id,bom_id,-sum(cost*confirm)*sets*(1+markup*1.0/100)+fee
			AS total,status FROM bom_item JOIN bom ON bom.id=bom_id WHERE bom_id
			IN (SELECT id FROM bom WHERE type=2 AND status>0) GROUP BY bom_id) 
			GROUP BY user_id,status ORDER BY user_id`)
		for _, a := range account {
			b := ab[a.UserID]
			switch a.Status {
			case 1:
				b.due = a.Total
			case 2:
				b.paid = a.Total
			}
			ab[a.UserID] = b
		}
		for i, u := range users {
			b := ab[u.ID]
			users[i].Paid = b.paid
			users[i].Due = b.due
		}
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
	if u.ID == 0 {
		res := db.MustExec(`INSERT INTO user (name,login,client,memo) VALUES
			(?,?,?,?)`, u.Name, u.Login, u.Client, u.Memo)
		id, _ := res.LastInsertId()
		u.ID = int(id)
	} else {
		if u.ID == 1 {
			u.Name = "管理员"
			u.Login = "admin"
		}
		cmd := `UPDATE user SET name=?,login=?,markup=?,memo=?`
		args := []interface{}{u.Name, u.Login, u.Markup, u.Memo}
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

func SetAccessToken(uid int, tok string, upd time.Time) {
	if tok != "" {
		tok = fmt.Sprintf("%d,%s", upd.Unix(), tok)
	}
	db.MustExec(`UPDATE user SET session=? WHERE id=?`, tok, uid)
}
