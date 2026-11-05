// ApprovalStepper.tsx
// 详情页顶部横向步骤条：提交 → 各审批级 → 处理中 → 完成。节点状态由工单状态与审批链派生。
import { Check } from '@phosphor-icons/react';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { Ticket } from '../ticketTypes';
import styles from './ApprovalStepper.module.css';

type NodeState = 'done' | 'current' | 'upcoming' | 'rejected';

interface StepNode {
  key: string;
  label: string;
  sub?: string;
  state: NodeState;
}

export function ApprovalStepper({ ticket }: { ticket: Ticket }) {
  const { t } = useI18n();
  const { status, approvalChain } = ticket;
  const firstPendingIdx = approvalChain.findIndex((s) => s.status === 'pending');
  const pastApproval = status === 'processing' || status === 'done';

  const nodes: StepNode[] = [
    {
      key: 'submitted',
      label: t('servicePortal.ticket.status.submitted'),
      sub: t('servicePortal.ticket.approval.stepperApplicant'),
      state: status === 'submitted' ? 'current' : 'done',
    },
  ];

  approvalChain.forEach((step, i) => {
    const name = step.approver.name;
    let state: NodeState = 'upcoming';
    let suffix: MessageKey | null = 'servicePortal.ticket.approval.stepperWaiting';
    if (step.status === 'approved' || pastApproval) {
      state = 'done';
      suffix = null;
    } else if (i === firstPendingIdx && status === 'approving') {
      state = 'current';
      suffix = 'servicePortal.ticket.approval.stepperCurrent';
    } else if (i === firstPendingIdx && status === 'rejected') {
      state = 'rejected';
      suffix = 'servicePortal.ticket.approval.stepperRejected';
    }
    nodes.push({
      key: `approval-${i}`,
      label: t(step.roleI18nKey as MessageKey),
      sub: suffix ? `${name} · ${t(suffix)}` : name,
      state,
    });
  });

  nodes.push(
    {
      key: 'processing',
      label: t('servicePortal.ticket.status.processing'),
      state: status === 'done' ? 'done' : status === 'processing' ? 'current' : 'upcoming',
    },
    {
      key: 'done',
      label: t('servicePortal.ticket.status.done'),
      state: status === 'done' ? 'done' : 'upcoming',
    },
  );

  return (
    <ol className={styles.stepper} aria-label={t('servicePortal.ticket.approval.stepperLabel')}>
      {nodes.map((node) => (
        <li
          key={node.key}
          className={styles.node}
          data-state={node.state}
          aria-current={node.state === 'current' ? 'step' : undefined}
        >
          <span className={styles.dot} aria-hidden="true">
            {node.state === 'done' ? <Check size={11} weight="bold" /> : null}
          </span>
          <span className={styles.label}>{node.label}</span>
          {node.sub ? <span className={styles.sub}>{node.sub}</span> : null}
        </li>
      ))}
    </ol>
  );
}
