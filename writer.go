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
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"sync"
)

var maxLevel = Level(math.MaxInt32)
var fileFlag = os.O_CREATE | os.O_APPEND | os.O_WRONLY

// Errors is used to wrap more than one error.
type Errors []error

func (es Errors) Error() string {
	buf := getBuilder()
	es.WriteTo(buf)
	s := buf.String()
	putBuilder(buf)
	return s
}

// WriteTo implements io.WriteTo.
func (es Errors) WriteTo(w io.Writer) (n int64, err error) {
	for i, e := range es {
		if e != nil {
			m, _err := fmt.Fprintf(w, "[%d]%s; ", i, e.Error())
			if _err != nil {
				return n + int64(m), _err
			}
			n += int64(m)
		}
	}
	return
}

// Writer is the interface to write the log to the underlying storage.
type Writer interface {
	io.Closer
	Write(level Level, data []byte) (n int, err error)
}

type ioWriter struct {
	w Writer
}

// Writer implements io.Writer.
func (w ioWriter) Write(p []byte) (int, error) {
	return w.w.Write(maxLevel, p)
}

// Close implements io.Closer.
func (w ioWriter) Close() error {
	return w.Close()
}

// FromWriter converts Writer to io.WriteCloser.
func FromWriter(w Writer) io.WriteCloser {
	return ioWriter{w: w}
}

type writerFunc struct {
	close func() error
	write func(Level, []byte) (int, error)
}

// Write implements Writer#Write().
func (w writerFunc) Write(level Level, data []byte) (int, error) {
	return w.write(level, data)
}

// Close implements io.Closer.
func (w writerFunc) Close() error {
	if w.close != nil {
		return w.close()
	}
	return nil
}

// WriterFunc adapts a function with w to Writer.
//
// If giving the close function, it will be called when closing the writer.
// Or do nothing.
func WriterFunc(w func(Level, []byte) (int, error), close ...func() error) Writer {
	var closer func() error
	if len(close) > 0 {
		closer = close[0]
	}
	return writerFunc{close: closer, write: w}
}

// DiscardWriter discards all the data.
func DiscardWriter() Writer {
	return WriterFunc(func(Level, []byte) (int, error) { return 0, nil })
}

// LevelWriter filters the logs whose level is less than lvl.
func LevelWriter(lvl Level, w Writer) Writer {
	return WriterFunc(func(level Level, p []byte) (int, error) {
		if level < lvl {
			return 0, nil
		}
		return w.Write(level, p)
	}, w.Close)
}

// SafeWriter is guaranteed that only a single writing operation can proceed
// at a time.
//
// The returned Writer supports io.Closer, which will be forwarded to w.Close().
//
// It's necessary for thread-safe concurrent writes.
func SafeWriter(w Writer) Writer {
	var mu sync.Mutex
	return WriterFunc(func(level Level, p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return w.Write(level, p)
	}, w.Close)
}

// StreamWriter converts io.Writer to Writer.
//
// If w has implemented io.Closer, the returned writer will forward the Close
// calling to it. Or do nothing.
func StreamWriter(w io.Writer) Writer {
	return WriterFunc(func(level Level, p []byte) (int, error) {
		return w.Write(p)
	}, func() error {
		if closer, ok := w.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	})
}

// NetWriter opens a socket to the given address and writes the log
// over the connection.
//
// Notice: it will be wrapped by SafeWriter, so it's thread-safe.
func NetWriter(network, addr string) (Writer, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return SafeWriter(StreamWriter(conn)), nil
}

// ReopenWriter returns a writer that can be closed then re-opened,
// which is used for logrotate typically.
//
// Notice: It's thread-safe.
func ReopenWriter(factory func() (w io.WriteCloser, reopen <-chan bool, err error)) Writer {
	var w io.WriteCloser
	var err error
	var reopen <-chan bool
	var lock sync.Mutex

	return WriterFunc(func(level Level, p []byte) (int, error) {
		lock.Lock()
		defer lock.Unlock()

		if reopen == nil {
			if w, reopen, err = factory(); err != nil {
				w = nil
				reopen = nil
				return 0, err
			}
		} else if w == nil {
			return 0, errors.New("the writer has been closed")
		}

		select {
		case <-reopen:
			w.Close()
			if w, reopen, err = factory(); err != nil {
				w = nil
				reopen = nil
				return 0, err
			}
		default:
		}
		return w.Write(p)
	}, func() (err error) {
		lock.Lock()
		defer lock.Unlock()
		if w != nil {
			err = w.Close()
			w = nil
		}
		return
	})
}

// MultiWriter writes one data to more than one destination.
func MultiWriter(outs ...Writer) Writer {
	_len := len(outs)
	return WriterFunc(func(level Level, p []byte) (n int, err error) {
		var errs Errors
		for i, out := range outs {
			if m, e := out.Write(level, p); e != nil {
				n += m
				if errs == nil {
					errs = make(Errors, _len)
				}
				errs[i] = e
			} else {
				n += m
			}
		}
		return n, errs
	}, func() error {
		var errs Errors
		for i, w := range outs {
			if e := w.Close(); e != nil {
				if errs == nil {
					errs = make(Errors, _len)
				}
				errs[i] = e
			}
		}
		return errs
	})
}

