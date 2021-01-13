package cache

import (
	"fmt"
	"reflect"
	"sync"
	"time"
	"zldface_server/config"
)

type Lockabler interface {
	LockID() string
}

type Locker sync.Locker

var Locks = sync.Map{}

func Mutex(l Lockabler) Locker {
	key := fmt.Sprintf("%s%s%s", reflect.TypeOf(l).String(), "#", l.LockID())
	var m interface{}
	if config.Debug {
		m = &sync.Mutex{}
	} else {
		m = &RedisLock{lockKey: key, timeout: time.Second * 10}
	}
	r, _ := Locks.LoadOrStore(key, m)
	return r.(Locker)
}
