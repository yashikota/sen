import { describe, expect, it } from 'vite-plus/test';
import { formatStamp } from './time.ts';

describe('formatStamp', () => {
  it('renders UTC instants in the workspace timezone', () => {
    const out = formatStamp('2026-09-01T00:00:00Z', 'Asia/Tokyo');
    expect(out).toContain('2026-09-01');
    expect(out).toContain('09:00');
  });

  it('falls back to the raw value when the instant is unusable', () => {
    expect(formatStamp('not-a-date', 'Asia/Tokyo')).toBe('not-a-date');
  });

  it('falls back to UTC when the timezone is unknown', () => {
    const out = formatStamp('2026-09-01T00:00:00Z', 'Not/A_Zone');
    expect(out).toContain('2026-09-01');
    expect(out).toContain('00:00');
  });

  it('renders a New York evening for a UTC midnight instant', () => {
    const out = formatStamp('2026-09-01T00:00:00Z', 'America/New_York');
    expect(out).toContain('2026-08-31');
    expect(out).toContain('20:00');
  });
});
