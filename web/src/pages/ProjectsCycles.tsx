import { Link, useLoaderData, useNavigate, useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { api } from '../api.ts';
import { IssueList } from '../components/IssueList.tsx';
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
              placeholder="New project"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </form>
        </div>
        {projects.length === 0 ? (
          <div className="empty">No projects yet.</div>
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
  const [selected, setSelected] = useState<string | null>(null);
  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>{data.project.name}</h1>
          <span className="badge">{data.project.status}</span>
        </div>
        <div className="detail">
          <p className="muted">{data.project.description || 'No description'}</p>
          <IssueList issues={data.issues} selectedId={selected} onSelect={setSelected} />
        </div>
      </section>
      <section className="pane">
        <div className="empty">{slug}</div>
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
          <div className="empty">No cycles yet.</div>
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
  const [selected, setSelected] = useState<string | null>(null);
  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Cycle {data.cycle.number}</h1>
          <span className="badge">{data.cycle.status}</span>
        </div>
        <IssueList issues={data.issues} selectedId={selected} onSelect={setSelected} />
      </section>
      <section className="pane">
        <div className="detail">
          <div className="muted">
            {data.cycle.startsAt} — {data.cycle.endsAt}
          </div>
        </div>
      </section>
    </div>
  );
}
