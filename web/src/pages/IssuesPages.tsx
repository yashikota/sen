import {
  useLoaderData,
  useNavigate,
  useParams,
  useRouter,
  useSearch,
} from '@tanstack/react-router';
import { useState } from 'react';
import { api, parseIssueSearch, searchToFilter, type IssueSearch } from '../api.ts';
import { IssueBoard, IssueList } from '../components/IssueList.tsx';
import { IssueDetail } from '../components/IssueDetail.tsx';
import { IssueFilters } from '../components/IssueFilters.tsx';
import type { Cycle, Issue, Label, Project } from '../types.ts';

type IssueListData = {
  issues: Issue[];
  projects: Project[];
  cycles: Cycle[];
  labels: Label[];
};

function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 48);
}

function compactSearch(next: IssueSearch): IssueSearch {
  return parseIssueSearch({
    status: next.status ?? '',
    project: next.project ?? '',
    cycle: next.cycle ?? '',
    priority: next.priority ?? '',
    labels: next.labels ?? '',
  });
}

function matchesFind(issue: Issue, q: string): boolean {
  const n = q.trim().toLowerCase();
  if (!n) {
    return true;
  }
  return issue.title.toLowerCase().includes(n) || issue.identifier.toLowerCase().includes(n);
}

export function IssuesPage() {
  const data = useLoaderData({ from: '/issues' }) as IssueListData;
  const search = useSearch({ from: '/issues' }) as IssueSearch;
  const navigate = useNavigate();
  const router = useRouter();
  const [find, setFind] = useState('');
  const issues = (data.issues ?? []).filter((i) => matchesFind(i, find));
  const [selected, setSelected] = useState<string | null>(issues[0]?.identifier ?? null);
  if (selected && !issues.some((i) => i.identifier === selected)) {
    setSelected(issues[0]?.identifier ?? null);
  }

  async function saveView(name: string) {
    const filter = searchToFilter(search);
    const view = await api.createView({
      name,
      slug: slugify(name) || `view-${Date.now()}`,
      display: 'list',
      status: filter.status ?? null,
      project: filter.project ?? null,
      cycle: filter.cycle ?? null,
      labels: filter.labels ?? [],
      priority: filter.priority ?? null,
    });
    await router.invalidate();
    await navigate({ to: '/views/$slug', params: { slug: view.slug } });
  }

  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Issues</h1>
          <span className="muted">{issues.length}</span>
        </div>
        <IssueFilters
          search={search}
          projects={data.projects}
          cycles={data.cycles}
          labels={data.labels}
          onChange={(next) => void navigate({ to: '/issues', search: compactSearch(next) })}
          onSaveView={saveView}
          find={find}
          onFind={setFind}
        />
        <IssueList issues={issues} selectedId={selected} onSelect={setSelected} />
      </section>
      <section className="pane">
        {selected ? (
          <IssueDetail identifier={selected} />
        ) : (
          <div className="empty">
            Select an issue, or press <span className="kbd">c</span> to create.
          </div>
        )}
      </section>
    </div>
  );
}

export function IssueRoutePage() {
  const { identifier } = useParams({ from: '/issues/$identifier' });
  const issues = (useLoaderData({ from: '/issues/$identifier' }) as Issue[] | null) ?? [];
  const navigate = useNavigate();
  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Issues</h1>
          <span className="muted">{issues.length}</span>
        </div>
        <IssueList
          issues={issues}
          selectedId={identifier}
          onSelect={(id) =>
            void navigate({ to: '/issues/$identifier', params: { identifier: id } })
          }
        />
      </section>
      <section className="pane">
        <IssueDetail identifier={identifier} />
      </section>
    </div>
  );
}

export function BoardPage() {
  const data = useLoaderData({ from: '/board' }) as IssueListData;
  const search = useSearch({ from: '/board' }) as IssueSearch;
  const navigate = useNavigate();
  const router = useRouter();
  const [find, setFind] = useState('');
  const issues = (data.issues ?? []).filter((i) => matchesFind(i, find));

  async function saveView(name: string) {
    const filter = searchToFilter(search);
    const view = await api.createView({
      name,
      slug: slugify(name) || `view-${Date.now()}`,
      display: 'board',
      status: filter.status ?? null,
      project: filter.project ?? null,
      cycle: filter.cycle ?? null,
      labels: filter.labels ?? [],
      priority: filter.priority ?? null,
    });
    await router.invalidate();
    await navigate({ to: '/views/$slug', params: { slug: view.slug } });
  }

  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>Board</h1>
          <span className="muted">Drag to move and reorder</span>
        </div>
        <IssueFilters
          search={search}
          projects={data.projects}
          cycles={data.cycles}
          labels={data.labels}
          onChange={(next) => void navigate({ to: '/board', search: compactSearch(next) })}
          onSaveView={saveView}
          find={find}
          onFind={setFind}
        />
        {issues.length === 0 ? (
          <div className="empty">
            No issues. Press <span className="kbd">c</span> to create.
          </div>
        ) : (
          <IssueBoard
            issues={issues}
            onOpen={(id) =>
              void navigate({ to: '/issues/$identifier', params: { identifier: id } })
            }
            onMove={(id, status, sortOrder) => {
              void api.patchIssue(id, { status, sortOrder }).then(() => router.invalidate());
            }}
          />
        )}
      </section>
    </div>
  );
}
