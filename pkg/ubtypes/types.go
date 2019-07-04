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
