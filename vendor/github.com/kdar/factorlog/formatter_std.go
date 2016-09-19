package factorlog

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/mgutz/ansi"
)

const pathSeparator = '/'

// Can hold 63 flags
type fmtVerb uint64

const (
	vSTRING fmtVerb = 1 << iota
	vSEVERITY
	vSeverity
	vseverity
	vSEV
	vSev
	vsev
	vS
	vs
	vDate
	vTime
	vUnix
	vUnixNano
	vFullFile
	vFile
	vShortFile
	vLine
	vFullFunction
	vPkgFunction
	vFunction
	vColor
	vMessage
	vSafeMessage
)

const (
	// If formatter.flags is set to any of these, we need runtime.Caller
	vRUNTIME_CALLER = int(vFullFile |
		vFile |
		vShortFile |
		vLine |
		vFullFunction |
		vPkgFunction |
		vFunction)
)

const (
	fTime_Default  = 0
	fTime_Provided = 1 << iota
	fTime_LogDate
	fTime_StampMilli
	fTime_StampMicro
	fTime_StampNano
)

var (
	formatRe = regexp.MustCompile(`%{([A-Za-z]+)(?:\s(.*?))?}`)
	argsRe   = regexp.MustCompile(`(?:"(.*?)")`)
	verbMap  = map[string]fmtVerb{
		"SEVERITY":     vSEVERITY,
		"Severity":     vSeverity,
		"severity":     vseverity,
		"SEV":          vSEV,
		"Sev":          vSev,
		"sev":          vsev,
		"S":            vS,
		"s":            vs,
		"Date":         vDate,
		"Time":         vTime,
		"Unix":         vUnix,
		"UnixNano":     vUnixNano,
		"FullFile":     vFullFile,
		"File":         vFile,
		"ShortFile":    vShortFile,
		"Line":         vLine,
		"FullFunction": vFullFunction,
		"PkgFunction":  vPkgFunction,
		"Function":     vFunction,
		"Color":        vColor,
		"Message":      vMessage,
		"SafeMessage":  vSafeMessage,
	}
	timeMap = map[string]int{
		"15:04:05":           fTime_Default,
		"2006/01/02":         fTime_LogDate,
		"15:04:05.000":       fTime_StampMilli,
		"15:04:05.000000":    fTime_StampMicro,
		"15:04:05.000000000": fTime_StampNano,
	}
)

type part struct {
	verb  fmtVerb
	value string
	args  []string
	flags int
}

type StdFormatter struct {
	// the original format
	frmt string
	// a slice depicting each part of the format
	// we build the final []byte from this
	parts []*part
	// temporary buffer to help in formatting.
	// initialized by newFormatter
	tmp []byte
	// temporary buffer used for safe messages.
	stmp []byte
	// flags represents all the verbs we used.
	// this is useful in speeding things up like
	// not calling runtime.Caller if we don't have
	// a format string that requires it
	flags int
}

// Available verbs:
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
func NewStdFormatter(frmt string) *StdFormatter {
	f := &StdFormatter{
		frmt: frmt,
		tmp:  make([]byte, 64),
		stmp: make([]byte, 0, 64),
	}

	matches := formatRe.FindAllStringSubmatchIndex(frmt, -1)
	prev := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		verb := frmt[m[2]:m[3]]

		// Try to get any arguments passed
		var args []string
		if m[4] != -1 {
			allargs := frmt[m[4]:m[5]]
			pargs := argsRe.FindAllStringSubmatch(allargs, -1)
			for _, arg := range pargs {
				args = append(args, arg[1])
			}
		}

		if start > prev {
			f.appendString(frmt[prev:start])
		}

		if v, ok := verbMap[verb]; ok {
			switch v {
			case vColor:
				if len(args) > 0 {
					if args[0] == "reset" {
						f.appendString(ansi.Reset)
					} else {
						code := ansi.ColorCode(args[0])
						if len(args) == 2 {
							// If we have two arguments, that means they
							// specified a severity this color applies to.
							// So we have to add the part.
							severity := StringToSeverity(args[1])
							f.parts = append(f.parts, &part{
								verb:  vColor,
								value: code,
								flags: int(severity),
							})
						} else {
							// We only got one argument so we can just append
							// the code as a string.
							f.appendString(code)
						}
					}
				}
			case vTime:
				f.flags |= int(v)
				if len(args) > 0 {
					// Some optimizations for Time
					opt_part := &part{
						verb: v,
					}
					if ftime, ok := timeMap[args[0]]; ok {
						opt_part.flags = ftime
						f.parts = append(f.parts, opt_part)
					} else {
						opt_part.flags = fTime_Provided
						opt_part.args = args
						f.parts = append(f.parts, opt_part)
					}
				} else {
					f.appendDefault(v, args)
				}
			default:
				f.appendDefault(v, args)
			}
		}

		prev = end
	}

	if frmt[prev:] != "" {
		f.appendString(frmt[prev:])
	}

	return f
}

