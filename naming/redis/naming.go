package redis

import (
	"github.com/whatvn/denny/naming"
	"time"
)

func (r *redis) Register(addr string, ttl int) error {
	var (
		ticker  = time.NewTicker(time.Second * time.Duration(ttl))
		err     error
		svcPath = "/" + naming.Prefix + "/" + r.serviceName + "/" + addr
	)

	r.Infof("register %s with registy", svcPath)
	err = r.register(addr, ttl)
	if err != nil {
		r.Errorf("error %v", err)
		return err
	}

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				_ = r.register(addr, ttl)
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

func (r *redis) register(addr string, ttl int) error {
	var (
		svcPath = "/" + naming.Prefix + "/" + r.serviceName + "/" + addr
	)

	existCmd := r.cli.Exists(svcPath)
	val, err := existCmd.Result()

	if err != nil {
		return err
	}

	if val != 0 {
		// increase expired time
		touchCmd := r.cli.Expire(svcPath, time.Duration(ttl*2)*time.Second)
		return touchCmd.Err()
	}
	setCmd := r.cli.Set(svcPath, addr, time.Duration(ttl*2)*time.Second)
	return setCmd.Err()
}

func (r *redis) UnRegister(addr string) error {
	var (
		svcPath = "/" + naming.Prefix + "/" + r.serviceName + "/" + addr
	)
	r.shutdown <- "stop"
	return r.cli.Del(svcPath).Err()
}
