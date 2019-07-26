package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

type Log struct {
	*logrus.Entry
	sync.Mutex
	step int32
}

func New(funcName string) *Log {
	return &Log{
		Entry: logrus.WithField("funcName", funcName),
	}
}

func (l *Log) addStep() {
	l.Lock()
	defer l.Unlock()
	l.step += 1
}

func (l *Log) AddLine(line string) {
	l.addStep()
	step := fmt.Sprintf("step_%d", l.step)
	l.Entry = l.Entry.WithField(step, line)
}
