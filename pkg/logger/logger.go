package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

func New(debug bool) *Logger {
	var debugOutput io.Writer = io.Discard
	if debug {
		debugOutput = os.Stdout
	}

	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.errorLogger.Fatal(v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}
