package factorlog

import (
	"time"
)

var UcSeverityStrings = [...]string{
	"NONE",
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"CRITICAL",
	"STACK",
	"FATAL",
	"PANIC",
}

var UcShortSeverityStrings = [...]string{
	"NONE",
	"TRAC",
	"DEBG",
	"INFO",
	"WARN",
	"EROR",
	"CRIT",
	"STAK",
	"FATL",
	"PANC",
}

var UcShortestSeverityStrings = [...]string{
	"",
	"T",
	"D",
	"I",
	"W",
	"E",
	"C",
	"S",
	"F",
	"P",
}

var LcSeverityStrings = [...]string{
	"none",
	"trace",
	"debug",
	"info",
	"warn",
	"error",
	"critical",
	"stack",
	"fatal",
	"panic",
}

var LcShortSeverityStrings = [...]string{
	"none",
	"trac",
	"debg",
	"info",
	"warn",
	"eror",
	"crit",
	"stak",
	"fatl",
	"panc",
}

var LcShortestSeverityStrings = [...]string{
	"n",
	"t",
	"d",
	"i",
	"w",
	"e",
	"c",
	"s",
	"f",
	"p",
}

var CapSeverityStrings = [...]string{
	"None",
	"Trace",
	"Debug",
	"Info",
	"Warn",
	"Error",
	"Critical",
	"Stack",
	"Fatal",
	"Panic",
}

var CapShortSeverityStrings = [...]string{
	"None",
	"Trac",
	"Debg",
	"Info",
	"Warn",
	"Eror",
	"Crit",
	"Stak",
	"Fatl",
	"Panc",
}

// Convert an uppercase string to a severity.
func StringToSeverity(s string) Severity {
	sev := Severity(1)
	for _, v := range UcSeverityStrings {
		if v == s {
			return sev
		}
		sev <<= 1
	}

	return -1
}

// We use this function to convert a severity to an index
// that we can then use to index variables like UcSeverityStrings.
// Why don't we use maps? This is almost 3x faster.
func SeverityToIndex(sev Severity) int {
	count := 0
	for ; sev > 1; count++ {
		sev >>= 1
	}
	return count
}

// Interface to format anything
type Formatter interface {
	// Formats LogRecord and returns the []byte that will
	// be written by the log. This is not inherently thread
	// safe but FactorLog uses a mutex before calling this.
	Format(context LogContext) []byte

	// Returns true if we should call runtime.Caller because
	// we have a format that requires it. We do this because
	// it is expensive.
	ShouldRuntimeCaller() bool
}

// Structure used to hold the data used for formatting
type LogContext struct {
	Time     time.Time
	Severity Severity
	File     string
	Line     int
	Function string
	Pid      int
	Format   *string
	Args     []interface{}
}

// // GetStack returns a stack trace from the runtime
// // if all is true, all goroutines are included
// func GetStack(all bool) []byte {
// 	n := 10000
// 	if all {
// 		n = 100000
// 	}
// 	var trace []byte
// 	for i := 0; i < 5; i++ {
// 		trace = make([]byte, n)
// 		nbytes := runtime.Stack(trace, all)
// 		if nbytes < len(trace) {
// 			return trace[:nbytes]
// 		}
// 		n *= 2
// 	}
// 	return trace
// }

const digits = "0123456789"

// twoDigits converts an integer d to its ascii representation
// i is the destination index in buf
func TwoDigits(buf *[]byte, i, d int) {
	(*buf)[i+1] = digits[d%10]
	d /= 10
	(*buf)[i] = digits[d%10]
}

// nDigits converts an integer d to its ascii representation
// n is how many digits to use
// i is the destination index in buf
func NDigits(buf *[]byte, n, i, d int) {
	// reverse order
	for j := n - 1; j >= 0; j-- {
		(*buf)[i+j] = digits[d%10]
		d /= 10
	}
}

const ddigits = `0001020304050607080910111213141516171819` +
	`2021222324252627282930313233343536373839` +
	`4041424344454647484950515253545556575859` +
	`6061626364656667686970717273747576777879` +
	`8081828384858687888990919293949596979899`

// itoa converts an integer d to its ascii representation
// i is the deintation index in buf
// algorithm from https://www.facebook.com/notes/facebook-engineering/three-optimization-tips-for-c/10151361643253920
func Itoa(buf *[]byte, i, d int) int {
	j := len(*buf)

	for d >= 100 {
		// Integer division is slow, so we do it by 2
		index := (d % 100) * 2
		d /= 100
		j--
		(*buf)[j] = ddigits[index+1]
		j--
		(*buf)[j] = ddigits[index]
	}

	if d < 10 {
		j--
		(*buf)[j] = byte(int('0') + d)
		return copy((*buf)[i:], (*buf)[j:])
	}

	index := d * 2
	j--
	(*buf)[j] = ddigits[index+1]
	j--
	(*buf)[j] = ddigits[index]

	return copy((*buf)[i:], (*buf)[j:])
}

// I64toa is the same as itoa but for 64bit integers
func I64toa(buf *[]byte, i int, d int64) int {
	j := len(*buf)

	for d >= 100 {
		// Integer division is slow, so we do it by 2
		index := (d % 100) * 2
		d /= 100
		j--
		(*buf)[j] = ddigits[index+1]
		j--
		(*buf)[j] = ddigits[index]
	}

	if d < 10 {
		j--
		(*buf)[j] = byte(int64('0') + d)
		return copy((*buf)[i:], (*buf)[j:])
	}

	index := d * 2
	j--
	(*buf)[j] = ddigits[index+1]
	j--
	(*buf)[j] = ddigits[index]

	return copy((*buf)[i:], (*buf)[j:])
}

// Ui64toa is the same as itoa but for 64bit unsigned integers
func Ui64toa(buf *[]byte, i int, d uint64) int {
	j := len(*buf)

	for d >= 100 {
		// Integer division is slow, so we do it by 2
		index := (d % 100) * 2
		d /= 100
		j--
		(*buf)[j] = ddigits[index+1]
		j--
		(*buf)[j] = ddigits[index]
	}

	if d < 10 {
		j--
		(*buf)[j] = byte(uint64('0') + d)
		return copy((*buf)[i:], (*buf)[j:])
	}

	index := d * 2
	j--
	(*buf)[j] = ddigits[index+1]
	j--
	(*buf)[j] = ddigits[index]

	return copy((*buf)[i:], (*buf)[j:])
}
