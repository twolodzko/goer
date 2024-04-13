package pids

import (
	"fmt"

	"github.com/twolodzko/goer/types"
)

// Process id, messages can be send and received by it.
// Internally, it is Go's channel.
// Pid is a channel for communication between processes.
type Pid struct {
	channel chan types.Expr
}

// Initialize new pid.
func NewPid() Pid {
	msg := make(chan types.Expr)
	return Pid{msg}
}

// Messages to receive messages.
func (p Pid) Messages() <-chan types.Expr {
	return p.channel
}

// Send the message to pid.
func (p Pid) Send(msg types.Expr) {
	go func() {
		defer func() { recover() }()
		p.channel <- msg
	}()
}

// Close the channel opened for the process.
func (p Pid) Close() {
	close(p.channel)
}

func (this Pid) Equal(other Pid) bool {
	return this.channel == other.channel
}

func (p Pid) String() string {
	return fmt.Sprintf("<%v>", p.channel)
}
