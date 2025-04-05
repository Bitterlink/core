// Package db Package for the database
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const MaxOpenMySQLConnections = 25
const MaxIdleMySQLConnections = MaxOpenMySQLConnections
const MySQLConnectionMaxLifetime = 10 * time.Minute

func ConnectDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// --- Provide Defaults (Optional, useful for local dev) ---
	if dbUser == "" {
		dbUser = "admin" // Replace with your local user if needed
		log.Println("WARN: DB_USER not set, using default 'admin'")
	}
	if dbPassword == "" {
		dbPassword = "a"
		log.Println("WARN: DB_PASSWORD not set, using default (CHANGE THIS)")
	}
	if dbHost == "" {
		dbHost = "127.0.0.1" 
		log.Println("WARN: DB_HOST not set, using default '127.0.0.1'")
	}
	if dbPort == "" {
		dbPort = "3306"
		log.Println("WARN: DB_PORT not set, using default '3306'")
	}
	if dbName == "" {
		dbName = "ping" 
		log.Println("WARN: DB_NAME not set, using default 'ping'")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	dbPool, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("ERROR: Failed to prepare database connection pool: %v", err)
		return nil, fmt.Errorf("failed to prepare database connection pool: %w", err)
	}

	dbPool.SetMaxOpenConns(MaxOpenMySQLConnections)
	dbPool.SetMaxIdleConns(MaxIdleMySQLConnections)
	dbPool.SetConnMaxLifetime(MySQLConnectionMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = dbPool.PingContext(ctx)
	if err != nil {
		err := dbPool.Close()
		if err != nil {
			return nil, err
		}
		log.Printf("ERROR: Failed to connect to database: %v", err)
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	log.Println("INFO: Database connection pool established successfully.")
	return dbPool, nil
}