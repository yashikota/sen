import { useNavigate } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { sortOrderForDrop } from '../board.ts';
import { isOverdue, localToday } from '../due.ts';
import { actionFromKeyboard } from '../keymap.ts';
import { ISSUE_STATUSES, PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type { Issue, IssueStatus } from '../types.ts';

type Props = {
  issues: Issue[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  openOnSelect?: boolean;
};

export function IssueList({ issues, selectedId, onSelect, openOnSelect = true }: Props) {
  const navigate = useNavigate();
  const ids = issues.map((i) => i.identifier);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const action = actionFromKeyboard(e);
      if (action !== 'move-down' && action !== 'move-up' && action !== 'open') {
        return;
      }
      if (ids.length === 0) {
        return;
      }
      e.preventDefault();
      const idx = Math.max(0, ids.indexOf(selectedId ?? ''));
      if (action === 'move-down') {
        const id = ids[Math.min(idx + 1, ids.length - 1)] ?? ids[0];
        if (id) {
          onSelect(id);
        }
      }
      if (action === 'move-up') {
        const id = ids[Math.max(idx - 1, 0)] ?? ids[0];
        if (id) {
          onSelect(id);
        }
      }
      if (action === 'open') {
        const id = selectedId ?? ids[0];
        if (id) {
          if (openOnSelect) {
            void navigate({ to: '/issues/$identifier', params: { identifier: id } });
          } else {
            onSelect(id);
          }
        }
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [ids, navigate, onSelect, openOnSelect, selectedId]);

  useEffect(() => {
    if (selectedId) {
      window.dispatchEvent(new CustomEvent('sen:issue', { detail: selectedId }));
    }
  }, [selectedId]);

  if (issues.length === 0) {
    return (
      <div className="empty">
        No issues. Press <span className="kbd">c</span> to create.
      </div>
    );
  }

  const today = localToday();

  return (
    <div className="list" role="listbox" aria-label="Issues">
      {issues.map((issue) => {
        const childCount = issues.filter((i) => i.parentId === issue.id).length;
        const overdue = isOverdue(issue.dueDate, today);
        return (
          <button
            type="button"
            key={issue.identifier}
            role="option"
            aria-selected={issue.identifier === selectedId}
            data-status={issue.status}
            className={`row ${issue.identifier === selectedId ? 'selected' : ''} ${overdue ? 'overdue' : ''}`}
            style={{ paddingLeft: 12 + (issue.depth ?? 0) * 16 }}
            onClick={() => {
              onSelect(issue.identifier);
              if (openOnSelect) {
                void navigate({
                  to: '/issues/$identifier',
                  params: { identifier: issue.identifier },
                });
              }
            }}
          >
            <span className="rail" />
            <span className="ident">{issue.identifier}</span>
            <span className="row-title">{issue.title}</span>
            <span className="row-meta">
              {issue.projectSlug ? <span className="badge">{issue.projectSlug}</span> : null}
              {issue.cycleNumber ? <span className="badge">C{issue.cycleNumber}</span> : null}
              {issue.dueDate ? (
                <span className={`badge ${overdue ? 'urgent' : ''}`}>
                  {issue.dueDate.slice(0, 10)}
                </span>
              ) : null}
              {childCount > 0 ? <span className="badge">{childCount}</span> : null}
              {issue.labels.slice(0, 3).map((l) => (
                <span key={l.id} className="pip" style={{ background: l.color }} title={l.name} />
              ))}
              {issue.priority > 0 ? (
                <span className={`badge ${issue.priority === 1 ? 'urgent' : ''}`}>
                  {PRIORITY_LABEL[issue.priority]}
                </span>
              ) : null}
              <span className="badge">{STATUS_LABEL[issue.status]}</span>
            </span>
          </button>
        );
      })}
    </div>
  );
}

type BoardProps = {
  issues: Issue[];
  onOpen: (id: string) => void;
  onMove: (id: string, status: IssueStatus, sortOrder: number) => void;
};

function columnIssues(issues: Issue[], status: IssueStatus): Issue[] {
  return issues
    .filter((i) => i.status === status)
    .slice()
    .sort((a, b) => a.sortOrder - b.sortOrder || a.number - b.number);
}

export function IssueBoard({ issues, onOpen, onMove }: BoardProps) {
  const [dragId, setDragId] = useState<string | null>(null);

  function drop(status: IssueStatus, beforeId: string | null) {
    if (!dragId) {
      return;
    }
    const col = columnIssues(issues, status);
    onMove(dragId, status, sortOrderForDrop(col, dragId, beforeId));
    setDragId(null);
  }

  return (
    <div className="board">
      {ISSUE_STATUSES.map((status) => (
        <div
          className={`column ${dragId ? 'droppable' : ''}`}
          key={status}
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => {
            e.preventDefault();
            drop(status, null);
          }}
        >
          <h2>
            {STATUS_LABEL[status]}
            <span className="muted">{columnIssues(issues, status).length}</span>
          </h2>
          {columnIssues(issues, status).map((issue) => (
            <button
              type="button"
              className={`card ${isOverdue(issue.dueDate, localToday()) ? 'overdue' : ''}`}
              key={issue.identifier}
              draggable
              onDragStart={() => setDragId(issue.identifier)}
              onDragEnd={() => setDragId(null)}
              onDragOver={(e) => {
                e.preventDefault();
                e.stopPropagation();
              }}
              onDrop={(e) => {
                e.preventDefault();
                e.stopPropagation();
                drop(status, issue.identifier);
              }}
              onClick={() => onOpen(issue.identifier)}
            >
              <div className="card-top">
                <span className="ident">{issue.identifier}</span>
                {issue.priority > 0 ? (
                  <span className={`badge ${issue.priority === 1 ? 'urgent' : ''}`}>
                    {PRIORITY_LABEL[issue.priority]}
                  </span>
                ) : null}
              </div>
              <div className="card-title">{issue.title}</div>
              <div className="row-meta">
                {issue.projectSlug ? <span className="badge">{issue.projectSlug}</span> : null}
                {issue.dueDate ? (
                  <span
                    className={`badge ${isOverdue(issue.dueDate, localToday()) ? 'urgent' : ''}`}
                  >
                    {issue.dueDate.slice(0, 10)}
                  </span>
                ) : null}
                {issue.labels.slice(0, 4).map((l) => (
                  <span key={l.id} className="pip" style={{ background: l.color }} title={l.name} />
                ))}
              </div>
            </button>
          ))}
        </div>
      ))}
    </div>
  );
}
