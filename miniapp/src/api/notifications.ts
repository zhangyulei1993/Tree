import { request } from './client'

export interface NotificationItem { id: number | string; category: string; title: string; content: string; route?: string; read: boolean; createdAt: string }
export interface NotificationPage { items: NotificationItem[]; total: number; unreadCount: number; page: number; pageSize: number }

export function listNotifications(page = 1, pageSize = 20) { return request<NotificationPage>('/notifications', { params: { page, pageSize } }) }
export function markNotificationRead(notificationId: number | string) { return request<{ read: boolean }>(`/notifications/${notificationId}/read`, { method: 'POST' }) }
export function markAllNotificationsRead() { return request<{ read: boolean }>('/notifications/read-all', { method: 'POST' }) }
