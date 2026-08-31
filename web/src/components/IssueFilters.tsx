import { useEffect, useRef, useState } from 'react';
import { ISSUE_STATUSES, PRIORITY_LABEL, STATUS_LABEL } from '../types.ts';
import type { Cycle, Label, Project } from '../types.ts';
import type { IssueSearch } from '../api.ts';

type Props = {
  search: IssueSearch;
  projects: Project[];
  cycles: Cycle[];
  labels: Label[];
  onChange: (next: IssueSearch) => void;
  onSaveView?: (name: string) => Promise<void>;
  find?: string;
  onFind?: (q: string) => void;
};

export function IssueFilters({
  search,
  projects,
  cycles,
  labels,
  onChange,
  onSaveView,
  find,
  onFind,
}: Props) {
  const [viewName, setViewName] = useState('');
  const findRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    function onFindKey() {
      findRef.current?.focus();
      findRef.current?.select();
    }
    window.addEventListener('sen:find', onFindKey);
    return () => window.removeEventListener('sen:find', onFindKey);
  }, []);
  const selectedLabels = (search.labels ?? '')
    .split(',')
    .map((n) => n.trim())
    .filter(Boolean);

  function set(patch: IssueSearch) {
    onChange({ ...search, ...patch });
  }

  return (
    <div className="toolbar" role="search" aria-label="Issue filters">
      <select
        aria-label="Filter status"
        value={search.status ?? ''}
        onChange={(e) => set({ status: e.target.value || undefined })}
      >
        <option value="">Any status</option>
        {ISSUE_STATUSES.map((s) => (
          <option key={s} value={s}>
            {STATUS_LABEL[s]}
          </option>
        ))}
      </select>
      <select
        aria-label="Filter project"
        value={search.project ?? ''}
        onChange={(e) => set({ project: e.target.value || undefined })}
      >
        <option value="">Any project</option>
        {projects.map((p) => (
          <option key={p.slug} value={p.slug}>
            {p.name}
          </option>
        ))}
      </select>
      <select
        aria-label="Filter cycle"
        value={search.cycle ?? ''}
        onChange={(e) => set({ cycle: e.target.value ? Number(e.target.value) : undefined })}
      >
        <option value="">Any cycle</option>
        {cycles.map((c) => (
          <option key={c.id} value={c.number}>
            Cycle {c.number}
          </option>
        ))}
      </select>
      <select
        aria-label="Filter priority"
        value={search.priority ?? ''}
        onChange={(e) =>
          set({ priority: e.target.value === '' ? undefined : Number(e.target.value) })
        }
      >
        <option value="">Any priority</option>
        {PRIORITY_LABEL.map((label, i) => (
          <option key={label} value={i}>
            {label}
          </option>
        ))}
      </select>
      {onFind ? (
        <input
          ref={findRef}
          aria-label="Find issues"
          placeholder="Find"
          value={find ?? ''}
          onChange={(e) => onFind(e.target.value)}
        />
      ) : null}
      <div className="chips toolbar-chips" role="group" aria-label="Filter labels">
        {labels.map((l) => {
          const on = selectedLabels.includes(l.name);
          return (
            <button
              type="button"
              key={l.id}
              className={`chip ${on ? 'on' : ''}`}
              aria-pressed={on}
              style={{ '--chip': l.color } as React.CSSProperties}
              onClick={() => {
                const next = on
                  ? selectedLabels.filter((n) => n !== l.name)
                  : [...selectedLabels, l.name];
                set({ labels: next.length ? next.join(',') : undefined });
              }}
            >
              {l.name}
            </button>
          );
        })}
      </div>
      {onSaveView ? (
        <form
          className="toolbar-save"
          onSubmit={(e) => {
            e.preventDefault();
            const name = viewName.trim();
            if (!name) {
              return;
            }
            void onSaveView(name).then(() => setViewName(''));
          }}
        >
          <input
            aria-label="New view name"
            placeholder="Save as view"
            value={viewName}
            onChange={(e) => setViewName(e.target.value)}
          />
        </form>
      ) : null}
    </div>
  );
}
