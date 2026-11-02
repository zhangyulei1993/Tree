package vo

import "time"

type Config struct {
	Enabled             bool      `json:"enabled"`
	RetentionDays       int       `json:"retentionDays"`
	MaxStorageMB        int       `json:"maxStorageMb"`
	MaxDiskUsagePercent int       `json:"maxDiskUsagePercent"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type Status struct {
	BackupCount      int        `json:"backupCount"`
	StorageBytes     int64      `json:"storageBytes"`
	LatestBackupAt   *time.Time `json:"latestBackupAt,omitempty"`
	DiskUsagePercent int        `json:"diskUsagePercent"`
	BackupDirectory  string     `json:"backupDirectory"`
}

type Overview struct {
	Config Config `json:"config"`
	Status Status `json:"status"`
}
