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
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
)

var fileFlag = os.O_CREATE | os.O_APPEND | os.O_WRONLY

// Writer is the interface to write the log to the underlying storage.
type Writer interface {
	Write(level Level, data []byte) (n int, err error)
}

// WriteCloser is the union of Writer and io.Closer.
type WriteCloser interface {
	Writer
	io.Closer
}

// WriteFlushCloser is the union of WriteCloser and Flusher.
type WriteFlushCloser interface {
	WriteCloser
	Flush() error
}

//////////////////////////////////////////////////////////////////////////////

type ioWriter struct {
	w Writer
}

// Writer implements io.Writer.
func (w ioWriter) Write(p []byte) (int, error) {
	return w.w.Write(LvlMax, p)
}

// Flush implements the interface { Flush() error }.
func (w ioWriter) Flush() error {
	if f := getFlush(w.w); f != nil {
		return f()
	}
	return nil
}

// Close implements io.Closer.
func (w ioWriter) Close() error {
	if c := getClose(w.w); c != nil {
		return c()
	}
	return nil
}

// ToIOWriter converts Writer to io.WriteCloser.
func ToIOWriter(w Writer) io.WriteCloser {
	return ioWriter{w: w}
}

//////////////////////////////////////////////////////////////////////////////

type writer struct {
	write func(Level, []byte) (int, error)
	flush func() error
	close func() error
}

func (w writer) Write(l Level, p []byte) (int, error) { return w.write(l, p) }
func (w writer) Close() error                         { return w.exec(w.close) }
func (w writer) Flush() error                         { return w.exec(w.flush) }
func (w writer) exec(f func() error) error {
	if f != nil {
		return f()
	}
	return nil
}

func getClose(v interface{}) func() error {
	if w, ok := v.(ioWriter); ok {
		v = w.w
	}

	if w, ok := v.(writer); ok {
		return w.close
	} else if c, ok := v.(io.Closer); ok {
		return c.Close
	}
	return nil
}

