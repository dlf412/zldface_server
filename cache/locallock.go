package cache

import (
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
	key := l.LockID()
	var m interface{}
	if !multi {
		m = &sync.Mutex{}
	} else {
		m = &RedisLock{lockKey: key, timeout: lock_expire}
	}
	r, _ := Locks.LoadOrStore(key, m)
	return r.(Locker)
}
