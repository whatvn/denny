package etcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/whatvn/denny/go_config/source"
	cetcd "go.etcd.io/etcd/clientv3"
)

type watcher struct {
	opts source.Options
	name string

	sync.RWMutex
	cs *source.ChangeSet

	ch   chan *source.ChangeSet
	exit chan bool
}

func newWatcher(key string, wc cetcd.Watcher, cs *source.ChangeSet, opts source.Options) (source.Watcher, error) {
	w := &watcher{
		opts: opts,
		name: "etcd",
		cs:   cs,
		ch:   make(chan *source.ChangeSet),
		exit: make(chan bool),
	}

	ch := wc.Watch(context.Background(), key)

	go w.run(wc, ch)

	return w, nil
}

func (w *watcher) handle(ev *cetcd.Event) {
	b := ev.Kv.Value
	// create new changeset
	cs := &source.ChangeSet{
		Timestamp: time.Now(),
		Source:    w.name,
		Data:      b,
		Format:    w.opts.Encoder.String(),
	}
	cs.Checksum = cs.Sum()

	// set base change set
	w.Lock()
	w.cs = cs
	w.Unlock()
	// send update
	w.ch <- cs
}

func (w *watcher) run(wc cetcd.Watcher, ch cetcd.WatchChan) {
	for {
		select {
		case rsp, ok := <-ch:
			if !ok {
				return
			}
			for _, ev := range rsp.Events {
				w.handle(ev)
			}
		case <-w.exit:
			_ = wc.Close()
			return
		}
	}
}

func (w *watcher) Next() (*source.ChangeSet, error) {
	select {
	case cs := <-w.ch:
		return cs, nil
	case <-w.exit:
		return nil, errors.New("watcher stopped")
	}
}

func (w *watcher) Stop() error {
	select {
	case <-w.exit:
		return nil
	default:
		close(w.exit)
	}
	return nil
}