func (f *StdFormatter) ShouldRuntimeCaller() bool {
	return f.flags&(vRUNTIME_CALLER) != 0
}

func (f *StdFormatter) appendString(s string) {
	if len(s) > 0 {
		f.parts = append(f.parts, &part{
			verb:  vSTRING,
			value: s,
		})
	}
}

func (f *StdFormatter) appendDefault(verb fmtVerb, args []string) {
	f.flags |= int(verb)
	f.parts = append(f.parts, &part{
		verb: verb,
		args: args,
	})
}

func (f *StdFormatter) Format(context LogContext) []byte {
	buf := &bytes.Buffer{}
	for _, p := range f.parts {
		switch p.verb {
		case vSTRING:
			buf.WriteString(p.value)
		case vSEVERITY:
			buf.WriteString(UcSeverityStrings[SeverityToIndex(context.Severity)])
		case vSeverity:
			buf.WriteString(CapSeverityStrings[SeverityToIndex(context.Severity)])
		case vseverity:
			buf.WriteString(LcSeverityStrings[SeverityToIndex(context.Severity)])
		case vSEV:
			buf.WriteString(UcShortSeverityStrings[SeverityToIndex(context.Severity)])
		case vSev:
			buf.WriteString(CapShortSeverityStrings[SeverityToIndex(context.Severity)])
		case vsev:
			buf.WriteString(LcShortSeverityStrings[SeverityToIndex(context.Severity)])
		case vS:
			buf.WriteString(UcShortestSeverityStrings[SeverityToIndex(context.Severity)])
		case vs:
			buf.WriteString(LcShortestSeverityStrings[SeverityToIndex(context.Severity)])
		case vDate:
			year, month, day := context.Time.Date()
			NDigits(&f.tmp, 4, 0, year)
			f.tmp[4] = '-'
			TwoDigits(&f.tmp, 5, int(month))
			f.tmp[7] = '-'
			TwoDigits(&f.tmp, 8, day)
			buf.Write(f.tmp[:10])
		case vTime:
			// Some optimization cases.
			switch {
			case p.flags == fTime_LogDate:
				year, month, day := context.Time.Date()
				NDigits(&f.tmp, 4, 0, year)
				f.tmp[4] = '/'
				TwoDigits(&f.tmp, 5, int(month))
				f.tmp[7] = '/'
				TwoDigits(&f.tmp, 8, day)
				buf.Write(f.tmp[:10])
			case p.flags&(fTime_StampMilli|fTime_StampMicro|fTime_StampNano) != 0:
				hour, min, sec := context.Time.Clock()
				TwoDigits(&f.tmp, 0, hour)
				f.tmp[2] = ':'
				TwoDigits(&f.tmp, 3, min)
				f.tmp[5] = ':'
				TwoDigits(&f.tmp, 6, sec)
				f.tmp[8] = '.'
				// Depending on what kind of stamp we're dealing with, we have
				// to output the correct precision.
				if p.flags == fTime_StampMilli {
					NDigits(&f.tmp, 3, 9, context.Time.Nanosecond()/1000000)
					buf.Write(f.tmp[:12])
				} else if p.flags == fTime_StampMicro {
					NDigits(&f.tmp, 6, 9, context.Time.Nanosecond()/1000)
					buf.Write(f.tmp[:15])
				} else if p.flags == fTime_StampNano {
					NDigits(&f.tmp, 9, 9, context.Time.Nanosecond())
					buf.Write(f.tmp[:18])
				}
			case p.flags == fTime_Provided:
				buf.WriteString(context.Time.Format(p.args[0]))
			default:
				hour, min, sec := context.Time.Clock()
				TwoDigits(&f.tmp, 0, hour)
				f.tmp[2] = ':'
				TwoDigits(&f.tmp, 3, min)
				f.tmp[5] = ':'
				TwoDigits(&f.tmp, 6, sec)
				buf.Write(f.tmp[:8])
			}
		case vUnix:
			n := I64toa(&f.tmp, 0, context.Time.Unix())
			buf.Write(f.tmp[:n])
		case vUnixNano:
			n := I64toa(&f.tmp, 0, context.Time.UnixNano())
			buf.Write(f.tmp[:n])
		case vFullFile:
			buf.WriteString(context.File)
		case vFile, vShortFile:
			file := context.File
			if len(file) == 0 {
				file = "???"
			} else {
				slash := len(file) - 1
				for ; slash >= 0; slash-- {
					if file[slash] == pathSeparator {
						break
					}
				}
				if slash >= 0 {
					file = file[slash+1:]
				}
			}

			if p.verb == vShortFile {
				file = file[:len(file)-3]
			}

			buf.WriteString(file)
		case vLine:
			n := Itoa(&f.tmp, 0, context.Line)
			buf.Write(f.tmp[:n])
		case vFullFunction:
			buf.WriteString(context.Function)
		case vPkgFunction:
			fun := context.Function
			slash := len(fun) - 1
			for ; slash >= 0; slash-- {
				if fun[slash] == pathSeparator {
					break
				}
			}
			if slash >= 0 {
				fun = fun[slash+1:]
			}

			buf.WriteString(fun)
		case vFunction:
			fun := context.Function

			slash := len(fun) - 1
			lastDot := -1
			for ; slash >= 0; slash-- {
				if fun[slash] == pathSeparator {
					break
				} else if fun[slash] == '.' {
					lastDot = slash
				}
			}

			fun = fun[lastDot+1:]
			buf.WriteString(fun)
		case vColor:
			// We must have args when we get here because of
			// the parser ensuring this. No need testing for it.
			if Severity(p.flags) == context.Severity {
				buf.WriteString(p.value)
			}
		case vMessage:
			if context.Format != nil {
				buf.WriteString(fmt.Sprintf(*context.Format, context.Args...))
			} else {
				buf.WriteString(fmt.Sprint(context.Args...))
			}
		case vSafeMessage:
			message := ""
			if context.Format != nil {
				message = fmt.Sprintf(*context.Format, context.Args...)
			} else {
				message = fmt.Sprint(context.Args...)
			}

			f.stmp = f.stmp[:0]
			l := len(message)
			ca := cap(f.stmp)
			if l > ca {
				f.stmp = make([]byte, 0, l)
			} else if ca > 8000 { // don't let memory usage get too big
				f.stmp = f.stmp[0:0:l]
			}

			for _, c := range message {
				if int(c) < 32 {
					f.tmp[0] = '\\'
					f.tmp[1] = 'x'
					TwoDigits(&f.tmp, 2, int(c))
					f.stmp = append(f.stmp, f.tmp[:4]...)
				} else {
					f.stmp = append(f.stmp, byte(c))
				}
			}
			buf.Write(f.stmp)
		}
	}

	b := buf.Bytes()
	if buf.Len() > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}

	return b
}
