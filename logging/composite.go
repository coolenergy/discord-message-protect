package logging

import (
	"fmt"
	"sync"
)

type CompositeLogger struct {
	sync.Mutex
	loggers []ILogger
}

func NewCompositeLogger(loggers ...ILogger) *CompositeLogger {
	return &CompositeLogger{
		loggers: loggers,
	}
}

func (c *CompositeLogger) Logf(level LogLevel, format string, args ...interface{}) {
	c.Log(level, fmt.Sprintf(format, args...))
}

func (c *CompositeLogger) Log(level LogLevel, message string) {
	c.Lock()
	defer c.Unlock()

	for i := 0; i < len(c.loggers); i++ {
		c.loggers[i].Log(level, message)
	}
}

func (c *CompositeLogger) Error(format string, args ...interface{}) {
	c.Logf(Error, format, args...)
}

func (c *CompositeLogger) Rotate() {
	for i := 0; i < len(c.loggers); i++ {
		c.loggers[i].Rotate()
	}
}

func (c *CompositeLogger) Warn(format string, args ...interface{}) {
	c.Logf(Warn, format, args...)
}

func (c *CompositeLogger) Infof(format string, args ...interface{}) {
	c.Logf(Info, format, args...)
}

func (c *CompositeLogger) Info(message string) {
	c.Log(Info, message)
}

func (c *CompositeLogger) Debugf(format string, args ...interface{}) {
	c.Logf(Debug, format, args...)
}

func (c *CompositeLogger) Debug(message string) {
	c.Log(Debug, message)
}

func (c *CompositeLogger) Trace(format string, args ...interface{}) {
	c.Logf(Trace, format, args...)
}

func (c *CompositeLogger) Write(p []byte) (n int, err error) {
	c.Lock()
	defer c.Unlock()

	count := 0
	for i := 0; i < len(c.loggers); i++ {
		count, err = c.loggers[i].Write(p)
		if err != nil {
			return 0, err
		}
	}

	return count, err
}

func (c *CompositeLogger) WriteRaw(p []byte) (n int, err error) {
	c.Lock()
	defer c.Unlock()

	count := 0
	for i := 0; i < len(c.loggers); i++ {
		count, err = c.loggers[i].WriteRaw(p)
		if err != nil {
			return 0, err
		}
	}

	return count, err
}

func (c *CompositeLogger) WriteRawString(p string) (n int, err error) {
	c.Lock()
	defer c.Unlock()

	count := 0
	for i := 0; i < len(c.loggers); i++ {
		count, err = c.loggers[i].WriteRawString(p)
		if err != nil {
			return 0, err
		}
	}

	return count, err
}

func (c *CompositeLogger) GetMinLevel() LogLevel {
	c.Lock()
	defer c.Unlock()
	minLogger := Fatal
	for i := 0; i < len(c.loggers); i++ {
		currentLevel := c.loggers[i].GetMinLevel()
		if currentLevel == Trace {
			return Trace
		}
		if currentLevel < minLogger {
			minLogger = currentLevel
		}
	}

	return minLogger
}

func (c *CompositeLogger) SetMinLevel(level LogLevel) {
	for i := 0; i < len(c.loggers); i++ {
		c.loggers[i].SetMinLevel(level)
	}
}
