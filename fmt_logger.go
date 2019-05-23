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

import "io"

// FmtLoggerError represents a logger based on the % foramtter.
type FmtLoggerError interface {
	Writer() io.Writer

	Trace(format string, args ...interface{}) error
	Debug(format string, args ...interface{}) error
	Info(format string, args ...interface{}) error
	Warn(format string, args ...interface{}) error
	Error(format string, args ...interface{}) error
	Panic(format string, args ...interface{}) error
	Fatal(format string, args ...interface{}) error
}

// FmtLogger represents a logger based on the % foramtter.
type FmtLogger interface {
	Writer() io.Writer

	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Panic(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// ToFmtLoggerError converts Logger to FmtLoggerError.
func ToFmtLoggerError(logger Logger) FmtLoggerError {
	return fmtLoggerError{logger: logger.WithDepth(1)}
}

// ToFmtLogger converts Logger to FmtLogger.
func ToFmtLogger(logger Logger) FmtLogger {
	return fmtLogger{logger: logger.WithDepth(1)}
}

////////////////////////////////////////////////////////////////////////////

type fmtLogger struct {
	logger Logger
}

func (l fmtLogger) Writer() io.Writer {
	return FromWriter(l.logger.GetWriter())
}

func (l fmtLogger) Trace(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

func (l fmtLogger) Debug(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l fmtLogger) Info(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l fmtLogger) Warn(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l fmtLogger) Error(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l fmtLogger) Panic(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

func (l fmtLogger) Fatal(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

///////////////////////////////////////////////////////////////////////////

type fmtLoggerError struct {
	logger Logger
}

func (l fmtLoggerError) Writer() io.Writer {
	return FromWriter(l.logger.GetWriter())
}

func (l fmtLoggerError) Trace(format string, args ...interface{}) error {
	l.logger.Trace().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Debug(format string, args ...interface{}) error {
	l.logger.Debug().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Info(format string, args ...interface{}) error {
	l.logger.Info().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Warn(format string, args ...interface{}) error {
	l.logger.Warn().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Error(format string, args ...interface{}) error {
	l.logger.Error().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Panic(format string, args ...interface{}) error {
	l.logger.Panic().Msgf(format, args...)
	return nil
}

func (l fmtLoggerError) Fatal(format string, args ...interface{}) error {
	l.logger.Fatal().Msgf(format, args...)
	return nil
}
