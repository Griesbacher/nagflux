package factorlog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	pid = 0
)

// Level represents the level of verbosity.
type Level int32

func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// Severity represents the severity of the log.
type Severity int32

func (l *Severity) get() Severity {
	return Severity(atomic.LoadInt32((*int32)(l)))
}

func (l *Severity) set(val Severity) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

const (
	NONE Severity = 1 << iota
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	CRITICAL
	STACK
	FATAL
	PANIC
)

var (
	maxint32 = ^uint32(0) >> 1
	//maxuint32 = ^uint32(0)
	//maxint64 = ^uint64(0) >> 1
	//maxuint64 = ^uint64(0)
)

type Logger interface {
	Output(sev Severity, calldepth int, v ...interface{}) error
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Traceln(v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Critical(v ...interface{})
	Criticalf(format string, v ...interface{})
	Criticalln(v ...interface{})
	Stack(v ...interface{})
	Stackf(format string, v ...interface{})
	Stackln(v ...interface{})
	Log(sev Severity, v ...interface{})

	//Log verbosity
	V(level Level) Verbose
	SetVerbosity(level Level)
	IsV(level Level) bool

	// golang's log interface
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}

// FactorLog is a logging object that outputs data to an io.Writer.
// Each write is threadsafe.
type FactorLog struct {
	mu         sync.Mutex // ensures atomic writes; protects the following fields
	out        io.Writer  // destination for output
	formatter  Formatter
	verbosity  Level
	severities Severity
}

// New creates a FactorLog with the given output and format.
func New(out io.Writer, formatter Formatter) *FactorLog {
	return &FactorLog{out: out, formatter: formatter, severities: Severity(maxint32)}
}

// just like Go's log.std
var std = New(os.Stderr, NewStdFormatter("%{Date} %{Time} %{Message}"))

// Sets the verbosity level of this log. Use IsV() or V() to
// utilize verbosity.
func (l *FactorLog) SetVerbosity(level Level) {
	l.verbosity.set(level)
}

// SetSeverities sets which severities this log will output for.
// Example:
//   l.SetSeverities(INFO|DEBUG)
func (l *FactorLog) SetSeverities(sev Severity) {
	l.severities.set(sev)
}

// SetMinMaxSeverity sets the minimum and maximum severities this
// log will output for.
// Example:
//   l.SetMinMaxSeverity(INFO, ERROR)
func (l *FactorLog) SetMinMaxSeverity(min Severity, max Severity) {
	if min > max || max < min {
		min, max = max, min
	}

	if max > PANIC {
		max = PANIC
	}

	if min < NONE {
		min = NONE
	}

	sev := Severity(0)
	for s := min; s <= max; s <<= 1 {
		sev |= s
	}

	l.severities.set(sev)
}

// Output will write to the writer with the given severity, calldepth,
// and string. calldepth is only used if the format requires a call to
// runtime.Caller.
func (l *FactorLog) Output(sev Severity, calldepth int, v ...interface{}) error {
	return l.output(sev, calldepth+1, nil, v...)
}

func (l *FactorLog) output(sev Severity, calldepth int, format *string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if sev&l.severities.get() == 0 {
		return nil
	}

	context := LogContext{
		Time:     time.Now(),
		Severity: sev,
		Pid:      pid,
		Format:   format,
		Args:     v,
	}

	if l.formatter.ShouldRuntimeCaller() {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		pc, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			me := runtime.FuncForPC(pc)
			if me != nil {
				context.Function = me.Name()
			}
		}

		context.File = file
		context.Line = line

		l.mu.Lock()
	}

	_, err := l.out.Write(l.formatter.Format(context))

	// If severity is STACK, output the stack.
	if sev == STACK {
		l.out.Write(GetStack(calldepth + 1))
	}

	return err
}

// SetOutput sets the output destination for this logger.
func (l *FactorLog) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetFormatter sets the formatter for this logger.
func (l *FactorLog) SetFormatter(f Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = f
}

// IsV tests whether the verbosity is of a certain level.
// Returns a bool.
// Example:
//    if log.IsV(2) {
//      log.Info("some info")
//    }
func (l *FactorLog) IsV(level Level) bool {
	if l.verbosity.get() >= level {
		return true
	}

	return false
}

