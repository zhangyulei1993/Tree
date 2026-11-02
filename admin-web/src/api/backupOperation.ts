import { apiClient, unwrapData } from '@/api/client'
import type { ApiResponse } from '@/types/api'

export interface BackupOperationConfig {
  enabled: boolean
  retentionDays: number
  maxStorageMb: number
  maxDiskUsagePercent: number
  updatedAt: string
}

export interface BackupOperationStatus {
  backupCount: number
  storageBytes: number
  latestBackupAt?: string | null
  diskUsagePercent: number
  backupDirectory: string
}

export interface BackupOperationOverview {
  config: BackupOperationConfig
  status: BackupOperationStatus
}

export async function getBackupOperationOverview() {
  return unwrapData<BackupOperationOverview>(await apiClient.get<ApiResponse<BackupOperationOverview>>('/admin/backup-operation-config'))
}

export async function updateBackupOperationConfig(input: Omit<BackupOperationConfig, 'updatedAt'>) {
  return unwrapData<BackupOperationConfig>(await apiClient.put<ApiResponse<BackupOperationConfig>>('/admin/backup-operation-config', input))
}
