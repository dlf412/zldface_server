package cache

import (
	"fmt"
	"go.uber.org/zap"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/utils"
)

func LoadAllFeatures() {
	// 加载所有的features, 用redis得hashset存储
	groups := []model.FaceGroup{}
	config.DB.Find(&groups)
FOR:
	for _, g := range groups {
		hkey := fmt.Sprintf("face_group#%s", g.Gid)
		features := g.FaceFeatures()
		if err := config.RedisCli.HSet(config.Rctx, hkey, features).Err(); err != nil {
			if err.Error() == "ERR wrong number of arguments for 'hset' command" {
				for k, v := range features {
					if err := config.RedisCli.HSet(config.Rctx, hkey, k, v).Err(); err != nil {
						config.Logger.Error("加载人脸特征出错", zap.String("group", g.Gid), zap.Error(err))
						continue FOR
					}
				}
				config.Logger.Info("加载人脸特征成功", zap.String("group", g.Gid), zap.Int("count", len(features)))
			} else {
				config.Logger.Error("加载人脸特征出错", zap.String("group", g.Gid), zap.Error(err))
			}
		} else {
			config.Logger.Info("加载人脸特征成功", zap.String("group", g.Gid), zap.Int("count", len(features)))
		}
	}
}

func GetGroupFeatures(group *model.FaceGroup) map[string]interface{} {
	hkey := fmt.Sprintf("face_group#%s", group.Gid)
	vals, _ := config.RedisCli.HGetAll(config.Rctx, hkey).Result()
	if vals == nil {
		return group.FaceFeatures()
	} else {
		var f = map[string]interface{}{}
		for k, v := range vals {
			f[k] = interface{}(utils.Str2bytes(v))
		}
		return f
	}
}

func DelGroupFeatures(gid string, users []model.FaceUser) (err error) {
	hkey := fmt.Sprintf("face_group#%s", gid)
	uids := make([]string, len(users))
	for idx, u := range users {
		uids[idx] = u.Uid
	}
	return config.RedisCli.HDel(config.Rctx, hkey, uids...).Err()
}

func AddGroupFeatures(gid string, users []model.FaceUser) (err error) {
	hkey := fmt.Sprintf("face_group#%s", gid)
	for _, u := range users {
		if u.FaceFeature == nil || len(u.FaceFeature) != 1032 {
			continue
		}
		err = config.RedisCli.HSet(config.Rctx, hkey, u.Uid, u.FaceFeature).Err()
		if err != nil {
			config.Logger.Error("group增加人脸特征失败", zap.String("group", gid), zap.String("user", u.Uid))
			return
		}
	}
	return
}

func UpdateUserFeature(user *model.FaceUser) (err error) {
	groups := []model.FaceGroup{}
	config.DB.Model(user).Association("Groups").Find(&groups)
	for _, g := range groups {
		hkey := fmt.Sprintf("face_group#%s", g.Gid)
		err = config.RedisCli.HSet(config.Rctx, hkey, user.Uid, user.FaceFeature).Err()
		if err != nil {
			config.Logger.Error("更新人脸特征失败", zap.String("group", g.Gid), zap.String("user", user.Uid))
		}
	}
	return
}
