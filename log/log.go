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

func New() *Log {
	logTime := time.Now().Format(time.RFC3339)
	return &Log{
		Entry: logrus.WithField("Time", logTime),
	}
}

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

func (l *Log) AddLog(line string, format ...interface{}) {
	l.addStep()
	step := fmt.Sprintf("STEP_%d", l.step)
	if len(format) > 0 {
		logLine := fmt.Sprintf(line, format)
		l.Entry = l.Entry.WithField(step, logLine)
		return
	}
	l.Entry = l.Entry.WithField(step, line)
}

func (l *Log) WithField(field string, value interface{}) {
	l.Entry = l.Entry.WithField(field, value)
}

func (l *Log) WithFields(fields map[string]interface{}) {
	l.Entry = l.Entry.WithFields(fields)
}
