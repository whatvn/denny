package etcd

import (
	"context"
	"github.com/whatvn/denny/naming"
	"go.etcd.io/etcd/clientv3"
	"time"
)

// Register starts etcd client and check for registered key, if it's not available
// in etcd storage, it will write service host:port to etcd and start watching to keep writing
// if its data is not available again
func (r *etcd) Register(addr string, ttl int) error {

	var (
		ticker  = time.NewTicker(time.Second * time.Duration(ttl))
		err     error
		svcPath = "/" + naming.Prefix + "/" + r.serviceName + "/" + addr
	)

	r.Infof("register %s with registy", svcPath)
	err = r.register(addr, ttl)
	if err != nil {
		r.Errorf("error %v", err)
	}

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				resp, err := r.cli.Get(context.Background(), svcPath)
				if err != nil {
					r.Errorf("error %v", err)
				} else if resp.Count == 0 {
					err = r.register(addr, ttl)
					if err != nil {
						r.Errorf("error %v", err)
					}
				}
			case _ = <-r.shutdown:
				// receive message from shutdown channel
				// will stop current thread and stop ticker to prevent thread leak
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (r *etcd) register(addr string, ttl int) error {
	leaseResp, err := r.cli.Grant(context.Background(), int64(ttl))
	if err != nil {
		return err
	}

	_, err = r.cli.Put(context.Background(), "/"+naming.Prefix+"/"+r.serviceName+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	_, err = r.cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister deletes itself in etcd storage
// also stop watch goroutine and ticker inside it
func (r *etcd) UnRegister(name string, addr string) error {
	_, _ = r.cli.Delete(context.Background(), "/"+naming.Prefix+"/"+name+"/"+addr)
	r.shutdown <- "stop"
	return nil
}
