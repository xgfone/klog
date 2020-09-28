# klog [![Build Status](https://travis-ci.org/xgfone/klog.svg?branch=master)](https://travis-ci.org/xgfone/klog) [![GoDoc](https://godoc.org/github.com/xgfone/klog?status.svg)](http://godoc.org/github.com/xgfone/klog) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/klog/master/LICENSE)

Package `klog` provides an simple, flexible, extensible, powerful and structured logging tool based on the level, which has done the better balance between the flexibility and the performance. It is inspired by [log15](https://github.com/inconshreveable/log15), [logrus](https://github.com/sirupsen/logrus), [go-kit](https://github.com/go-kit/kit) and [zerolog](github.com/rs/zerolog).

**API has been stable.** The current is `v4.x` and support Go `1.7+`.


## Features

- The better performance.
- Lazy evaluation of expensive operations.
- Simple, Flexible, Extensible, Powerful and Structured.
- Avoid to allocate the memory on heap as far as possible.
- Child loggers which inherit and add their own private context.
- Built-in support for logging to files, syslog, and the network. See `Writer`.

`klog` supports two kinds of the logger interfaces:
```go
// Logger is the Key-Value logger interface.
type Logger interface {
	Trace(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(mst string, fields ...Field)
}

// Loggerf is the format logger interface.
type Loggerf interface {
	Tracef(msgfmt string, args ...interface{})
	Debugf(msgfmt string, args ...interface{})
	Infof(msgfmt string, args ...interface{})
	Warnf(msgfmt string, args ...interface{})
	Errorf(msgfmt string, args ...interface{})
	Fatalf(msgfmt string, args ...interface{})
}
```

## Example

```go
package main

import (
	"os"

	"github.com/xgfone/klog/v4"
)

func main() {
	opts := []klog.EncoderOption{
		klog.Quote(),
		klog.EncodeTime("t"),
		klog.EncodeLevel("lvl"),
		klog.EncodeLogger("logger"),
	}

	log := klog.New("name").
		WithEncoder(klog.TextEncoder(klog.SafeWriter(klog.StreamWriter(os.Stdout)), opts...)).
		WithLevel(klog.LvlWarn).
		WithCtx(klog.Caller("caller"))

	log.Info("log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
	log.Error("log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))

	// Output:
	// t=1601185933 logger=name lvl=ERROR caller=main.go:23 key1=value1 key2=value2 msg="log msg"
}
```

`klog` supplies the default global logger and some convenient functions based on the level:
```go
// Emit the log with the fields.
func Trace(msg string, fields ...Field)
func Debug(msg string, fields ...Field)
func Info(msg string, fields ...Field)
func Warn(msg string, fields ...Field)
func Error(msg string, fields ...Field)
func Fatal(msg string, fields ...Field)

// Emit the log with the formatter.
func Printf(format string, args ...interface{})
func Tracef(format string, args ...interface{})
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
func Ef(err error, format string, args ...interface{})
```

For example,
```go
package main

import (
	"fmt"

	"github.com/xgfone/klog/v4"
)

func main() {
	// Initialize the default logger.
	klog.DefalutLogger = klog.WithLevel(klog.LvlWarn).WithCtx(klog.Caller("caller"))

	// Emit the log with the fields.
	klog.Info("msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
	klog.Error("msg", klog.F("key1", "value1"), klog.F("key2", "value2"))

	// Emit the log with the formatter.
	klog.Infof("%s log msg", "infof")
	klog.Errorf("%s log msg", "errorf")
	klog.Ef(fmt.Errorf("error"), "%s log msg", "errorf")

	// Output:
	// t=2020-09-27T23:52:35.63282+08:00 lvl=ERROR caller=main.go:15 key1=value1 key2=value2 msg=msg
	// t=2020-09-27T23:52:35.64482+08:00 lvl=ERROR caller=main.go:19 msg="errorf log msg"
	// t=2020-09-27T23:52:35.64482+08:00 lvl=ERROR caller=main.go:20 err=error msg="errorf log msg"
}
```


### Encoder

```go
type Encoder interface {
	// Writer returns the writer.
	Writer() Writer

	// SetWriter resets the writer.
	SetWriter(Writer)

	// Encode encodes the log record and writes it into the writer.
	Encode(Record)
}
```

This pakcage has implemented four kinds of encoders, `NothingEncoder`, `TextEncoder`, `JSONEncoder` and `LevelEncoder`. It will use `TextEncoder` by default.


### Writer

```go
type Writer interface {
	WriteLevel(level Level, data []byte) (n int, err error)
	io.Closer
}
```

All implementing the interface `Writer` are a Writer.

There are some built-in writers, such as `DiscardWriter`, `FailoverWriter`, `LevelWriter`, `NetWriter`, `SafeWriter`, `SplitWriter`, `StreamWriter`, `FileWriter`. `FileWriter` uses `SizedRotatingFile` to write the log to the file rotated based on the size.

```go
package main

import (
	"fmt"

	"github.com/xgfone/klog/v4"
)

func main() {
	file, err := klog.FileWriter("test.log", "100M", 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	klog.DefalutLogger.Encoder.SetWriter(file)
	klog.Info("hello world", klog.F("key", "value"))

	// Output to file test.log:
	// t=2020-09-27T23:56:04.0691608+08:00 lvl=INFO key=value msg="hello world"
}
```


### Lazy evaluation

`Field` supports the lazy evaluation, such as `F("key", func() interface{} {return "value"})`. And there are some built-in lazy `Field`, such as `Caller()`, `CallerStack()`.


## Performance

The log framework itself has no any performance costs and the key of the bottleneck is the encoder.

```
Dell Vostro 3470
Intel Core i5-7400 3.0GHz
8GB DDR4 2666MHz
Windows 10
Go 1.13.4
```

**Benchmark Package:**

|               Function               |      ops      | ns/op | bytes/opt | allocs/op
|--------------------------------------|--------------:|------:|-----------|----------
|BenchmarkKlogNothingEncoder-4         | 273, 440, 714 | 4     |     0     |    0
|BenchmarkKlogTextEncoder-4            |  30, 770, 728 | 43    |     0     |    0
|BenchmarkKlogJSONEncoder-4            |  41, 626, 033 | 27    |     0     |    0
|BenchmarkKlogTextEncoder10CtxFields-4 |  10, 344, 880 | 149   |     0     |    0
|BenchmarkKlogJSONEncoder10CtxFields-4 |   7, 692, 381 | 165   |     0     |    0
