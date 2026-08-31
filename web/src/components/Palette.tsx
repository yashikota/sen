import { useEffect, useState } from 'react';
import type { Command } from '../commands.ts';

type Props = {
  query: string;
  onQuery: (q: string) => void;
  commands: Command[];
  onPick: (id: string) => void;
  onClose: () => void;
};

export function Palette({ query, onQuery, commands, onPick, onClose }: Props) {
  const [active, setActive] = useState(0);

  useEffect(() => {
    setActive(0);
  }, [query, commands.length]);

  return (
    <div className="overlay" onClick={onClose}>
      <div
        className="palette"
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        onClick={(e) => e.stopPropagation()}
      >
        <input
          autoFocus
          aria-label="Command search"
          placeholder="Type a command or search…"
          value={query}
          onChange={(e) => onQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'ArrowDown') {
              e.preventDefault();
              setActive((n) => Math.min(n + 1, Math.max(commands.length - 1, 0)));
            }
            if (e.key === 'ArrowUp') {
              e.preventDefault();
              setActive((n) => Math.max(n - 1, 0));
            }
            if (e.key === 'Enter') {
              e.preventDefault();
              const id = commands[active]?.id;
              if (id) {
                onPick(id);
              }
            }
            if (e.key === 'Escape') {
              e.preventDefault();
              onClose();
            }
          }}
        />
        <div className="palette-list" role="listbox">
          {commands.map((c, i) => (
            <button
              type="button"
              key={c.id + i}
              role="option"
              aria-selected={i === active}
              className={`palette-item ${i === active ? 'active' : ''}`}
              onMouseEnter={() => setActive(i)}
              onClick={() => onPick(c.id)}
            >
              <span>{c.title}</span>
              {c.hint ? <span className="kbd">{c.hint}</span> : null}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
