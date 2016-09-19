package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	log "github.com/kdar/factorlog"
)

// If you know exactly what your format is going to be, this is really
// good for performance.
// Outputs: 2006/01/02 15:04:05 Message
type CustomFormatter struct {
	tmp []byte
}

func NewCustomFormatter() *CustomFormatter {
	return &CustomFormatter{make([]byte, 64)}
}

// Return false. We don't want the source of the call.
func (f *CustomFormatter) ShouldRuntimeCaller() bool {
	return false
}

func (f *CustomFormatter) Format(context log.LogContext) []byte {
	buf := &bytes.Buffer{}

	t := time.Now()

	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	// Write the year in 2006/01/02 format
	log.NDigits(&f.tmp, 4, 0, year)
	f.tmp[4] = '/'
	log.TwoDigits(&f.tmp, 5, int(month))
	f.tmp[7] = '/'
	log.TwoDigits(&f.tmp, 8, day)

	f.tmp[10] = ' '

	// Write the time in 15:04:05 format
	log.TwoDigits(&f.tmp, 11, hour)
	f.tmp[13] = ':'
	log.TwoDigits(&f.tmp, 14, min)
	f.tmp[16] = ':'
	log.TwoDigits(&f.tmp, 17, sec)

	f.tmp[19] = ' '

	// Write what we have thus far in tmp to our buffer
	buf.Write(f.tmp[:20])

	message := ""
	if context.Format != nil {
		message = fmt.Sprintf(*context.Format, context.Args...)
	} else {
		message = fmt.Sprint(context.Args...)
	}

	// Write the message to our buffer
	buf.WriteString(message)

	// If we don't have a newline, put one. All formatters must
	// do this.
	l := len(message)
	if l > 0 && message[l-1] != '\n' {
		buf.WriteRune('\n')
	}

	return buf.Bytes()
}

func main() {
	log := log.New(os.Stdout, NewCustomFormatter())
	log.Println("Custom formatter")
}
