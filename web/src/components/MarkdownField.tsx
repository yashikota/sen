import { useState } from 'react';
import { renderMarkdown } from '../markdown.ts';

type Props = {
  value: string;
  placeholder?: string;
  onChange: (value: string) => void;
  onSave: (value: string) => void;
};

export function MarkdownField({ value, placeholder, onChange, onSave }: Props) {
  const [mode, setMode] = useState<'edit' | 'preview'>('edit');
  return (
    <div className="md-field">
      <div className="seg" role="tablist" aria-label="Body">
        <button
          type="button"
          role="tab"
          aria-selected={mode === 'edit'}
          className={mode === 'edit' ? 'on' : ''}
          onClick={() => setMode('edit')}
        >
          Edit
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={mode === 'preview'}
          className={mode === 'preview' ? 'on' : ''}
          onClick={() => setMode('preview')}
        >
          Preview
        </button>
      </div>
      {mode === 'edit' ? (
        <textarea
          className="body-input"
          aria-label="Markdown body"
          value={value}
          placeholder={placeholder ?? 'Write markdown…'}
          onChange={(e) => onChange(e.target.value)}
          onBlur={() => onSave(value)}
        />
      ) : (
        <div
          className="md"
          // HTML is escaped by renderMarkdown before tags are added.
          dangerouslySetInnerHTML={{
            __html: value.trim() ? renderMarkdown(value) : '<p class="muted">Empty</p>',
          }}
        />
      )}
    </div>
  );
}
