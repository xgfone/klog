// Copyright 2020 xgfone
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
	WriteLevel(level Level, data []byte) (n int, err error)
	io.Closer
}

type writerLevel interface {
	WriteLevel(level Level, data []byte) (n int, err error)
}

type writerFunc struct {
	write func(Level, []byte) (int, error)
	close func() error
}

func (w writerFunc) WriteLevel(l Level, p []byte) (int, error) { return w.write(l, p) }
func (w writerFunc) Close() error                              { return w.close() }

// WriterFunc adapts the function to Writer.
func WriterFunc(write func(Level, []byte) (int, error), close ...func() error) Writer {
	closer := func() error { return nil }
	if len(close) != 0 {
		closer = close[0]
	}
	return writerFunc{write, closer}
}

//////////////////////////////////////////////////////////////////////////////

type ioWriter struct {
	Level Level
	Writer
}

func (w ioWriter) Write(p []byte) (int, error) { return w.WriteLevel(w.Level, p) }

// ToIOWriter converts Writer to io.WriteCloser with the level.
//
// lvl is LvlInfo by default, which is only useful when w is the writer
// like SyslogWriter. Or it should be ignored.
func ToIOWriter(w Writer, lvl ...Level) io.WriteCloser {
	level := LvlInfo
	if len(lvl) > 0 {
		level = lvl[0]
	}
	return ioWriter{Writer: w, Level: level}
}

//////////////////////////////////////////////////////////////////////////////

type streamWriter struct {
	io.Writer
}

func (w streamWriter) WriteLevel(l Level, p []byte) (int, error) {
	if wl, ok := w.Writer.(writerLevel); ok {
		return wl.WriteLevel(l, p)
	}
	return w.Writer.Write(p)
}

func (w streamWriter) Close() (err error) {
	if c, ok := w.Writer.(io.Closer); ok {
		err = c.Close()
	}
	return
}

// StreamWriter converts io.Writer to Writer.
func StreamWriter(w io.Writer) Writer { return streamWriter{w} }

//////////////////////////////////////////////////////////////////////////////

// DiscardWriter discards all the data.
func DiscardWriter() Writer {
	return WriterFunc(func(l Level, p []byte) (int, error) { return len(p), nil })
}

// LevelWriter filters the logs whose level is less than lvl.
func LevelWriter(lvl Level, w Writer) Writer {
	return WriterFunc(func(level Level, p []byte) (n int, err error) {
		if level >= lvl {
			n, err = w.WriteLevel(level, p)
		}
		return
	}, w.Close)
}

// SafeWriter is guaranteed that only a single writing operation can proceed
// at a time.
//
// It's necessary for thread-safe concurrent writes.
func SafeWriter(w Writer) Writer {
	var mu sync.Mutex
	return WriterFunc(func(level Level, p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return w.WriteLevel(level, p)
	}, w.Close)
}

// BufferWriter returns a new Writer to write all logs to a buffer
// which flushes into the wrapped writer whenever it is available for writing.
func BufferWriter(w Writer, bufferSize int) Writer {
	bw := bufio.NewWriterSize(ToIOWriter(w), bufferSize)
	sw := StreamWriter(bw)
	return WriterFunc(sw.WriteLevel, func() error { bw.Flush(); return w.Close() })
}

// NetWriter opens a socket to the given address and writes the log
// over the connection.
func NetWriter(network, addr string) (Writer, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return StreamWriter(conn), nil
}

// FailoverWriter writes all log records to the first writer specified,
// but will failover and write to the second writer if the first writer
// has failed, and so on for all writers specified.
//
// For example, you might want to log to a network socket, but failover to
// writing to a file if the network fails, and then to standard out
// if the file write fails.
func FailoverWriter(writers ...Writer) Writer {
	_len := len(writers)
	return WriterFunc(func(level Level, p []byte) (n int, err error) {
		for i := 0; i < _len; i++ {
			if n, err = writers[i].WriteLevel(level, p); err == nil {
				return
			}
		}
		return
	}, func() error {
		for i := 0; i < _len; i++ {
			writers[i].Close()
		}
		return nil
	})
}

// SplitWriter returns a level-separated writer, which will write the log record
// into the separated writer.
func SplitWriter(getWriter func(Level) Writer) Writer {
	return WriterFunc(func(level Level, p []byte) (n int, err error) {
		if w := getWriter(level); w != nil {
			n, err = w.WriteLevel(level, p)
		}
		return
	}, nil)
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
	var w io.WriteCloser = os.Stdout
	if filename != "" {
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
		if w, err = NewSizedRotatingFile(filename, int(size), filenum); err != nil {
			return nil, err
		}
	}

	return StreamWriter(w), nil
}

// NewSizedRotatingFile returns a new SizedRotatingFile, which is not thread-safe.
//
// The default permission of the log file is 0644.
func NewSizedRotatingFile(filename string, size, count int,
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

	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

// SizedRotatingFile is a file rotating logging writer based on the size.
type SizedRotatingFile struct {
	file        *os.File
	filePerm    os.FileMode
	filename    string
	maxSize     int
	backupCount int
	nbytes      int
}

// Close implements io.Closer.
func (f *SizedRotatingFile) Close() error { return f.close() }

// Flush flushes the data to the underlying disk.
func (f *SizedRotatingFile) Flush() error { return f.file.Sync() }

// Write implements io.Writer.
func (f *SizedRotatingFile) Write(data []byte) (n int, err error) {
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
			return fmt.Errorf("failed to close the rotating file '%s': %s", f.filename, err)
		}

		if !fileIsExist(f.filename) {
			return nil
		} else if n, err := fileSize(f.filename); err != nil {
			return fmt.Errorf("failed to get the size of the rotating file '%s': %s",
				f.filename, err)
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
					return fmt.Errorf("failed to rename the rotating file '%s' to '%s': %s",
						sfn, dfn, err)
				}
			}
		}
		dfn := f.filename + ".1"
		if fileIsExist(dfn) {
			if err = os.Remove(dfn); err != nil {
				return fmt.Errorf("failed to remove the rotating file '%s': %s", dfn, err)
			}
		}
		if fileIsExist(f.filename) {
			if err = os.Rename(f.filename, dfn); err != nil {
				return fmt.Errorf("failed to rename the rotating file '%s' to '%s': %s",
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

	panic(fmt.Errorf("step must not be 0"))
}
