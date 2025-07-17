package domain_util

import (
	"sync"
	"sync/atomic"
)

type TaskProgress struct {
	ID             string
	TotalFiles     int32
	WalkedFiles    int32
	ProcessedFiles int32
	Mu             sync.Mutex
	Initialized    bool
	Status         string
}

func (tp *TaskProgress) AddTotalFiles(count int) {
	tp.Mu.Lock()
	defer tp.Mu.Unlock()
	atomic.AddInt32(&tp.TotalFiles, int32(count))
	tp.Initialized = true
}
