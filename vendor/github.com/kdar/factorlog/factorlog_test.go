package factorlog

import (
	"bytes"
	"log"
	"regexp"
	"strings"
	"testing"
)

var (
	// Test to make sure these types satisfy the Logger interface.
	_ Logger = &FactorLog{}
	_ Logger = Verbose{}
	// too bad this doesn't work
	//_ Logger = factorlog
)

var logTests = []struct {
	frmt string
	in   string
	out  []byte
}{
	{
		// we can't use every verb here, because the test will fail
		"%{FullFunction} [%{SEVERITY}:%{SEV}:%{File}:%{ShortFile}] %%{Message}%",
		"hello there!",
		[]byte("github.com/kdar/factorlog.TestLog [ERROR:EROR:factorlog_test.go:factorlog_test] %hello there!%\n"),
	},
	{
		"%{Message} %{File}",
		"hello there!",
		[]byte("hello there! factorlog_test.go\n"),
	},
}

func TestLog(t *testing.T) {
	buf := &bytes.Buffer{}
	for _, tt := range logTests {
		buf.Reset()
		f := New(buf, NewStdFormatter(tt.frmt))
		f.Errorln(tt.in)
		if !bytes.Equal(tt.out, buf.Bytes()) {
			t.Fatalf("\nexpected: %#v\ngot:      %#v", string(tt.out), buf.String())
		}
	}
}

func TestSetOutput(t *testing.T) {
	str := "hey\n"

	buf1 := &bytes.Buffer{}
	l := New(buf1, NewStdFormatter("%{Message}"))
	buf2 := &bytes.Buffer{}
	l.SetOutput(buf2)
	l.Print(str)

	if buf1.Len() > 0 {
		t.Fatal("the first buffer is suppose to be empty, but it's not")
	}

	if str != buf2.String() {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", str, buf2.String())
	}
}

func TestSetFormatter(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf, NewStdFormatter("%{Message}"))
	f := NewStdFormatter("2 %{Message}")

	l.Print("hey")

	if buf.String() != "hey\n" {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", "hey\n", buf.String())
	}

	buf.Reset()
	l.SetFormatter(f)
	l.Print("hey")

	if buf.String() != "2 hey\n" {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", "2 hey\n", buf.String())
	}
}

func TestOutputStack(t *testing.T) {
	r := regexp.MustCompile(`[\t\s]+.*?: f.output\(STACK, 1, nil, "hellothere"\)`)
	buf := &bytes.Buffer{}
	f := New(buf, NewStdFormatter("%{Message}"))

	f.output(STACK, 1, nil, "hellothere")
	lines := strings.Split(buf.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected stack trace to be at least 3 lines")
	}
	if lines[0] != "hellothere" {
		t.Fatalf("expected first line of Stack to be 'hellothere'")
	}
	if !r.MatchString(lines[2]) {
		t.Fatalf("regexp `%s` didn't match `%s`", r.String(), lines[2])
	}
}

func TestVerbosity(t *testing.T) {
	buf := &bytes.Buffer{}
	f := New(buf, NewStdFormatter("%{Message}"))

	f.SetVerbosity(2)
	f.V(3).Info("should not appear")
	if buf.Len() > 0 {
		t.Fatal("Verbosity set to 3, Info() called with verbosity of 3. Yet, we still got a log.")
	}

	buf.Reset()
	f.SetVerbosity(4)
	f.V(3).Info("should appear")
	if buf.Len() == 0 {
		t.Fatal("Verbosity set to 4, Info() called with verbosity of 3. We should have got a log.")
	}

}

