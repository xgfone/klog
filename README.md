# klog [![Build Status](https://travis-ci.org/xgfone/klog.svg?branch=master)](https://travis-ci.org/xgfone/klog) [![GoDoc](https://godoc.org/github.com/xgfone/klog?status.svg)](http://godoc.org/github.com/xgfone/klog) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/klog/master/LICENSE)

Package `klog` provides an simple, flexible, extensible, powerful and structured logging tool based on the level, which has done the better balance between the flexibility and the performance. It is inspired by [log15](https://github.com/inconshreveable/log15), [logrus](https://github.com/sirupsen/logrus), [go-kit](https://github.com/go-kit/kit), [logger](https://github.com/xgfone/logger) and [log](github.com/go-playground/log).

See the [GoDoc](https://godoc.org/github.com/xgfone/klog).

**API has been stable.** The current is `v1.x`.


## Prerequisite

Now `klog` requires Go `1.x`.


## Features

- Lazy evaluation of expensive operations.
- Simple, Flexible, Extensible, Powerful and Structured.
- Built-in support for logging to files, syslog, and the network.
- Child loggers which inherit and add their own private context.
- Support lots of `Writer`, such as `StreamWriter`, `NetWriter`, `SyslogWriter`, `SizedRotatingFile`, etc.


## Example

```go
package main

import "github.com/xgfone/klog"

func main() {
	log := klog.New().WithLevel(klog.LvlWarn)

	log.Info().K("key", "value").Msg("don't output")
	log.Error().K("key", "value").Msgf("will output %s", "placeholder")

	// Output:
	// t=2019-05-22T16:05:56.2318273+08:00 lvl=ERROR key=value msg=will output placeholder
}
```

Furthermore, `klog` has built in a global logger, `Std`, which is equal to `klog.New()`, and you can use it and its exported function, `Trace()`, `Debug`, `Info`, `Warn`, `Error`, `Panic`, `Fatal`, or `L(level)`. **Suggestion:** You should use these functions instead.

```go
package main

import "github.com/xgfone/klog"

func main() {
	klog.Std = klog.Std.WithLevel(klog.LvlWarn)

	klog.Info().K("key", "value").Msg("don't output")
	klog.Error().K("key", "value").Msgf("will output %s", "placeholder")

	// Output:
	// t=2019-05-22T16:05:56.2318273+08:00 lvl=ERROR key=value msg=will output placeholder
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

	// Output:
	// t=2019-05-22T16:41:03.1281272+08:00 lvl=INFO caller=main.go:9 stack=[main.go:9] msg=hello world
}
```

### Hook

```go
type Hook func(name string, level Level) bool
```

You can use the hook to filter or count logs. There are two built-in hooks, `DisableLogger` and `EnableLogger`.

```go
package main

import (
	"os"

	"github.com/xgfone/klog"
)

func main() {
	var debug bool
	for _, env := range os.Environ() {
		if env == "debug=on" || env == "debug=1" {
			debug = true
			break
		}
	}

	if !debug {
		klog.Std = klog.Std.WithHook(klog.DisableLogger("debug"))
	}

	log := klog.Std.WithName("debug")
	log.Info().Msg("hello world")

	// $ go run main.go  # No output
	// $ debug=on go run main.go
	// t=2019-05-22T17:07:20.2504266+08:00 logger=debug lvl=INFO msg=hello world
}
```

## Performance

The log framework itself has no any performance costs and the key of the bottleneck is the encoder.

|  test   | ops | ns/op | bytes/op | allocs/op
|---------|-----|-------|----------|-----------
|BenchmarkKlogNothingEncoder-4     | 10000000  |  149 ns/op | **32 B/op** |  **1 allocs/op**
|BenchmarkKlogTextEncoder-4        |  5000000  |  281 ns/op | **32 B/op** |  **1 allocs/op**
|BenchmarkKlogJSONEncoder-4        |  5000000  |  313 ns/op | **32 B/op** |  **1 allocs/op**
|BenchmarkKlogStdJSONEncoder-4     |  1000000  | 1043 ns/op | 1455 B/op   | 22 allocs/op

**Notice:** The once memory allocation, `32 B/op` and `1 allocs/op`, is due to the slice type `[]Field`.
