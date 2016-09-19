package factorlog

import (
	"testing"
	"time"
)

var fmtTestsContext = LogContext{
	Time:     time.Unix(0, 1389223634123456789).In(time.UTC),
	Severity: PANIC,
	File:     "path/to/testing.go",
	Line:     391,
	Format:   nil,
	Args:     []interface{}{"hello there!"},
	Function: "some crazy/path.path/pkg.(*Type).Function",
	Pid:      1234,
}

var std2FmtTests = []struct {
	context LogContext
	frmt    string
	out     string
}{
	{
		fmtTestsContext,
		"%{SEVERITY}",
		"PANIC\n",
	},
	{
		fmtTestsContext,
		"%{Severity}",
		"Panic\n",
	},
	{
		fmtTestsContext,
		"%{severity}",
		"panic\n",
	},
	{
		fmtTestsContext,
		"%{SEV}",
		"PANC\n",
	},
	{
		fmtTestsContext,
		"%{Sev}",
		"Panc\n",
	},
	{
		fmtTestsContext,
		"%{sev}",
		"panc\n",
	},
	{
		fmtTestsContext,
		"%{S}",
		"P\n",
	},
	{
		fmtTestsContext,
		"%{s}",
		"p\n",
	},
	{
		fmtTestsContext,
		"",
		"",
	},
	{
		fmtTestsContext,
		"%{Date}",
		"2014-01-08\n",
	},
	{
		fmtTestsContext,
		"%{Time}",
		"23:27:14\n",
	},
	{
		fmtTestsContext,
		`%{Time "15:04:05"}`,
		"23:27:14\n",
	},
	{
		fmtTestsContext,
		`%{Time "2006/01/02"}`,
		"2014/01/08\n",
	},
	{
		fmtTestsContext,
		`%{Time "15:04:05.000"}`,
		"23:27:14.123\n",
	},
	{
		fmtTestsContext,
		`%{Time "15:04:05.000000"}`,
		"23:27:14.123456\n",
	},
	{
		fmtTestsContext,
		`%{Time "15:04:05.000000000"}`,
		"23:27:14.123456789\n",
	},
	{
		fmtTestsContext,
		"%{Unix}",
		"1389223634\n",
	},
	{
		fmtTestsContext,
		"%{UnixNano}",
		"1389223634123456789\n",
	},
	{
		fmtTestsContext,
		"%{FullFile}",
		"path/to/testing.go\n",
	},
	{
		fmtTestsContext,
		"%{File}",
		"testing.go\n",
	},
	{
		fmtTestsContext,
		"%{ShortFile}",
		"testing\n",
	},
	{
		fmtTestsContext,
		"%{Line}",
		"391\n",
	},
	{
		fmtTestsContext,
		"%{FullFunction}",
		"some crazy/path.path/pkg.(*Type).Function" + "\n",
	},
	{
		fmtTestsContext,
		"%{PkgFunction}",
		"pkg.(*Type).Function\n",
	},
	{
		fmtTestsContext,
		"%{Function}",
		"(*Type).Function\n",
	},
	{
		LogContext{Function: "main.main"},
		"%{FullFunction}",
		"main.main\n",
	},
	{
		LogContext{Function: "main.main"},
		"%{PkgFunction}",
		"main.main\n",
	},
	{
		LogContext{Function: "main.main"},
		"%{Function}",
		"main\n",
	},
	{
		LogContext{Function: "main.Type.main"},
		"%{FullFunction}",
		"main.Type.main\n",
	},
	{
		LogContext{Function: "main.Type.main"},
		"%{PkgFunction}",
		"main.Type.main\n",
	},
	{
		LogContext{Function: "main.Type.main"},
		"%{Function}",
		"Type.main\n",
	},
	{
		LogContext{Function: ""},
		"%{FullFunction}",
		"",
	},
	{
		LogContext{Function: ""},
		"%{PkgFunction}",
		"",
	},
	{
		LogContext{Function: ""},
		"%{Function}",
		"",
	},
	{
		fmtTestsContext,
		`%{Color "red"}`,
		"\x1b[31m\n",
	},
	{
		fmtTestsContext,
		`%{Color "red+b:blue"}`,
		"\x1b[1;31;44m\n",
	},
	{
		fmtTestsContext,
		`%{Color "yellow"}hey%{Color "reset"}`,
		"\x1b[33mhey\x1b[0m\n",
	},
	{
		fmtTestsContext,
		`%{Color "red" "PANIC"}`,
		"\x1b[31m\n",
	},
	{
		fmtTestsContext,
		`%{Color "red" "INFO"}`,
		"",
	},
	{
		fmtTestsContext,
		"%{Message}",
		"hello there!\n",
	},
	{
		fmtTestsContext,
		"%{Message} %{File}:%{Line}",
		"hello there! testing.go:391\n",
	},
	{
		fmtTestsContext,
		"just text here",
		"just text here\n",
	},
	{
		fmtTestsContext,
		"hey%{Date}%{Time}there",
		"hey2014-01-0823:27:14there\n",
	},
	{
		LogContext{Args: []interface{}{"hey\x08\x08\x08there"}},
		"%{SafeMessage}",
		"hey\\x08\\x08\\x08there\n",
	},
}

func TestStdFormatter(t *testing.T) {
	for _, tt := range std2FmtTests {
		f := NewStdFormatter(tt.frmt)
		out := string(f.Format(tt.context))
		if tt.out != out {
			t.Fatalf("\nfor: %v\nexpected: %#v\ngot:      %#v", tt.frmt, tt.out, out)
		}
	}
}

func TestStdShouldRuntimeCaller(t *testing.T) {
	f := NewStdFormatter("[%{Date} %{Time}]")
	if f.ShouldRuntimeCaller() {
		t.Fatalf("Formatter should not need to call runtime.Caller().")
	}
	f = NewStdFormatter("%{Line}")
	if !f.ShouldRuntimeCaller() {
		t.Fatalf("Formatter should need to call runtime.Caller().")
	}
}

func BenchmarkStdFormatter(b *testing.B) {
	// var m runtime.MemStats

	f := NewStdFormatter(`[%{Date} %{Time}] [%{SEVERITY}:%{File}:%{Line}] %{Message}`)

	// l := len(fmtTestsContext.Message)
	for x := 0; x < b.N; x++ {
		f.Format(fmtTestsContext)
		// if x%500 == 0 {
		// 	fmtTestsContext.Message = fmtTestsContext.Message[:l]
		// 	l--
		// 	if l < 0 {
		// 		l = 0
		// 	}
		// }
	}

	// runtime.ReadMemStats(&m)
	// fmt.Printf("%d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc,
	// 	m.HeapIdle, m.HeapReleased)
}
