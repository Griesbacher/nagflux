FactorLog
=========

FactorLog is a fast logging infrastructure for Go that provides numerous logging functions for whatever your style may be. It could easily be a replacement for Go's log in the standard library (though it doesn't support functions such as `SetFlags()`).

It has a modular formatter interface, and even have a GELF and glog formatter in the [contrib](https://github.com/kdar/factorlog-contrib). 

Documentation here: [http://godoc.org/github.com/kdar/factorlog](http://godoc.org/github.com/kdar/factorlog)

![factorlog](http://puu.sh/6jPEt.png "FactorLog")

## Features

- Various log severities: TRACE, DEBUG, INFO, WARN, ERROR, CRITICAL, STACK, FATAL, PANIC
- Configurable, formattable logger. Check [factorlog-contrib](https://github.com/kdar/factorlog-contrib) for examples.
- Modular formatter. Care about speed? Use [GlogFormatter](https://github.com/kdar/factorlog-contrib/tree/master/glog) or roll your own.
- Designed with speed in mind (it's really fast).
- Many logging functions to fit your style of logging. (Trace, Tracef, Traceln, etc...)
- Supports colors.
- Settable verbosity like [glog](https://github.com/golang/glog).
- Filter by severity.
- Used in a production system, so it will get some love.

## Motivation

There are many good logging libraries out there but none of them worked the way I wanted them to. For example, some libraries have an API like `log.Info()`, but behind the scenes it's using fmt.Sprintf. What this means is I couldn't do things like `log.Info(err)`. I would instead have to do `log.Info("%s", err.Error())`. This kept biting me as I was coding. In FactorLog, `log.Info` behaves exactly like `fmt.Sprint`. `log.Infof` uses `fmt.Sprintf` and `log.Infoln` uses `fmt.Sprintln`.

I really like [glog](https://github.com/golang/glog), but I don't like that it takes over your command line arguments. I may implement more of its features into FactorLog.

I also didn't want a library that read options from a configuration file. You could easily handle that yourself if you wanted to. FactorLog doesn't include any code for logging to different backends (files, syslog, etc...). The reason for this is I was structuring this after [http://12factor.net/](http://12factor.net/). There are many programs out there that can parse your log output from stdout/stderr and redirect it to the appropriate place. However, FactorLog doesn't prevent you from writing this code yourself. I think it would be better if there was a third party library that did backend work itself. Then every logging library could benefit from it.

## Examples

You can use the API provided by the package itself just like Go's log does:

```
package main

import (
  log "github.com/kdar/factorlog"
)

func main() {
  log.Println("Hello there!")
}
```

Or, you can make your own log from `FactorLog`:

```
package main

import (
  "github.com/kdar/factorlog"
  "os"
)

func main() {
  log := factorlog.New(os.Stdout, factorlog.NewStdFormatter("%{Date} %{Time} %{File}:%{Line} %{Message}"))
  log.Println("Hello there!")
}
```

Check the examples/ directory for more.

## Documentation

[http://godoc.org/github.com/kdar/factorlog](http://godoc.org/github.com/kdar/factorlog)

## API Stability

I'm using this in software that is rapidly changing, so I haven't stabilized the API of this just yet.
