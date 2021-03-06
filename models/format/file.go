package format

import (
	"bytes"
	"fmt"
	"ghost/services/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	. "ghost/global"
)

// fileLogWriter implements LoggerInterface.
// It writes messages by lines limit, file size limit, or time frequency.
type fileLogWriter struct {
	sync.RWMutex // write log order by order and  atomic incr maxLinesCurLines and maxSizeCurSize
	// The opened file
	filename   string `json:"filename"`
	fileWriter *os.File

	// Rotate at line
	maxLines         int `json:"maxlines"`
	maxLinesCurLines int

	// Rotate at size
	maxSize        int `json:"maxsize"`
	maxSizeCurSize int

	// Rotate daily
	daily         bool `json:"daily"`
	dailyOpenDate int
	dailyOpenTime time.Time

	rotate bool   `json:"rotate"`
	perm   string `json:"perm"`

	fileNameOnly, suffix string // like "project.log", project is fileNameOnly and .log is suffix
}

// newFileWriter create a FileLogWriter returning as LoggerInterface.
func newFileWriter(filename string, maxlines, maxsize int) (*fileLogWriter, error) {
	w := &fileLogWriter{
		filename: filename,
		daily:    true,
		maxLines: maxlines,
		maxSize:  maxsize,
		rotate:   true,
		perm:     "0660",
	}

	w.suffix = filepath.Ext(w.filename)
	w.fileNameOnly = strings.TrimSuffix(w.filename, w.suffix)
	if w.suffix == "" {
		w.suffix = ".log"
	}
	err := w.startLogger()
	if err != nil {
		return nil, err
	}

	return w, nil
}

// start file logger. create log file and set to locker-inside file writer.
func (w *fileLogWriter) startLogger() error {
	file, err := w.createLogFile()
	if err != nil {
		return err
	}
	if w.fileWriter != nil {
		w.fileWriter.Close()
	}
	w.fileWriter = file
	return w.initFd()
}

func (w *fileLogWriter) needRotate(size int, day int) bool {
	return (w.maxLines > 0 && w.maxLinesCurLines >= w.maxLines) ||
		(w.maxSize > 0 && w.maxSizeCurSize >= w.maxSize) ||
		(w.daily && day != w.dailyOpenDate)

}

// WriteMsg write logger message into file.
func (w *fileLogWriter) WriteMsg(when time.Time, msg string) error {
	_, d := utils.FormatTimeHeader(when)
	if w.rotate {
		w.RLock()
		if w.needRotate(len(msg), d) {
			w.RUnlock()
			w.Lock()
			if w.needRotate(len(msg), d) {
				if err := w.doRotate(when); err != nil {
					Log.Critical("FileLogWriter err (%q) : %s", w.filename, err)
					//fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
				}
			}
			w.Unlock()
		} else {
			w.RUnlock()
		}
	}

	w.Lock()
	_, err := w.fileWriter.Write([]byte(msg))
	if err == nil {
		w.maxLinesCurLines++
		w.maxSizeCurSize += len(msg)
	}
	w.Unlock()
	return err
}

func (w *fileLogWriter) createLogFile() (*os.File, error) {
	// Open the log file
	perm, err := strconv.ParseInt(w.perm, 8, 64)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(w.filename, os.FileMode(perm))
	}
	return fd, err
}

func (w *fileLogWriter) initFd() error {
	fd := w.fileWriter
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s", err)
	}
	w.maxSizeCurSize = int(fInfo.Size())
	w.dailyOpenTime = time.Now()
	w.dailyOpenDate = w.dailyOpenTime.Day()
	w.maxLinesCurLines = 0
	if w.daily {
		go w.dailyRotate(w.dailyOpenTime)
	}
	if fInfo.Size() > 0 {
		count, err := w.lines()
		if err != nil {
			return err
		}
		w.maxLinesCurLines = count
	}
	return nil
}

func (w *fileLogWriter) dailyRotate(openTime time.Time) {
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	<-tm.C
	w.Lock()
	if w.needRotate(0, time.Now().Day()) {
		if err := w.doRotate(time.Now()); err != nil {
			Log.Critical("dailyRotate FileLogWriter err (%q) : %s", w.filename, err)
			//fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		}
	}
	w.Unlock()
}

func (w *fileLogWriter) lines() (int, error) {
	fd, err := os.Open(w.filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// DoRotate means it need to write file in new file.
// new file name like xx.2013-01-01.log (daily) or xx.001.log (by line or size)
func (w *fileLogWriter) doRotate(logTime time.Time) error {
	// file exists
	// Find the next available number
	num := 1
	fName := ""

	_, err := os.Lstat(w.filename)
	if err != nil {
		//even if the file is not exist or other ,we should RESTART the logger
		goto RESTART_LOGGER
	}

	if w.maxLines > 0 || w.maxSize > 0 {
		for ; err == nil && num <= 999; num++ {
			fName = w.fileNameOnly + fmt.Sprintf(".%s.%03d%s", logTime.Format("2006-01-02"), num, w.suffix)
			_, err = os.Lstat(fName)
		}
	} else {
		fName = fmt.Sprintf("%s.%s%s", w.fileNameOnly, w.dailyOpenTime.Format("2006-01-02"), w.suffix)
		_, err = os.Lstat(fName)
		for ; err == nil && num <= 999; num++ {
			fName = w.fileNameOnly + fmt.Sprintf(".%s.%03d%s", w.dailyOpenTime.Format("2006-01-02"), num, w.suffix)
			_, err = os.Lstat(fName)
		}
	}
	// return error if the last file checked still existed
	if err == nil {
		return fmt.Errorf("Rotate: Cannot find free log number to rename %s", w.filename)
	}

	// close fileWriter before rename
	w.fileWriter.Close()

	// Rename the file to its new found name
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(w.filename, fName)
	if err != nil {
		goto RESTART_LOGGER
	}
	err = os.Chmod(fName, os.FileMode(0440))
	// re-start logger
RESTART_LOGGER:
	startLoggerErr := w.startLogger()
	if startLoggerErr != nil {
		return fmt.Errorf("Rotate StartLogger: %s", startLoggerErr)
	}
	if err != nil {
		return fmt.Errorf("Rotate: %s", err)
	}
	return nil
}

// Destroy close the file description, close file writer.
func (w *fileLogWriter) Destroy() {
	w.fileWriter.Close()
}

// Flush flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (w *fileLogWriter) Flush() {
	w.fileWriter.Sync()
}
