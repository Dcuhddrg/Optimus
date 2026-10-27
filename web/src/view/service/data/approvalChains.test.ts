import { describe, expect, it } from 'vitest';

import { buildApprovalChain } from './approvalChains';

describe('buildApprovalChain', () => {
  it('account 类服务返回 1 级链（直属主管），全部 pending', () => {
    const chain = buildApprovalChain('account.role-grant');
    expect(chain).toHaveLength(1);
    expect(chain[0].roleI18nKey).toBe('servicePortal.ticket.approval.role.lineManager');
    expect(chain[0].status).toBe('pending');
    expect(chain[0].approver.name).toBeTruthy();
    expect(chain[0].decidedAt).toBeUndefined();
  });

  it('cloud 类服务返回 3 级链', () => {
    const chain = buildApprovalChain('cloud.compute-resize');
    expect(chain).toHaveLength(3);
    expect(chain.every((s) => s.status === 'pending')).toBe(true);
  });

  it('network 类服务返回 2 级链', () => {
    expect(buildApprovalChain('network.acl')).toHaveLength(2);
  });

  it('null 或未知服务返回默认 1 级链', () => {
    expect(buildApprovalChain(null)).toHaveLength(1);
    expect(buildApprovalChain('does.not.exist')).toHaveLength(1);
  });

  it('返回独立副本，改一处不影响下次', () => {
    const a = buildApprovalChain('account.role-grant');
    a[0].status = 'approved';
    const b = buildApprovalChain('account.role-grant');
    expect(b[0].status).toBe('pending');
  });
});
