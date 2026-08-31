import { describe, expect, it } from 'vite-plus/test';
import { isOverdue, localToday } from './due.ts';

describe('isOverdue', () => {
  it('is true when the due day is before today', () => {
    expect(isOverdue('2026-08-01', '2026-09-01')).toBe(true);
  });

  it('is false on the due day and when unset', () => {
    expect(isOverdue('2026-09-01', '2026-09-01')).toBe(false);
    expect(isOverdue(null, '2026-09-01')).toBe(false);
    expect(isOverdue(undefined, '2026-09-01')).toBe(false);
  });

  it('is false when the due day is in the future', () => {
    expect(isOverdue('2026-09-02', '2026-09-01')).toBe(false);
  });
});

describe('localToday', () => {
  it('formats a local calendar date as YYYY-MM-DD', () => {
    expect(localToday(new Date(2026, 8, 1, 23, 59, 0))).toBe('2026-09-01');
  });
});
