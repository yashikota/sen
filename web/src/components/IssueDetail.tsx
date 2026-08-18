import { useEffect, useState } from 'react';
import { api } from '../api.ts';
import { ISSUE_STATUSES, PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type { Activity, Comment, Cycle, Issue, Project } from '../types.ts';

type Props = {
  identifier: string;
};

export function IssueDetail({ identifier }: Props) {
  const [issue, setIssue] = useState<Issue | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [draft, setDraft] = useState('');
  const [error, setError] = useState('');

  async function reload() {
    const [iss, com, act, proj, cyc] = await Promise.all([
      api.issue(identifier),
      api.comments(identifier),
      api.activities(identifier),
      api.projects(),
      api.cycles(),
    ]);
    setIssue(iss);
    setComments(com);
    setActivities(act);
    setProjects(proj);
    setCycles(cyc);
  }

  useEffect(() => {
    void reload().catch((e: unknown) => setError(e instanceof Error ? e.message : 'load failed'));
  }, [identifier]);

  async function patch(body: Record<string, unknown>) {
    const next = await api.patchIssue(identifier, body);
    setIssue(next);
    setActivities(await api.activities(identifier));
  }

  if (error) {
    return <div className="error">{error}</div>;
  }
  if (!issue) {
    return <div className="empty">Loading…</div>;
  }

  return (
    <div className="detail">
      <div className="ident">{issue.identifier}</div>
      <input
        className="title-input"
        value={issue.title}
        onChange={(e) => setIssue({ ...issue, title: e.target.value })}
        onBlur={() => void patch({ title: issue.title })}
      />
      <div className="props">
        <select value={issue.status} onChange={(e) => void patch({ status: e.target.value })}>
          {ISSUE_STATUSES.map((s) => (
            <option key={s} value={s}>
              {STATUS_LABEL[s]}
            </option>
          ))}
        </select>
        <select
          value={issue.priority}
          onChange={(e) => void patch({ priority: Number(e.target.value) })}
        >
          {PRIORITY_LABEL.map((label, i) => (
            <option key={label} value={i}>
              {label}
            </option>
          ))}
        </select>
        <select
          value={issue.projectId ?? ''}
          onChange={(e) =>
            void patch({
              projectId: e.target.value ? Number(e.target.value) : null,
            })
          }
        >
          <option value="">No project</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
        <select
          value={issue.cycleId ?? ''}
          onChange={(e) =>
            void patch({
              cycleId: e.target.value ? Number(e.target.value) : null,
            })
          }
        >
          <option value="">No cycle</option>
          {cycles.map((c) => (
            <option key={c.id} value={c.id}>
              Cycle {c.number}
            </option>
          ))}
        </select>
      </div>
      <textarea
        className="body-input"
        value={issue.body}
        placeholder="Write markdown…"
        onChange={(e) => setIssue({ ...issue, body: e.target.value })}
        onBlur={() => void patch({ body: issue.body })}
      />
      <div>
        <div className="muted">Comments</div>
        <div className="comments">
          {comments.map((c) => (
            <div className="comment" key={c.id}>
              <div className="muted">{c.createdAt}</div>
              <div>{c.body}</div>
            </div>
          ))}
          <textarea
            className="field"
            rows={3}
            placeholder="Comment"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                const body = draft.trim();
                if (!body) {
                  return;
                }
                void api.addComment(identifier, body).then(async () => {
                  setDraft('');
                  setComments(await api.comments(identifier));
                  setActivities(await api.activities(identifier));
                });
              }
            }}
          />
          <span className="muted">Mod+Enter to send</span>
        </div>
      </div>
      <div>
        <div className="muted">Activity</div>
        <div className="comments">
          {activities.map((a) => (
            <div className="comment" key={a.id}>
              <span className="ident">{a.action}</span> <span className="muted">{a.createdAt}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
