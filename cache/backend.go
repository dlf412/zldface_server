package cache

import (
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/utils"
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
					res, _ := config.RedisCli.BRPop(config.Rctx, time.Second*10, "update_face_queue").Result()
					for idx, q := range res {
						if (idx & 1) == 1 {
							t := strings.SplitN(q, "@#$", 2)
							groups := []model.FaceGroup{}
							config.DB.Raw(`SELECT a.gid FROM face_groups AS a INNER JOIN face_users AS b INNER JOIN face_group_users AS c ON a.id = c.face_group_id AND b.id = c.face_user_id WHERE b.uid = ?;`,
								t[0]).Select("gid").Scan(&groups)
							for _, g := range groups {
								hkey := fmt.Sprintf("face_group#%s", g.Gid)
								err := setGroupFaces(hkey, map[string]interface{}{t[0]: utils.Str2bytes(t[1])})
								if err != nil {
									config.Logger.Error("更新人脸特征失败", zap.String("group", g.Gid), zap.String("user", t[0]))
								}
							}
						}
					}
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
