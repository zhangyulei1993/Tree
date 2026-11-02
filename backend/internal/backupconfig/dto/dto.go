package dto

type UpdateRequest struct {
	Enabled             *bool `json:"enabled" binding:"required"`
	RetentionDays       int   `json:"retentionDays" binding:"required"`
	MaxStorageMB        int   `json:"maxStorageMb" binding:"required"`
	MaxDiskUsagePercent int   `json:"maxDiskUsagePercent" binding:"required"`
}
