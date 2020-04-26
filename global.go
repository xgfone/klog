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
	"os"
	"path/filepath"
)

type originLoggerWrapper struct{ ExtLogger }

func (o originLoggerWrapper) SetEncoder(enc Encoder) {
	o.ExtLogger.SetEncoder(enc)
	defaultLogger.SetEncoder(enc)
}
func (o originLoggerWrapper) SetLevel(lvl Level) {
	o.ExtLogger.SetLevel(lvl)
	defaultLogger.SetLevel(lvl)
}

var defaultLogger ExtLogger
var originLogger originLoggerWrapper

func init() { SetDefaultLogger(New("")) }

// UpdateDefaultLogger updates the default logger with NameToLevel(level)
// and FileWriter(filePath, fileSize, fileNum) if filePath is not "".
func UpdateDefaultLogger(level, filePath, fileSize string, fileNum int) error {
	if filePath != "" {
		os.MkdirAll(filepath.Dir(filePath), 0755)
		wc, err := FileWriter(filePath, fileSize, fileNum)
		if err != nil {
			return err
		}
		GetEncoder().SetWriter(wc)
	}
	SetLevel(NameToLevel(level))
	return nil
}

// GetDefaultLogger returns the default logger.
func GetDefaultLogger() ExtLogger { return originLogger }

// SetDefaultLogger sets the default logger to l.
func SetDefaultLogger(l ExtLogger) {
	originLogger = originLoggerWrapper{l}
	defaultLogger = l.WithDepth(1)
}

// GetEncoder returns the encoder of the default logger.
func GetEncoder() Encoder { return defaultLogger.Encoder() }

// SetEncoder resets the encoder of the default logger, which is TextEncoder by default.
func SetEncoder(enc Encoder) { originLogger.SetEncoder(enc) }

// SetLevel resets the level of the default logger, which is LvlDebug by default.
func SetLevel(lvl Level) { originLogger.SetLevel(lvl) }

// WithCtx returns a new ExtLogger based on the default logger with the fields.
func WithCtx(fields ...Field) ExtLogger { return originLogger.WithCtx(fields...) }

// WithName returns a new ExtLogger based on the default logger with the name.
func WithName(name string) ExtLogger { return originLogger.WithName(name) }

// WithLevel returns a new ExtLogger based on the default logger with the level.
func WithLevel(level Level) ExtLogger { return originLogger.WithLevel(level) }

// WithEncoder returns a new ExtLogger based on the default logger with the encoder.
func WithEncoder(enc Encoder) ExtLogger { return originLogger.WithEncoder(enc) }

// WithDepth returns a new ExtLogger based on the default logger with the depth.
func WithDepth(depth int) ExtLogger { return originLogger.WithDepth(depth) }

// Log emits the log with the level by the default logger.
func Log(level Level, msg string, fields ...Field) { defaultLogger.Log(level, msg, fields...) }

// Trace is equal to Log(LvlTrace, msg, field...).
func Trace(msg string, fields ...Field) { defaultLogger.Log(LvlTrace, msg, fields...) }

// Debug is equal to Log(LvlDebug, msg, field...).
func Debug(msg string, fields ...Field) { defaultLogger.Log(LvlDebug, msg, fields...) }

// Info is equal to Log(LvlInfo, msg, field...).
func Info(msg string, fields ...Field) { defaultLogger.Log(LvlInfo, msg, fields...) }

// Warn is equal to Log(LvlWarn, msg, field...).
func Warn(msg string, fields ...Field) { defaultLogger.Log(LvlWarn, msg, fields...) }

// Error is equal to Log(LvlError, msg, field...).
func Error(msg string, fields ...Field) { defaultLogger.Log(LvlError, msg, fields...) }

// Panic is equal to Log(LvlCrit, msg, fields...), then panic.
func Panic(msg string, fields ...Field) {
	defaultLogger.Log(LvlCrit, msg, fields...)
	panic(fmt.Errorf("%s: %s", LvlCrit.String(), msg))
}

// Fatal is equal to Log(LvlEmerg, msg, fields...), then call os.Exit(1) to exit.
func Fatal(msg string, fields ...Field) {
	defaultLogger.Log(LvlEmerg, msg, fields...)
	os.Exit(1)
}

// Ef is equal to Kv("err", err).Log(LvlError, Sprintf(format, args...)).
func Ef(err error, format string, args ...interface{}) {
	defaultLogger.Log(LvlError, Sprintf(format, args...), F("err", err))
}

// Tracef is equal to Log(LvlTrace, Sprintf(format, args...)).
func Tracef(format string, args ...interface{}) { defaultLogger.Log(LvlTrace, Sprintf(format, args...)) }

// Debugf is equal to Log(LvlDebug, Sprintf(format, args...)).
func Debugf(format string, args ...interface{}) { defaultLogger.Log(LvlDebug, Sprintf(format, args...)) }

// Infof is equal to Log(LvlInfo, Sprintf(format, args...)).
func Infof(format string, args ...interface{}) { defaultLogger.Log(LvlInfo, Sprintf(format, args...)) }

// Warnf is equal to Log(LvlWarn, Sprintf(format, args...)).
func Warnf(format string, args ...interface{}) { defaultLogger.Log(LvlWarn, Sprintf(format, args...)) }

// Errorf is equal to Log(LvlError, Sprintf(format, args...)).
func Errorf(format string, args ...interface{}) { defaultLogger.Log(LvlError, Sprintf(format, args...)) }

// Printf is equal to Infof(format, args...).
func Printf(format string, args ...interface{}) { defaultLogger.Log(LvlInfo, Sprintf(format, args...)) }

// Panicf is equal to Log(LvlCrit, Sprintf(format, args...)), then panic.
func Panicf(format string, args ...interface{}) {
	msg := Sprintf(format, args...)
	defaultLogger.Log(LvlCrit, msg)
	panic(fmt.Errorf("%s: %s", LvlCrit.String(), msg))
}

// Fatalf is equal to Log(LvlEmerg, Sprintf(format, args...)),
// then call os.Exit(1) to exit.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Log(LvlEmerg, Sprintf(format, args...))
	os.Exit(1)
}
