# klog [![Build Status](https://travis-ci.org/xgfone/klog.svg?branch=master)](https://travis-ci.org/xgfone/klog) [![GoDoc](https://godoc.org/github.com/xgfone/klog?status.svg)](http://godoc.org/github.com/xgfone/klog) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/klog/master/LICENSE)

Package `klog` provides an simple, flexible, extensible, powerful and structured logging tool based on the level, which has done the better balance between the flexibility and the performance. It is inspired by [log15](https://github.com/inconshreveable/log15), [logrus](https://github.com/sirupsen/logrus), [go-kit](https://github.com/go-kit/kit), [logger](https://github.com/xgfone/logger) and [log](https://github.com/go-playground/log).

See the [GoDoc](https://godoc.org/github.com/xgfone/klog).

**API has been stable.** The current is `v1.x`.


## Prerequisite

Now `klog` requires Go `1.x`.


## Features

- The better performance.
- Lazy evaluation of expensive operations.
- The memory allocation is based on stack, not heap.
- Simple, Flexible, Extensible, Powerful and Structured.
- Child loggers which inherit and add their own private context.
- Built-in support for logging to files, syslog, and the network. See `Writer`.


## Example

```go
package main

import "github.com/xgfone/klog"

func main() {
	log := klog.New().WithLevel(klog.LvlWarn).WithKv("key1", "value1")

	field1 := klog.Field{Key: "key2", Value: "value2"}
	field2 := klog.Field{Key: "key3", Value: "value3"}

	// Style 1:
	log.Info().K("key4", "value4").Print("don't output")
	log.Error(field1, field2).K("key4", "value4").Printf("will output %s", "placeholder")

	// Style 2:
	log.K("key4", "value4").Infof("don't output")
	log.F(field1, field2).K("key4", "value4").Errorf("will output %s", "placeholder")
	log.Warnf("output '%s' log", "WARN") // You can emit log directly without key-value pairs.

	// Output:
	// t=2019-05-24T09:31:27.2592259+08:00 lvl=ERROR key1=value1 key2=value2 key3=value3 key4=value4 msg=will output placeholder
	// t=2019-05-24T09:31:27.2712334+08:00 lvl=ERROR key1=value1 key2=value2 key3=value3 key4=value4 msg=will output placeholder
	// t=2019-05-24T09:31:27.2712334+08:00 lvl=WARN key1=value1 msg=output 'WARN' log
}
```

Notice: `klog` supplies two kinds of log styles, `Log`(__**Style 1**__) and `LLog` (__**Style 2**__).

Furthermore, `klog` has built in a global logger, `Std`, which is equal to `klog.New()`, and you can use it and its exported function. **Suggestion:** You should use these functions instead.

**Style 1:** `F()`, `K()`, `Tracef()`, `Debugf()`, `Infof()`, `Warnf()`, `Errorf()`, `Panicf()`, `Fatalf()`, or `Lf(level)`.

**Style 2:** `Trace()`, `Debug()`, `Info()`, `Warn()`, `Error()`, `Panic()`, `Fatal()`, or `L(level)`.

```go
package main

import "github.com/xgfone/klog"

func main() {
	klog.Std = klog.Std.WithLevel(klog.LvlWarn)
	// Or
	// klog.SetLevel(klog.LvlWarn)

	// Style 1:
	klog.Info().K("key", "value").Msg("don't output")
	klog.Error().K("key", "value").Msgf("will output %s", "placeholder")

	// Style 2:
	klog.K("key", "value").Infof("don't output")
	klog.K("key", "value").Errorf("will output %s", "placeholder")
	klog.Warnf("output '%s' log", "WARN") // You can emit log directly without key-value pairs.

	// Output:
	// t=2019-05-24T09:46:50.9758631+08:00 lvl=ERROR key=value msg=will output placeholder
	// t=2019-05-24T09:46:50.9868622+08:00 lvl=ERROR key=value msg=will output placeholder
	// t=2019-05-24T09:46:50.9868622+08:00 lvl=WARN msg=output 'WARN' log
}
```

### Inherit the context of the parent logger

```go
package main

import "github.com/xgfone/klog"

func main() {
	parent := klog.New().WithKv("parent", 123)
	child := parent.WithKv("child", 456)
	child.Info().Msgf("hello %s", "world")

	// Output:
	// t=2019-05-22T16:14:56.3740242+08:00 lvl=INFO parent=123 child=456 msg=hello world
}
```

### Encoder

```go
type Encoder func(buf *Builder, r Record) error
```

This pakcage has implemented four kinds of encoders, `NothingEncoder`, `TextEncoder`, `JSONEncoder` and `StdJSONEncoder`. It will use `TextEncoder` by default, but you can set it to others.

```go
package main

import "github.com/xgfone/klog"

func main() {
	log := klog.New().WithEncoder(klog.JSONEncoder())
	log.Info().K("key1", "value1").K("key2", "value2").Msg("hello world")

	// Output:
	// {"t":"2019-05-22T16:19:03.2972646+08:00","lvl":"INFO","key1":"value1","key2":"value2","msg":"hello world"}
}
```

### Writer

```go
type Writer interface {
	io.Closer
	Write(level Level, data []byte) (n int, err error)
}
```

All implementing the interface `Writer` are a Writer.

There are some built-in writers, such as `DiscardWriter`, `LevelWriter`, `SafeWriter`, `StreamWriter`, `MultiWriter`, `FailoverWriter`, `ReopenWriter`, `NetWriter`, `SyslogWriter` and`SyslogNetWriter`. It also supplies a rotating-size file writer `SizedRotatingFile`.

```go
package main

import (
	"fmt"

	"github.com/xgfone/klog"
)

func main() {
	file, err := klog.NewSizedRotatingFile("test.log", 1024*1024*100, 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	log := klog.New(klog.StreamWriter(file))
	log.Info().K("key", "value").Msg("hello world")

	// Output to file test.log:
	// t=2019-05-22T16:34:09.0685161+08:00 lvl=INFO key=value msg=hello world
}
```

If you want to use `SizedRotatingFile` as the writer, `NewSimpleLogger` maybe is your better choice.

```go
package main

import "github.com/xgfone/klog"

func main() {
	log, _ := klog.NewSimpleLogger("warn", "test.log", "100M", 100)

	log.Info().Print("don't output")
	log.Error().Printf("will output %s %s", "key", "value")

	// Output to test.log:
	// t=2019-05-23T17:20:45.0741975+08:00 lvl=ERROR msg=will output key value
}
```

### Lazy evaluation

```go
type Valuer func(Record) (v interface{})
```

If the type of a certain value is `Valuer`, the logger engine will call it to get the corresponding value firstly before calling the encoder. There are some built-in `Valuer`, such as `Caller()`, `CallerStack()`, `LineNo()`, `LineNoAsInt()`, `FuncName()`, `FuncFullName()`, `FileName()`, `FileLongName()` and `Package()`.

```go
package main

import "github.com/xgfone/klog"

func main() {
	log := klog.Std.WithKv("caller", klog.Caller())
	log.Info().K("stack", klog.CallerStack()).Msg("hello world")
	// Or
	// klog.AddKv("caller", klog.Caller())
	// klog.Info().K("stack", klog.CallerStack()).Msg("hello world")

	// Output:
	// t=2019-05-22T16:41:03.1281272+08:00 lvl=INFO caller=main.go:7 stack=[main.go:7] msg=hello world
}
```

### Hook

```go
type Hook func(name string, level Level) bool
```

You can use the hook to filter or count logs. There are four built-in hooks, `DisableLogger`, `EnableLogger`, `DisableLoggerFromEnv` and `EnableLoggerFromEnv`.

```go
package main

import "github.com/xgfone/klog"

func main() {
	klog.Std = klog.Std.WithHook(klog.EnableLoggerFromEnv("mod"))
	log := klog.Std.WithName("debug")
	log.Info().Msg("hello world")
	// Or
	// klog.SetHook(klog.EnableLoggerFromEnv("mod")).SetName("debug")
	// klog.Info().Msg("hello world")

	// $ go run main.go  # No output
	// $ mod=debug=1 go run main.go
	// t=2019-05-22T17:07:20.2504266+08:00 logger=debug lvl=INFO msg=hello world
}
```

## Performance

The log framework itself has no any performance costs and the key of the bottleneck is the encoder.

### Test 1

```
MacBook Pro(Retina, 13-inch, Mid 2014)
Intel Core i5 2.6GHz
8GB DDR3 1600MHz
macOS Mojave
```

|                test                 |    ops    |     ns/op    |   bytes/op   |    allocs/op
|-------------------------------------|-----------|--------------|--------------|-----------------
|BenchmarkKlog**L**NothingEncoder-4   |  5000000  |   274 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**TextEncoder-4      |  3000000  |   556 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**JSONEncoder-4      |  3000000  |   530 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**StdJSONEncoder-4   |  1000000  |  2190 ns/op  |  1441 B/op   |  22 allocs/op
|BenchmarkKlog**F**NothingEncoder-4   | 10000000  |   189 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**TextEncoder-4      |  3000000  |   457 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**JSONEncoder-4      |  3000000  |   513 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**StdJSONEncoder-4   |  1000000  |  2177 ns/op  |  1441 B/op   |  22 allocs/op


### Test 2

```
Dell Vostro 3470
Intel Core i5-7400 3.0GHz
8GB DDR4 2666MHz
Windows 10
```

|                test                 |    ops    |     ns/op    |   bytes/op   |    allocs/op
|-------------------------------------|-----------|--------------|--------------|-----------------
|BenchmarkKlog**L**NothingEncoder-4   | 10000000  |   235 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**TextEncoder-4      |  5000000  |   331 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**JSONEncoder-4      |  3000000  |   448 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**L**StdJSONEncoder-4   |  1000000  |  1239 ns/op  |  1454 B/op   |   22 allocs/op
|BenchmarkKlog**F**NothingEncoder-4   | 10000000  |   162 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**TextEncoder-4      |  5000000  |   315 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**JSONEncoder-4      |  3000000  |   436 ns/op  |  **32 B/op** |  **1 allocs/op**
|BenchmarkKlog**F**StdJSONEncoder-4   |  1000000  |  1091 ns/op  |  1454 B/op   |   22 allocs/op


**Notice:**
1. **L** and **F** respectively represents **Log** and **LLog** interface.
2. The once memory allocation, `32 B/op` and `1 allocs/op`, is due to the slice type `[]Field`.
