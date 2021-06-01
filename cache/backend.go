package cache

import (
	"bytes"
	"go.uber.org/zap"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"zldface_server/config"
	"zldface_server/utils"
)

type fileStream struct {
	path   string
	buffer *bytes.Buffer
}

type gidUid struct {
	gid string
	uid string
}

var saveFileTasks = make(chan fileStream, 100)
var hotFeatureTasks = make(chan gidUid, 100)

func UnsafeSaveFile(path string, src io.Reader) {
	b := bytes.NewBuffer(make([]byte, 0))
	io.Copy(b, src)
	saveFileTasks <- fileStream{path: path, buffer: b}
}

func HotFeautre(gid, uid string) {
	if config.MultiPoint {
		hotFeatureTasks <- gidUid{gid: gid, uid: uid}
	}
}

// 可以启动goroutine进行一些后台作业， 比如异步接口的处理，定时器触发等一些任务
func BeRun() {
	// 单节点模式需要先加载特征到内存
	if !config.MultiPoint {
		LoadAllFeatures()
	}
	tmpDir := os.TempDir()
	go func() {
		for {
			select {
			case gu := <-hotFeatureTasks:
				AddGroupFeatures(gu.gid, gu.uid)
			case fs := <-saveFileTasks:
				go func() {
					buf := fs.buffer.Bytes()
					_, err := utils.SaveReader(fs.buffer, fs.path)
					if err != nil {
						config.Logger.Warn("保存文件失败, 将保存到临时目录下",
							zap.String("tmpDir", tmpDir),
							zap.String("path", fs.path),
							zap.Error(err))
						{ // 保存到tmp中, 如果还保存失败则彻底丢失文件
							dirs := strings.Split(fs.path, "/")
							l := len(dirs)
							var tmpFile string
							if l >= 4 {
								tmpFile = path.Join(tmpDir, dirs[l-4], dirs[l-3], dirs[l-2], dirs[l-1])
							} else {
								tmpFile = path.Join(tmpDir, dirs[l-1])
							}
							err = utils.SaveBytes(buf, tmpFile)
							if err != nil {
								config.Logger.Error("系统出现异常导致文件丢失",
									zap.String("tmpFile", tmpFile), zap.Error(err))
							}
						}
					}
					fs.buffer.Reset()
				}()
			case <-time.After(time.Second * 300): // 处理一些定时器逻辑
				config.Logger.Info("time after trigger, do nothing...")
			}
		}
	}() //异步保存人脸匹配照，不阻塞接口
}
