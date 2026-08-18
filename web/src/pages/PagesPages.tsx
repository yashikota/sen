import { Link, useLoaderData, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { api } from '../api.ts';
import type { Page } from '../types.ts';

export function PagesPage() {
  const pages = useLoaderData({ from: '/pages' }) as Page[];
  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>Pages</h1>
          <span className="muted">Press p</span>
        </div>
        {pages.length === 0 ? (
          <div className="empty">No pages. Press p to create an ADR.</div>
        ) : (
          <div className="list">
            {pages.map((p) => (
              <Link className="row" key={p.slug} to="/pages/$slug" params={{ slug: p.slug }}>
                <span className="rail" />
                <span className="ident">{p.status}</span>
                <span>{p.title}</span>
                <span className="badge">{p.slug}</span>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export function PageDetailPage() {
  const { slug } = useParams({ from: '/pages/$slug' });
  const initial = useLoaderData({ from: '/pages/$slug' }) as Page;
  const [page, setPage] = useState(initial);
  const [tagDraft, setTagDraft] = useState(initial.tags.join(', '));

  useEffect(() => {
    setPage(initial);
    setTagDraft(initial.tags.join(', '));
  }, [initial]);

  async function save(body: Record<string, unknown>) {
    const next = await api.patchPage(slug, body);
    setPage(next);
  }

  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>{page.slug}</h1>
          <select
            className="field"
            value={page.status}
            onChange={(e) => void save({ status: e.target.value })}
          >
            <option value="proposed">proposed</option>
            <option value="accepted">accepted</option>
            <option value="deprecated">deprecated</option>
            <option value="superseded">superseded</option>
          </select>
        </div>
        <div className="detail">
          <input
            className="title-input"
            value={page.title}
            onChange={(e) => setPage({ ...page, title: e.target.value })}
            onBlur={() => void save({ title: page.title })}
          />
          <input
            className="field"
            placeholder="tags, comma separated"
            value={tagDraft}
            onChange={(e) => setTagDraft(e.target.value)}
            onBlur={() =>
              void save({
                tags: tagDraft
                  .split(',')
                  .map((t) => t.trim())
                  .filter(Boolean),
              })
            }
          />
          <textarea
            className="body-input"
            value={page.body}
            placeholder="Markdown body (frontmatter is added on push)"
            onChange={(e) => setPage({ ...page, body: e.target.value })}
            onBlur={() => void save({ body: page.body })}
          />
        </div>
      </section>
    </div>
  );
}
