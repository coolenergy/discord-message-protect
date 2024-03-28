package logging

import (
	"github.com/logrusorgru/aurora"
)

type LogLevel int

const (
	Trace LogLevel = 0
	Debug LogLevel = 1
	Info  LogLevel = 2
	Warn  LogLevel = 3
	Error LogLevel = 4
	Fatal LogLevel = 5
)

// Taken from aurora::color.go
var shiftFg = 16                  // shift for foreground (starting from 16th bit)
var flagFg aurora.Color = 1 << 14 // presence flag (14th bit)

var LogColors = map[LogLevel]aurora.Color{
	Fatal: aurora.RedFg,
	Error: aurora.BrightFg | aurora.RedFg,
	Warn:  aurora.BoldFm | aurora.YellowFg,
	Debug: aurora.GreenFg,
	Info:  aurora.CyanFg | aurora.BrightFg,
	// Gray, taken from Gray() n from 0 to 24
	// Trace: (aurora.Color(uint((232+24) << shiftFg)) | flagFg,
}

type ILogger interface {
	Log(level LogLevel, message string)
	Logf(level LogLevel, format string, args ...interface{})
	Error(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Info(message string)
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(message string)
	Trace(format string, args ...interface{})
	Write(p []byte) (int, error)
	WriteRaw(buffer []byte) (int, error)
	WriteRawString(buffer string) (int, error)
	GetMinLevel() LogLevel
	SetMinLevel(level LogLevel)
	Rotate()
}
