package log

import (
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

func (l *Log) addStep() {
	l.Lock()
	defer l.Unlock()
	l.step += 1
}

func (l *Log) AddLog(line string) {
	l.addStep()
	step := fmt.Sprintf("STEP_%d", l.step)
	l.Entry = l.Entry.WithField(step, line)
}

func (l *Log) WithField(field string, value interface{}) {
	l.Entry = l.Entry.WithField(field, value)
}
