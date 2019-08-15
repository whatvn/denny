package log

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Log struct {
	*logrus.Entry
	sync.Mutex
	step int32
}

type TextFormatter = logrus.TextFormatter
type JSONFormatter = logrus.JSONFormatter
type Formatter = logrus.Formatter

// New return a new log object with log start time
func New(formatter ...Formatter) *Log {
	if len(formatter) > 0 {
		logrus.SetFormatter(formatter[0])
	}
	logTime := time.Now().Format(time.RFC3339)
	return &Log{
		Entry: logrus.WithField("Time", logTime),
	}
}

// ToJsonString convert an object into json string to beautify log
// return nil if marshalling error
func (l *Log) ToJsonString(input interface{}) string {
	if bytes, err := json.Marshal(input); err == nil {
		return string(bytes)
	}
	return ""
}

func (l *Log) addStep() {
	l.Lock()
	defer l.Unlock()
	l.step += 1
}

func (l *Log) init() {
	if l.Entry == nil {
		logTime := time.Now().Format(time.RFC3339)
		l.Entry = logrus.WithField("Time", logTime)
	}
}

// AddLog add a new field to log with step = current step + 1
func (l *Log) AddLog(line string, format ...interface{}) {
	l.init()
	l.addStep()
	step := fmt.Sprintf("STEP_%d", l.step)
	if len(format) > 0 {
		logLine := fmt.Sprintf(line, format)
		l.Entry = l.Entry.WithField(step, logLine)
		return
	}
	l.Entry = l.Entry.WithField(step, line)
}

// WithField a a new key = value to log with key = field, value = value
func (l *Log) WithField(field string, value interface{}) {
	l.Entry = l.Entry.WithField(field, value)
}

// WithFields add multiple key/value to log: key1 = value1, key2 = value2
func (l *Log) WithFields(fields map[string]interface{}) {
	l.Entry = l.Entry.WithFields(fields)
}
