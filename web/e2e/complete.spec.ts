import { expect, test, type APIResponse } from '@playwright/test';

async function json<T>(res: APIResponse): Promise<T> {
  if (!res.ok()) {
    throw new Error(`${res.status()} ${await res.text()}`);
  }
  return (await res.json()) as T;
}

test('adhoc filter, custom label, and delete', async ({ page, request }) => {
  const stamp = `${Date.now()}`;
  const keepTitle = `Keep ${stamp}`;
  const hideTitle = `Hide ${stamp}`;
  await json(await request.post('/api/issues', { data: { title: keepTitle, status: 'todo' } }));
  await json(await request.post('/api/issues', { data: { title: hideTitle, status: 'done' } }));

  await page.goto('/issues');
  await page.getByLabel('Filter status').selectOption('todo');
  await expect(page).toHaveURL(/status=todo/);
  const list = page.getByRole('listbox', { name: 'Issues' });
  await expect(list.getByRole('option', { name: new RegExp(keepTitle) })).toBeVisible();
  await expect(list.getByRole('option', { name: new RegExp(hideTitle) })).toHaveCount(0);

  await page.getByLabel('New view name').fill(`Todo ${stamp}`);
  await page.getByLabel('New view name').press('Enter');
  await expect(page).toHaveURL(new RegExp(`/views/todo-${stamp}`));
  await expect(page.getByRole('heading', { name: `Todo ${stamp}` })).toBeVisible();

  await page.goto(`/issues`);
  await page.getByRole('option', { name: new RegExp(keepTitle) }).click();
  await expect(page).toHaveURL(/\/issues\/SEN-/);
  const label = `Harbor ${stamp}`;
  await page.getByLabel('New label').fill(label);
  await page.getByLabel('New label').press('Enter');
  await expect(page.getByRole('button', { name: label, exact: true })).toHaveAttribute(
    'aria-pressed',
    'true',
  );

  page.once('dialog', (dialog) => dialog.accept());
  await page.getByRole('button', { name: 'Delete', exact: true }).click();
  await expect(page).toHaveURL(/\/issues/);
  await expect(page.getByRole('option', { name: new RegExp(keepTitle) })).toHaveCount(0);
});
