import { describe, expect, it } from 'vite-plus/test';
import { renderMarkdown } from './markdown.ts';

describe('renderMarkdown', () => {
  it('escapes html then renders emphasis and links', () => {
    const html = renderMarkdown('see **bold** and *em* plus [x](https://example.com) <script>');
    expect(html).toContain('<strong>bold</strong>');
    expect(html).toContain('<em>em</em>');
    expect(html).toContain('href="https://example.com"');
    expect(html).toContain('&lt;script&gt;');
    expect(html).not.toContain('<script>');
  });

  it('renders headings, lists, and fenced code', () => {
    const html = renderMarkdown('# Title\n\n- a\n- b\n\n```\ncode <x>\n```\n');
    expect(html).toContain('<h1>Title</h1>');
    expect(html).toContain('<li>a</li>');
    expect(html).toContain('<pre><code>');
    expect(html).toContain('code &lt;x&gt;');
  });

  it('rejects javascript urls', () => {
    const html = renderMarkdown('[x](javascript:alert(1))');
    expect(html).not.toContain('javascript:');
    expect(html).toContain('href="#"');
  });

  it('renders h2, h3, and paragraph line breaks', () => {
    const html = renderMarkdown('## Section\n\n### Sub\n\nline one\nline two');
    expect(html).toContain('<h2>Section</h2>');
    expect(html).toContain('<h3>Sub</h3>');
    expect(html).toContain('<p>line one<br />line two</p>');
  });

  it('keeps mailto, hash, and root-relative links', () => {
    const html = renderMarkdown('[m](mailto:a@b.c) [h](#top) [p](/issues)');
    expect(html).toContain('href="mailto:a@b.c"');
    expect(html).toContain('href="#top"');
    expect(html).toContain('href="/issues"');
  });

  it('renders empty markdown as empty html', () => {
    expect(renderMarkdown('')).toBe('');
    expect(renderMarkdown('\n\n')).toBe('');
  });
});
