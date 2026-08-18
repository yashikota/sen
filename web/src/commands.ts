export type Command = {
  id: string;
  title: string;
  hint?: string;
  keywords?: string;
};

export const STATIC_COMMANDS: Command[] = [
  { id: 'new-issue', title: 'Create issue', hint: 'c', keywords: 'new' },
  { id: 'new-page', title: 'Create page', hint: 'p', keywords: 'new adr' },
  { id: 'goto-issues', title: 'Go to Issues', keywords: 'list' },
  { id: 'goto-board', title: 'Go to Board', keywords: 'kanban' },
  { id: 'goto-projects', title: 'Go to Projects' },
  { id: 'goto-cycles', title: 'Go to Cycles' },
  { id: 'goto-pages', title: 'Go to Pages', keywords: 'adr docs' },
  { id: 'set-status-backlog', title: 'Set status: Backlog', hint: 's' },
  { id: 'set-status-todo', title: 'Set status: Todo', hint: 's' },
  { id: 'set-status-in_progress', title: 'Set status: In Progress', hint: 's' },
  { id: 'set-status-done', title: 'Set status: Done', hint: 's' },
  { id: 'set-status-canceled', title: 'Set status: Canceled', hint: 's' },
];

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