// FailoverWriter writes all log records to the first handler specified,
// but will failover and write to the second handler if the first handler
// has failed, and so on for all handlers specified.
//
// For example, you might want to log to a network socket, but failover to
// writing to a file if the network fails, and then to standard out
// if the file write fails.
func FailoverWriter(outs ...Writer) Writer {
	_len := len(outs)
	return WriterFunc(func(level Level, p []byte) (n int, err error) {
		for _, out := range outs {
			if n, err = out.Write(level, p); err == nil {
				return
			}
		}
		return
	}, func() error {
		var errs Errors
		for i, w := range outs {
			if e := w.Close(); e != nil {
				if errs == nil {
					errs = make(Errors, _len)
				}
				errs[i] = e
			}
		}
		return errs
	})
}

// FileWriter returns a writer based the file, which uses NewSizedRotatingFile
// to generate the file writer. If filename is "", however, it will return
// a os.Stdout writer instead.
//
// filesize is parsed by ParseSize to get the size of the log file.
// If it is "", it is "100M" by default.
//
// filenum is the number of the log file. If it is 0 or negative,
// it will be reset to 100.
//
// Notice: if the directory in where filename is does not exist, it will be
// created automatically.
func FileWriter(filename, filesize string, filenum int) (Writer, error) {
	if filename == "" {
		return SafeWriter(StreamWriter(os.Stdout)), nil
	}

	if filesize == "" {
		filesize = "100M"
	}
	size, err := ParseSize(filesize)
	if err != nil {
		return nil, err
	} else if filenum < 1 {
		filenum = 100
	}

	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := NewSizedRotatingFile(filename, int(size), filenum)
	if err != nil {
		return nil, err
	}
	AppendCleaner(func() { file.Close() })
	return StreamWriter(file), nil
}

// NewSizedRotatingFile returns a new SizedRotatingFile.
//
// It is thread-safe for concurrent writes.
//
// The default permission of the log file is 0644.
func NewSizedRotatingFile(filename string, size, count int,
	mode ...os.FileMode) (*SizedRotatingFile, error) {

	var _mode os.FileMode = 0644
	if len(mode) > 0 && mode[0] > 0 {
		_mode = mode[0]
	}

	w := SizedRotatingFile{
		filename:    filename,
		filePerm:    _mode,
		maxSize:     size,
		backupCount: count,
	}

	if err := w.open(); err != nil {
		return nil, err
	}
	return &w, nil
}

// SizedRotatingFile is a file rotating logging handler based on the size.
type SizedRotatingFile struct {
	lock sync.Mutex
	file *os.File

	filePerm    os.FileMode
	filename    string
	maxSize     int
	backupCount int
	nbytes      int
}

// Close implements io.Closer.
func (f *SizedRotatingFile) Close() (err error) {
	f.lock.Lock()
	err = f.close()
	f.lock.Unlock()
	return
}

// Flush flushes the data to the underlying disk.
func (f *SizedRotatingFile) Flush() error {
	return f.file.Sync()
}

// Write implements io.Writer.
func (f *SizedRotatingFile) Write(data []byte) (n int, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.file == nil {
		return 0, errors.New("the file has been closed")
	}

	if f.nbytes+len(data) > f.maxSize {
		if err = f.doRollover(); err != nil {
			return
		}
	}

	if n, err = f.file.Write(data); err != nil {
		return
	}

	f.nbytes += n
	return
}

func (f *SizedRotatingFile) open() (err error) {
	file, err := os.OpenFile(f.filename, fileFlag, f.filePerm)
	if err != nil {
		return
	}

	info, err := file.Stat()
	if err != nil {
		return
	}

	f.nbytes = int(info.Size())
	f.file = file
	return
}

func (f *SizedRotatingFile) close() (err error) {
	err = f.file.Close()
	f.file = nil
	return
}

func (f *SizedRotatingFile) doRollover() (err error) {
	if f.backupCount > 0 {
		if err = f.close(); err != nil {
			return fmt.Errorf("Rotating: close failed: %s", err)
		}

		if !fileIsExist(f.filename) {
			return nil
		} else if n, err := fileSize(f.filename); err != nil {
			return fmt.Errorf("Rotating: failed to get the size: %s", err)
		} else if n == 0 {
			return nil
		}

		for _, i := range ranges(f.backupCount-1, 0, -1) {
			sfn := fmt.Sprintf("%s.%d", f.filename, i)
			dfn := fmt.Sprintf("%s.%d", f.filename, i+1)
			if fileIsExist(sfn) {
				if fileIsExist(dfn) {
					os.Remove(dfn)
				}
				if err = os.Rename(sfn, dfn); err != nil {
					return fmt.Errorf("Rotating: failed to rename %s -> %s: %s",
						sfn, dfn, err)
				}
			}
		}
		dfn := f.filename + ".1"
		if fileIsExist(dfn) {
			if err = os.Remove(dfn); err != nil {
				return fmt.Errorf("Rotating: failed to remove %s: %s", dfn, err)
			}
		}
		if fileIsExist(f.filename) {
			if err = os.Rename(f.filename, dfn); err != nil {
				return fmt.Errorf("Rotating: failed to rename %s -> %s: %s",
					f.filename, dfn, err)
			}
		}
		err = f.open()
	}
	return
}

func fileIsExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// fileSize returns the size of the file as how many bytes.
func fileSize(fp string) (int64, error) {
	f, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

func ranges(start, stop, step int) (r []int) {
	if step > 0 {
		for start < stop {
			r = append(r, start)
			start += step
		}
		return
	} else if step < 0 {
		for start > stop {
			r = append(r, start)
			start += step
		}
		return
	}

	panic(fmt.Errorf("The step must not be 0"))
}
