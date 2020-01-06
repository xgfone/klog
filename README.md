# klog [![Build Status](https://travis-ci.org/xgfone/klog.svg?branch=master)](https://travis-ci.org/xgfone/klog) [![GoDoc](https://godoc.org/github.com/xgfone/klog?status.svg)](http://godoc.org/github.com/xgfone/klog) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/klog/master/LICENSE)

Package `klog` provides an simple, flexible, extensible, powerful and structured logging tool based on the level, which has done the better balance between the flexibility and the performance. It is inspired by [log15](https://github.com/inconshreveable/log15), [logrus](https://github.com/sirupsen/logrus), [go-kit](https://github.com/go-kit/kit) and [zerolog](github.com/rs/zerolog).

**API has been stable.** The current is `v3.x` and support Go `1.x`.


## Features

- The better performance.
- Lazy evaluation of expensive operations.
- Simple, Flexible, Extensible, Powerful and Structured.
- Avoid to allocate the memory on heap as far as possible.
- Child loggers which inherit and add their own private context.
- Built-in support for logging to files, syslog, and the network. See `Writer`.


## Example

```go
package main

import (
	"os"

	"github.com/xgfone/klog/v2"
)

func main() {
	opts := []klog.EncoderOption{
		klog.Quote(),
		klog.EncodeTime("t"),
		klog.EncodeLevel("lvl"),
		klog.EncodeLogger("logger"),
	}

	log := klog.New("loggername").
		WithEncoder(klog.TextEncoder(klog.SafeWriter(klog.StreamWriter(os.Stdout)), opts...)).
		WithLevel(klog.LvlWarn).
		WithCtx(klog.F("caller", klog.Caller()))

	log.Log(klog.LvlInfo, "log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
	log.Log(klog.LvlError, "log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))

	// Output:
	// t=1574056059 logger=loggername lvl=ERROR caller=main.go:23 key1=value1 key2=value2 msg="log msg"
}
```

`klog` supplies the default global logger and some convenient functions based on the level:
```go
// Emit the log with the fields.
func Log(level Level, msg string, fields ...Field)
func Trace(msg string, fields ...Field)
func Debug(msg string, fields ...Field)
func Info(msg string, fields ...Field)
func Warn(msg string, fields ...Field)
func Error(msg string, fields ...Field)

// Emit the log with the formatter.
func Printf(format string, args ...interface{})
func Tracef(format string, args ...interface{})
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Ef(err error, format string, args ...interface{})
```

For example,
```go
package main

import (
	"fmt"

	"github.com/xgfone/klog/v2"
)

func main() {
	// Initialize the default logger.
	log := klog.WithLevel(klog.LvlWarn).WithCtx(klog.F("caller", klog.Caller()))
	klog.SetDefaultLogger(log)

	// Emit the log with the fields.
	klog.Info("msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
	klog.Error("msg", klog.F("key1", "value1"), klog.F("key2", "value2"))

	// Emit the log with the formatter.
	klog.Infof("%s log msg", "infof")
	klog.Errorf("%s log msg", "errorf")
	klog.Ef(fmt.Errorf("error"), "%s log msg", "errorf")

	// Output:
	// t=2019-11-18T14:01:08.7345586+08:00 lvl=ERROR caller=main.go:15 key1=value1 key2=value2 msg="msg"
	// t=2019-11-18T14:01:08.735969+08:00 lvl=ERROR caller=main.go:18 msg="errorf log msg"
	// t=2019-11-18T14:01:08.7360115+08:00 lvl=ERROR caller=main.go:19 err=error msg="errorf log msg"
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

This pakcage has implemented four kinds of encoders, `NothingEncoder`, `TextEncoder`, `JSONEncoder`. It will use `TextEncoder` by default, but you can set it to others by `SetEncoder` or `WithEncoder`.


### Writer

```go
type Writer interface {
	Write(level Level, data []byte) (n int, err error)
}
```

All implementing the interface `Writer` are a Writer.

There are some built-in writers, such as `DiscardWriter`, `FailoverWriter`, `LevelWriter`, `NetWriter`, `SafeWriter`, `SplitWriter`, `StreamWriter`, `FileWriter`. `FileWriter` uses `SizedRotatingFile` to write the log to the file rotated based on the size.

```go
package main

import (
	"fmt"

	"github.com/xgfone/klog/v2"
)

func main() {
	file, err := klog.FileWriter("test.log", "100M", 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	klog.GetEncoder().SetWriter(file)
	klog.Info("hello world", klog.F("key", "value"))

	// Output to file test.log:
	// t=2019-11-18T14:18:01.479374+08:00 lvl=INFO key=value msg="hello world"
}
```

You can use `WriterFunc` or `WriteCloserFunc` to implement the interface `Writer` or `WriteCloser`.


### Lazy evaluation

`Field` supports the lazy evaluation, and you can use it by `LazyField`, such as `LazyField("key", func() interface{} {return "value"})`. And there are some built-in lazy `Field`, such as `Caller()`, `CallerStack()`.


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

|               Function               |    ops    |   ns/op   | bytes/opt |  allocs/op
|--------------------------------------|-----------|-----------|-----------|-------------
|BenchmarkKlogNothingEncoder-4         | 194001346 | 6 ns/op   |  0 B/op   | 0 allocs/op
|BenchmarkKlogTextEncoder-4            |  42973323 | 24 ns/op  |  0 B/op   | 0 allocs/op
|BenchmarkKlogJSONEncoder-4            |  57550428 | 21 ns/op  |  0 B/op   | 0 allocs/op
|BenchmarkKlogTextEncoder10CtxFields-4 |  10026812 | 107 ns/op |  0 B/op   | 0 allocs/op
|BenchmarkKlogJSONEncoder10CtxFields-4 |   6574923 | 221 ns/op |  0 B/op   | 0 allocs/op
