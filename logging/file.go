package logging

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

type FileLogger struct {
	LogFd    *lumberjack.Logger
	MinLevel LogLevel
}

type FileLoggerOptions struct {
	AbsPath    string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

func NewFileLogger(absPath string) *FileLogger {
	return &FileLogger{
		LogFd: &lumberjack.Logger{
			Filename:   absPath,
			MaxSize:    50, // megabytes
			MaxBackups: 10,
			MaxAge:     28, //days
		},
	}
}

func NewFileLoggerWithOptions(options *FileLoggerOptions) *FileLogger {
	maxSize := 50
	if options.MaxSize > 0 {
		maxSize = options.MaxSize
	}

	maxAge := 28
	if options.MaxAge > 0 {
		maxAge = options.MaxAge
	}

	maxBackups := 3
	if options.MaxBackups > -1 {
		maxBackups = options.MaxBackups
	}

	return &FileLogger{
		LogFd: &lumberjack.Logger{
			Filename:   options.AbsPath,
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
		},
	}
}

func (f *FileLogger) Logf(level LogLevel, format string, args ...interface{}) {
	f.Log(level, fmt.Sprintf(format, args...))
}

func (f *FileLogger) Log(level LogLevel, message string) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	var err error
	if level == Error {
		_, err = f.LogFd.Write([]byte("Error - "))
	} else if level == Warn {
		_, err = f.LogFd.Write([]byte("Warn - "))
	} else if level == Info {
		_, err = f.LogFd.Write([]byte("Info - "))
	} else if level == Debug {
		_, err = f.LogFd.Write([]byte("Debug - "))
	} else if level == Trace {
		_, err = f.LogFd.Write([]byte("Trace - "))
	}

	if err != nil {
		fmt.Printf("An error occurred Logging to File - %v\n", err)
	}

	_, _ = f.LogFd.Write([]byte(message))
}

func (f *FileLogger) Error(format string, args ...interface{}) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Error - "))
	_, _ = f.LogFd.Write([]byte(fmt.Sprintf(format, args...)))
}

func (f *FileLogger) Rotate() {
	f.LogFd.Rotate()
}

func (f *FileLogger) Warn(format string, args ...interface{}) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Warning - "))
	_, _ = f.LogFd.Write([]byte(fmt.Sprintf(format, args...)))
}

func (f *FileLogger) Infof(format string, args ...interface{}) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Info - "))
	_, _ = f.LogFd.Write([]byte(fmt.Sprintf(format, args...)))
}

func (f *FileLogger) Debugf(format string, args ...interface{}) {
	f.Debug(fmt.Sprintf(format, args...))
}

func (f *FileLogger) Debug(message string) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Debug - "))
	_, _ = f.LogFd.Write([]byte(message))
}

func (f *FileLogger) Info(message string) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Info - "))
	_, _ = f.LogFd.Write([]byte(message))
}

func (f *FileLogger) Trace(format string, args ...interface{}) {
	_, _ = f.Write([]byte(fmt.Sprintf(format, args...)))
}

func (f *FileLogger) Write(p []byte) (int, error) {
	dateTime := time.Now().Format("02/01/2006 15:04:05 ")
	_, _ = f.LogFd.Write([]byte(dateTime))
	_, _ = f.LogFd.Write([]byte("Trace - "))
	count, err := f.LogFd.Write(p)
	return count, err
}

func (f *FileLogger) WriteRaw(buffer []byte) (int, error) {
	return f.LogFd.Write(buffer)
}

func (f *FileLogger) WriteRawString(buffer string) (int, error) {
	count, err := f.LogFd.Write([]byte(buffer))
	return count, err
}

func (f *FileLogger) GetMinLevel() LogLevel {
	return f.MinLevel
}

func (f *FileLogger) SetMinLevel(level LogLevel) {
	f.MinLevel = level
}
