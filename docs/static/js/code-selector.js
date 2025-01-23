document.addEventListener('DOMContentLoaded', inject)

function inject() {
  document.querySelectorAll('.code-selector').forEach(injectTabs)
}

function injectTabs(codeSelector) {
  const nav = codeSelector.querySelector('ul.nav')

  const langs = []
  const tabs = []
  let currentLang = null
  for (const element of codeSelector.childNodes) {
    let toAdd = null

    if (element.nodeType === Node.COMMENT_NODE) {
      // if it's a comment like <!-- go -->, then set the currentLang to go and add it to langs
      // end will cancel the currentLang, so you can continue to write common documentation.
      const lang = element.data.trim()
      if (lang === 'end') {
        currentLang = null
      } else {
        currentLang = lang
        toAdd = { lang, element }
      }
    } else if (currentLang) {
      // if we are in a currentLang, then add the element to it.
      toAdd = { lang: currentLang, element }
    }

    if (!toAdd) {
      continue
    }

    langs.push(toAdd)
    if (!tabs.includes(toAdd.lang)) {
      tabs.push(toAdd.lang)
    }
  }

  // Let us put this in the codeSelector element, so that we later check if the new selected language is in the list first.
  codeSelector.langs = langs

  const saved = localStorage.getItem('code-lang')
  let selected = tabs[0]
  if (saved) {
    // Only select the saved language if it's in the list.
    const found = tabs.find((lang) => lang === saved)
    if (found) {
      selected = found
    }
  }

  for (const lang of tabs) {
    const li = document.createElement('li')
    li.classList.add('nav-item')

    const a = document.createElement('a')
    a.classList.add('nav-link')
    if (selected === lang) {
      a.classList.add('active')
      a.setAttribute('aria-current', 'page')
    }

    a.href = '#'
    a.textContent = capitalize(lang)
    a.lang = lang

    a.addEventListener('click', (e) => {
      e.preventDefault()
      changeLanguage(lang)
    })

    li.appendChild(a)
    nav.appendChild(li)
  }

  for (const { lang, element } of langs) {
    if (element.classList && selected !== lang) {
      element.classList.add('d-none')
    }
  }
}

function capitalize(str) {
  return str[0].toUpperCase() + str.slice(1)
}

function changeLanguage(lang) {
  localStorage.setItem('code-lang', lang)

  for (const codeSelector of document.querySelectorAll('.code-selector')) {
    const langs = codeSelector.langs
    if (!langs) {
      console.error('Missing langs property on codeSelector', codeSelector)
      continue
    }

    const selected = langs.find((l) => l.lang === lang)
    if (!selected) {
      // This tab group doesn't have the selected language--all good.
      continue
    }

    // Show/hide each element within the codeSelector
    for (const l of langs) {
      if (!l.element.classList) {
        continue
      }
      if (l.lang === selected.lang) {
        l.element.classList.remove('d-none')
      } else {
        l.element.classList.add('d-none')
      }
    }

    // Update the active tab
    for (const a of codeSelector.querySelectorAll('a.nav-link')) {
      a.classList.remove('active')
      a.removeAttribute('aria-current')
      if (a.lang === lang) {
        a.classList.add('active')
        a.setAttribute('aria-current', 'page')
      }
    }
  }
}
