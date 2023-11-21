package main

import (
	"time"
)

func CreateNewBackupDir() string {
	timestamp := time.Now().Format("05-04-15_01-02-2006")
	new_backup_dir := GetBackupDir() + "/" + timestamp
	CreateDir(new_backup_dir)

	return new_backup_dir
}
