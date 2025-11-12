package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitDB 初始化数据库连接并创建表
func InitDB(dbPath string) (*sql.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 连接数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

// createTables 创建所需的数据库表
func createTables(db *sql.DB) error {
	// 文件元数据表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS file_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resource_type VARCHAR(10) NOT NULL,      -- 'image' 或 'file'
			resource_key VARCHAR(255) UNIQUE NOT NULL, -- image_key 或 file_key
			original_name VARCHAR(255),
			file_size INTEGER,
			md5_hash VARCHAR(32),
			sha256_hash VARCHAR(64),
			upload_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			upload_user_id VARCHAR(64),
			feishu_api_response TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_resource_key ON file_metadata(resource_key);
		CREATE INDEX IF NOT EXISTS idx_md5_hash ON file_metadata(md5_hash);
		CREATE INDEX IF NOT EXISTS idx_sha256_hash ON file_metadata(sha256_hash);
		CREATE INDEX IF NOT EXISTS idx_upload_time ON file_metadata(upload_time);
	`)
	if err != nil {
		return err
	}

	// 用户查询记录表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_search_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone_number VARCHAR(20) NOT NULL,
			search_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			result_count INTEGER,
			user_ids TEXT,  -- JSON格式存储查询结果
			search_result TEXT  -- 完整的搜索结果
		);
		CREATE INDEX IF NOT EXISTS idx_phone_number ON user_search_logs(phone_number);
		CREATE INDEX IF NOT EXISTS idx_search_time ON user_search_logs(search_time);
	`)

	log.Println("Database tables created successfully")
	return err
}