import { STATUS_LABEL } from './types.ts';
import type { IssueStatus } from './types.ts';

export function formatActivity(action: string, payload: Record<string, unknown>): string {
  if (action === 'created') {
    const id = typeof payload.identifier === 'string' ? payload.identifier : '';
    return id ? `Created ${id}` : 'Created';
  }
  if (action === 'status_changed') {
    const from = typeof payload.from === 'string' ? payload.from : '';
    const to = typeof payload.to === 'string' ? payload.to : '';
    const fromLabel = STATUS_LABEL[from as IssueStatus] ?? from;
    const toLabel = STATUS_LABEL[to as IssueStatus] ?? to;
    return `Status ${fromLabel} → ${toLabel}`;
  }
  if (action === 'commented') {
    return 'Added a note';
  }
  return action;
}
