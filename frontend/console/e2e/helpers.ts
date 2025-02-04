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

  // Ensure editor is ready and focused
  await editor.click()

  // Clear the editor by filling with empty content
  await editor.fill('')
  await expect(editor).toHaveText('')

  // Fill with the new content
  await editor.fill(content)

  // Verify content matches by normalizing both strings
  const normalizeJSON = (str: string) => {
    try {
      return JSON.stringify(JSON.parse(str))
    } catch {
      return str.replace(/\s+/g, '')
    }
  }

  const editorContent = await editor.textContent()
  expect(normalizeJSON(editorContent || '')).toBe(normalizeJSON(content))
}
