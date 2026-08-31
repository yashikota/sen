import { useLoaderData, useNavigate, useParams, useRouter } from '@tanstack/react-router';
import { useState } from 'react';
import { api } from '../api.ts';
import { IssueDetail } from '../components/IssueDetail.tsx';
import { IssueBoard, IssueList } from '../components/IssueList.tsx';
import { ISSUE_STATUSES, PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type { Cycle, Issue, Label, Project, View } from '../types.ts';

export function ViewPage() {
  const { slug } = useParams({ from: '/views/$slug' });
  const data = useLoaderData({ from: '/views/$slug' }) as {
    view: View;
    issues: Issue[];
    projects: Project[];
    cycles: Cycle[];
    labels: Label[];
  };
  const router = useRouter();
  const navigate = useNavigate();
  const issues = data.issues ?? [];
  const [view, setView] = useState(data.view);
  const [selected, setSelected] = useState<string | null>(issues[0]?.identifier ?? null);

  if (view.slug !== data.view.slug || view.updatedAt !== data.view.updatedAt) {
    setView(data.view);
    setSelected((data.issues ?? [])[0]?.identifier ?? null);
  }

  async function save(body: Record<string, unknown>) {
    const next = await api.patchView(slug, body);
    setView(next);
    await router.invalidate();
  }

  return (
    <div className={view.display === 'board' ? 'main single' : 'main'}>
      <section className="pane">
        <div className="pane-head">
          <h1>{view.name}</h1>
          <span className="muted">{issues.length}</span>
          <button
            type="button"
            className="ghost danger"
            onClick={() => {
              if (!window.confirm(`Delete view ${view.name}?`)) {
                return;
              }
              void api.deleteView(slug).then(async () => {
                await router.invalidate();
                await navigate({ to: '/issues', search: {} });
              });
            }}
          >
            Delete
          </button>
        </div>
        <div className="props view-filters">
          <label>
            <span className="sr-only">View name</span>
            <input
              aria-label="View name"
              value={view.name}
              onChange={(e) => setView({ ...view, name: e.target.value })}
              onBlur={() => void save({ name: view.name })}
            />
          </label>
          <label>
            <span className="sr-only">View display</span>
            <select
              aria-label="View display"
              value={view.display}
              onChange={(e) => void save({ display: e.target.value })}
            >
              <option value="list">List</option>
              <option value="board">Board</option>
            </select>
          </label>
          <label>
            <span className="sr-only">View status</span>
            <select
              aria-label="View status"
              value={view.status ?? ''}
              onChange={(e) => void save({ status: e.target.value })}
            >
              <option value="">Any status</option>
              {ISSUE_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {STATUS_LABEL[s]}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span className="sr-only">View project</span>
            <select
              aria-label="View project"
              value={view.project ?? ''}
              onChange={(e) => void save({ project: e.target.value })}
            >
              <option value="">Any project</option>
              {data.projects.map((p) => (
                <option key={p.slug} value={p.slug}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span className="sr-only">View cycle</span>
            <select
              aria-label="View cycle"
              value={view.cycle ?? ''}
              onChange={(e) => void save({ cycle: e.target.value ? Number(e.target.value) : 0 })}
            >
              <option value="">Any cycle</option>
              {data.cycles.map((c) => (
                <option key={c.id} value={c.number}>
                  Cycle {c.number}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span className="sr-only">View priority</span>
            <select
              aria-label="View priority"
              value={view.priority ?? ''}
              onChange={(e) =>
                void save({
                  priority: e.target.value === '' ? -1 : Number(e.target.value),
                })
              }
            >
              <option value="">Any priority</option>
              {PRIORITY_LABEL.map((label, i) => (
                <option key={label} value={i}>
                  {label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="chips" role="group" aria-label="View labels">
          {data.labels.map((l) => {
            const on = view.labels.includes(l.name);
            return (
              <button
                type="button"
                key={l.id}
                className={`chip ${on ? 'on' : ''}`}
                aria-pressed={on}
                style={{ '--chip': l.color } as React.CSSProperties}
                onClick={() => {
                  const next = on
                    ? view.labels.filter((n) => n !== l.name)
                    : [...view.labels, l.name];
                  void save({ labels: next });
                }}
              >
                {l.name}
              </button>
            );
          })}
        </div>
        {view.display === 'board' ? (
          <IssueBoard
            issues={issues}
            onOpen={(id) =>
              void navigate({ to: '/issues/$identifier', params: { identifier: id } })
            }
            onMove={(id, status, sortOrder) => {
              void api.patchIssue(id, { status, sortOrder }).then(() => router.invalidate());
            }}
          />
        ) : (
          <IssueList issues={issues} selectedId={selected} onSelect={setSelected} />
        )}
      </section>
      {view.display === 'list' ? (
        <section className="pane">
          {selected ? (
            <IssueDetail identifier={selected} />
          ) : (
            <div className="empty">Select an issue</div>
          )}
        </section>
      ) : null}
    </div>
  );
}
