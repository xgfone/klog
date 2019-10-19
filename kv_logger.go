package klog

import (
	"fmt"
	"io"
)

// KvLogger is a key-value logger interface.
type KvLogger interface {
	Writer() io.Writer

	Tracef(msg string, kvs ...interface{})
	Debugf(msg string, kvs ...interface{})
	Infof(msg string, kvs ...interface{})
	Warnf(msg string, kvs ...interface{})
	Errorf(msg string, kvs ...interface{})
	Panicf(msg string, kvs ...interface{})
	Fatalf(msg string, kvs ...interface{})
}

// ToKvLogger converts the Logger to KvLogger.
func ToKvLogger(logger Logger) KvLogger {
	return kvLogger{logger: logger.WithDepth(2)}
}

type kvLogger struct {
	logger Logger
}

func (l kvLogger) Writer() io.Writer {
	return FromWriter(l.logger.GetWriter())
}

func (l kvLogger) emit(log Log, msg string, kvs []interface{}) {
	_len := len(kvs)
	if _len == 0 {
		log.Printf(msg)
		return
	}

	if _len%2 != 0 {
		panic(fmt.Errorf("KvLogger: the length '%d' of kvs is not even", _len))
	}

	for i := 0; i < _len; i += 2 {
		if s, ok := kvs[i].(string); ok {
			log = log.K(s, kvs[i+1])
		} else {
			panic(fmt.Errorf("KvLogger: the %dth key-value is not string", i))
		}
	}

	log.Printf(msg)
}

func (l kvLogger) Tracef(msg string, kvs ...interface{}) {
	l.emit(l.logger.Trace(), msg, kvs)
}

func (l kvLogger) Debugf(msg string, kvs ...interface{}) {
	l.emit(l.logger.Debug(), msg, kvs)
}

func (l kvLogger) Infof(msg string, kvs ...interface{}) {
	l.emit(l.logger.Info(), msg, kvs)
}

func (l kvLogger) Warnf(msg string, kvs ...interface{}) {
	l.emit(l.logger.Warn(), msg, kvs)
}

func (l kvLogger) Errorf(msg string, kvs ...interface{}) {
	l.emit(l.logger.Error(), msg, kvs)
}

func (l kvLogger) Panicf(msg string, kvs ...interface{}) {
	l.emit(l.logger.Panic(), msg, kvs)
}

func (l kvLogger) Fatalf(msg string, kvs ...interface{}) {
	l.emit(l.logger.Fatal(), msg, kvs)
}
