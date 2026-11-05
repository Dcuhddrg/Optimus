// web/src/view/service/data/approvalChains.ts
// 按服务类别的 mock 审批链与审批人目录。创建工单时用 buildApprovalChain 快照进工单。

import type { ApprovalStep, ServiceCategory } from '../ticketTypes';
import { getCatalogById } from './serviceCatalog';

interface ApproverDef {
  roleI18nKey: string;
  approver: { name: string; email: string };
}

const APPROVERS: Record<'lineManager' | 'security' | 'deptHead', ApproverDef> = {
  lineManager: {
    roleI18nKey: 'servicePortal.ticket.approval.role.lineManager',
    approver: { name: '王明', email: 'wang.ming@optimus.io' },
  },
  security: {
    roleI18nKey: 'servicePortal.ticket.approval.role.security',
    approver: { name: '李强', email: 'li.qiang@optimus.io' },
  },
  deptHead: {
    roleI18nKey: 'servicePortal.ticket.approval.role.deptHead',
    approver: { name: '张伟', email: 'zhang.wei@optimus.io' },
  },
};

type RoleKey = keyof typeof APPROVERS;

const CHAIN_TEMPLATES: Record<ServiceCategory, RoleKey[]> = {
  account: ['lineManager'],
  cloud: ['lineManager', 'security', 'deptHead'],
  network: ['lineManager', 'security'],
  env: ['lineManager', 'deptHead'],
  support: ['lineManager', 'security', 'deptHead'],
};

const DEFAULT_CHAIN: RoleKey[] = ['lineManager'];

// 返回独立副本，避免工单改写共享的模板。
export function buildApprovalChain(serviceId: string | null): ApprovalStep[] {
  const category = serviceId ? getCatalogById(serviceId)?.category : undefined;
  const roleKeys = (category && CHAIN_TEMPLATES[category]) || DEFAULT_CHAIN;
  return roleKeys.map((key) => {
    const def = APPROVERS[key];
    return {
      roleI18nKey: def.roleI18nKey,
      approver: { ...def.approver },
      status: 'pending' as const,
    };
  });
}