func getFlush(v interface{}) func() error {
	if w, ok := v.(ioWriter); ok {
		v = w.w
	}

	if w, ok := v.(writer); ok {
		return w.flush
	} else if f, ok := v.(interface{ Flush() error }); ok {
		return f.Flush
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

// WriterFunc adapts the function to Writer.
func WriterFunc(f func(Level, []byte) (int, error)) Writer {
	return writer{write: f}
}

// WriteCloserFunc adapts the write and close function to WriteCloser.
//
// close may be nil, which will do nothing when calling the Close method.
func WriteCloserFunc(write func(Level, []byte) (int, error), close func() error) WriteCloser {
	return writer{write: write, close: close}
}

// WriteFlushCloserFunc adapts the write, flush and close function to WriteFlushCloser.
//
// flush and close may be nil, which will do nothing when calling the Flush
// or Close method.
func WriteFlushCloserFunc(write func(Level, []byte) (int, error), flush, close func() error) WriteFlushCloser {
	return writer{write: write, flush: flush, close: close}
}

// ToWriteCloser converts Writer to WriteCloser.
//
// If the writer has no the method Close, it does nothing.
func ToWriteCloser(w Writer) WriteCloser {
	return writer{write: w.Write, close: getClose(w)}
}

// ToWriteFlushCloser converts the Writer to WriteFlushCloser.
//
// If the writer has no the methods Close and Flush, they do nothing.
func ToWriteFlushCloser(w Writer) WriteFlushCloser {
	return writer{write: w.Write, flush: getFlush(w), close: getClose(w)}
}

func flushAllWriters(writers ...Writer) error {
	for _, w := range writers {
		if f, ok := w.(interface{ Flush() error }); ok {
			f.Flush()
		}
	}
	return nil
}

func closeAllWriters(writers ...Writer) error {
	for _, w := range writers {
		if c, ok := w.(io.Closer); ok {
			c.Close()
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

// DiscardWriter discards all the data.
func DiscardWriter() WriteFlushCloser {
	return WriteFlushCloserFunc(func(Level, []byte) (int, error) { return 0, nil }, nil, nil)
}

// LevelWriter filters the logs whose level is less than lvl.
func LevelWriter(lvl Level, w Writer) WriteFlushCloser {
	return WriteFlushCloserFunc(func(level Level, p []byte) (int, error) {
		if level.Priority() < lvl.Priority() {
			return 0, nil
		}
		return w.Write(level, p)
	}, getFlush(w), getClose(w))
}

// SafeWriter is guaranteed that only a single writing operation can proceed
// at a time.
//
// It's necessary for thread-safe concurrent writes.
func SafeWriter(w Writer) WriteFlushCloser {
	var mu sync.Mutex
	return WriteFlushCloserFunc(func(level Level, p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return w.Write(level, p)
	}, getFlush(w), getClose(w))
}

// StreamWriter converts io.Writer to Writer.
func StreamWriter(w io.Writer) WriteFlushCloser {
	return WriteFlushCloserFunc(func(level Level, p []byte) (int, error) {
		return w.Write(p)
	}, getFlush(w), getClose(w))
}

// BufferWriter returns a new WriteFlushCloser to write all logs to a buffer
// which flushes into the wrapped writer whenever it is available for writing.
//
// It uses SafeWriter to write all logs to the buffer thread-safely.
// So the first argument w may not be thread-safe.
func BufferWriter(w Writer, bufferSize int) WriteFlushCloser {
	bw := bufio.NewWriterSize(ToIOWriter(w), bufferSize)
	sw := SafeWriter(StreamWriter(bw))
	return WriteFlushCloserFunc(sw.Write, bw.Flush, getClose(w))
}

// NetWriter opens a socket to the given address and writes the log
// over the connection.
//
// Notice: it will be wrapped by SafeWriter, so it's thread-safe.
func NetWriter(network, addr string) (WriteFlushCloser, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return ToWriteFlushCloser(SafeWriter(StreamWriter(conn))), nil
}

// FailoverWriter writes all log records to the first writer specified,
// but will failover and write to the second writer if the first writer
// has failed, and so on for all writers specified.
//
// For example, you might want to log to a network socket, but failover to
// writing to a file if the network fails, and then to standard out
// if the file write fails.
func FailoverWriter(writers ...Writer) WriteFlushCloser {
	return WriteFlushCloserFunc(func(level Level, p []byte) (n int, err error) {
		for _, w := range writers {
			if n, err = w.Write(level, p); err == nil {
				return
			}
		}
		return
	},
		func() error { return flushAllWriters(writers...) },
		func() error { return closeAllWriters(writers...) })
}

// SplitWriter returns a level-separated writer, which will write the log record
// into the separated writer.
func SplitWriter(getWriter func(Level) Writer, flush ...func() error) WriteFlushCloser {
	var flusher func() error
	if len(flush) > 0 {
		flusher = flush[0]
	}

	return WriteFlushCloserFunc(func(level Level, p []byte) (n int, err error) {
		if w := getWriter(level); w != nil {
			n, err = w.Write(level, p)
		}
		return
	}, flusher, nil)
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
func FileWriter(filename, filesize string, filenum int) (WriteFlushCloser, error) {
	return fileWriter(filename, filesize, filenum, true)
}

// FileWriterWithoutLock is the same as FileWriter, but not use the lock
// to ensure that it's thread-safe to write the log.
func FileWriterWithoutLock(filename, filesize string, filenum int) (WriteFlushCloser, error) {
	return fileWriter(filename, filesize, filenum, false)
}

func fileWriter(filename, filesize string, filenum int,
	lock bool) (WriteFlushCloser, error) {
	if filename == "" {
		return ToWriteFlushCloser(SafeWriter(StreamWriter(os.Stdout))), nil
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
	var file *SizedRotatingFile
	if lock {
		file, err = NewSizedRotatingFile(filename, int(size), filenum)
	} else {
		file, err = NewSizedRotatingFileWithoutLock(filename, int(size), filenum)
	}
	if err != nil {
		return nil, err
	}
	return ToWriteFlushCloser(StreamWriter(file)), nil
}

// NewSizedRotatingFile returns a new SizedRotatingFile.
//
// It is thread-safe for concurrent writes.
//
// The default permission of the log file is 0644.
func NewSizedRotatingFile(filename string, size, count int,
	mode ...os.FileMode) (*SizedRotatingFile, error) {
	return newSizedRotatingFile(filename, size, count, true, mode...)
}

// NewSizedRotatingFileWithoutLock is equal to NewSizedRotatingFile,
// But not use the lock to ensure that it's thread-safe to write the log.
func NewSizedRotatingFileWithoutLock(filename string, size, count int,
	mode ...os.FileMode) (*SizedRotatingFile, error) {
	return newSizedRotatingFile(filename, size, count, false, mode...)
}

func newSizedRotatingFile(filename string, size, count int, lock bool,
	mode ...os.FileMode) (*SizedRotatingFile, error) {
	var _mode os.FileMode = 0644
	if len(mode) > 0 && mode[0] > 0 {
		_mode = mode[0]
	}

	w := &SizedRotatingFile{
		filename:    filename,
		filePerm:    _mode,
		maxSize:     size,
		backupCount: count,
	}

	if lock {
		w.lock = new(sync.Mutex)
	}

	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

// SizedRotatingFile is a file rotating logging writer based on the size.
type SizedRotatingFile struct {
	lock *sync.Mutex
	file *os.File

	filePerm    os.FileMode
	filename    string
	maxSize     int
	backupCount int
	nbytes      int
}

func (f *SizedRotatingFile) locked() {
	if f.lock != nil {
		f.lock.Lock()
	}
}

func (f *SizedRotatingFile) unlocked() {
	if f.lock != nil {
		f.lock.Unlock()
	}
}

// Close implements io.Closer.
func (f *SizedRotatingFile) Close() error {
	f.locked()
	defer f.unlocked()
	return f.close()
}

// Flush flushes the data to the underlying disk.
func (f *SizedRotatingFile) Flush() error {
	f.locked()
	defer f.unlocked()
	return f.file.Sync()
}

// Write implements io.Writer.
func (f *SizedRotatingFile) Write(data []byte) (n int, err error) {
	f.locked()
	defer f.unlocked()

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
