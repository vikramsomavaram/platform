/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package echo_logger

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"time"
)

// EchoLogger extend logrus.EchoLogger
type EchoLogger struct {
	*logrus.Logger
}

// Singleton logger
var singletonLogger = &EchoLogger{
	Logger: logrus.New(),
}

// Logger return singleton logger
func Logger() *EchoLogger {
	return singletonLogger
}

// Print output message of print level
func Print(i ...interface{}) {
	singletonLogger.Print(i...)
}

// Printf output format message of print level
func Printf(format string, i ...interface{}) {
	singletonLogger.Printf(format, i...)
}

// Printj output json of print level
func Printj(j log.JSON) {
	singletonLogger.Printj(j)
}

// Debug output message of debug level
func Debug(i ...interface{}) {
	singletonLogger.Debug(i...)
}

// Debugf output format message of debug level
func Debugf(format string, args ...interface{}) {
	singletonLogger.Debugf(format, args...)
}

// Debugj output json of debug level
func Debugj(j log.JSON) {
	singletonLogger.Debugj(j)
}

// Info output message of info level
func Info(i ...interface{}) {
	singletonLogger.Info(i...)
}

// Infof output format message of info level
func Infof(format string, args ...interface{}) {
	singletonLogger.Infof(format, args...)
}

// Infoj output json of info level
func Infoj(j log.JSON) {
	singletonLogger.Infoj(j)
}

// Warn output message of warn level
func Warn(i ...interface{}) {
	singletonLogger.Warn(i...)
}

// Warnf output format message of warn level
func Warnf(format string, args ...interface{}) {
	singletonLogger.Warnf(format, args...)
}

// Warnj output json of warn level
func Warnj(j log.JSON) {
	singletonLogger.Warnj(j)
}

// Error output message of error level
func Error(i ...interface{}) {
	singletonLogger.Error(i...)
}

// Errorf output format message of error level
func Errorf(format string, args ...interface{}) {
	singletonLogger.Errorf(format, args...)
}

// Errorj output json of error level
func Errorj(j log.JSON) {
	singletonLogger.Errorj(j)
}

// Fatal output message of fatal level
func Fatal(i ...interface{}) {
	singletonLogger.Fatal(i...)
}

// Fatalf output format message of fatal level
func Fatalf(format string, args ...interface{}) {
	singletonLogger.Fatalf(format, args...)
}

// Fatalj output json of fatal level
func Fatalj(j log.JSON) {
	singletonLogger.Fatalj(j)
}

// Panic output message of panic level
func Panic(i ...interface{}) {
	singletonLogger.Panic(i...)
}

// Panicf output format message of panic level
func Panicf(format string, args ...interface{}) {
	singletonLogger.Panicf(format, args...)
}

// Panicj output json of panic level
func Panicj(j log.JSON) {
	singletonLogger.Panicj(j)
}

// To logrus.Level
func toLogrusLevel(level log.Lvl) logrus.Level {
	switch level {
	case log.DEBUG:
		return logrus.DebugLevel
	case log.INFO:
		return logrus.InfoLevel
	case log.WARN:
		return logrus.WarnLevel
	case log.ERROR:
		return logrus.ErrorLevel
	}

	return logrus.InfoLevel
}

// To Echo.log.lvl
func toEchoLevel(level logrus.Level) log.Lvl {
	switch level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.InfoLevel:
		return log.INFO
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	}

	return log.OFF
}

// Output return logger io.Writer
func (l *EchoLogger) Output() io.Writer {
	return l.Out
}

// SetOutput logger io.Writer
func (l *EchoLogger) SetOutput(w io.Writer) {
	l.Out = w
}

// Level return logger level
func (l *EchoLogger) Level() log.Lvl {
	return toEchoLevel(l.Logger.Level)
}

// SetLevel logger level
func (l *EchoLogger) SetLevel(v log.Lvl) {
	l.Logger.Level = toLogrusLevel(v)
}

