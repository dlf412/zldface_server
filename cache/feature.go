package cache

import (
	"go.uber.org/zap"
	"strings"
	"time"
	"zldface_server/config"
	"zldface_server/model"
)

const group_prefix = "##face_group##"
const user_prefix = "##face_user##"

var groupFace = map[string]map[string]interface{}{} // group
var userFace = map[string]interface{}{}             // user

func setUserFace(uid string, f []byte) (err error) {
	if !config.MultiPoint {
		if uf, ok := userFace[uid]; ok {
			*uf.(*[]byte) = f // 地址指向的值变更
		} else {
			userFace[uid] = &f // 取地址
		}
	} else {
		return config.RedisCli.Set(config.Rctx, user_prefix+uid, f, time.Hour*100000).Err()
	}
	return
}

// group 关联user face的地址
func setGroupFace(gid string, uids ...string) (err error) {
	if !config.MultiPoint {
		for _, uid := range uids {
			face := userFace[uid]
			if face != nil {
				if groupFace[gid] == nil {
					groupFace[gid] = make(map[string]interface{})
				}
				groupFace[gid][uid] = face
			}
		}
	} else {
		slice := make([]interface{}, len(uids))
		for i, u := range uids {
			slice[i] = interface{}(user_prefix + u)
		}
		err = config.RedisCli.SAdd(config.Rctx, group_prefix+gid, slice...).Err()
		if err != nil {
			return
		}
	}
	return
}

func delGroupFace(gid string, uids ...string) error {
	if config.MultiPoint {
		slice := make([]interface{}, len(uids))
		for i, u := range uids {
			slice[i] = interface{}(user_prefix + u)
		}
		err := config.RedisCli.SRem(config.Rctx, group_prefix+gid, slice...).Err()
		return err
	} else {
		faces := groupFace[gid]
		for _, u := range uids {
			delete(faces, u)
		}
		return nil
	}
}

func getGroupFace(gid string) (res map[string]interface{}) {
	if !config.MultiPoint {
		return groupFace[gid]
	} else {
		uids, err := config.RedisCli.SMembers(config.Rctx, group_prefix+gid).Result()
		if err != nil {
			return
		}
		faces, err := config.RedisCli.MGet(config.Rctx, uids...).Result()
		if err != nil {
			return
		}

		res = make(map[string]interface{}, len(uids))
		for idx, uid := range uids {
			key := strings.TrimLeft(uid, user_prefix)
			if faces[idx] != nil {
				res[key] = faces[idx]
			}
		}
		return
	}
}

func LoadAllFeatures() {
	// 加载所有的features到内存或分布式缓存里
	users := []model.FaceUser{}
	config.DB.Preload("Groups").Find(&users)
	for idx := range users { // range是copy内存的，user struct占用内存较多，所以只遍历index
		u := users[idx]
		if len(u.FaceFeature) != 1032 {
			continue // 无效的face忽略掉
		}
		if err := setUserFace(u.Uid, u.FaceFeature); err != nil {
			config.Logger.Error("加载用户人脸特征发生错误", zap.String("user", u.Uid), zap.Error(err))
			return
		}
		for _, g := range u.Groups {
			if err := setGroupFace(g.Gid, u.Uid); err != nil {
				config.Logger.Error("加载分组人脸特征发生错误", zap.String("group", g.Gid), zap.Error(err))
				return
			}
		}
	}
	config.Logger.Info("加载用户人脸特征到内存成功", zap.Int("用户总数", len(users)))
}

func GetGroupFeatures(group *model.FaceGroup) map[string]interface{} {
	res := getGroupFace(group.Gid)
	if len(res) == 0 {
		return group.FaceFeatures()
	}
	return res
}

func DelGroupFeatures(gid string, uids ...string) (err error) {
	return delGroupFace(gid, uids...)
}

func AddGroupFeatures(gid string, uids ...string) (err error) {
	return setGroupFace(gid, uids...)
}

func UpdateUserFeature(uid string, feature []byte) (err error) {
	return setUserFace(uid, feature)
}
