// web/src/view/service/data/taskTypes.ts

export type TaskStatus = 'pending' | 'in_progress' | 'done';

export interface TaskEvent {
  at: string; // ISO timestamp
  status: TaskStatus;
  i18nNoteKey?: string;
  actor?: string; // displayName，便于时间线展示
}

export interface OpsTask {
  id: string; // 'TASK-YYYY-NNNN'
  ticketId: string; // 关联工单
  title: string; // 创建时从工单拷贝
  serviceId: string | null;
  status: TaskStatus;
  assignee: { name: string; email: string } | null;
  createdAt: string; // ISO
  claimedAt?: string;
  completedAt?: string;
  timeline: TaskEvent[];
}

export type TaskStorageStatus = 'persisted' | 'memory-only';
