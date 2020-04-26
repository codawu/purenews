package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const FileMode = 0644

// FileLogger is an io.WriteCloser that writes to the specified filename.
type FileLogger struct {
	Filename string
	mu       sync.Mutex
	file     *os.File
}

func NewFileLogger(filename string) *FileLogger {
	return &FileLogger{Filename: filename}
}

// Write implements io.Writer.
func (logger *FileLogger) Write(p []byte) (n int, err error) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.file == nil {
		if err = logger.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	n, err = logger.file.Write(p)
	return n, err
}
func (logger *FileLogger) Sync() error {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	return logger.file.Sync()
}

// Reopen implements io.Closer, and reopens the current logfile by simply closing it.
func (logger *FileLogger) Reopen() error {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	return logger.close()
}

// openExistingOrNew opens the logfile if it exists. If there is no such file,
// a new file is created. Must be called when l.mu mutex is acquired.
func (logger *FileLogger) openExistingOrNew() error {
	_, err := os.Stat(logger.Filename)
	if os.IsNotExist(err) {
		return logger.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %q", err)
	}

	file, err := os.OpenFile(logger.Filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// If we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return logger.openNew()
	}
	logger.file = file
	return nil
}

// openNew opens a new log file for writing, moving any old log file out of the
// way. This methods assumes the file has already been closed.
func (logger *FileLogger) openNew() error {
	err := os.MkdirAll(logger.dir(), 0755)
	fmt.Println(logger.dir())
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %q", err)
	}

	// We use truncate here because this should only get called when we've closed
	// the renamed (by logrotate) file ourselves. If someone else creates the file
	// in the meantime, just wipe out the contents.
	fmt.Println(logger.Filename)
	f, err := os.OpenFile(logger.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(FileMode))
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	logger.file = f
	return nil
}

// dir returns the directory for the current filename.
func (logger *FileLogger) dir() string {
	return filepath.Dir(logger.Filename)
}

// close closes the file if it is open. Must be called when l.mu mutex is acquired.
func (logger *FileLogger) close() error {
	if logger.file == nil {
		return nil
	}
	err := logger.file.Close()
	logger.file = nil
	return err
}