// V tests whether the verbosity is of a certain level,
// and returns a Verbose object that allows you to
// chain calls. This is a convenience function and should
// be avoided if you care about raw performance (use IsV()
// instead).
// Example:
//   log.V(2).Info("some info")
func (l *FactorLog) V(level Level) Verbose {
	if l.verbosity.get() >= level {
		return Verbose{true, l}
	}

	return Verbose{false, l}
}

// Trace is equivalent to Print with severity TRACE.
func (l *FactorLog) Trace(v ...interface{}) {
	l.output(TRACE, 2, nil, v...)
}

// Tracef is equivalent to Printf with severity TRACE.
func (l *FactorLog) Tracef(format string, v ...interface{}) {
	l.output(TRACE, 2, &format, v...)
}

// Traceln is equivalent to Println with severity TRACE.
func (l *FactorLog) Traceln(v ...interface{}) {
	l.output(TRACE, 2, nil, v...)
}

// Debug is equivalent to Print with severity DEBUG.
func (l *FactorLog) Debug(v ...interface{}) {
	l.output(DEBUG, 2, nil, v...)
}

// Debugf is equivalent to Printf with severity DEBUG.
func (l *FactorLog) Debugf(format string, v ...interface{}) {
	l.output(DEBUG, 2, &format, v...)
}

// Debugln is equivalent to Println with severity DEBUG.
func (l *FactorLog) Debugln(v ...interface{}) {
	l.output(DEBUG, 2, nil, v...)
}

// Info is equivalent to Print with severity INFO.
func (l *FactorLog) Info(v ...interface{}) {
	l.output(INFO, 2, nil, v...)
}

// Infof is equivalent to Printf with severity INFO.
func (l *FactorLog) Infof(format string, v ...interface{}) {
	l.output(INFO, 2, &format, v...)
}

// Infoln is equivalent to Println with severity INFO.
func (l *FactorLog) Infoln(v ...interface{}) {
	l.output(INFO, 2, nil, v...)
}

// Warn is equivalent to Print with severity WARN.
func (l *FactorLog) Warn(v ...interface{}) {
	l.output(WARN, 2, nil, v...)
}

// Warnf is equivalent to Printf with severity WARN.
func (l *FactorLog) Warnf(format string, v ...interface{}) {
	l.output(WARN, 2, &format, v...)
}

// Warnln is equivalent to Println with severity WARN.
func (l *FactorLog) Warnln(v ...interface{}) {
	l.output(WARN, 2, nil, v...)
}

// Error is equivalent to Print with severity ERROR.
func (l *FactorLog) Error(v ...interface{}) {
	l.output(ERROR, 2, nil, v...)
}

// Errorf is equivalent to Printf with severity ERROR.
func (l *FactorLog) Errorf(format string, v ...interface{}) {
	l.output(ERROR, 2, &format, v...)
}

// Errorln is equivalent to Println with severity ERROR.
func (l *FactorLog) Errorln(v ...interface{}) {
	l.output(ERROR, 2, nil, v...)
}

// Critical is equivalent to Print with severity CRITICAL.
func (l *FactorLog) Critical(v ...interface{}) {
	l.output(CRITICAL, 2, nil, v...)
}

// Criticalf is equivalent to Printf with severity CRITICAL.
func (l *FactorLog) Criticalf(format string, v ...interface{}) {
	l.output(CRITICAL, 2, &format, v...)
}

// Criticalln is equivalent to Println with severity CRITICAL.
func (l *FactorLog) Criticalln(v ...interface{}) {
	l.output(CRITICAL, 2, nil, v...)
}

// Stack is equivalent to Print() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stack(v ...interface{}) {
	l.output(STACK, 2, nil, v...)
}

// Stackf is equivalent to Printf() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stackf(format string, v ...interface{}) {
	l.output(STACK, 2, &format, v...)
}

// Stackln is equivalent to Println() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stackln(v ...interface{}) {
	l.output(STACK, 2, nil, v...)
}

// Log calls l.output to print to the logger. Uses fmt.Sprint.
func (l *FactorLog) Log(sev Severity, v ...interface{}) {
	l.output(sev, 2, nil, v...)
}

