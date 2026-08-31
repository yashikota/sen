import { describe, expect, it } from 'vite-plus/test';
import { formatActivity } from './activity.ts';

describe('formatActivity', () => {
  it('describes a create', () => {
    expect(formatActivity('created', { identifier: 'SEN-1' })).toBe('Created SEN-1');
  });

  it('describes a status change', () => {
    expect(formatActivity('status_changed', { from: 'todo', to: 'done' })).toBe(
      'Status Todo → Done',
    );
  });

  it('describes a note', () => {
    expect(formatActivity('commented', {})).toBe('Added a note');
  });

  it('falls back to the raw action', () => {
    expect(formatActivity('mystery', {})).toBe('mystery');
  });

  it('describes a create without an identifier', () => {
    expect(formatActivity('created', {})).toBe('Created');
  });

  it('passes through unknown status values', () => {
    expect(formatActivity('status_changed', { from: 'ready', to: 'shipped' })).toBe(
      'Status ready → shipped',
    );
  });
});
