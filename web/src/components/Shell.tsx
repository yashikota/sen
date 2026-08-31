import { Link, Outlet, useNavigate, useRouter, useRouterState } from '@tanstack/react-router';
import { BookText, CircleDot, Filter, Kanban, Layers, ListTodo } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { api } from '../api.ts';
import { STATIC_COMMANDS, cycleCommands, filterCommands, projectCommands } from '../commands.ts';
import { actionFromKeyboard } from '../keymap.ts';
import { formatStamp } from '../time.ts';
import { ISSUE_STATUSES, PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type {
  Cycle,
  Diagnostic,
  Issue,
  IssueStatus,
  Project,
  SearchHit,
  View,
  Workspace,
} from '../types.ts';
import { Palette } from './Palette.tsx';
import { ShortcutHelp } from './ShortcutHelp.tsx';

function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 48);
}

export function Shell() {
  const navigate = useNavigate();
  const router = useRouter();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [diagnostics, setDiagnostics] = useState<Diagnostic[]>([]);
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [views, setViews] = useState<View[]>([]);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [hits, setHits] = useState<SearchHit[]>([]);
  const [createIssue, setCreateIssue] = useState(false);
  const [createPage, setCreatePage] = useState(false);
  const [createView, setCreateView] = useState(false);
  const [issueTitle, setIssueTitle] = useState('');
  const [issueStatus, setIssueStatus] = useState<IssueStatus>('todo');
  const [issuePriority, setIssuePriority] = useState(0);
  const [issueProjectId, setIssueProjectId] = useState('');
  const [issueCycleId, setIssueCycleId] = useState('');
  const [helpOpen, setHelpOpen] = useState(false);
  const [projects, setProjects] = useState<Project[]>([]);
  const [pageTitle, setPageTitle] = useState('');
  const [viewName, setViewName] = useState('');
  const [error, setError] = useState('');
  const [focusedIssue, setFocusedIssue] = useState<string | null>(null);

  const loadWorkspace = useCallback(async () => {
    try {
      const [ws, diags, cyc, vs, proj] = await Promise.all([
        api.workspace(),
        api.diagnostics(),
        api.cycles(),
        api.views(),
        api.projects(),
      ]);
      setWorkspace(ws);
      setDiagnostics(diags);
      setCycles(cyc);
      setViews(vs);
      setProjects(proj);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load workspace');
    }
  }, []);

  useEffect(() => {
    void loadWorkspace();
  }, [loadWorkspace, pathname]);

  useEffect(() => {
    if (!createIssue) {
      return;
    }
    void api
      .projects()
      .then(setProjects)
      .catch(() => setProjects([]));
  }, [createIssue]);

  useEffect(() => {
    function onFocus(e: Event) {
      const id = (e as CustomEvent<string>).detail;
      if (id) {
        setFocusedIssue(id);
      }
    }
    window.addEventListener('sen:issue', onFocus);
    return () => window.removeEventListener('sen:issue', onFocus);
  }, []);

  useEffect(() => {
    function onCreate(e: Event) {
      const detail = (e as CustomEvent<{ projectId?: number; cycleId?: number }>).detail ?? {};
      if (detail.projectId) {
        setIssueProjectId(String(detail.projectId));
      }
      if (detail.cycleId) {
        setIssueCycleId(String(detail.cycleId));
      }
      setCreateIssue(true);
    }
    window.addEventListener('sen:create-issue', onCreate);
    return () => window.removeEventListener('sen:create-issue', onCreate);
  }, []);

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
    : focusedIssue;

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
        case 'new-view':
          setCreateView(true);
          return;
        case 'goto-issues':
          await navigate({ to: '/issues', search: {} });
          return;
        case 'goto-board':
          await navigate({ to: '/board', search: {} });
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
        case 'goto-active-cycle': {
          const active = cycles.find((c) => c.status === 'active');
          if (active) {
            await navigate({
              to: '/cycles/$number',
              params: { number: String(active.number) },
            });
          } else {
            await navigate({ to: '/cycles' });
          }
          return;
        }
        case 'copy-identifier':
          if (currentIdentifier) {
            await navigator.clipboard.writeText(currentIdentifier).catch(() => undefined);
          }
          return;
        case 'keyboard-help':
          setHelpOpen(true);
          return;
        default:
          break;
      }
      if (id.startsWith('set-status-') && currentIdentifier) {
        const status = id.replace('set-status-', '') as IssueStatus;
        await api.patchIssue(currentIdentifier, { status });
        window.dispatchEvent(new Event('sen:refresh'));
        await router.invalidate();
        await navigate({
          to: '/issues/$identifier',
          params: { identifier: currentIdentifier },
        });
      }
      if (id.startsWith('assign-cycle:') && currentIdentifier) {
        const raw = id.slice('assign-cycle:'.length);
        const cycleId = raw === 'none' ? null : Number(raw);
        await api.patchIssue(currentIdentifier, { cycleId });
        window.dispatchEvent(new Event('sen:refresh'));
        await router.invalidate();
      }
      if (id.startsWith('assign-project:') && currentIdentifier) {
        const raw = id.slice('assign-project:'.length);
        const projectId = raw === 'none' ? null : Number(raw);
        await api.patchIssue(currentIdentifier, { projectId });
        window.dispatchEvent(new Event('sen:refresh'));
        await router.invalidate();
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
      if (id.startsWith('open-view:')) {
        await navigate({
          to: '/views/$slug',
          params: { slug: id.slice('open-view:'.length) },
        });
      }
    },
    [currentIdentifier, cycles, navigate, router],
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
        setCreateView(false);
        setHelpOpen(false);
        return;
      }
      if (action === 'help') {
        e.preventDefault();
        setHelpOpen((v) => !v);
        return;
      }
      if (action === 'find') {
        e.preventDefault();
        window.dispatchEvent(new Event('sen:find'));
        return;
      }
      if (paletteOpen || createIssue || createPage || createView || helpOpen) {
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
        void api.patchIssue(currentIdentifier, { priority: n }).then(() => {
          window.dispatchEvent(new Event('sen:refresh'));
          return router.invalidate();
        });
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [paletteOpen, createIssue, createPage, createView, helpOpen, currentIdentifier, router]);

  const commands = [
    ...hits.map((h) => ({
      id: `open-${h.kind}:${h.id}`,
      title: `${h.kind} ${h.id}  ${h.title}`,
    })),
    ...filterCommands(STATIC_COMMANDS, query),
    ...(currentIdentifier ? filterCommands(cycleCommands(cycles), query) : []),
    ...(currentIdentifier ? filterCommands(projectCommands(projects), query) : []),
  ];

  async function submitIssue() {
    const title = issueTitle.trim();
    if (!title) {
      return;
    }
    const issue: Issue = await api.createIssue({
      title,
      status: issueStatus,
      priority: issuePriority,
      projectId: issueProjectId ? Number(issueProjectId) : undefined,
      cycleId: issueCycleId ? Number(issueCycleId) : undefined,
    });
    setIssueTitle('');
    setIssueStatus('todo');
    setIssuePriority(0);
    setIssueProjectId('');
    setIssueCycleId('');
    setCreateIssue(false);
    await router.invalidate();
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
    await router.invalidate();
    await navigate({ to: '/pages/$slug', params: { slug: page.slug } });
  }

  async function submitView() {
    const name = viewName.trim();
    if (!name) {
      return;
    }
    const slug = slugify(name) || `view-${Date.now()}`;
    const view = await api.createView({ name, slug, display: 'list' });
    setViewName('');
    setCreateView(false);
    await loadWorkspace();
    await router.invalidate();
    await navigate({ to: '/views/$slug', params: { slug: view.slug } });
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">sen</div>
          <div className="brand-sub">{workspace?.name ?? 'workspace'}</div>
        </div>
        <nav aria-label="Primary">
          <Link
            to="/issues"
            search={{}}
            className="nav-link"
            activeProps={{ className: 'nav-link active' }}
          >
            <ListTodo size={14} aria-hidden /> Issues
          </Link>
          <Link
            to="/board"
            search={{}}
            className="nav-link"
            activeProps={{ className: 'nav-link active' }}
          >
            <Kanban size={14} aria-hidden /> Board
          </Link>
          <Link to="/projects" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <Layers size={14} aria-hidden /> Projects
          </Link>
          <Link to="/cycles" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <CircleDot size={14} aria-hidden /> Cycles
          </Link>
          {cycles
            .filter((c) => c.status === 'active')
            .map((c) => (
              <Link
                key={c.number}
                to="/cycles/$number"
                params={{ number: String(c.number) }}
                className="nav-link nav-sub"
                activeProps={{ className: 'nav-link nav-sub active' }}
              >
                Cycle {c.number} active
              </Link>
            ))}
          <Link to="/pages" className="nav-link" activeProps={{ className: 'nav-link active' }}>
            <BookText size={14} aria-hidden /> Pages
          </Link>
        </nav>
        <div className="nav-section">
          <div className="muted">Views</div>
          {views.map((v) => (
            <Link
              key={v.slug}
              to="/views/$slug"
              params={{ slug: v.slug }}
              className="nav-link"
              activeProps={{ className: 'nav-link active' }}
            >
              <Filter size={14} aria-hidden /> {v.name}
            </Link>
          ))}
          <button type="button" className="ghost nav-add" onClick={() => setCreateView(true)}>
            New view
          </button>
        </div>
        <button type="button" className="ghost palette-btn" onClick={() => setPaletteOpen(true)}>
          Command palette <span className="kbd">Mod+K</span>
        </button>
        <button type="button" className="ghost palette-btn" onClick={() => setHelpOpen(true)}>
          Shortcuts <span className="kbd">?</span>
        </button>
        {diagnostics.length > 0 ? (
          <div className="diag" role="status" aria-label="Workspace diagnostics">
            <div className="muted">Diagnostics</div>
            {diagnostics.map((d) => (
              <div key={`${d.path}:${d.code}`} title={d.message}>
                <span className="ident">{d.path}</span> {d.message}
              </div>
            ))}
          </div>
        ) : null}
        {workspace ? (
          <form
            className="workspace-card"
            onSubmit={(e) => {
              e.preventDefault();
              void api
                .patchWorkspace({
                  name: workspace.name,
                  ghcrRef: workspace.ghcrRef,
                  timezone: workspace.timezone,
                })
                .then(setWorkspace)
                .catch((err: unknown) =>
                  setError(err instanceof Error ? err.message : 'save failed'),
                );
            }}
          >
            <label className="muted" htmlFor="ws-name">
              Workspace
            </label>
            <input
              id="ws-name"
              value={workspace.name}
              onChange={(e) => setWorkspace({ ...workspace, name: e.target.value })}
            />
            <label className="sr-only" htmlFor="ws-ghcr">
              GHCR reference
            </label>
            <input
              id="ws-ghcr"
              value={workspace.ghcrRef}
              placeholder="ghcr.io/user/sen"
              onChange={(e) => setWorkspace({ ...workspace, ghcrRef: e.target.value })}
            />
            <label className="sr-only" htmlFor="ws-tz">
              Timezone
            </label>
            <input
              id="ws-tz"
              aria-label="Timezone"
              value={workspace.timezone}
              placeholder="UTC"
              onChange={(e) => setWorkspace({ ...workspace, timezone: e.target.value })}
            />
            {workspace.dirty ? (
              <div className="dirty" role="status">
                Unpushed
              </div>
            ) : workspace.lastPushedAt ? (
              <div className="muted">
                Pushed {formatStamp(workspace.lastPushedAt, workspace.timezone)}
              </div>
            ) : (
              <div className="muted">Never pushed</div>
            )}
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
      {helpOpen ? <ShortcutHelp onClose={() => setHelpOpen(false)} /> : null}
      {createIssue ? (
        <div className="overlay" onClick={() => setCreateIssue(false)}>
          <div
            className="dialog"
            role="dialog"
            aria-modal="true"
            aria-label="Create issue"
            onClick={(e) => e.stopPropagation()}
          >
            <input
              autoFocus
              aria-label="Issue title"
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
            <div className="dialog-fields">
              <label>
                <span>Status</span>
                <select
                  aria-label="Issue status"
                  value={issueStatus}
                  onChange={(e) => setIssueStatus(e.target.value as IssueStatus)}
                >
                  {ISSUE_STATUSES.map((s) => (
                    <option key={s} value={s}>
                      {STATUS_LABEL[s]}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span>Priority</span>
                <select
                  aria-label="Issue priority"
                  value={issuePriority}
                  onChange={(e) => setIssuePriority(Number(e.target.value))}
                >
                  {PRIORITY_LABEL.map((label, i) => (
                    <option key={label} value={i}>
                      {label}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span>Project</span>
                <select
                  aria-label="Issue project"
                  value={issueProjectId}
                  onChange={(e) => setIssueProjectId(e.target.value)}
                >
                  <option value="">No project</option>
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span>Cycle</span>
                <select
                  aria-label="Issue cycle"
                  value={issueCycleId}
                  onChange={(e) => setIssueCycleId(e.target.value)}
                >
                  <option value="">No cycle</option>
                  {cycles.map((c) => (
                    <option key={c.id} value={c.id}>
                      Cycle {c.number}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          </div>
        </div>
      ) : null}
      {createPage ? (
        <div className="overlay" onClick={() => setCreatePage(false)}>
          <div
            className="dialog"
            role="dialog"
            aria-modal="true"
            aria-label="Create page"
            onClick={(e) => e.stopPropagation()}
          >
            <input
              autoFocus
              aria-label="Page title"
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
      {createView ? (
        <div className="overlay" onClick={() => setCreateView(false)}>
          <div
            className="dialog"
            role="dialog"
            aria-modal="true"
            aria-label="Create view"
            onClick={(e) => e.stopPropagation()}
          >
            <input
              autoFocus
              aria-label="View name"
              placeholder="View name"
              value={viewName}
              onChange={(e) => setViewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  void submitView();
                }
              }}
            />
          </div>
        </div>
      ) : null}
    </div>
  );
}
