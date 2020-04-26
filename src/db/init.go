package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupSchema() {
	db.MustExec(`CREATE TABLE IF NOT EXISTS "user" (
		"id"	 INTEGER PRIMARY KEY AUTOINCREMENT,
		"name"	 TEXT NOT NULL,
		"login"	 TEXT NOT NULL UNIQUE,
		"otpkey" TEXT NOT NULL DEFAULT ""
	)`)
}

func setupParams() {
	//TODO:  设置固定参数，例如bom类型等
}

func Initialize(fn string) {
	var err error
	db, err = sqlx.Connect("sqlite3", "file:"+fn+"?cache=shared")
	assert(err)
	db.SetMaxOpenConns(1)
	setupSchema()
	setupParams()
}

var db *sqlx.DB
