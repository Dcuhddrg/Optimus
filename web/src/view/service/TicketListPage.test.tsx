import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { LanguageProvider } from '../../locale/LanguageProvider';
import { setStoredLocale } from '../../locale/storage';
import { ticketMockEngine } from './data/ticketMockEngine';
import { ticketStore } from './data/ticketStore';
import { TicketListPage } from './TicketListPage';

let container: HTMLDivElement;
let root: Root;

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  setStoredLocale('zh-CN');
  ticketMockEngine.__resetForTest();
  ticketStore.__resetForTest();
  container = document.createElement('div');
  document.body.appendChild(container);
  root = createRoot(container);
});

afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
  ticketMockEngine.__resetForTest();
});

describe('TicketListPage 审批进度', () => {
  it('approving 行显示进度、当前角色和审批人', async () => {
    const id = ticketStore.createTicket({
      serviceId: 'cloud.compute-resize',
      title: '扩容',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    }).ticket.id;
    ticketStore.advance(id, 'submitted', 'approving');
    ticketStore.advanceApproval(id);

    await act(async () =>
      root.render(
        <LanguageProvider>
          <MemoryRouter>
            <TicketListPage />
          </MemoryRouter>
        </LanguageProvider>,
      ),
    );

    expect(container.textContent).toContain('审批 1/3 · 安全审批：李强');
  });
});
