import { expect, test, type APIResponse } from '@playwright/test';

async function json<T>(res: APIResponse): Promise<T> {
  if (!res.ok()) {
    throw new Error(`${res.status()} ${await res.text()}`);
  }
  return (await res.json()) as T;
}

test('shortcuts, find, dirty, and project-scoped create', async ({ page, request }) => {
  const stamp = `${Date.now()}`;
  const needle = `Needle ${stamp}`;
  const hay = `Haystack ${stamp}`;
  await json(await request.post('/api/issues', { data: { title: needle, status: 'todo' } }));
  await json(await request.post('/api/issues', { data: { title: hay, status: 'todo' } }));
  const project = await json<{ slug: string }>(
    await request.post('/api/projects', { data: { name: `Pier ${stamp}`, slug: `pier-${stamp}` } }),
  );

  await page.goto('/issues');
  await page.getByRole('heading', { name: 'Issues' }).click();
  await expect(page.getByText('Unpushed')).toBeVisible();

  await page.keyboard.press('?');
  await expect(page.getByRole('dialog', { name: 'Keyboard shortcuts' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('dialog', { name: 'Keyboard shortcuts' })).toHaveCount(0);

  await page.keyboard.press('/');
  await expect(page.getByLabel('Find issues')).toBeFocused();
  await page.getByLabel('Find issues').fill('Needle');
  const list = page.getByRole('listbox', { name: 'Issues' });
  await expect(list.getByRole('option', { name: new RegExp(needle) })).toBeVisible();
  await expect(list.getByRole('option', { name: new RegExp(hay) })).toHaveCount(0);

  await page.goto(`/projects/${project.slug}`);
  await page.getByRole('button', { name: 'New issue' }).click();
  await expect(page.getByRole('dialog', { name: 'Create issue' })).toBeVisible();
  await expect(page.getByLabel('Issue project')).not.toHaveValue('');
});
