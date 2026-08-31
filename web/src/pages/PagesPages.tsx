import { Link, useLoaderData, useNavigate, useParams, useRouter } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { api } from '../api.ts';
import { MarkdownField } from '../components/MarkdownField.tsx';
import { PAGE_STATUSES } from '../types.ts';
import type { Page, Project } from '../types.ts';

function pageDepth(pages: Page[], page: Page): number {
  let depth = 0;
  let parentId = page.parentId;
  const byId = new Map(pages.map((p) => [p.id, p]));
  const seen = new Set<number>();
  while (parentId) {
    if (seen.has(parentId)) {
      break;
    }
    seen.add(parentId);
    const parent = byId.get(parentId);
    if (!parent) {
      break;
    }
    depth += 1;
    parentId = parent.parentId;
  }
  return depth;
}

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
          <div className="empty">
            No pages. Press <span className="kbd">p</span> to create an ADR.
          </div>
        ) : (
          <div className="list">
            {pages.map((p) => (
              <Link
                className="row"
                key={p.slug}
                to="/pages/$slug"
                params={{ slug: p.slug }}
                style={{ paddingLeft: 12 + pageDepth(pages, p) * 16 }}
              >
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
  const router = useRouter();
  const navigate = useNavigate();
  const [page, setPage] = useState(initial);
  const [pages, setPages] = useState<Page[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [tagDraft, setTagDraft] = useState(initial.tags.join(', '));

  useEffect(() => {
    setPage(initial);
    setTagDraft(initial.tags.join(', '));
  }, [initial]);

  useEffect(() => {
    void Promise.all([api.pages(), api.projects()]).then(([all, proj]) => {
      setPages(all);
      setProjects(proj);
    });
  }, [slug]);

  async function save(body: Record<string, unknown>) {
    const next = await api.patchPage(slug, body);
    setPage(next);
    await router.invalidate();
  }

  return (
    <div className="main single">
      <section className="pane">
        <div className="pane-head">
          <h1>{page.slug}</h1>
          <select
            className="field"
            aria-label="Page status"
            value={page.status}
            onChange={(e) => void save({ status: e.target.value })}
          >
            {PAGE_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="ghost danger"
            onClick={() => {
              if (!window.confirm(`Delete page ${page.slug}?`)) {
                return;
              }
              void api.deletePage(slug).then(async () => {
                await router.invalidate();
                await navigate({ to: '/pages' });
              });
            }}
          >
            Delete
          </button>
        </div>
        <div className="detail">
          <input
            className="title-input"
            aria-label="Page title"
            value={page.title}
            onChange={(e) => setPage({ ...page, title: e.target.value })}
            onBlur={() => void save({ title: page.title })}
          />
          <div className="props">
            <label>
              <span className="sr-only">Parent page</span>
              <select
                aria-label="Parent page"
                value={page.parentId ?? ''}
                onChange={(e) =>
                  void save({
                    parentId: e.target.value ? Number(e.target.value) : null,
                  })
                }
              >
                <option value="">No parent</option>
                {pages
                  .filter((p) => p.slug !== slug)
                  .map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.title}
                    </option>
                  ))}
              </select>
            </label>
            <label>
              <span className="sr-only">Project</span>
              <select
                aria-label="Page project"
                value={page.projectId ?? ''}
                onChange={(e) =>
                  void save({
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
            </label>
            <label>
              <span className="sr-only">Document date</span>
              <input
                type="date"
                aria-label="Document date"
                value={page.date?.slice(0, 10) ?? ''}
                onChange={(e) => void save({ date: e.target.value ? e.target.value : null })}
              />
            </label>
          </div>
          <input
            className="field"
            aria-label="Tags"
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
          <MarkdownField
            value={page.body}
            placeholder="Markdown body"
            onChange={(body) => setPage({ ...page, body })}
            onSave={(body) => void save({ body })}
          />
        </div>
      </section>
    </div>
  );
}
