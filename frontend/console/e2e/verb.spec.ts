import { expect, ftlTest } from './ftl-test'

ftlTest.beforeEach(async ({ page }) => {
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()

  const moduleEcho = page.locator('div.cursor-pointer').getByText('echo').nth(0)
  await moduleEcho.click()

  const verbEcho = page.locator('div#decl-echo')
  await verbEcho.click()

  await expect(page).toHaveURL(/\/modules\/echo\/verb\/echo/)
})

ftlTest('shows verb form', async ({ page }) => {
  await expect(page.getByText('CALL', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('echo.echo')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Verb Schema', { exact: true })).toBeVisible()
  await expect(page.getByText('JSONSchema', { exact: true })).toBeVisible()
})
