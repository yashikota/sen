import { describe, expect, it } from 'vite-plus/test';
import { STATIC_COMMANDS, filterCommands } from './commands.ts';

describe('filterCommands', () => {
  it('returns all commands for an empty query', () => {
    expect(filterCommands(STATIC_COMMANDS, '')).toHaveLength(STATIC_COMMANDS.length);
  });

  it('matches title and keywords', () => {
    const hits = filterCommands(STATIC_COMMANDS, 'adr');
    expect(hits.map((c) => c.id)).toContain('new-page');
    expect(hits.map((c) => c.id)).toContain('goto-pages');
  });

  it('matches status commands', () => {
    const hits = filterCommands(STATIC_COMMANDS, 'done');
    expect(hits).toHaveLength(1);
    expect(hits[0]?.id).toBe('set-status-done');
  });
});
