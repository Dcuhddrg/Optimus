// web/src/view/service/ticketTypes.ts

export type ServiceCategory = 'account' | 'cloud' | 'network' | 'env' | 'support';

export interface ServiceCatalogItem {
  id: string;
  category: ServiceCategory;
  i18nKey: string; // e.g. 'servicePortal.catalog.accountRoleGrant'
  iconToken: 'shield' | 'cloud' | 'network' | 'env' | 'support';
  tags: string[]; // i18n leaf keys, e.g. 'servicePortal.tag.approval'
  estimatedSla: string; // i18n leaf key
}

export type FieldType = 'text' | 'textarea' | 'select' | 'date' | 'multiselect';

export interface RequestFormFieldOption {
  value: string;
  i18nLabelKey: string;
}

export interface RequestFormField {
  name: string;
  type: FieldType;
  i18nLabelKey: string;
  required: boolean;
  options?: RequestFormFieldOption[];
  placeholderI18nKey?: string;
  maxLength?: number;
}

export type ServiceFormSchemas = Record<string, RequestFormField[]>;
export const GENERIC_FORM_KEY = '__generic';

export type TicketStatus = 'submitted' | 'approving' | 'processing' | 'done' | 'rejected';

export type ApprovalStepStatus = 'pending' | 'approved';

export interface ApprovalStep {
  roleI18nKey: string; // 角色名 i18n leaf key
  approver: { name: string; email: string };
  status: ApprovalStepStatus;
  decidedAt?: string; // ISO，通过时写入
}

export interface TicketEvent {
  at: string; // ISO timestamp
  status: TicketStatus;
  i18nNoteKey?: string;
  actor?: string; // 审批人 displayName
  roleI18nKey?: string; // 审批角色
}

export interface Ticket {
  id: string; // 'TKT-YYYY-NNNN'
  serviceId: string | null;
  title: string;
  formValues: Record<string, unknown>;
  status: TicketStatus;
  createdAt: string; // ISO
  timeline: TicketEvent[];
  applicant: { name: string; email: string };
}

export type StorageStatus = 'persisted' | 'memory-only';

export interface CreateTicketInput {
  serviceId: string | null;
  title: string;
  formValues: Record<string, unknown>;
  applicant: { name: string; email: string };
}

export interface CreateTicketResult {
  ticket: Ticket;
  storageStatus: StorageStatus;
}
