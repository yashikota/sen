const ROWS: { keys: string; action: string }[] = [
  { keys: 'Mod+K', action: 'Command palette' },
  { keys: 'c', action: 'Create issue' },
  { keys: 'p', action: 'Create page' },
  { keys: '/', action: 'Find in the current list' },
  { keys: 'j / k', action: 'Move selection' },
  { keys: 'Enter', action: 'Open selected issue' },
  { keys: 's', action: 'Set status' },
  { keys: '1–4', action: 'Set priority' },
  { keys: 'Esc', action: 'Close dialogs' },
  { keys: '?', action: 'This help' },
];

type Props = {
  onClose: () => void;
};

export function ShortcutHelp({ onClose }: Props) {
  return (
    <div className="overlay" onClick={onClose}>
      <div
        className="dialog help-dialog"
        role="dialog"
        aria-modal="true"
        aria-label="Keyboard shortcuts"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="help-head">
          <h2>Keyboard</h2>
          <button type="button" className="ghost" onClick={onClose}>
            Close
          </button>
        </div>
        <dl className="help-list">
          {ROWS.map((row) => (
            <div key={row.keys} className="help-row">
              <dt>
                <span className="kbd">{row.keys}</span>
              </dt>
              <dd>{row.action}</dd>
            </div>
          ))}
        </dl>
      </div>
    </div>
  );
}
