// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package klog

import (
	"fmt"
	"io"
	"strings"
)

// Predefine some levels.
//
// You can define yourself level.
//
// For LvlFatal or higher, it will emit the log and call the clean functions
// firstly, then the program will exit by calling `os.Exit(1)`.
//
// For LvlPanic or higher, it will emit the log firstly, then panic with Record,
// but the field Fields is reset to nil.
const (
	LvlTrace Level = iota * 100
	LvlDebug
	LvlInfo
	LvlWarn
	LvlError
	LvlPanic
	LvlFatal
)

// Levels is the pre-defined level names, but you can reset and override them.
var Levels = map[Level]string{
	LvlTrace: "TRACE",
	LvlDebug: "DEBUG",
	LvlInfo:  "INFO",
	LvlWarn:  "WARN",
	LvlError: "ERROR",
	LvlPanic: "PANIC",
	LvlFatal: "FATAL",
}

// Level represents a level. The bigger the value, the higher the level.
type Level int32

func (l Level) String() string {
	return Levels[l]
}

// WriteTo implements io.WriterTo.
func (l Level) WriteTo(out io.Writer) (int64, error) {
	n, err := io.WriteString(out, l.String())
	return int64(n), err
}

// MarshalJSON implements json.Marshaler.
func (l Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.String() + `"`), nil
}

// NameToLevel returns the Level by the name, which is case Insensitive.
//
// If not panic, it will return `LvlInfo` instead if no level named `name`.
func NameToLevel(name string, defaultPanic ...bool) Level {
	for k, v := range Levels {
		if v == name {
			return k
		}
	}

	uname := strings.ToUpper(name)
	for k, v := range Levels {
		if v == uname {
			return k
		}
	}

	if len(defaultPanic) > 0 && defaultPanic[0] {
		panic(fmt.Errorf("unknown level name '%s'", name))
	}

	return LvlInfo
}
