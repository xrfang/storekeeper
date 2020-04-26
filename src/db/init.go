package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupSchema() {
	db.MustExec(`CREATE TABLE IF NOT EXISTS "user" (
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL,
        "login"  TEXT NOT NULL UNIQUE,
        "otpkey" TEXT NOT NULL DEFAULT ""
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "herb" (
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL UNIQUE,
        "pinyin" TEXT NOT NULL DEFAULT "",
        "alias"  INTEGER NOT NULL DEFAULT 0,
        "stock"  INTEGER NOT NULL DEFAULT 0,
        "unit"   TEXT NOT NULL,
        "cost"   NUMERIC NOT NULL DEFAULT 0
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "sku" (
        "caption" TEXT NOT NULL,
        "base"    TEXT NOT NULL DEFAULT "",
        "count"   INTEGER NOT NULL DEFAULT 1,
        PRIMARY KEY("caption")
    ) WITHOUT ROWID`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom_type" (
        "id"      INTEGER,
        "caption" TEXT NOT NULL UNIQUE,
        "class"   INTEGER NOT NULL,
        PRIMARY KEY("id")
    ) WITHOUT ROWID`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom" (
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "type"    INTEGER NOT NULL,
        "user_id" INTEGER NOT NULL,
        "money"   NUMERIC NOT NULL DEFAULT 0,
        "memo"    TEXT NOT NULL DEFAULT "",
        "status"  INTEGER NOT NULL DEFAULT 0,
        "created" INTEGER NOT NULL,
        "updated" INTEGER NOT NULL
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom_item" (
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "bom_id"  INTEGER NOT NULL,
        "herb_id" INTEGER NOT NULL,
        "unit"    TEXT NOT NULL,
        "count"   INTEGER NOT NULL,
        "status"  INTEGER NOT NULL
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
