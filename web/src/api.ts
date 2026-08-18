import type {
  Activity,
  Comment,
  Cycle,
  Issue,
  Label,
  Page,
  Project,
  SearchHit,
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
  issues: (q = '') => req<Issue[]>(`/api/issues${q}`),
  issue: (id: string) => req<Issue>(`/api/issues/${id}`),
  createIssue: (body: { title: string; status?: string; priority?: number }) =>
    req<Issue>('/api/issues', { method: 'POST', body: JSON.stringify(body) }),
  patchIssue: (id: string, body: Record<string, unknown>) =>
    req<Issue>(`/api/issues/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
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
  search: (q: string) => req<SearchHit[]>(`/api/search?q=${encodeURIComponent(q)}`),
};
