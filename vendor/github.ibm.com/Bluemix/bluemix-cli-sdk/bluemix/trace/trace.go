package trace

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type Closer interface {
	Close() error
}

type PrinterCloser interface {
	Printer
	Closer
}

type NullLogger struct{}

func (l *NullLogger) Print(v ...interface{})                 {}
func (l *NullLogger) Printf(format string, v ...interface{}) {}
func (l *NullLogger) Println(v ...interface{})               {}

type loggerImpl struct {
	*log.Logger
	c io.WriteCloser
}

func (loggerImpl *loggerImpl) Close() error {
	if loggerImpl.c != nil {
		return loggerImpl.c.Close()
	}
	return nil
}

func newLoggerImpl(out io.Writer, prefix string, flag int) *loggerImpl {
	l := log.New(out, prefix, flag)
	c := out.(io.WriteCloser)
	return &loggerImpl{
		Logger: l,
		c:      c,
	}
}

var Logger Printer = NewLogger("")

// NewLogger returns a printer for the given trace setting.
func NewLogger(bluemix_trace string) Printer {
	switch strings.ToLower(bluemix_trace) {
	case "", "false":
		return new(NullLogger)
	case "true":
		return NewStdLogger()
	default:
		return NewFileLogger(bluemix_trace)
	}
}

// NewStdLogger return a printer that writes to StdOut.
func NewStdLogger() PrinterCloser {
	return newLoggerImpl(os.Stderr, "", 0)
}

// NewFileLoffer return a printer that writes to the given file path.
func NewFileLogger(path string) PrinterCloser {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		logger := NewStdLogger()
		logger.Printf("An error occurred when creating log file '%s':\n%v\n\n", path, err)
		return logger
	}
	return newLoggerImpl(file, "", 0)
}

// Santitize returns a clean string with sentive user data in the input
// replaced by PRIVATE_DATA_PLACEHOLDER.
func Sanitize(input string) (sanitized string) {
	var sanitizeJson = func(propertyName string, json string) string {
		regex := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"[^"]*"`, propertyName))
		return regex.ReplaceAllString(json, fmt.Sprintf(`"%s":"%s"`, propertyName, PRIVATE_DATA_PLACEHOLDER()))
	}

	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized = re.ReplaceAllString(input, "Authorization: "+PRIVATE_DATA_PLACEHOLDER())
	re = regexp.MustCompile(`(?m)^X-Auth-Token: .*`)
	sanitized = re.ReplaceAllString(sanitized, "X-Auth-Token: "+PRIVATE_DATA_PLACEHOLDER())
	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PRIVATE_DATA_PLACEHOLDER()+"&")

	sanitized = sanitizeJson("access_token", sanitized)
	sanitized = sanitizeJson("refresh_token", sanitized)
	sanitized = sanitizeJson("token", sanitized)
	sanitized = sanitizeJson("password", sanitized)
	sanitized = sanitizeJson("oldPassword", sanitized)

	return
}

// PRIVATE_DATA_PLACEHOLDER returns the text to replace the sentive data.
func PRIVATE_DATA_PLACEHOLDER() string {
	return "[PRIVATE DATA HIDDEN]"
}
