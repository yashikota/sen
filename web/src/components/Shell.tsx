import { Link, Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { BookText, CircleDot, Kanban, Layers, ListTodo } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { api } from '../api.ts';
import { STATIC_COMMANDS, filterCommands } from '../commands.ts';
import { actionFromKeyboard } from '../keymap.ts';
import type { Issue, IssueStatus, SearchHit, Workspace } from '../types.ts';
import { Palette } from './Palette.tsx';

function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 48);
}

export function Shell() {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [hits, setHits] = useState<SearchHit[]>([]);
  const [createIssue, setCreateIssue] = useState(false);
  const [createPage, setCreatePage] = useState(false);
  const [issueTitle, setIssueTitle] = useState('');
  const [pageTitle, setPageTitle] = useState('');
  const [error, setError] = useState('');

  const loadWorkspace = useCallback(async () => {
    try {
      setWorkspace(await api.workspace());
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load workspace');
    }
  }, []);

  useEffect(() => {
    void loadWorkspace();
  }, [loadWorkspace]);

  useEffect(() => {
    if (!query.trim()) {
      setHits([]);
      return;
    }
    const t = window.setTimeout(() => {
      void api
        .search(query)
        .then(setHits)
        .catch(() => setHits([]));
    }, 120);
    return () => window.clearTimeout(t);
  }, [query]);

  const currentIdentifier = pathname.startsWith('/issues/SEN-')
    ? pathname.slice('/issues/'.length)
    : null;

  const runCommand = useCallback(
    async (id: string) => {
      setPaletteOpen(false);
      setQuery('');
      switch (id) {
        case 'new-issue':
          setCreateIssue(true);
          return;
        case 'new-page':
          setCreatePage(true);
          return;
        case 'goto-issues':
          await navigate({ to: '/issues' });
          return;
        case 'goto-board':
          await navigate({ to: '/board' });
          return;
        case 'goto-projects':
          await navigate({ to: '/projects' });
          return;
        case 'goto-cycles':
          await navigate({ to: '/cycles' });
          return;
        case 'goto-pages':
          await navigate({ to: '/pages' });
          return;
        default:
          break;
      }
      if (id.startsWith('set-status-') && currentIdentifier) {
        const status = id.replace('set-status-', '') as IssueStatus;
        await api.patchIssue(currentIdentifier, { status });
        await navigate({
          to: '/issues/$identifier',
          params: { identifier: currentIdentifier },
        });
      }
      if (id.startsWith('open-issue:')) {
        await navigate({
          to: '/issues/$identifier',
          params: { identifier: id.slice('open-issue:'.length) },
        });
      }
      if (id.startsWith('open-project:')) {
        await navigate({
          to: '/projects/$slug',
          params: { slug: id.slice('open-project:'.length) },
        });
      }
      if (id.startsWith('open-page:')) {
        await navigate({
          to: '/pages/$slug',
          params: { slug: id.slice('open-page:'.length) },
        });
      }
    },
    [currentIdentifier, navigate],
  );

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const action = actionFromKeyboard(e);
      if (!action) {
        return;
      }
      if (action === 'palette') {
        e.preventDefault();
        setPaletteOpen((v) => !v);
        return;
      }
      if (action === 'escape') {
        setPaletteOpen(false);
        setCreateIssue(false);
        setCreatePage(false);
        return;
      }
      if (paletteOpen || createIssue || createPage) {
        return;
      }
      if (action === 'new-issue') {
        e.preventDefault();
        setCreateIssue(true);
      }
      if (action === 'new-page') {
        e.preventDefault();
        setCreatePage(true);
      }
      if (action === 'status' && currentIdentifier) {
        e.preventDefault();
        setPaletteOpen(true);
        setQuery('status');
      }
      if (action.startsWith('priority-') && currentIdentifier) {
        const n = Number(action.slice(-1));
        void api.patchIssue(currentIdentifier, { priority: n });
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [paletteOpen, createIssue, createPage, currentIdentifier]);

  const commands = [
    ...hits.map((h) => ({
      id: `open-${h.kind}:${h.id}`,
      title: `${h.kind} ${h.id}  ${h.title}`,
    })),
    ...filterCommands(STATIC_COMMANDS, query),
  ];

  async function submitIssue() {
    const title = issueTitle.trim();
    if (!title) {
      return;
    }
    const issue: Issue = await api.createIssue({ title });
    setIssueTitle('');
    setCreateIssue(false);
    await navigate({
      to: '/issues/$identifier',
      params: { identifier: issue.identifier },
    });
  }

  async function submitPage() {
    const title = pageTitle.trim();
    if (!title) {
      return;
    }
    const slug = slugify(title) || `page-${Date.now()}`;
    const page = await api.createPage({ title, slug });
    setPageTitle('');
    setCreatePage(false);
    await navigate({ to: '/pages/$slug', params: { slug: page.slug } });
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">sen</div>
          <div className="brand-sub">{workspace?.name ?? 'workspace'}</div>
        </div>
        <nav>
          <Link to="/issues" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <ListTodo size={14} /> Issues
          </Link>
          <Link to="/board" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <Kanban size={14} /> Board
          </Link>
          <Link to="/projects" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <Layers size={14} /> Projects
          </Link>
          <Link to="/cycles" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <CircleDot size={14} /> Cycles
          </Link>
          <Link to="/pages" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <BookText size={14} /> Pages
          </Link>
        </nav>
        {workspace ? (
          <form
            className="workspace-card"
            onSubmit={(e) => {
              e.preventDefault();
              void api
                .patchWorkspace({
                  name: workspace.name,
                  ghcrRef: workspace.ghcrRef,
                })
                .then(setWorkspace)
                .catch((err: unknown) =>
                  setError(err instanceof Error ? err.message : 'save failed'),
                );
            }}
          >
            <label className="muted">Workspace</label>
            <input
              value={workspace.name}
              onChange={(e) => setWorkspace({ ...workspace, name: e.target.value })}
            />
            <input
              value={workspace.ghcrRef}
              placeholder="ghcr.io/user/sen"
              onChange={(e) => setWorkspace({ ...workspace, ghcrRef: e.target.value })}
            />
            <button className="ghost" type="submit">
              Save
            </button>
          </form>
        ) : null}
      </aside>
      <div>
        {error ? <div className="error">{error}</div> : null}
        <Outlet />
      </div>
      {paletteOpen ? (
        <Palette
          query={query}
          onQuery={setQuery}
          commands={commands}
          onPick={(id) => void runCommand(id)}
          onClose={() => setPaletteOpen(false)}
        />
      ) : null}
      {createIssue ? (
        <div className="overlay" onClick={() => setCreateIssue(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <input
              autoFocus
              placeholder="Issue title"
              value={issueTitle}
              onChange={(e) => setIssueTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  void submitIssue();
                }
              }}
            />
          </div>
        </div>
      ) : null}
      {createPage ? (
        <div className="overlay" onClick={() => setCreatePage(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <input
              autoFocus
              placeholder="Page title"
              value={pageTitle}
              onChange={(e) => setPageTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  void submitPage();
                }
              }}
            />
          </div>
        </div>
      ) : null}
    </div>
  );
}
