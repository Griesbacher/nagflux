// FactorLog is a logging infrastructure for Go that provides numerous
// logging functions for whatever your style may be. It could easily
// be a replacement for Go's log in the standard library (though it
// doesn't support functions such as `SetFlags()`).
//
// Basic usage:
//   import log "github.com/kdar/factorlog"
//   log.Print("Hello there!")
//
// Setting your own format:
//   import os
//   import "github.com/kdar/factorlog"
//   log := factorlog.New(os.Stdout, factorlog.NewStdFormatter("%{Date} %{Time} %{File}:%{Line} %{Message}"))
//   log.Print("Hello there!")
//
// Setting the verbosity and testing against it:
//   import os
//   import "github.com/kdar/factorlog"
//   log := factorlog.New(os.Stdout, factorlog.NewStdFormatter("%{Date} %{Time} %{File}:%{Line} %{Message}"))
//   log.SetVerbosity(2)
//   log.V(1).Print("Will print")
//   log.V(3).Print("Will not print")
//
// If you care about performance, you can test for verbosity this way:
//   if log.IsV(1) {
//     log.Print("Hello there!")
//   }
//
// For more usage examples, check the examples/ directory.
//
// Format verbs:
//   %{SEVERITY} - TRACE, DEBUG, INFO, WARN, ERROR, CRITICAL, STACK, FATAL, PANIC
//   %{Severity} - Trace, Debug, Info, Warn, Error, Critical, Stack, Fatal, Panic
//   %{severity} - trace, debug, info, warn, error, critical, stack, fatal, panic
//   %{SEV} - TRAC, DEBG, INFO, WARN, EROR, CRIT, STAK, FATL, PANC
//   %{Sev} - Trac, Debg, Info, Warn, Eror, Crit, Stak, Fatl, Panc
//   %{sev} - trac, debg, info, warn, eror, crit, stak, fatl, panc
//   %{S} - T, D, I, W, E, C, S, F, P
//   %{s} - t, d, i, w, e, c, s, f, p
//   %{Date} - Shorthand for 2006-01-02
//   %{Time} - Shorthand for 15:04:05
//   %{Time "<fmt>"} - Specify a format (read time.Format for details).
//                     Optimized formats: 2006/01/02, 15:04:05.000, 15:04:05.000000, 15:04:05.000000000
//   %{Unix} - Returns the number of seconds elapsed since January 1, 1970 UTC.
//   %{UnixNano} - Returns the number of nanoseconds elapsed since January 1, 1970 UTC.
//   %{FullFile} - Full source file path (e.g. /dev/project/file.go).
//   %{File} - The source file name (e.g. file.go).
//   %{ShortFile} - The short source file name (file without .go).
//   %{Line} - The source line number.
//   %{FullFunction} - The full source function including path. (e.g. /dev/project.(*Type).Function)
//   %{PkgFunction} - The source package and function (e.g. project.(*Type).Function)
//   %{Function} - The source function name (e.g. (*Type).Function)
//   %{Color "<fmt>"} - Specify a color (uses https://github.com/mgutz/ansi)
//   %{Color "<fmt>" "<severity>"} - Specify a color for a given severity (e.g. %{Color "red" "ERROR"})
//   %{Message} - The message.
//   %{SafeMessage} - Safe message. It will escape any character below ASCII 32. This helps prevent
//                    attacks like using 0x08 to backspace log entries.
//
// Example colors (see https://github.com/mgutz/ansi for more examples):
//   Added to mgutz/ansi:
//     %{Color "reset"}          - reset colors
//   Supported by mgutz/ansi
//     %{Color "red"}            - red
//     %{Color "red+b"}          - red bold
//     %{Color "red+B"}          - red blinking
//     %{Color "red+u"}          - red underline
//     %{Color "red+bh"}         - red bold bright
//     %{Color "red:white"}      - red on white
//     %{Color "red+b:white+h"}  - red bold on white bright
//     %{Color "red+B:white+h"}  - red blink on white bright
//
// All logging functions ending in "ln" are merely convenience functions
// and won't actually output another newline. This allows the
// formatters to handle a newline however they like (e.g. if you wanted to
// make a formatter that would output a streamed format without newlines).
//
package factorlog