// Print calls l.output to print to the logger. Uses fmt.Sprint.
func (l *FactorLog) Print(v ...interface{}) {
	l.output(DEBUG, 2, nil, v...)
}

// Print calls l.output to print to the logger. Uses fmt.Sprintf.
func (l *FactorLog) Printf(format string, v ...interface{}) {
	l.output(DEBUG, 2, &format, v...)
}

// Println calls l.output to print to the logger. Uses fmt.Sprint.
// This is more of a convenience function. If you really want
// to output an extra newline at the end, just append \n.
func (l *FactorLog) Println(v ...interface{}) {
	l.output(DEBUG, 2, nil, v...)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func (l *FactorLog) Fatal(v ...interface{}) {
	l.output(FATAL, 2, nil, v...)
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func (l *FactorLog) Fatalf(format string, v ...interface{}) {
	l.output(FATAL, 2, &format, v...)
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func (l *FactorLog) Fatalln(v ...interface{}) {
	l.output(FATAL, 2, nil, v...)
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func (l *FactorLog) Panic(v ...interface{}) {
	l.output(PANIC, 2, nil, v...)
	panic(fmt.Sprint(v...))
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (l *FactorLog) Panicf(format string, v ...interface{}) {
	l.output(PANIC, 2, &format, v...)
	panic(fmt.Sprintf(format, v...))
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (l *FactorLog) Panicln(v ...interface{}) {
	l.output(PANIC, 2, nil, v...)
	panic(fmt.Sprint(v...))
}

// Verbose is a structure that enables syntatic sugar
// when testing for verbosity and calling a log function.
// See FactorLog.V().
type Verbose struct {
	True   bool
	logger *FactorLog
}

func (b Verbose) Output(sev Severity, calldepth int, v ...interface{}) error {
	if b.True {
		return b.logger.output(TRACE, calldepth, nil, v...)
	}

	return nil
}

func (b Verbose) Trace(v ...interface{}) {
	if b.True {
		b.logger.output(TRACE, 2, nil, v...)
	}
}

func (b Verbose) Tracef(format string, v ...interface{}) {
	if b.True {
		b.logger.output(TRACE, 2, &format, v...)
	}
}

func (b Verbose) Traceln(v ...interface{}) {
	if b.True {
		b.logger.output(TRACE, 2, nil, v...)
	}
}

func (b Verbose) Debug(v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, nil, v...)
	}
}

func (b Verbose) Debugf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, &format, v...)
	}
}

func (b Verbose) Debugln(v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, nil, v...)
	}
}

func (b Verbose) Info(v ...interface{}) {
	if b.True {
		b.logger.output(INFO, 2, nil, v...)
	}
}

func (b Verbose) Infof(format string, v ...interface{}) {
	if b.True {
		b.logger.output(INFO, 2, &format, v...)
	}
}

func (b Verbose) Infoln(v ...interface{}) {
	if b.True {
		b.logger.output(INFO, 2, nil, v...)
	}
}

func (b Verbose) Warn(v ...interface{}) {
	if b.True {
		b.logger.output(WARN, 2, nil, v...)
	}
}

func (b Verbose) Warnf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(WARN, 2, &format, v...)
	}
}

func (b Verbose) Warnln(v ...interface{}) {
	if b.True {
		b.logger.output(WARN, 2, nil, v...)
	}
}

func (b Verbose) Error(v ...interface{}) {
	if b.True {
		b.logger.output(ERROR, 2, nil, v...)
	}
}

func (b Verbose) Errorf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(ERROR, 2, &format, v...)
	}
}

func (b Verbose) Errorln(v ...interface{}) {
	if b.True {
		b.logger.output(ERROR, 2, nil, v...)
	}
}

func (b Verbose) Critical(v ...interface{}) {
	if b.True {
		b.logger.output(CRITICAL, 2, nil, v...)
	}
}

func (b Verbose) Criticalf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(CRITICAL, 2, &format, v...)
	}
}

func (b Verbose) Criticalln(v ...interface{}) {
	if b.True {
		b.logger.output(CRITICAL, 2, nil, v...)
	}
}

