package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupSchema() {
	db.MustExec(`CREATE TABLE IF NOT EXISTS "user" ( -- 系统用户表
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL,                      -- 姓名
        "login"  TEXT NOT NULL UNIQUE,               -- 登录标识（如email、手机号或用户名）
        "otpkey" TEXT NOT NULL DEFAULT "",           -- OTP密钥（只有主账户有密钥，可以登录）
        "client" INTEGER NOT NULL DEFAULT 0,         -- 0表示主账户，非0表示那个主账户的客户
        "memo"   TEXT NOT NULL DEFAULT ""            -- 备注，可用于保存地址等信息     
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "herb" ( -- 商品表
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "name"   TEXT NOT NULL UNIQUE,               -- 品名
        "pinyin" TEXT NOT NULL DEFAULT "",           -- 拼音首字母索引
        "alias"  INTEGER NOT NULL DEFAULT 0,         -- 别名ID索引（0表示没有别名）
        "stock"  INTEGER NOT NULL DEFAULT 0,         -- 存货数量（如果alias不为0，该值一定为0）
        "unit"   TEXT NOT NULL,                      -- 存货计量单位
        "cost"   NUMERIC NOT NULL DEFAULT 0          -- 平均成本单价
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "sku" (  -- 库存单元表
        "caption" TEXT NOT NULL,                     -- 单元名称
        "base"    TEXT NOT NULL DEFAULT "",          -- 关联基本单元（为空表示该单元即基本单元）
        "count"   INTEGER NOT NULL DEFAULT 1,        -- 包含的基本单元数量
        PRIMARY KEY("caption")
    ) WITHOUT ROWID`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom_type" ( -- 单据类型表
        "id"      INTEGER,
        "caption" TEXT NOT NULL UNIQUE,                  -- 单据类型名称
        "class"   INTEGER NOT NULL,                      -- 单据性质（0=入库；1=出库）
        PRIMARY KEY("id")
    ) WITHOUT ROWID`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom" (  -- 单据表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "type"    INTEGER NOT NULL,                  -- 单据类型
        "user_id" INTEGER NOT NULL,                  -- 操作用户ID
        "amount"  NUMERIC NOT NULL DEFAULT 0,        -- 总金额
        "markup"  TEXT NOT NULL DEFAULT 0,           -- 定价策略描述
        "fee"     NUMERIC NOT NULL DEFAULT 0,        -- 额外费用（不含在总金额内，如运费）
        "memo"    TEXT NOT NULL DEFAULT "",          -- 备注
        "status"  INTEGER NOT NULL DEFAULT 0,        -- 状态（0表示未完成；1表示完成）
        "created" INTEGER NOT NULL,                  -- 创建时间戳
        "updated" INTEGER NOT NULL                   -- 最后更新时间戳
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "bom_item" ( -- 单据条目表
        "id"      INTEGER PRIMARY KEY AUTOINCREMENT,
        "bom_id"  INTEGER NOT NULL,                      -- 单据ID
        "herb_id" INTEGER NOT NULL,                      -- 商品ID
        "unit"    TEXT NOT NULL,                         -- 计量单位
        "count"   INTEGER NOT NULL,                      -- 数量
        "status"  INTEGER NOT NULL                       -- 状态（0表示出入库未完成；1表示完成）
    )`)
	db.MustExec(`CREATE TABLE IF NOT EXISTS "markup" ( -- 定价策略表
        "id"     INTEGER PRIMARY KEY AUTOINCREMENT,
        "multi"  NUMERIC NOT NULL DEFAULT 1,           -- 成本的倍数
        "plus"   NUMERIC NOT NULL DEFAULT 0            -- 加固定金额
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