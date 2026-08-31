export type Command = {
  id: string;
  title: string;
  hint?: string;
  keywords?: string;
};

export const STATIC_COMMANDS: Command[] = [
  { id: 'new-issue', title: 'Create issue', hint: 'c', keywords: 'new' },
  { id: 'new-page', title: 'Create page', hint: 'p', keywords: 'new adr' },
  { id: 'new-view', title: 'Create view', keywords: 'new filter saved' },
  { id: 'goto-issues', title: 'Go to Issues', keywords: 'list' },
  { id: 'goto-board', title: 'Go to Board', keywords: 'kanban' },
  { id: 'goto-projects', title: 'Go to Projects' },
  { id: 'goto-cycles', title: 'Go to Cycles' },
  { id: 'goto-pages', title: 'Go to Pages', keywords: 'adr docs' },
  { id: 'goto-active-cycle', title: 'Go to active cycle', keywords: 'sprint current' },
  { id: 'copy-identifier', title: 'Copy identifier', keywords: 'id sen clipboard' },
  { id: 'keyboard-help', title: 'Keyboard shortcuts', hint: '?', keywords: 'help keys' },
  { id: 'set-status-backlog', title: 'Set status: Backlog', hint: 's' },
  { id: 'set-status-todo', title: 'Set status: Todo', hint: 's' },
  { id: 'set-status-in_progress', title: 'Set status: In Progress', hint: 's' },
  { id: 'set-status-done', title: 'Set status: Done', hint: 's' },
  { id: 'set-status-canceled', title: 'Set status: Canceled', hint: 's' },
];

export function cycleCommands(cycles: { id: number; number: number; status: string }[]): Command[] {
  const cmds = cycles.map((c) => ({
    id: `assign-cycle:${c.id}`,
    title: `Assign to Cycle ${c.number}`,
    keywords: `cycle ${c.status} sprint`,
  }));
  cmds.push({
    id: 'assign-cycle:none',
    title: 'Remove from cycle',
    keywords: 'unassign cycle none',
  });
  return cmds;
}

export function projectCommands(projects: { id: number; name: string; slug: string }[]): Command[] {
  const cmds = projects.map((p) => ({
    id: `assign-project:${p.id}`,
    title: `Assign to ${p.name}`,
    keywords: `project ${p.slug}`,
  }));
  cmds.push({
    id: 'assign-project:none',
    title: 'Remove from project',
    keywords: 'unassign project none',
  });
  return cmds;
}

export function filterCommands(commands: Command[], query: string): Command[] {
  const q = query.trim().toLowerCase();
  if (!q) {
    return commands;
  }
  return commands.filter((c) => {
    const hay = `${c.title} ${c.id} ${c.keywords ?? ''}`.toLowerCase();
    return hay.includes(q);
  });
}
