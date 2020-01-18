package ubtypes

import (
	"time"
)

type Backup struct {
	Started time.Time
	Target  BackupTarget
}

type BackupTarget struct {
	ServiceName   string   `json:"service_name"`
	BackupCommand []string `json:"backup_command"`
	TaskId        string   `json:"task_id,omitempty"`
}

// makes a backup struct with "now" as start timestamp
func BackupForTarget(target BackupTarget) Backup {
	return Backup{
		Started: time.Now(),
		Target:  target,
	}
}
