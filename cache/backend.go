package cache

import (
	"zldface_server/config"
)

// 可以进行一些后台作业， 比如异步接口的处理，定时器触发等一些任务
func BeRun() {
	if !config.MultiPoint || config.MultiPoint {
		LoadAllFeatures()
	}
}
