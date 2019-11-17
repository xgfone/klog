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
	"math"
	"strings"
)

// Predefine some levels.
var (
	LvlTrace = NewLevel("TRACE", 0)
	LvlDebug = NewLevel("DEBUG", 100)
	LvlInfo  = NewLevel("INFO", 200)
	LvlWarn  = NewLevel("WARN", 300)
	LvlError = NewLevel("ERROR", 400)
	LvlCrit  = NewLevel("CRIT", 500)
	LvlEmerg = NewLevel("EMERG", 600)
	LvlMax   = NewLevel("MAX", math.MaxInt32)
)

// Level represents the logger level.
type Level interface {
	// Priority returns the priority of the level.
	// The bigger the level, the higher the priority.
	Priority() int

	// String returns the name of the level.
	String() string
}

func levelIsLess(lvl1, lvl2 Level) bool {
	return lvl1.Priority() < lvl2.Priority()
}

// NewLevel returns a new level, which has also implemented the interface
// io.WriterTo.
func NewLevel(name string, priority int) Level {
	return namedLevel{name: name, prio: priority}
}

type namedLevel struct {
	name string
	prio int
}

func (l namedLevel) String() string {
	return l.name
}

func (l namedLevel) Priority() int {
	return l.prio
}

func (l namedLevel) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, l.name)
	return int64(n), err
}

// Levels is a set of Level, which is used to get the level by the name.
var Levels map[string]Level

// NameToLevel returns the Level by the name, which is case Insensitive.
//
// If not panic, it will return `LvlInfo` instead if no level named `name`.
func NameToLevel(name string, defaultPanic ...bool) Level {
	for n, lvl := range Levels {
		if strings.ToUpper(n) == strings.ToUpper(name) {
			return lvl
		}
	}

	if len(defaultPanic) > 0 && defaultPanic[0] {
		panic(fmt.Errorf("unknown level name '%s'", name))
	}

	return LvlInfo
}
