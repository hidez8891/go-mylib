// Package die provides error handling.
package die

import (
	"fmt"
	"runtime"
)

type message struct {
	caller string
	line   int
	msg    interface{}
}

// If raise msg if cond is true.
func If(cond bool, msg interface{}) {
	dieIf(cond, msg)
}

// IfErr raise err if err is not nil.
func IfErr(err error) {
	dieIf(err != nil, err)
}

// dieIF raise *message if cond is true.
func dieIf(cond bool, msg interface{}) {
	if cond {
		_, file, line, _ := runtime.Caller(2)
		panic(&message{
			caller: file,
			line:   line,
			msg:    msg,
		})
	}
}

// Revival rescue raise message.
func Revival(f func(interface{})) {
	revival(
		recover(),
		func(msg *message) {
			f(msg.msg)
		},
		func(other interface{}) {
			f(other)
		},
	)
}

// RevivalErr rescue raise error.
func RevivalErr(f func(error)) {
	revival(
		recover(),
		func(msg *message) {
			if err, ok := msg.msg.(error); ok {
				f(err)
			} else {
				f(fmt.Errorf("%v", msg.msg))
			}
		},
		func(other interface{}) {
			f(fmt.Errorf("%v", other))
		},
	)
}

// RevivalDetail rescue raise message and get file/line infomation.
func RevivalDetail(f func(string, int, interface{})) {
	revival(
		recover(),
		func(msg *message) {
			f(msg.caller, msg.line, msg.msg)
		},
		func(other interface{}) {
			f("", -1, other)
		},
	)
}

// revival executes f1() if m is *message, otherwise executes f2().
// if m is nil, skip execution.
func revival(m interface{}, f1 func(*message), f2 func(interface{})) {
	if m != nil {
		if msg, ok := m.(*message); ok {
			f1(msg)
		} else {
			f2(m)
		}
	}
}
