import { expect, test } from '@playwright/test'

test('shows active modules', async ({ page }) => {
  await page.goto('/')
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()
  await expect(page).toHaveURL(/\/modules$/)

  await page.waitForSelector('[data-module-row]')
  const moduleNames = await page.$$eval('[data-module-row]', (elements) => elements.map((el) => el.getAttribute('data-module-row')))

  const expectedModuleNames = ['cron', 'time', 'pubsub', 'http', 'echo']
  expect(moduleNames).toEqual(expect.arrayContaining(expectedModuleNames))
})

test('tapping on a module navigates to the module page', async ({ page }) => {
  await page.goto('/modules')
  await page.locator(`[data-module-row="echo"]`).click()
  await expect(page).toHaveURL(/\/modules\/echo$/)
})

test('tapping through module tree navigates to decl page', async ({ page }) => {
  await page.goto('/modules')
  await page.locator(`[data-module-row="echo"]`).click()
  await expect(page).toHaveURL(/\/modules\/echo$/)

  // Expand the verb group for echo
  await page.locator('#module-echo-tree-group').click()
  await page.locator('#module-tree-module-echo .group:has-text("verb")').click()
  await page.locator('a#decl-echo').click()

  await expect(page).toHaveURL(/\/modules\/echo\/verb\/echo$/)
})

test('clearing type filter works correctly', async ({ page }) => {
  await page.goto('/modules')
  await page.waitForSelector('[data-module-row]')

  // Open the filter dropdown and clear the selections.
  await page.getByRole('button', { name: 'Filter types...' }).click()
  await page.getByRole('button', { name: 'Deselect all' }).click()

  await page.locator('body').click()

  // Verify tree is empty when no types are selected
  await expect(page.locator('[data-test-id="module-tree-group"]')).toHaveCount(0)
})