func (b Verbose) Stack(v ...interface{}) {
	if b.True {
		b.logger.output(STACK, 2, nil, v...)
	}
}

func (b Verbose) Stackf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(STACK, 2, &format, v...)
	}
}

func (b Verbose) Stackln(v ...interface{}) {
	if b.True {
		b.logger.output(STACK, 2, nil, v...)
	}
}

func (b Verbose) Log(sev Severity, v ...interface{}) {
	if b.True {
		b.logger.output(sev, 2, nil, v...)
	}
}

func (b Verbose) Print(v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, nil, v...)
	}
}

func (b Verbose) Printf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, &format, v...)
	}
}

func (b Verbose) Println(v ...interface{}) {
	if b.True {
		b.logger.output(DEBUG, 2, nil, v...)
	}
}

func (b Verbose) Fatal(v ...interface{}) {
	if b.True {
		b.logger.output(FATAL, 2, nil, v...)
		os.Exit(1)
	}
}

func (b Verbose) Fatalf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(FATAL, 2, &format, v...)
		os.Exit(1)
	}
}

func (b Verbose) Fatalln(v ...interface{}) {
	if b.True {
		b.logger.output(FATAL, 2, nil, v...)
		os.Exit(1)
	}
}

func (b Verbose) Panic(v ...interface{}) {
	if b.True {
		b.logger.output(PANIC, 2, nil, v...)
		panic(fmt.Sprint(v...))
	}
}

