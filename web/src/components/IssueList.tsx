import { useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';
import { actionFromKeyboard } from '../keymap.ts';
import { PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type { Issue } from '../types.ts';

type Props = {
  issues: Issue[];
  selectedId: string | null;
  onSelect: (id: string) => void;
};

export function IssueList({ issues, selectedId, onSelect }: Props) {
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
        onSelect(ids[Math.min(idx + 1, ids.length - 1)] ?? ids[0]);
      }
      if (action === 'move-up') {
        onSelect(ids[Math.max(idx - 1, 0)] ?? ids[0]);
      }
      if (action === 'open') {
        const id = selectedId ?? ids[0];
        if (id) {
          void navigate({ to: '/issues/$identifier', params: { identifier: id } });
        }
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [ids, navigate, onSelect, selectedId]);

  if (issues.length === 0) {
    return (
      <div className="empty">
        No issues. Press <span className="kbd">c</span> to create.
      </div>
    );
  }

  return (
    <div className="list">
      {issues.map((issue) => (
        <div
          key={issue.identifier}
          className={`row ${issue.identifier === selectedId ? 'selected' : ''}`}
          onClick={() => {
            onSelect(issue.identifier);
            void navigate({
              to: '/issues/$identifier',
              params: { identifier: issue.identifier },
            });
          }}
        >
          <span className="rail" />
          <span className="ident">{issue.identifier}</span>
          <span>{issue.title}</span>
          <span className={`badge ${issue.priority === 1 ? 'urgent' : ''}`}>
            {STATUS_LABEL[issue.status]}
            {issue.priority > 0 ? ` · ${PRIORITY_LABEL[issue.priority]}` : ''}
          </span>
        </div>
      ))}
    </div>
  );
}
