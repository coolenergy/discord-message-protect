package logging

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"os"
)

type ConsoleLogger struct {
	MinLevel LogLevel
}

func (c *ConsoleLogger) Logf(level LogLevel, format string, args ...interface{}) {

	if c.MinLevel > level {
		return
	}

	if level == Trace {
		if len(args) == 0 {
			fmt.Print(format)
		} else {
			fmt.Printf(format, args...)
		}
		return
	}

	if color, ok := LogColors[level]; ok {
		if len(args) == 0 {
			fmt.Print(aurora.Colorize(format, color))
		} else {
			fmt.Print(aurora.Colorize(fmt.Sprintf(format, args...), color))
		}
	} else {
		if len(args) == 0 {
			fmt.Print(format)
		} else {
			fmt.Printf(format, args...)
		}
	}
}

func (c *ConsoleLogger) Log(level LogLevel, message string) {

	if c.MinLevel > level {
		return
	}

	if level == Trace {
		fmt.Print(message)
		return
	}

	if color, ok := LogColors[level]; ok {
		fmt.Print(aurora.Colorize(message, color))
	} else {
		fmt.Print(message)
	}
}

func (c *ConsoleLogger) Error(format string, args ...interface{}) {
	c.Logf(Error, format, args...)
}

func (c *ConsoleLogger) Warn(format string, args ...interface{}) {
	c.Logf(Warn, format, args...)
}

func (c *ConsoleLogger) Infof(format string, args ...interface{}) {
	c.Logf(Info, format, args...)
}

func (c *ConsoleLogger) Info(message string) {
	c.Log(Info, message)
}

func (c *ConsoleLogger) Debugf(format string, args ...interface{}) {
	c.Logf(Debug, format, args...)
}
func (c *ConsoleLogger) Debug(message string) {
	c.Log(Debug, message)
}

func (c *ConsoleLogger) Trace(format string, args ...interface{}) {
	c.Logf(Trace, format, args...)
}

func (c *ConsoleLogger) Write(p []byte) (n int, err error) {
	count, err := os.Stdout.Write(p)
	return count, err
}

func (c *ConsoleLogger) WriteRaw(buffer []byte) (int, error) {
	count, err := os.Stdout.Write(buffer)
	return count, err
}

func (c *ConsoleLogger) WriteRawString(buffer string) (int, error) {
	count, err := os.Stdout.Write([]byte(buffer))
	return count, err
}

func (c *ConsoleLogger) GetMinLevel() LogLevel {
	return c.MinLevel
}

func (c *ConsoleLogger) SetMinLevel(level LogLevel) {
	c.MinLevel = level
}

func (c *ConsoleLogger) Rotate() {

}
