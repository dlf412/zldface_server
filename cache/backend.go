package cache

import (
	"time"
	"zldface_server/config"
	"zldface_server/model"
)

var UpdateUserCh = make(chan *model.FaceUser, 1000)
var AddUserCh = make(chan map[string][]model.FaceUser, 1000)
var DelUserCh = make(chan map[string][]string, 1000)

func Run() {

	retries := map[string]*model.FaceUser{}
	for {
		select {
		case u := <-UpdateUserCh:
			for k, v := range retries {
				if err := UpdateUserFeature(v); err == nil {
					// 加入到重试里面
					delete(retries, k)
				} else {
					break
				}
			}
			if err := UpdateUserFeature(u); err != nil {
				// 加入到重试里面
				retries[u.Uid] = u
			}
		case <-time.After(5 * time.Second):
			config.Logger.Info("processing retries")
			for k, v := range retries {
				if err := UpdateUserFeature(v); err == nil {
					// 加入到重试里面
					delete(retries, k)
				} else {
					break
				}
			}

			//case u := <-AddUserCh:
			//	for k, v:= range u {
			//		if err := AddGroupFeatures(k, v); err != nil {
			//			config.Logger.Error(err.Error())
			//		}
			//	}
			//case u := <-DelUserCh:
			//	for k, v:= range u {
			//		if err := DelGroupFeatures(k, v); err != nil {
			//			config.Logger.Error(err.Error())
			//		}
			//	}

		}
	}
}
