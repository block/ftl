import { type Page, expect } from '@playwright/test'

export const navigateToModule = async (page: Page, moduleName: string) => {
  await page.goto(`/modules/${moduleName}`)
  await expect(page).toHaveURL(new RegExp(`/modules/${moduleName}`))
}

export const navigateToDecl = async (page: Page, moduleName: string, declName: string) => {
  await page.goto(`/modules/${moduleName}/verb/${declName}`)
  await expect(page).toHaveURL(new RegExp(`/modules/${moduleName}/verb/${declName}`))
}

export const pressShortcut = async (page: Page, key: string) => {
  // Get the platform-specific modifier key
  const isMac = await page.evaluate(() => navigator.userAgent.includes('Mac'))
  const modifier = isMac ? 'Meta' : 'Control'

  await page.keyboard.down(modifier)
  await page.keyboard.press(key)
  await page.keyboard.up(modifier)
}

export const setVerbRequestBody = async (page: Page, content: string) => {
  const editor = page.locator('#body-editor .cm-content[contenteditable="true"]')
  await expect(editor).toBeVisible()

  await editor.click()
  await editor.page().keyboard.press('Control+A')
  await editor.page().keyboard.press('Delete')
  await editor.fill(content)
}
