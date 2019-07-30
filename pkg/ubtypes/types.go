package ubtypes

import (
	"time"
)

type Backup struct {
	Started time.Time
	Target  BackupTarget
}

type BackupTarget struct {
	ServiceName   string
	TaskId        string
	BackupCommand string
}

// makes a backup struct with "now" as start timestamp
func BackupForTarget(target BackupTarget) Backup {
	return Backup{
		Started: time.Now(),
		Target:  target,
	}
}
