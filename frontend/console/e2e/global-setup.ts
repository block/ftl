import { type FullConfig, chromium } from '@playwright/test'

const globalSetup = async (config: FullConfig) => {
  console.log('Waiting for server to be available...')

  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()
  await page.goto('http://localhost:8899/modules')

  console.log('Waiting for modules to be ready...')
  const moduleNames = ['time', 'echo', 'cron', 'http', 'pubsub']
  await page.waitForFunction(
    (modules) => {
      const readyModules = modules.filter((module) => {
        const moduleElement = document.querySelector(`li#module-tree-module-${module}`)
        if (!moduleElement) return false

        const greenDot = moduleElement.querySelector('.bg-green-400')
        return greenDot !== null
      })
      console.log('Ready modules:', readyModules.join(', '))
      return readyModules.length === modules.length
    },
    moduleNames,
    { timeout: 240000 },
  )

  console.log('All modules are ready!')

  await browser.close()
}

export default globalSetup
