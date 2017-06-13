package s_timer

import (
	"sync/atomic"
	"time"
)

const (
	mainChanLength = 1024
)

const (
	afterFunc = iota
	afterFuncWithAfterFuncFinishedCallback
	execute
	stop
	stopWithStopFinishedCallback
	reset
	resetWithResetFinishedCallback
	remove
	removeWithRemoveFinishedCallback
)

type event struct {
	_type int
	arg   interface{}
}

type afterFuncEvent struct {
	id uint64
	d  time.Duration
	f  func()
}

type afterFuncWithAfterFuncFinishedCallbackEvent struct {
	id uint64
	d  time.Duration
	f  func()
	cb func(uint64)
}

type executeEvent struct {
	id uint64
	f  func()
}

type stopEvent struct {
	id uint64
}

type stopWithStopFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type resetEvent struct {
	id uint64
	d  time.Duration
}

type resetWithResetFinishedCallbackEvent struct {
	id uint64
	d  time.Duration
	f  func(bool)
}

type removeEvent struct {
	id uint64
}

type removeWithRemoveFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type Timer struct {
	allocID  uint64
	mainChan chan *event
	timers   map[uint64]*time.Timer
}

func New() *Timer {
	t := new(Timer)
	t.allocID = 0
	t.mainChan = make(chan *event, mainChanLength)
	t.timers = make(map[uint64]*time.Timer)
	return t
}

func (t *Timer) GetMainChan() <-chan *event {
	return t.mainChan
}

func (t *Timer) Deal(e *event) {
	switch e._type {
	case afterFunc:
		arg := e.arg.(*afterFuncEvent)
		t.timers[arg.id] = time.AfterFunc(arg.d, func() {
			t.mainChan <- &event{
				_type: execute,
				arg: &executeEvent{
					id: arg.id,
					f:  arg.f,
				},
			}
		})
	case afterFuncWithAfterFuncFinishedCallback:
		arg := e.arg.(*afterFuncWithAfterFuncFinishedCallbackEvent)
		t.timers[arg.id] = time.AfterFunc(arg.d, func() {
			t.mainChan <- &event{
				_type: execute,
				arg: &executeEvent{
					id: arg.id,
					f:  arg.f,
				},
			}
		})
		arg.cb(arg.id)
	case execute:
		arg := e.arg.(*executeEvent)
		arg.f()
		delete(t.timers, arg.id)
	case stop:
		arg := e.arg.(*stopEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Stop()
		}
	case stopWithStopFinishedCallback:
		arg := e.arg.(*stopWithStopFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Stop()
		}
		arg.f(result)
	case reset:
		arg := e.arg.(*resetEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Reset(arg.d)
		}
	case resetWithResetFinishedCallback:
		arg := e.arg.(*resetWithResetFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Reset(arg.d)
		}
		arg.f(result)
	case remove:
		arg := e.arg.(*removeEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Stop()
			delete(t.timers, arg.id)
		}
	case removeWithRemoveFinishedCallback:
		arg := e.arg.(*removeWithRemoveFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			timer.Stop()
			delete(t.timers, arg.id)
			result = true
		}
		arg.f(result)
	}
}

func (t *Timer) AfterFunc(d time.Duration, f func()) uint64 {
	id := atomic.AddUint64(&t.allocID, 1)
	go func() {
		t.mainChan <- &event{
			_type: afterFunc,
			arg: &afterFuncEvent{
				id: id,
				d:  d,
				f:  f,
			},
		}
	}()
	return id
}

func (t *Timer) AfterFuncWithAfterFuncFinishedCallback(d time.Duration, f func(), cb func(uint64)) uint64 {
	id := atomic.AddUint64(&t.allocID, 1)
	go func() {
		t.mainChan <- &event{
			_type: afterFuncWithAfterFuncFinishedCallback,
			arg: &afterFuncWithAfterFuncFinishedCallbackEvent{
				id: id,
				d:  d,
				f:  f,
				cb: cb,
			},
		}
	}()
	return id
}

func (t *Timer) Stop(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: stop,
			arg: &stopEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) StopWithStopFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: stopWithStopFinishedCallback,
			arg: &stopWithStopFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Reset(id uint64, d time.Duration) {
	go func() {
		t.mainChan <- &event{
			_type: reset,
			arg: &resetEvent{
				id: id,
				d:  d,
			},
		}
	}()
}

func (t *Timer) ResetWithResetFinishedCallback(id uint64, d time.Duration, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: resetWithResetFinishedCallback,
			arg: &resetWithResetFinishedCallbackEvent{
				id: id,
				d:  d,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Remove(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: remove,
			arg: &removeEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) RemoveWithRemoveFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: removeWithRemoveFinishedCallback,
			arg: &removeWithRemoveFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}
