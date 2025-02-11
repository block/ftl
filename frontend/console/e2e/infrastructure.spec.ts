import { expect, test } from '@playwright/test'

test('shows infrastructure', async ({ page }) => {
  await page.goto('/')
  const infrastructureNavItem = page.getByRole('link', { name: 'Infrastructure' })
  await infrastructureNavItem.click()
  await expect(page).toHaveURL(/\/infrastructure\/deployments$/)

  const deploymentsTab = await page.getByRole('button', { name: 'Deployments' })
  await expect(deploymentsTab).toBeVisible()

  const buildEventsTab = await page.getByRole('button', { name: 'Build Engine Events' })
  await expect(buildEventsTab).toBeVisible()
})
