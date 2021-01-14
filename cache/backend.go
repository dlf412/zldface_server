package cache

import (
	"time"
	"zldface_server/config"
	"zldface_server/model"
)

var UpdateUserCh = make(chan *model.FaceUser, 100)

func BeRun() {
	if config.MultiPoint {
		go func() {
			for {
				var rLock = RedisLock{
					lockKey: "##master_server##",
					value:   nil,
					timeout: time.Second * 30,
					loop:    time.Second * 10,
				}
				rLock.Lock()
				defer rLock.Unlock()
				for {
					if !rLock.Keep() {
						break
					} // 保持锁
				}
			}
		}()
	} else {
		LoadAllFeatures()
		go faceUpdateRun()
	}
}

func faceUpdateRun() {

	retries := map[string]*model.FaceUser{}
	for {
		select {
		case u := <-UpdateUserCh:
			for k, v := range retries {
				if err := UpdateUserFeature(v); err == nil {
					delete(retries, k)
				} else {
					break
				}
			}
			if err := UpdateUserFeature(u); err != nil {
				// 加入到重试里面
				retries[u.Uid] = u
			}
		case <-time.After(10 * time.Second): // 如空闲超过10秒检查重试
			for k, v := range retries {
				if err := UpdateUserFeature(v); err == nil {
					delete(retries, k)
				} else {
					break
				}
			}
		}
	}
}