func (b Verbose) Panicf(format string, v ...interface{}) {
	if b.True {
		b.logger.output(PANIC, 2, &format, v...)
		panic(fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Panicln(v ...interface{}) {
	if b.True {
		b.logger.output(PANIC, 2, nil, v...)
		panic(fmt.Sprint(v...))
	}
}

func (b Verbose) IsV(level Level) bool {
	if b.logger.verbosity.get() >= level {
		return true
	}

	return false
}

func (b Verbose) V(level Level) Verbose {
	if b.logger.verbosity.get() >= level {
		return Verbose{true, b.logger}
	}

	return Verbose{false, b.logger}
}

func (b Verbose) SetVerbosity(level Level) {
	b.logger.SetVerbosity(level)
}

// Global functions for the package. Uses a standard
// logger just like Go's log package.

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// SetFormatter sets the formatter for the standard logger.
func SetFormatter(f Formatter) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.formatter = f
}

func SetVerbosity(level Level) {
	std.SetVerbosity(level)
}

func SetSeverities(sev Severity) {
	std.SetSeverities(sev)
}

func SetMinMaxSeverity(min Severity, max Severity) {
	std.SetMinMaxSeverity(min, max)
}

func IsV(level Level) bool {
	if std.verbosity.get() >= level {
		return true
	}

	return false
}

func V(level Level) Verbose {
	if std.verbosity.get() >= level {
		return Verbose{true, std}
	}

	return Verbose{false, std}
}

func Trace(v ...interface{}) {
	std.output(TRACE, 2, nil, v...)
}

func Tracef(format string, v ...interface{}) {
	std.output(TRACE, 2, &format, v...)
}

func Traceln(v ...interface{}) {
	std.output(TRACE, 2, nil, v...)
}

func Debug(v ...interface{}) {
	std.output(DEBUG, 2, nil, v...)
}

func Debugf(format string, v ...interface{}) {
	std.output(DEBUG, 2, &format, v...)
}

func Debugln(v ...interface{}) {
	std.output(DEBUG, 2, nil, v...)
}

func Info(v ...interface{}) {
	std.output(INFO, 2, nil, v...)
}

func Infof(format string, v ...interface{}) {
	std.output(INFO, 2, &format, v...)
}

func Infoln(v ...interface{}) {
	std.output(INFO, 2, nil, v...)
}

func Warn(v ...interface{}) {
	std.output(WARN, 2, nil, v...)
}

func Warnf(format string, v ...interface{}) {
	std.output(WARN, 2, &format, v...)
}

func Warnln(v ...interface{}) {
	std.output(WARN, 2, nil, v...)
}

func Error(v ...interface{}) {
	std.output(ERROR, 2, nil, v...)
}

func Errorf(format string, v ...interface{}) {
	std.output(ERROR, 2, &format, v...)
}

func Errorln(v ...interface{}) {
	std.output(ERROR, 2, nil, v...)
}

func Critical(v ...interface{}) {
	std.output(CRITICAL, 2, nil, v...)
}

func Criticalf(format string, v ...interface{}) {
	std.output(CRITICAL, 2, &format, v...)
}

func Criticalln(v ...interface{}) {
	std.output(CRITICAL, 2, nil, v...)
}

func Stack(v ...interface{}) {
	std.output(STACK, 2, nil, v...)
}

func Stackf(format string, v ...interface{}) {
	std.output(STACK, 2, &format, v...)
}

func Stackln(v ...interface{}) {
	std.output(STACK, 2, nil, v...)
}

func Log(sev Severity, v ...interface{}) {
	std.output(sev, 2, nil, v...)
}

func Print(v ...interface{}) {
	std.output(DEBUG, 2, nil, v...)
}

func Printf(format string, v ...interface{}) {
	std.output(DEBUG, 2, &format, v...)
}

func Println(v ...interface{}) {
	std.output(DEBUG, 2, nil, v...)
}

func Fatal(v ...interface{}) {
	std.output(FATAL, 2, nil, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.output(FATAL, 2, &format, v...)
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	std.output(FATAL, 2, nil, v...)
	os.Exit(1)
}

func Panic(v ...interface{}) {
	std.output(PANIC, 2, nil, v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	std.output(PANIC, 2, &format, v...)
	panic(fmt.Sprintf(format, v...))
}

func Panicln(v ...interface{}) {
	std.output(PANIC, 2, nil, v...)
	panic(fmt.Sprint(v...))
}

func init() {
	pid = os.Getpid()
}

// Creates a logger that outputs to nothing
type NullLogger struct{}

func (NullLogger) Output(sev Severity, calldepth int, v ...interface{}) error { return nil }
func (NullLogger) Trace(v ...interface{})                                     {}
func (NullLogger) Tracef(format string, v ...interface{})                     {}
func (NullLogger) Traceln(v ...interface{})                                   {}
func (NullLogger) Debug(v ...interface{})                                     {}
func (NullLogger) Debugf(format string, v ...interface{})                     {}
func (NullLogger) Debugln(v ...interface{})                                   {}
func (NullLogger) Info(v ...interface{})                                      {}
func (NullLogger) Infof(format string, v ...interface{})                      {}
func (NullLogger) Infoln(v ...interface{})                                    {}
func (NullLogger) Warn(v ...interface{})                                      {}
func (NullLogger) Warnf(format string, v ...interface{})                      {}
func (NullLogger) Warnln(v ...interface{})                                    {}
func (NullLogger) Error(v ...interface{})                                     {}
func (NullLogger) Errorf(format string, v ...interface{})                     {}
func (NullLogger) Errorln(v ...interface{})                                   {}
func (NullLogger) Critical(v ...interface{})                                  {}
func (NullLogger) Criticalf(format string, v ...interface{})                  {}
func (NullLogger) Criticalln(v ...interface{})                                {}
func (NullLogger) Stack(v ...interface{})                                     {}
func (NullLogger) Stackf(format string, v ...interface{})                     {}
func (NullLogger) Stackln(v ...interface{})                                   {}
func (NullLogger) Log(sev Severity, v ...interface{})                         {}
func (NullLogger) Print(v ...interface{})                                     {}
func (NullLogger) Printf(format string, v ...interface{})                     {}
func (NullLogger) Println(v ...interface{})                                   {}
func (NullLogger) Fatal(v ...interface{})                                     {}
func (NullLogger) Fatalf(format string, v ...interface{})                     {}
func (NullLogger) Fatalln(v ...interface{})                                   {}
func (NullLogger) Panic(v ...interface{})                                     {}
func (NullLogger) Panicf(format string, v ...interface{})                     {}
func (NullLogger) Panicln(v ...interface{})                                   {}
func (NullLogger) V(level Level) Verbose                                      { return Verbose{} }
func (NullLogger) SetVerbosity(level Level)                                   {}
func (NullLogger) IsV(level Level) bool                                       { return false }
