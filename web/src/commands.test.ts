import { describe, expect, it } from 'vite-plus/test';
import { STATIC_COMMANDS, cycleCommands, filterCommands, projectCommands } from './commands.ts';

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

  it('matches by command id', () => {
    const hits = filterCommands(STATIC_COMMANDS, 'keyboard-help');
    expect(hits.map((c) => c.id)).toEqual(['keyboard-help']);
  });

  it('returns no matches for unknown text', () => {
    expect(filterCommands(STATIC_COMMANDS, 'assignee')).toEqual([]);
  });
});

describe('cycleCommands', () => {
  it('builds assign and remove commands from cycles', () => {
    const cmds = cycleCommands([
      { id: 10, number: 1, status: 'active' },
      { id: 11, number: 2, status: 'upcoming' },
    ]);
    expect(cmds.map((c) => c.id)).toEqual([
      'assign-cycle:10',
      'assign-cycle:11',
      'assign-cycle:none',
    ]);
    expect(cmds[0]?.title).toBe('Assign to Cycle 1');
  });
});

describe('projectCommands', () => {
  it('builds assign and remove commands from projects', () => {
    const cmds = projectCommands([
      { id: 4, name: 'Harbor', slug: 'harbor' },
      { id: 5, name: 'Dock', slug: 'dock' },
    ]);
    expect(cmds.map((c) => c.id)).toEqual([
      'assign-project:4',
      'assign-project:5',
      'assign-project:none',
    ]);
    expect(cmds[0]?.title).toBe('Assign to Harbor');
  });

  it('still offers remove when there are no projects', () => {
    expect(projectCommands([]).map((c) => c.id)).toEqual(['assign-project:none']);
  });
});

describe('cycleCommands empty', () => {
  it('still offers remove when there are no cycles', () => {
    expect(cycleCommands([]).map((c) => c.id)).toEqual(['assign-cycle:none']);
  });
});
