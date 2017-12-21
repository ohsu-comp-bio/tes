package tes

// State constants for convenience
const (
	Unknown       = State_UNKNOWN
	Queued        = State_QUEUED
	Initializing  = State_INITIALIZING
	Running       = State_RUNNING
	Paused        = State_PAUSED
	Complete      = State_COMPLETE
	ExecutorError = State_EXECUTOR_ERROR
	SystemError   = State_SYSTEM_ERROR
	Canceled      = State_CANCELED
)

// View constants for convenience
const (
	Minimal = TaskView_MINIMAL
	Basic   = TaskView_BASIC
	Full    = TaskView_FULL
)
