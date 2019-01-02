package main

const (
	backupCommandEnvKey = "BACKUP_COMMAND"
)

type BackupTarget struct {
	ServiceName   string
	ContainerId   string
	BackupCommand string
}
