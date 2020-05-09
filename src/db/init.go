package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupSchema() {
	tx := db.MustBegin()
	defer func() { assert(tx.Commit()) }()
	tx.MustExec(`PRAGMA recursive_triggers=0`)
	tx.MustExec(`PRAGMA foreign_keys=ON`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "user" (           -- 系统用户表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"    TEXT NOT NULL,                               -- 姓名
        "login"   TEXT NOT NULL UNIQUE,                        -- 登录标识（如email、手机号或用户名）
        "otpkey"  TEXT NOT NULL DEFAULT "",                    -- OTP密钥（只有主账户有密钥，可以登录）
        "client"  INTEGER NOT NULL DEFAULT 0,                  -- 0表示主账户，非0表示那个主账户的客户
        "memo"    TEXT NOT NULL DEFAULT "",                    -- 备注，可用于保存地址等信息     
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 最后更新时间戳
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS updated AFTER UPDATE ON "user"
        FOR EACH ROW BEGIN UPDATE "user" SET updated=CURRENT_TIMESTAMP WHERE
        id=OLD.id; END`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "goods" ( -- 货品表
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL UNIQUE,               -- 品名
        "pinyin" TEXT NOT NULL DEFAULT "",           -- 拼音首字母索引
        "stock"  INTEGER NOT NULL DEFAULT 0,         -- 存货数量
        "cost"   NUMERIC NOT NULL DEFAULT 0          -- 平均成本单价
    )`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "bom" (            -- 单据表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "type"    INTEGER NOT NULL,                            -- 单据类型（1=入库；2=出库；3=盘点）
        "user_id" INTEGER NOT NULL,                            -- 操作用户ID
        "cost"    NUMERIC NOT NULL DEFAULT 0,                  -- 总成本
        "charge"  NUMERIC NOT NULL DEFAULT 0,                  -- 总价格
        "fee"     NUMERIC NOT NULL DEFAULT 0,                  -- 额外费用（不含在总金额内，如运费）
        "memo"    TEXT NOT NULL DEFAULT "",                    -- 备注
        "status"  INTEGER NOT NULL DEFAULT 0,                  -- 状态（1=未配齐；2=未发货；3=未收款；4=完成）
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间戳
        FOREIGN KEY(user_id) REFERENCES user(id)
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS updated AFTER UPDATE ON "bom"
        FOR EACH ROW BEGIN UPDATE "bom" SET updated=CURRENT_TIMESTAMP WHERE
        id=OLD.id; END`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "bom_item" (       -- 单据条目表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "bom_id"  INTEGER NOT NULL,                            -- 单据ID
        "gid"     INTEGER NOT NULL,                            -- 货品ID
        "gname"   TEXT NOT NULL,                               -- 货品名称
        "price"   NUMERIC NOT NULL DEFAULT 0,                  -- 单价
        "count"   INTEGER NOT NULL,                            -- 数量
        "status"  INTEGER NOT NULL DEFAULT 0,                  -- 状态（0表示出入库未完成；1表示完成）
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间戳
        FOREIGN KEY(bom_id) REFERENCES bom(id),
        FOREIGN KEY(gid) REFERENCES goods(id)
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS updated AFTER UPDATE ON "bom_item"
        FOR EACH ROW BEGIN UPDATE "bom_item" SET updated=CURRENT_TIMESTAMP WHERE
        id=OLD.id; END`)
	//添加管理员
	tx.MustExec(`INSERT OR IGNORE INTO "user" ("id","name","login")
        VALUES (1, "管理员", "admin")`)
}

func Initialize(fn string) {
	var err error
	db, err = sqlx.Connect("sqlite3", "file:"+fn+"?cache=shared")
	assert(err)
	db.SetMaxOpenConns(1)
	setupSchema()
}

var db *sqlx.DB
