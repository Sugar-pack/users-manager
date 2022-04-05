package logging

import (
	"fmt"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// StartDepth is a depth of stack trace,
	// because of log-wrapper, which is used here
	// this start depth is required to find caller correctly
	StartDepth int = 2
	// PathLen is a count of the directories to log, including file itself
	PathLen int = 2
	// DefaultFileNameLineKey is a field name in a logged record
	DefaultFileNameLineKey string = "where"
)

// GetFileLineHook prepares and returns filename line hook
func GetFileLineHook() log.Hook {
	return &FileLineHook{
		LogKeyName: DefaultFileNameLineKey,
	}
}

// FileLineHook contains caller's log settings
type FileLineHook struct {
	LogKeyName string `json:"field_name" yaml:"field_name"`
}

// Levels implements logrus's Hook interface
func (hook *FileLineHook) Levels() []log.Level {
	return log.AllLevels
}

// Fire implements logrus's Hook interface
func (hook *FileLineHook) Fire(entry *log.Entry) error {
	var (
		file string
		line int
	)
	for i := 0; i < 10; i++ {
		file, line = getCaller(StartDepth + i)
		if !strings.HasPrefix(file, "logrus") {
			break
		}
	}

	entry.Data[hook.LogKeyName] = fmt.Sprintf("%s:%d", file, line)
	return nil
}

func getCaller(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}

	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= PathLen {
				file = file[i+1:]
				break
			}
		}
	}

	return file, line
}
