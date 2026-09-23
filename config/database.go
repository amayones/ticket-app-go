package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func ConnectDatabase() *sql.DB {
	host := GetEnv("DB_HOST")
	port := GetEnv("DB_PORT")
	database := GetEnv("DB_DATABASE")
	username := GetEnv("DB_USERNAME")
	password := GetEnv("DB_PASSWORD")
	connectionString := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		username,
		password,
		host,
		port,
		database,
	)
	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
	fmt.Println("Database connected successfully!")
	return db
}
