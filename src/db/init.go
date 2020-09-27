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
        "markup"  NUMERIC NOT NULL DEFAULT -1,                 -- 溢价率（-1表示使用系统默认）
        "otpkey"  TEXT NOT NULL DEFAULT "",                    -- OTP密钥（只有主账户有密钥，可以登录）
        "client"  INTEGER NOT NULL DEFAULT 0,                  -- 0表示主账户，非0表示那个主账户的客户
        "memo"    TEXT NOT NULL DEFAULT "",                    -- 备注，可用于保存地址等信息     
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 最后更新时间戳
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS usrupd AFTER UPDATE ON "user"
        FOR EACH ROW BEGIN UPDATE "user" SET updated=CURRENT_TIMESTAMP WHERE
        id=OLD.id; END`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "access" ( -- 访问控制表
        "tok"     TEXT PRIMARY KEY,                    -- 令牌
        "uid"     INTEGER NOT NULL,                    -- 用户ID
        "upd"     DATETIME NOT NULL                    -- 创建时间戳
    )`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "goods" ( -- 药材表
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL UNIQUE,               -- 品名
        "pinyin" TEXT NOT NULL DEFAULT "",           -- 拼音首字母索引
        "stock"  NUMERIC NOT NULL DEFAULT 0,         -- 存货数量
        "cost"   NUMERIC NOT NULL DEFAULT 0,         -- 平均成本单价
        "batch"  NUMERIC NOT NULL DEFAULT 500,       -- 默认采购批量（克）
        "rack"   TEXT NOT NULL DEFAULT ""            -- 货架编号
    )`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "bom" (            -- 单据表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "type"    INTEGER NOT NULL,                            -- 单据类型（1=入库；2=出库；3=盘点）
        "user_id" INTEGER NOT NULL,                            -- 操作用户ID
        "markup"  NUMERIC NOT NULL DEFAULT 0,                  -- 加价百分比
        "fee"     NUMERIC NOT NULL DEFAULT 0,                  -- 额外费用（如运费、外购费）
        "sets"    INTEGER NOT NULL DEFAULT 1,                  -- 服数
        "memo"    TEXT NOT NULL DEFAULT "",                    -- 备注
        "status"  INTEGER NOT NULL DEFAULT 0,                  -- 状态（原始状态一律为0，其他状态各单据类型自行定义）
        "paid"    NUMERIC NOT NULL DEFAULT 0,                  -- 实际付款金额
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间戳
        FOREIGN KEY(user_id) REFERENCES user(id)
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS bomupd AFTER UPDATE ON "bom"
        FOR EACH ROW BEGIN UPDATE "bom" SET updated=CURRENT_TIMESTAMP WHERE
        id=OLD.id; END`)
	tx.MustExec(`CREATE TABLE IF NOT EXISTS "bom_item" (       -- 单据条目表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "bom_id"  INTEGER NOT NULL,                            -- 单据ID
        "gid"     INTEGER NOT NULL,                            -- 药材ID
        "gname"   TEXT NOT NULL,                               -- 药材名称
        "cost"    NUMERIC NOT NULL DEFAULT 0,                  -- 单位成本
        "request" NUMERIC NOT NULL,                            -- 需求数量
        "confirm" NUMERIC NOT NULL,                            -- 确认数量
        "flag"    INTEGER NOT NULL DEFAULT 0,                  -- 标志位
        "memo"    TEXT NOT NULL DEFAULT "",                    -- 备注
        "created" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间戳
        "updated" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间戳
        FOREIGN KEY(bom_id) REFERENCES bom(id),
        FOREIGN KEY(gid) REFERENCES goods(id)
    )`)
	tx.MustExec(`CREATE TRIGGER IF NOT EXISTS bisupd AFTER UPDATE ON "bom_item"
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
