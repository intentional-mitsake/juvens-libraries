package utils

import (
	"log"
	"os"
)

type Logger struct {
	iLog *log.Logger
	eLog *log.Logger
	wLog *log.Logger
}

func CreateLogger() *Logger {
	return &Logger{
		iLog: log.New(os.Stdout, "INFO: ", log.LstdFlags),
		eLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		wLog: log.New(os.Stdout, "WARNING: ", log.LstdFlags),
	}
}
func (l *Logger) Info(msg string) {
	l.iLog.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.eLog.Println(msg)
}
func (l *Logger) Warning(msg string) {
	l.wLog.Println(msg)
}
