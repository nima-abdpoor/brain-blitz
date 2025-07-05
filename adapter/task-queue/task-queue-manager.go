package task_queue

import (
	"context"
	"fmt"
	"time"
)

type Option interface {
	// String returns a string representation of the option.
	String() string

	// Type return the type of the option
	Type() OptionType

	// Value returns a value used to create this option.
	Value() interface{}
}

func (n retryOption) String() string     { return fmt.Sprintf("MaxRetry(%d)", int(n)) }
func (n retryOption) Type() OptionType   { return MaxRetryOpt }
func (n retryOption) Value() interface{} { return int(n) }

func (t processAtOption) String() string {
	return fmt.Sprintf("ProcessAt(%v)", time.Time(t).Format(time.UnixDate))
}
func (t processAtOption) Type() OptionType   { return ProcessAtOpt }
func (t processAtOption) Value() interface{} { return time.Time(t) }

func (d processInOption) String() string     { return fmt.Sprintf("ProcessIn(%v)", time.Duration(d)) }
func (d processInOption) Type() OptionType   { return ProcessInOpt }
func (d processInOption) Value() interface{} { return time.Duration(d) }

type (
	retryOption     int
	processAtOption time.Time
	processInOption time.Duration
)

type OptionType int

const (
	MaxRetryOpt OptionType = iota
	QueueOpt
	TimeoutOpt
	DeadlineOpt
	UniqueOpt
	ProcessAtOpt
	ProcessInOpt
	TaskIDOpt
	RetentionOpt
	GroupOpt
)

type HandlerFunc func(ctx context.Context, payload map[string]interface{}) error

type PublishNewTask interface {
	Publish(ctx context.Context, taskType string, payload any, opt ...Option) error
}

type ProcessTask interface {
	Process(ctx context.Context, handlers map[string]HandlerFunc) error
}
