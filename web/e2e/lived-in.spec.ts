import { expect, test, type APIResponse } from '@playwright/test';

async function json<T>(res: APIResponse): Promise<T> {
  if (!res.ok()) {
    throw new Error(`${res.status()} ${await res.text()}`);
  }
  return (await res.json()) as T;
}

test.describe('lived-in workspace', () => {
  test('nested issues, compound view, and palette search', async ({ page, request }) => {
    const stamp = `${Date.now()}`;
    const labels = await json<{ id: number; name: string }[]>(await request.get('/api/labels'));
    const bug = labels[0];
    if (!bug) {
      throw new Error('expected seeded labels');
    }
    const project = await json<{ id: number }>(
      await request.post('/api/projects', { data: { name: 'Harbor', slug: `harbor-${stamp}` } }),
    );
    const epicTitle = `Ship harbor ${stamp}`;
    const childTitle = `Write runbook ${stamp}`;
    const doneTitle = `Already shipped ${stamp}`;
    const viewName = `Harbor todo ${stamp}`;
    const epic = await json<{ id: number; identifier: string }>(
      await request.post('/api/issues', {
        data: {
          title: epicTitle,
          status: 'todo',
          priority: 1,
          projectId: project.id,
          labelIds: [bug.id],
        },
      }),
    );
    await json(
      await request.post('/api/issues', {
        data: {
          title: childTitle,
          status: 'todo',
          parentId: epic.id,
          projectId: project.id,
        },
      }),
    );
    await json(await request.post('/api/issues', { data: { title: doneTitle, status: 'done' } }));
    await json(
      await request.post('/api/views', {
        data: {
          name: viewName,
          slug: `harbor-todo-${stamp}`,
          display: 'list',
          status: 'todo',
          project: `harbor-${stamp}`,
        },
      }),
    );

    await page.goto('/issues');
    await expect(page.getByRole('heading', { name: 'Issues' })).toBeVisible();
    const issueList = page.getByRole('listbox', { name: 'Issues' });
    await expect(issueList.getByRole('option', { name: new RegExp(epicTitle) })).toBeVisible();
    await expect(issueList.getByRole('option', { name: new RegExp(childTitle) })).toBeVisible();
    await expect(issueList.getByRole('option', { name: new RegExp(doneTitle) })).toBeVisible();

    await page.getByRole('link', { name: viewName, exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`/views/harbor-todo-${stamp}`));
    const viewList = page.getByRole('listbox', { name: 'Issues' });
    await expect(viewList.getByRole('option', { name: new RegExp(epicTitle) })).toBeVisible();
    await expect(viewList.getByRole('option', { name: new RegExp(childTitle) })).toBeVisible();
    await expect(viewList.getByRole('option', { name: new RegExp(doneTitle) })).toHaveCount(0);

    await page.getByRole('button', { name: /Command palette/ }).click();
    const palette = page.getByRole('dialog', { name: 'Command palette' });
    await palette.getByLabel('Command search').fill(childTitle);
    await palette.getByRole('option', { name: new RegExp(childTitle) }).click();
    await expect(page).toHaveURL(/\/issues\/SEN-/);
    await expect(page.locator('.title-input')).toHaveValue(childTitle);
    await expect(page.getByLabel('Parent')).not.toHaveValue('');
  });
});
