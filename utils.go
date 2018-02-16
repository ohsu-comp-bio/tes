package tes

import (
	"bytes"
	"fmt"
	"github.com/getlantern/deepcopy"
	"github.com/golang/protobuf/jsonpb"
)

var mar = jsonpb.Marshaler{}

// Marshaler marshals tasks to indented JSON.
var Marshaler = jsonpb.Marshaler{
	Indent: "  ",
}

// MarshalToString marshals a task to an indented JSON string.
func MarshalToString(t *Task) (string, error) {
	if t == nil {
		return "", fmt.Errorf("can't marshal nil task")
	}
	return Marshaler.MarshalToString(t)
}

// Final returns true if the state is any of the final states:
//   complete, executor error, system error, canceled
func (s State) Final() bool {
	return s == Complete || s == ExecutorError || s == SystemError || s == Canceled
}

// Active returns true if the state is any of the active states:
//   queued, initializing, running
func (s State) Active() bool {
	return s == Queued || s == Initializing || s == Running
}

// GetBasicView returns the basic view of a task.
func (task *Task) GetBasicView() *Task {
	view := &Task{}
	deepcopy.Copy(view, task)

	// remove contents from inputs
	for _, v := range view.Inputs {
		v.Content = ""
	}

	// remove stdout and stderr from Task.Logs.Logs
	for _, tl := range view.Logs {
		for _, el := range tl.Logs {
			el.Stdout = ""
			el.Stderr = ""
		}
	}
	return view
}

// GetMinimalView returns the minimal view of a task.
func (task *Task) GetMinimalView() *Task {
	id := task.Id
	state := task.State
	return &Task{
		Id:    id,
		State: state,
	}
}

// GetTaskLog gets the task log entry at the given index "i".
// If the entry doesn't exist, empty logs will be appended up to "i".
func (task *Task) GetTaskLog(i int) *TaskLog {

	// Grow slice length if necessary
	for j := len(task.Logs); j <= i; j++ {
		task.Logs = append(task.Logs, &TaskLog{})
	}

	return task.Logs[i]
}

// GetExecLog gets the executor log entry at the given index "i".
// If the entry doesn't exist, empty logs will be appended up to "i".
func (task *Task) GetExecLog(attempt int, i int) *ExecutorLog {
	tl := task.GetTaskLog(attempt)

	// Grow slice length if necessary
	for j := len(tl.Logs); j <= i; j++ {
		tl.Logs = append(tl.Logs, &ExecutorLog{})
	}

	return tl.Logs[i]
}

func (task *Task) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	err := mar.Marshal(&b, task)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (task *Task) UnmarshalJSON(b []byte) error {
	by := bytes.NewBuffer(b)
	return jsonpb.Unmarshal(by, task)
}
