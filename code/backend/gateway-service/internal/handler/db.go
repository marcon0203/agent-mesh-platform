package handler

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresConnection 打开一个到 infra/docker-compose.yml 里 PostgreSQL 服务的连接池，
// 供 AuthService 查询 api_key 表使用。
func NewPostgresConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}