// SetHeader logger header
// Managed by Logrus itself
// This function do nothing
func (l *EchoLogger) SetHeader(h string) {
	// do nothing
}

// Formatter return logger formatter
func (l *EchoLogger) Formatter() logrus.Formatter {
	return l.Logger.Formatter
}

// SetFormatter logger formatter
// Only support logrus formatter
func (l *EchoLogger) SetFormatter(formatter logrus.Formatter) {
	l.Logger.Formatter = formatter
}

// Prefix return logger prefix
// This function do nothing
func (l *EchoLogger) Prefix() string {
	return ""
}

// SetPrefix logger prefix
// This function do nothing
func (l *EchoLogger) SetPrefix(p string) {
	// do nothing
}

// Print output message of print level
func (l *EchoLogger) Print(i ...interface{}) {
	l.Logger.Print(i...)
}

// Printf output format message of print level
func (l *EchoLogger) Printf(format string, args ...interface{}) {
	l.Logger.Printf(format, args...)
}

// Printj output json of print level
func (l *EchoLogger) Printj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Println(string(b))
}

// Debug output message of debug level
func (l *EchoLogger) Debug(i ...interface{}) {
	l.Logger.Debug(i...)
}

// Debugf output format message of debug level
func (l *EchoLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Debugj output message of debug level
func (l *EchoLogger) Debugj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Debugln(string(b))
}

// Info output message of info level
func (l *EchoLogger) Info(i ...interface{}) {
	l.Logger.Info(i...)
}

// Infof output format message of info level
func (l *EchoLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Infoj output json of info level
func (l *EchoLogger) Infoj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Infoln(string(b))
}

// Warn output message of warn level
func (l *EchoLogger) Warn(i ...interface{}) {
	l.Logger.Warn(i...)
}

// Warnf output format message of warn level
func (l *EchoLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Warnj output json of warn level
func (l *EchoLogger) Warnj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Warnln(string(b))
}

// Error output message of error level
func (l *EchoLogger) Error(i ...interface{}) {
	l.Logger.Error(i...)
}

// Errorf output format message of error level
func (l *EchoLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Errorj output json of error level
func (l *EchoLogger) Errorj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Errorln(string(b))
}

// Fatal output message of fatal level
func (l *EchoLogger) Fatal(i ...interface{}) {
	l.Logger.Fatal(i...)
}

// Fatalf output format message of fatal level
func (l *EchoLogger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

// Fatalj output json of fatal level
func (l *EchoLogger) Fatalj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Fatalln(string(b))
}

// Panic output message of panic level
func (l *EchoLogger) Panic(i ...interface{}) {
	l.Logger.Panic(i...)
}

// Panicf output format message of panic level
func (l *EchoLogger) Panicf(format string, args ...interface{}) {
	l.Logger.Panicf(format, args...)
}

// Panicj output json of panic level
func (l *EchoLogger) Panicj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Logger.Panicln(string(b))
}

// Logger returns a middleware that logs HTTP requests.
func LogrusLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {
				reqSize = "0"
			}

			l := stop.Sub(start)

			var b []byte
			if err != nil {
				b, _ := json.Marshal(err.Error())
				b = b[1 : len(b)-1]
			}

			logrus.WithFields(logrus.Fields{
				"httpRequest": map[string]interface{}{
					"time":          stop.Format(time.RFC3339),
					"id":            id,
					"remote_ip":     c.RealIP(),
					"status":        res.Status,
					"error":         b,
					"latency":       strconv.FormatInt(int64(l), 10),
					"latency_human": stop.Sub(start).String(),
					"bytes_in":      req.Header.Get(echo.HeaderContentLength),
					"bytes_out":     strconv.FormatInt(res.Size, 10),
					"method":        req.Method,
					"url":           req.URL.String(),
					"userAgent":     req.UserAgent(),
					"referrer":      req.Referer(),
				},
			}).Info()
			return err
		}
	}
}
