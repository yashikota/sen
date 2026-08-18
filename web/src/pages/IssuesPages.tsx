import { useLoaderData, useNavigate, useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { IssueDetail } from '../components/IssueDetail.tsx';
import { IssueList } from '../components/IssueList.tsx';
import type { Issue } from '../types.ts';
import { ISSUE_STATUSES, STATUS_LABEL } from '../types.ts';

export function IssuesPage() {
  const issues = useLoaderData({ from: '/issues' }) as Issue[];
  const [selected, setSelected] = useState<string | null>(issues[0]?.identifier ?? null);
  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Issues</h1>
          <span className="muted">{issues.length}</span>
        </div>
        <IssueList issues={issues} selectedId={selected} onSelect={setSelected} />
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

export function IssueRoutePage() {
  const { identifier } = useParams({ from: '/issues/$identifier' });
  const issues = useLoaderData({ from: '/issues/$identifier' }) as Issue[];
  return (
    <div className="main">
      <section className="pane">
        <div className="pane-head">
          <h1>Issues</h1>
        </div>
        <IssueList issues={issues} selectedId={identifier} onSelect={() => undefined} />
      </section>
      <section className="pane">
        <IssueDetail identifier={identifier} />
      </section>
    </div>
  );
}

export function BoardPage() {
  const issues = useLoaderData({ from: '/board' }) as Issue[];
  const navigate = useNavigate();
  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>Board</h1>
        </div>
        <div className="board">
          {ISSUE_STATUSES.map((status) => (
            <div className="column" key={status}>
              <h2>{STATUS_LABEL[status]}</h2>
              {issues
                .filter((i) => i.status === status)
                .map((issue) => (
                  <button
                    type="button"
                    className="card"
                    key={issue.identifier}
                    onClick={() =>
                      void navigate({
                        to: '/issues/$identifier',
                        params: { identifier: issue.identifier },
                      })
                    }
                  >
                    <div className="ident">{issue.identifier}</div>
                    <div>{issue.title}</div>
                  </button>
                ))}
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
