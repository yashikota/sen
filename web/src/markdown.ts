const SAFE_HREF = /^(https?:|mailto:|\/|#)/i;

export function escapeHtml(s: string): string {
  return s
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

const FENCE = '@@SENFENCE:';

export function renderMarkdown(src: string): string {
  const fences: string[] = [];
  let text = src.replaceAll('\r\n', '\n').replace(/```([\s\S]*?)```/g, (_, code: string) => {
    const i = fences.length;
    fences.push(
      `<pre><code>${escapeHtml(code.replace(/^\n/, '').replace(/\n$/, ''))}</code></pre>`,
    );
    return `${FENCE}${i}@@`;
  });
  text = escapeHtml(text);
  const blocks = text.split(/\n{2,}/);
  const html = blocks
    .map((block) => {
      const t = block.trim();
      if (!t) {
        return '';
      }
      if (t.startsWith(FENCE)) {
        return t;
      }
      if (t.startsWith('# ')) {
        return `<h1>${inline(t.slice(2))}</h1>`;
      }
      if (t.startsWith('## ')) {
        return `<h2>${inline(t.slice(3))}</h2>`;
      }
      if (t.startsWith('### ')) {
        return `<h3>${inline(t.slice(4))}</h3>`;
      }
      const lines = t.split('\n');
      if (lines.every((l) => l.startsWith('- ') || l.startsWith('* '))) {
        const items = lines.map((l) => `<li>${inline(l.slice(2))}</li>`).join('');
        return `<ul>${items}</ul>`;
      }
      return `<p>${inline(t).replaceAll('\n', '<br />')}</p>`;
    })
    .join('');
  return html.replace(/@@SENFENCE:(\d+)@@/g, (_, i: string) => fences[Number(i)] ?? '');
}

function inline(s: string): string {
  let out = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label: string, href: string) => {
    const safe = SAFE_HREF.test(href) ? href : '#';
    return `<a href="${safe}" rel="noreferrer">${label}</a>`;
  });
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  out = out.replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<em>$2</em>');
  return out;
}
