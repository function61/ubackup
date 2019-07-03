package main

import (
	"time"
)

const (
	backupCommandEnvKey = "BACKUP_COMMAND"
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
