import { type FullConfig, chromium } from '@playwright/test'

const globalSetup = async (config: FullConfig) => {
  console.log('Waiting for server to be available...')
  const startTime = Date.now()

  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  try {
    await page.goto('http://localhost:8899/modules')
    console.log(`Server available after ${(Date.now() - startTime) / 1000}s`)

    console.log('Waiting for modules to be ready...')
    const moduleNames = ['time', 'echo', 'cron', 'http', 'pubsub']
    const moduleStartTime = Date.now()

    await page.waitForFunction(
      (modules) => {
        const readyModules = modules.filter((module) => {
          const moduleElement = document.querySelector(`li#module-tree-module-${module}`)
          if (!moduleElement) {
            console.log(`Module ${module} element not found`)
            return false
          }

          const greenDot = moduleElement.querySelector('.bg-green-400')
          const isReady = greenDot !== null
          if (isReady) {
            const timestamp = new Date().toISOString()
            console.log(`[${timestamp}] Module ${module} is ready`)
          }
          return isReady
        })

        if (readyModules.length === modules.length) {
          const timestamp = new Date().toISOString()
          console.log(`[${timestamp}] All modules are ready!`)
          return true
        }

        const pendingModules = modules.filter((m) => !readyModules.includes(m))
        console.log('Still waiting for modules:', pendingModules.join(', '))
        return false
      },
      moduleNames,
      { timeout: 240000 },
    )

    console.log(`All modules ready after ${(Date.now() - moduleStartTime) / 1000}s`)
    console.log(`Total startup time: ${(Date.now() - startTime) / 1000}s`)
  } catch (error) {
    console.error('Error during startup:', error)
    throw error
  } finally {
    await browser.close()
  }
}

export default globalSetup
