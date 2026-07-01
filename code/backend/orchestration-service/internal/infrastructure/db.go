package infrastructure

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLConnection 打开一个到 infra/docker-compose.yml 里 MySQL 服务的连接池。
// dsn 形如 "root:agentmesh@tcp(127.0.0.1:3306)/agentmesh?parseTime=true"。
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

// rowScanner 抽象 *sql.Row 和 *sql.Rows 共有的 Scan 方法，
// 让单行/多行查询可以复用同一个字段映射函数。
type rowScanner interface {
	Scan(dest ...any) error
}
