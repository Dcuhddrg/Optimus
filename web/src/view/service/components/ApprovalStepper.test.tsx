import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { LanguageProvider } from '../../../locale/LanguageProvider';
import { setStoredLocale } from '../../../locale/storage';
import type { ApprovalStep, Ticket, TicketStatus } from '../ticketTypes';
import { ApprovalStepper } from './ApprovalStepper';

let container: HTMLDivElement;
let root: Root;

beforeEach(() => {
  setStoredLocale('zh-CN');
  container = document.createElement('div');
  document.body.appendChild(container);
  root = createRoot(container);
});

afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});

const step = (role: string, name: string, status: ApprovalStep['status']): ApprovalStep => ({
  roleI18nKey: `servicePortal.ticket.approval.role.${role}`,
  approver: { name, email: `${name}@x.io` },
  status,
});

function ticket(status: TicketStatus, chain: ApprovalStep[]): Ticket {
  return {
    id: 'TKT-2026-0001',
    serviceId: 'cloud.compute-resize',
    title: 't',
    formValues: {},
    status,
    createdAt: '2026-06-04T00:00:00.000Z',
    timeline: [],
    applicant: { name: 'a', email: 'a@x.io' },
    approvalChain: chain,
  };
}

async function render(t: Ticket) {
  await act(async () => root.render(<LanguageProvider><ApprovalStepper ticket={t} /></LanguageProvider>));
  return [...container.querySelectorAll('li')].map((li) => ({
    state: li.dataset.state,
    text: li.textContent,
  }));
}

describe('ApprovalStepper', () => {
  it('审批中：已过级为 done，第一个待审级为 current 并显示审批人', async () => {
    const nodes = await render(
      ticket('approving', [
        step('lineManager', '王明', 'approved'),
        step('security', '李强', 'pending'),
        step('deptHead', '张伟', 'pending'),
      ]),
    );
    expect(nodes.map((n) => n.state)).toEqual(['done', 'done', 'current', 'upcoming', 'upcoming', 'upcoming']);
    expect(nodes[2].text).toContain('安全审批');
    expect(nodes[2].text).toContain('李强 · 审批中');
    expect(nodes[3].text).toContain('张伟 · 待审批');
  });

  it('处理中：审批级全部 done，处理中为 current', async () => {
    const nodes = await render(ticket('processing', [step('lineManager', '王明', 'approved')]));
    expect(nodes.map((n) => n.state)).toEqual(['done', 'done', 'current', 'upcoming']);
  });

  it('空链（旧工单）只显示提交、处理中、完成', async () => {
    const nodes = await render(ticket('done', []));
    expect(nodes).toHaveLength(3);
    expect(nodes.every((n) => n.state === 'done')).toBe(true);
  });
});
