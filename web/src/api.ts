import type {
  Activity,
  Comment,
  Cycle,
  Diagnostic,
  Issue,
  Label,
  Page,
  Project,
  SearchHit,
  View,
  Workspace,
} from './types.ts';

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const res = await fetch(path, {
    ...init,
    headers,
  });
  if (res.status === 204) {
    return undefined as T;
  }
  const text = await res.text();
  const data = text ? (JSON.parse(text) as unknown) : null;
  if (!res.ok) {
    const err = data as { error?: string } | null;
    throw new Error(err?.error ?? res.statusText);
  }
  return data as T;
}

export const api = {
  workspace: () => req<Workspace>('/api/workspace'),
  patchWorkspace: (body: Partial<Workspace>) =>
    req<Workspace>('/api/workspace', {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  labels: () => req<Label[]>('/api/labels'),
  createLabel: (body: { name: string; color: string }) =>
    req<Label>('/api/labels', { method: 'POST', body: JSON.stringify(body) }),
  diagnostics: () => req<Diagnostic[]>('/api/diagnostics'),
  issues: (q = '') => req<Issue[]>(`/api/issues${q}`),
  issue: (id: string) => req<Issue>(`/api/issues/${id}`),
  createIssue: (body: {
    title: string;
    status?: string;
    priority?: number;
    parentId?: number;
    projectId?: number;
    cycleId?: number;
    labelIds?: number[];
  }) => req<Issue>('/api/issues', { method: 'POST', body: JSON.stringify(body) }),
  patchIssue: (id: string, body: Record<string, unknown>) =>
    req<Issue>(`/api/issues/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  deleteIssue: (id: string) => req<void>(`/api/issues/${id}`, { method: 'DELETE' }),
  comments: (id: string) => req<Comment[]>(`/api/issues/${id}/comments`),
  addComment: (id: string, body: string) =>
    req<Comment>(`/api/issues/${id}/comments`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),
  activities: (id: string) => req<Activity[]>(`/api/issues/${id}/activities`),
  projects: () => req<Project[]>('/api/projects'),
  project: (slug: string) => req<Project>(`/api/projects/${slug}`),
  createProject: (body: { name: string; slug: string; description?: string }) =>
    req<Project>('/api/projects', { method: 'POST', body: JSON.stringify(body) }),
  patchProject: (slug: string, body: Record<string, unknown>) =>
    req<Project>(`/api/projects/${slug}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  deleteProject: (slug: string) => req<void>(`/api/projects/${slug}`, { method: 'DELETE' }),
  cycles: () => req<Cycle[]>('/api/cycles'),
  cycle: (n: number) => req<Cycle>(`/api/cycles/${n}`),
  createCycle: (body: { startsAt: string; endsAt: string; status?: string }) =>
    req<Cycle>('/api/cycles', { method: 'POST', body: JSON.stringify(body) }),
  patchCycle: (n: number, body: Record<string, unknown>) =>
    req<Cycle>(`/api/cycles/${n}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  pages: () => req<Page[]>('/api/pages'),
  page: (slug: string) => req<Page>(`/api/pages/${slug}`),
  createPage: (body: {
    title: string;
    slug: string;
    body?: string;
    status?: string;
    tags?: string[];
  }) => req<Page>('/api/pages', { method: 'POST', body: JSON.stringify(body) }),
  patchPage: (slug: string, body: Record<string, unknown>) =>
    req<Page>(`/api/pages/${slug}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  deletePage: (slug: string) => req<void>(`/api/pages/${slug}`, { method: 'DELETE' }),
  views: () => req<View[]>('/api/views'),
  view: (slug: string) => req<View>(`/api/views/${slug}`),
  createView: (body: {
    name: string;
    slug: string;
    display?: string;
    status?: string | null;
    project?: string | null;
    cycle?: number | null;
    labels?: string[];
    priority?: number | null;
  }) => req<View>('/api/views', { method: 'POST', body: JSON.stringify(body) }),
  patchView: (slug: string, body: Record<string, unknown>) =>
    req<View>(`/api/views/${slug}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  deleteView: (slug: string) => req<void>(`/api/views/${slug}`, { method: 'DELETE' }),
  search: (q: string) => req<SearchHit[]>(`/api/search?q=${encodeURIComponent(q)}`),
};

export function issuesQuery(filter: {
  status?: string | null;
  project?: string | null;
  cycle?: number | null;
  labels?: string[] | null;
  priority?: number | null;
}): string {
  const q = new URLSearchParams();
  if (filter.status) {
    q.set('status', filter.status);
  }
  if (filter.project) {
    q.set('project', filter.project);
  }
  if (filter.cycle) {
    q.set('cycle', String(filter.cycle));
  }
  if (filter.labels?.length) {
    q.set('labels', filter.labels.join(','));
  }
  if (filter.priority != null && filter.priority >= 0) {
    q.set('priority', String(filter.priority));
  }
  const s = q.toString();
  return s ? `?${s}` : '';
}

export type IssueSearch = {
  status?: string;
  project?: string;
  cycle?: number;
  priority?: number;
  labels?: string;
};

export function parseIssueSearch(raw: Record<string, unknown>): IssueSearch {
  const out: IssueSearch = {};
  if (typeof raw.status === 'string' && raw.status) {
    out.status = raw.status;
  }
  if (typeof raw.project === 'string' && raw.project) {
    out.project = raw.project;
  }
  if (raw.cycle !== undefined && raw.cycle !== '') {
    const n = Number(raw.cycle);
    if (Number.isFinite(n) && n > 0) {
      out.cycle = n;
    }
  }
  if (raw.priority !== undefined && raw.priority !== '') {
    const n = Number(raw.priority);
    if (Number.isFinite(n) && n >= 0) {
      out.priority = n;
    }
  }
  if (typeof raw.labels === 'string' && raw.labels) {
    out.labels = raw.labels;
  }
  return out;
}

export function searchToFilter(search: IssueSearch): {
  status?: string;
  project?: string;
  cycle?: number;
  labels?: string[];
  priority?: number;
} {
  return {
    status: search.status,
    project: search.project,
    cycle: search.cycle,
    labels: search.labels
      ? search.labels
          .split(',')
          .map((n) => n.trim())
          .filter(Boolean)
      : undefined,
    priority: search.priority,
  };
}
