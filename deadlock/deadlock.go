package deadlock

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var Config = struct {
	DeadlockTimeout time.Duration
	StackDepth      int
}{
	DeadlockTimeout: time.Second * 10,
	StackDepth:      50,
}

type stack struct {
	callers []uintptr
}

func lock(lockFn func(), m *stack) {
	var callers = make([]uintptr, Config.StackDepth)
	var length = runtime.Callers(3, callers)
	callers = callers[:length]

	ch := make(chan struct{})
	go func() {
		lockFn()
		close(ch)
		m.callers = callers
	}()
	t := time.NewTimer(Config.DeadlockTimeout)
	defer t.Stop()
	select {
	case <-t.C:
		fmt.Printf("A timeout ocurred (%s) while trying to acquire a lock at:\n", Config.DeadlockTimeout)
		printStackTrace(callers)
		fmt.Println("\nLast lock was at:")
		printStackTrace(m.callers)
		os.Exit(1)
	case <-ch:
	}
}

type Mutex struct {
	stack
	mu sync.Mutex
}

func (m *Mutex) Lock() {
	lock(m.mu.Lock, &m.stack)
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
	(&m.stack).callers = nil
}

type RWMutex struct {
	stack
	mu sync.RWMutex
}

func (m *RWMutex) Lock() {
	lock(m.mu.Lock, &m.stack)
}

func (m *RWMutex) Unlock() {
	m.mu.Unlock()
	(&m.stack).callers = nil
}

func (m *RWMutex) RLock() {
	//lock(m.mu.RLock, &m.stack)
	m.mu.RLock()
}

func (m *RWMutex) RUnlock() {
	m.mu.RUnlock()
	//(&m.stack).callers = nil
}

func printStackTrace(callers []uintptr) {
	var pc uintptr
	for _, v := range callers {
		pc = v - 1
		fun := runtime.FuncForPC(pc)
		if fun == nil {
			fmt.Println("n/a")
		} else {
			var fnName = fun.Name()
			var idx = strings.LastIndex(fnName, ".")
			if idx > 0 {
				fnName = fnName[idx+1:]
			}

			var file, line = fun.FileLine(pc)
			fmt.Printf("%s:%v %s()\n", file, line, fnName)
		}
	}
}
