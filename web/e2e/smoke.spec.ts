import { expect, test } from '@playwright/test';

test('create issue, comment, and page', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/issues/);
  await expect(page.getByRole('heading', { name: 'Issues' })).toBeVisible();
  await page.getByRole('heading', { name: 'Issues' }).click();

  await page.keyboard.press('c');
  const issueTitle = page.getByPlaceholder('Issue title');
  await expect(issueTitle).toBeFocused();
  await issueTitle.fill('Smoke issue');
  await issueTitle.press('Enter');
  await expect(page.getByPlaceholder('Issue title')).toHaveCount(0);
  await expect(page).toHaveURL(/\/issues\/SEN-\d+/);
  const identifier = page.url().match(/SEN-\d+/)?.[0];
  if (!identifier) {
    throw new Error('expected issue identifier in the URL');
  }
  await expect(page.locator('.title-input')).toHaveValue('Smoke issue');

  await page.getByRole('button', { name: 'Bug' }).click();
  await expect(page.getByRole('button', { name: 'Bug' })).toHaveAttribute('aria-pressed', 'true');

  await page.getByLabel('Due date').fill('2026-09-01');
  await page.getByRole('tab', { name: 'Preview' }).click();
  await expect(page.getByText('Empty')).toBeVisible();
  await page.getByRole('tab', { name: 'Edit' }).click();
  await page.getByLabel('Markdown body').fill('## Goal\n\nShow **labels**.');
  await page.getByLabel('Markdown body').blur();
  await page.getByRole('tab', { name: 'Preview' }).click();
  await expect(page.getByRole('heading', { name: 'Goal' })).toBeVisible();

  const comment = page.getByLabel('New note');
  await comment.fill('looks good');
  await comment.press('Control+Enter');
  await expect(page.getByText('looks good')).toBeVisible();

  await page.getByRole('link', { name: 'Projects' }).click();
  await page.getByLabel('New project name').fill('Atlas');
  await page.getByLabel('New project name').press('Enter');
  await expect(page).toHaveURL(/\/projects\/atlas/);
  await expect(page.getByRole('heading', { name: 'Atlas' })).toBeVisible();

  await page.getByRole('link', { name: 'Cycles' }).click();
  await page.getByRole('button', { name: 'New cycle' }).click();
  await expect(page).toHaveURL(/\/cycles\/1/);
  await expect(page.getByRole('heading', { name: 'Cycle 1' })).toBeVisible();

  await page.goto(`/issues/${identifier}`);
  await expect(page.locator('.title-input')).toHaveValue('Smoke issue');
  await page.getByRole('button', { name: /Command palette/ }).click();
  await page.getByLabel('Command search').fill('Assign to Cycle');
  await page.getByRole('option', { name: 'Assign to Cycle 1' }).click();
  await expect(page.getByLabel('Cycle')).toHaveValue(/[1-9]/);

  await page.getByLabel('Status').selectOption('done');
  await page.getByRole('link', { name: 'Board' }).click();
  const doneCol = page.locator('.column', { has: page.getByRole('heading', { name: 'Done' }) });
  await expect(doneCol.getByRole('button', { name: new RegExp(identifier) })).toBeVisible();

  await page.getByRole('link', { name: 'Pages' }).click();
  await page.keyboard.press('p');
  const pageTitle = page.getByPlaceholder('Page title');
  await expect(pageTitle).toBeFocused();
  await pageTitle.fill('ADR 1');
  await pageTitle.press('Enter');
  await expect(page).toHaveURL(/\/pages\//);
  await expect(page.getByRole('textbox', { name: 'Page title' })).toHaveValue('ADR 1');
  await page.getByLabel('Page project').selectOption({ label: 'Atlas' });
  await expect(page.getByLabel('Page project')).not.toHaveValue('');
});

test('sub-issue and saved view', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/issues/);
  await expect(page.getByRole('heading', { name: 'Issues' })).toBeVisible();
  await page.getByRole('heading', { name: 'Issues' }).click();
  await page.keyboard.press('c');
  const issueTitle = page.getByPlaceholder('Issue title');
  await expect(issueTitle).toBeFocused();
  await issueTitle.fill('Parent job');
  await issueTitle.press('Enter');
  await expect(page.getByPlaceholder('Issue title')).toHaveCount(0);
  await expect(page.locator('.title-input')).toHaveValue('Parent job');

  await page.getByLabel('New sub-issue').fill('Child step');
  await page.getByLabel('New sub-issue').press('Enter');
  await expect(page.getByRole('button', { name: /Child step/ })).toBeVisible();

  await page.getByRole('link', { name: 'Issues' }).click();
  const issueList = page.getByRole('listbox', { name: 'Issues' });
  await expect(issueList.getByRole('option', { name: /Parent job/ })).toBeVisible();
  await expect(issueList.getByRole('option', { name: /Child step/ })).toBeVisible();

  await page.getByRole('button', { name: 'New view' }).click();
  const viewName = page.getByPlaceholder('View name');
  await expect(viewName).toBeFocused();
  await viewName.fill('Todos');
  await viewName.press('Enter');
  await expect(page).toHaveURL(/\/views\/todos/);
  await page.getByLabel('View status').selectOption('todo');
  await expect(page.getByLabel('View status')).toHaveValue('todo');
  await expect(page.getByRole('link', { name: 'Todos' })).toBeVisible();
});
