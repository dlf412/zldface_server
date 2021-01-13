package cache

import (
	"zldface_server/model"
)

var UpdateUserCh = make(chan *model.FaceUser, 100)
var AddUserCh = make(chan map[string][]model.FaceUser, 100)
var DelUserCh = make(chan map[string][]string, 100)