//these methods have to be implemented for a Verbose{} to satisfy a Logger interface
//You really should not be using it like this
func TestChainedVerbosity(t *testing.T) {
	buf := &bytes.Buffer{}
	f := New(buf, NewStdFormatter("%{Message}"))
	//the V() calls here shouldnt do anything
	f.V(8).V(6).SetVerbosity(5)
	if f.IsV(5) != true {
		t.Fatal("Set verbosity to 5 through a chained V() call, but it was not actually set")

	}
	f.SetVerbosity(1)
	if f.V(1234).V(51234).IsV(1) != true {
		t.Fatal("Set verbosity to 1, checked it through a chained V() call, should still be 1")

	}

	f.SetVerbosity(2)
	f.V(8).V(3).Info("should not appear")
	if buf.Len() > 0 {
		t.Fatal("Verbosity set to 3, Info() called with verbosity of 8 then changed to3. Yet, we still got a log.")
	}

	buf.Reset()
	f.SetVerbosity(4)
	f.V(1000).V(3).Info("should appear")
	if buf.Len() == 0 {
		t.Fatal("Verbosity set to 4, Info() called with verbosity of 100 then changed to 3. We should have got a log.")
	}

}

type sevTestType int

const (
	sevTest_Set sevTestType = iota
	sevTest_MinMax
)

type sevTestFunc func(l *FactorLog, v ...interface{})

var severitiesTests = []struct {
	typ     sevTestType
	min     Severity // also used for l.SetSeverities()
	max     Severity
	funName string
	fun     sevTestFunc
	output  bool
}{
	{sevTest_Set, INFO, 0, "Info", (*FactorLog).Info, true},
	{sevTest_Set, PANIC, 0, "Info", (*FactorLog).Info, false},
	{sevTest_MinMax, WARN, CRITICAL, "Info", (*FactorLog).Info, false},
	{sevTest_MinMax, WARN, CRITICAL, "Warn", (*FactorLog).Warn, true},
	{sevTest_MinMax, WARN, CRITICAL, "Error", (*FactorLog).Error, true},
	{sevTest_MinMax, WARN, CRITICAL, "Critical", (*FactorLog).Critical, true},
	{sevTest_MinMax, WARN, CRITICAL, "Stack", (*FactorLog).Stack, false},
}

func TestSeverities(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf, NewStdFormatter("%{Message}"))

	for _, tt := range severitiesTests {
		buf.Reset()
		if tt.typ == sevTest_Set {
			l.SetSeverities(tt.min)
			tt.fun(l, "hello")
			if tt.output && buf.Len() == 0 {
				t.Fatalf("Severity set to %s. Called %s(). We didn't get a log we expected.", UcSeverityStrings[SeverityToIndex(tt.min)], tt.funName)
			} else if !tt.output && buf.Len() > 0 {
				t.Fatalf("Severity set to %s. Called %s(). We got a log we didn't expect.", UcSeverityStrings[SeverityToIndex(tt.min)], tt.funName)
			}
		} else if tt.typ == sevTest_MinMax {
			l.SetMinMaxSeverity(tt.min, tt.max)
			tt.fun(l, "hello")
			if tt.output && buf.Len() == 0 {
				t.Fatalf("Severity set to %s-%s. Called %s(). We didn't get a log we expected.", UcSeverityStrings[SeverityToIndex(tt.min)], UcSeverityStrings[SeverityToIndex(tt.max)], tt.funName)
			} else if !tt.output && buf.Len() > 0 {
				t.Fatalf("Severity set to %s-%s. Called %s(). We got a log we didn't expect.", UcSeverityStrings[SeverityToIndex(tt.min)], UcSeverityStrings[SeverityToIndex(tt.max)], tt.funName)
			}
		}
	}
}

// Ensure `std`'s format is correct.
func TestStdFormat(t *testing.T) {
	output := std.formatter.Format(fmtTestsContext)
	expect := "2014-01-08 23:27:14 hello there!\n"
	if string(output) != expect {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", expect, string(output))
	}
}

func BenchmarkGoLogBuffer(b *testing.B) {
	buf := &bytes.Buffer{}
	l := log.New(buf, "", log.Ldate|log.Ltime|log.Lshortfile)
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		l.Print("hey")
	}
}

func BenchmarkFactorLogBuffer(b *testing.B) {
	buf := &bytes.Buffer{}
	l := New(buf, NewStdFormatter("%{Date} %{Time} %{File}:%{Line}: %{Message}"))
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		l.Info("hey")
	}
}
