import { Link, useLoaderData, useNavigate, useParams, useRouter } from '@tanstack/react-router';
import { useState } from 'react';
import { api } from '../api.ts';
import { IssueDetail } from '../components/IssueDetail.tsx';
import { IssueList } from '../components/IssueList.tsx';
import { CYCLE_STATUSES, PROJECT_STATUSES } from '../types.ts';
import type { Cycle, Issue, Project } from '../types.ts';

export function ProjectsPage() {
  const projects = useLoaderData({ from: '/projects' }) as Project[];
  const [name, setName] = useState('');
  const navigate = useNavigate();
  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>Projects</h1>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const n = name.trim();
              if (!n) {
                return;
              }
              const slug = n
                .toLowerCase()
                .replace(/[^a-z0-9]+/g, '-')
                .replace(/^-|-$/g, '');
              void api
                .createProject({ name: n, slug })
                .then((p) => navigate({ to: '/projects/$slug', params: { slug: p.slug } }));
            }}
          >
            <input
              className="field"
              aria-label="New project name"
              placeholder="New project"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </form>
        </div>
        {projects.length === 0 ? (
          <div className="empty">No projects yet. Name one above.</div>
        ) : (
          <div className="list">
            {projects.map((p) => (
              <Link className="row" key={p.slug} to="/projects/$slug" params={{ slug: p.slug }}>
                <span className="rail" />
                <span className="ident">{p.status}</span>
                <span>{p.name}</span>
                <span className="progress" style={{ width: 72 }}>
                  <span style={{ width: `${Math.round(p.progress * 100)}%` }} />
                </span>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export function ProjectDetailPage() {
  const { slug } = useParams({ from: '/projects/$slug' });
  const data = useLoaderData({ from: '/projects/$slug' }) as {
    project: Project;
    issues: Issue[];
  };
  const router = useRouter();
  const navigate = useNavigate();
  const [selected, setSelected] = useState<string | null>(data.issues[0]?.identifier ?? null);
  const [project, setProject] = useState(data.project);

  if (project.slug !== data.project.slug) {
    setProject(data.project);
    setSelected(data.issues[0]?.identifier ?? null);
  }

  async function save(body: Record<string, unknown>) {
    const next = await api.patchProject(slug, body);
    setProject(next);
    await router.invalidate();
  }

  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>{project.name}</h1>
          <select
            className="field"
            aria-label="Project status"
            value={project.status}
            onChange={(e) => void save({ status: e.target.value })}
          >
            {PROJECT_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="ghost"
            onClick={() =>
              window.dispatchEvent(
                new CustomEvent('sen:create-issue', { detail: { projectId: project.id } }),
              )
            }
          >
            New issue
          </button>
          <button
            type="button"
            className="ghost danger"
            onClick={() => {
              if (!window.confirm(`Delete project ${project.name}?`)) {
                return;
              }
              void api.deleteProject(slug).then(async () => {
                await router.invalidate();
                await navigate({ to: '/projects' });
              });
            }}
          >
            Delete
          </button>
        </div>
        <div className="detail">
          <textarea
            className="field"
            aria-label="Project description"
            placeholder="Description"
            value={project.description}
            onChange={(e) => setProject({ ...project, description: e.target.value })}
            onBlur={() => void save({ description: project.description })}
          />
          <div className="props">
            <label>
              <span className="muted">Start</span>
              <input
                type="date"
                aria-label="Start date"
                value={project.startDate?.slice(0, 10) ?? ''}
                onChange={(e) =>
                  void save(
                    e.target.value ? { startDate: e.target.value } : { clearStartDate: true },
                  )
                }
              />
            </label>
            <label>
              <span className="muted">Target</span>
              <input
                type="date"
                aria-label="Target date"
                value={project.targetDate?.slice(0, 10) ?? ''}
                onChange={(e) =>
                  void save(
                    e.target.value ? { targetDate: e.target.value } : { clearTargetDate: true },
                  )
                }
              />
            </label>
          </div>
          <IssueList
            issues={data.issues}
            selectedId={selected}
            onSelect={setSelected}
            openOnSelect={false}
          />
        </div>
      </section>
      <section className="pane">
        {selected ? (
          <IssueDetail identifier={selected} />
        ) : (
          <div className="empty">Select an issue</div>
        )}
      </section>
    </div>
  );
}

export function CyclesPage() {
  const cycles = useLoaderData({ from: '/cycles' }) as Cycle[];
  const navigate = useNavigate();
  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>Cycles</h1>
          <button
            className="ghost"
            type="button"
            onClick={() => {
              const start = new Date();
              const end = new Date(start.getTime() + 14 * 86400000);
              void api
                .createCycle({
                  startsAt: start.toISOString(),
                  endsAt: end.toISOString(),
                })
                .then((c) =>
                  navigate({
                    to: '/cycles/$number',
                    params: { number: String(c.number) },
                  }),
                );
            }}
          >
            New cycle
          </button>
        </div>
        {cycles.length === 0 ? (
          <div className="empty">No cycles yet. Start one to timebox work.</div>
        ) : (
          <div className="list">
            {cycles.map((c) => (
              <Link
                className="row"
                key={c.number}
                to="/cycles/$number"
                params={{ number: String(c.number) }}
              >
                <span className="rail" />
                <span className="ident">{c.number}</span>
                <span>
                  {c.status} · {c.startsAt.slice(0, 10)} → {c.endsAt.slice(0, 10)}
                </span>
                <span className="badge">{c.status}</span>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export function CycleDetailPage() {
  const data = useLoaderData({ from: '/cycles/$number' }) as {
    cycle: Cycle;
    issues: Issue[];
  };
  const router = useRouter();
  const [selected, setSelected] = useState<string | null>(data.issues[0]?.identifier ?? null);
  const [cycle, setCycle] = useState(data.cycle);
  const done = data.issues.filter((i) => i.status === 'done' || i.status === 'canceled').length;

  if (cycle.number !== data.cycle.number) {
    setCycle(data.cycle);
    setSelected(data.issues[0]?.identifier ?? null);
  }

  async function save(body: Record<string, unknown>) {
    const next = await api.patchCycle(cycle.number, body);
    setCycle(next);
    await router.invalidate();
  }

  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Cycle {cycle.number}</h1>
          <span className="muted">
            {done}/{data.issues.length}
          </span>
          <select
            className="field"
            aria-label="Cycle status"
            value={cycle.status}
            onChange={(e) => void save({ status: e.target.value })}
          >
            {CYCLE_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="ghost"
            onClick={() =>
              window.dispatchEvent(
                new CustomEvent('sen:create-issue', { detail: { cycleId: cycle.id } }),
              )
            }
          >
            New issue
          </button>
        </div>
        <div className="detail">
          <div className="muted">
            {cycle.startsAt.slice(0, 10)} — {cycle.endsAt.slice(0, 10)}
          </div>
          <IssueList
            issues={data.issues}
            selectedId={selected}
            onSelect={setSelected}
            openOnSelect={false}
          />
        </div>
      </section>
      <section className="pane">
        {selected ? (
          <IssueDetail identifier={selected} />
        ) : (
          <div className="empty">Select an issue</div>
        )}
      </section>
    </div>
  );
}
