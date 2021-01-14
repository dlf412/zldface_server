package cache

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Lockabler interface {
	LockID() string
}

type Locker sync.Locker

var Locks = sync.Map{}
var lock_expire = 10 * time.Second

func Mutex(l Lockabler, multi bool) Locker {
	key := fmt.Sprintf("%s%s%s", reflect.TypeOf(l).String(), "#", l.LockID())
	var m interface{}
	if !multi {
		m = &sync.Mutex{}
	} else {
		m = &RedisLock{lockKey: key, timeout: lock_expire}
	}
	r, _ := Locks.LoadOrStore(key, m)
	return r.(Locker)
}
