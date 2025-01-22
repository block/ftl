document.addEventListener('DOMContentLoaded', () => {
  const codeBlocks = document.querySelectorAll('pre code')

  for (const codeBlock of codeBlocks) {
    const pre = codeBlock.parentElement
    const copyButton = document.createElement('button')
    copyButton.className = 'copy-button'
    copyButton.innerHTML = 'Copy'

    // Add button to pre element
    pre.style.position = 'relative'
    pre.appendChild(copyButton)

    copyButton.addEventListener('click', async () => {
      if (!navigator.clipboard) {
        console.error('Clipboard API not available')
        return
      }

      try {
        await navigator.clipboard.writeText(codeBlock.textContent)
        copyButton.innerHTML = 'Copied!'
        setTimeout(() => {
          copyButton.innerHTML = 'Copy'
        }, 2000)
      } catch (error) {
        console.error('Copy failed', error)
        copyButton.innerHTML = 'Error!'
        setTimeout(() => {
          copyButton.innerHTML = 'Copy'
        }, 2000)
      }
    })
  }
})
