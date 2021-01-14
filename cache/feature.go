package cache

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/utils"
)

var a = sync.Map{}
var faceMap = map[string]map[string]interface{}{}

func setGroupFaces(hkey string, features map[string]interface{}) error {
	if config.RedisCli != nil {
		if err := config.RedisCli.HSet(config.Rctx, hkey, features).Err(); err != nil {
			if err.Error() == "ERR wrong number of arguments for 'hset' command" {
				for k, v := range features {
					if err := config.RedisCli.HSet(config.Rctx, hkey, k, v).Err(); err != nil {
						config.Logger.Error("加载人脸特征出错", zap.String("group", hkey), zap.Error(err))
						return err
					}
				}
			} else {
				return err
			}
		}
	} else {
		if vals, ok := faceMap[hkey]; ok {
			for k, v := range features {
				vals[k] = v
			}
		} else {
			faceMap[hkey] = features
		}
	}
	return nil
}

func LoadAllFeatures() {
	// 加载所有的features, 用redis得hashset存储
	groups := []model.FaceGroup{}
	config.DB.Find(&groups)

	for _, g := range groups {
		hkey := fmt.Sprintf("face_group#%s", g.Gid)
		features := g.FaceFeatures()
		err := setGroupFaces(hkey, features)
		if err != nil {
			config.Logger.Error("加载人脸特征出错", zap.String("group", g.Gid), zap.Error(err))
		} else {
			config.Logger.Info("加载人脸特征成功", zap.String("group", g.Gid), zap.Int("count", len(features)))
		}
	}
}

func GetGroupFeatures(group *model.FaceGroup) map[string]interface{} {
	hkey := fmt.Sprintf("face_group#%s", group.Gid)
	if config.RedisCli != nil {
		vals, err := config.RedisCli.HGetAll(config.Rctx, hkey).Result()
		if err != nil { // 缓存异常或为空直接用数据库
			return group.FaceFeatures()
		} else {
			var f = map[string]interface{}{}
			for k, v := range vals {
				f[k] = interface{}(utils.Str2bytes(v))
			}
			return f
		}
	} else {
		if vals, ok := faceMap[hkey]; ok {
			return vals
		} else {
			return group.FaceFeatures()
		}
	}

}

func DelGroupFeatures(gid string, users []model.FaceUser) (err error) {
	hkey := fmt.Sprintf("face_group#%s", gid)
	uids := make([]string, len(users))
	for idx, u := range users {
		uids[idx] = u.Uid
	}
	if config.RedisCli != nil {
		return config.RedisCli.HDel(config.Rctx, hkey, uids...).Err()
	} else {
		if vals, ok := faceMap[hkey]; ok {
			for _, u := range uids {
				delete(vals, u)
			}
		}
		return nil
	}
}

func AddGroupFeatures(gid string, users []model.FaceUser) (err error) {
	hkey := fmt.Sprintf("face_group#%s", gid)
	features := map[string]interface{}{}
	for _, u := range users {
		if u.FaceFeature == nil || len(u.FaceFeature) != 1032 {
			continue
		}
		features[u.Uid] = u.FaceFeature
	}
	err = setGroupFaces(hkey, features)
	if err != nil {
		config.Logger.Error("group增加人脸特征失败", zap.String("group", gid))
		return
	}
	return
}

func UpdateUserFeature(user *model.FaceUser) (err error) {
	groups := []model.FaceGroup{}
	config.DB.Model(user).Association("Groups").Find(&groups)
	for _, g := range groups {
		hkey := fmt.Sprintf("face_group#%s", g.Gid)
		err = setGroupFaces(hkey, map[string]interface{}{user.Uid: user.FaceFeature})
		if err != nil {
			config.Logger.Error("更新人脸特征失败", zap.String("group", g.Gid), zap.String("user", user.Uid))
		}
	}
	return
}
