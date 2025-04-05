// package loggin provides setup for file-based logging with rotation
package logging

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogging() {
	logDirectory := "logs"
	logFilename := "ping_app.log"
	logMaxSizeMB := 10
	logMaxBackups := 3
	logMaxAgeDays := 28
	compressRotated := true

	err := os.MkdirAll(logDirectory, 0750)
	if err != nil {
		log.Fatalf("FATAL: Failed to create log directory %s: %v", logDirectory, err)
	}

	logFilePath := filepath.Join(logDirectory, logFilename)

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    logMaxSizeMB, // megabytes
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAgeDays, // days
		Compress:   compressRotated, // Enable compression
		LocalTime:  true, // Use local time zone for timestamps in backup filenames
	}

	log.SetOutput(lumberjackLogger)

	// Optional: Configure standard log flags (add date, time, file/line number)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	log.Println("INFO: Logging configured successfully. Output directed to:", logFilePath)

}