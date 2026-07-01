package handler

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLConnection 打开一个到 infra/docker-compose.yml 里 MySQL 服务的连接池，
// 供 AuthService 查询 api_key 表使用。
func NewMySQLConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}
