package ubtypes

import (
	"io"
	"time"
)

type Snapshotter interface {
	// describes, for logging purposes, how the snapshot will be obtained
	Describe() string
	// snapshots a target into a sink. usable only once.
	CreateSnapshot(snapshotSink io.Writer) error
}

type Backup struct {
	Started time.Time
	Target  BackupTarget
}

type BackupTarget struct {
	ServiceName string
	Snapshotter Snapshotter
	TaskId      string
}

// makes a backup struct with "now" as start timestamp
func BackupForTarget(target BackupTarget) Backup {
	return Backup{
		Started: time.Now(),
		Target:  target,
	}
}

type customStreamSnapshotter struct {
	fn func(sink io.Writer) error
}

func CustomStream(fn func(sink io.Writer) error) Snapshotter {
	return &customStreamSnapshotter{fn}
}

func (c *customStreamSnapshotter) Describe() string {
	return "custom stream"
}

func (c *customStreamSnapshotter) CreateSnapshot(sink io.Writer) error {
	return c.fn(sink)
}
