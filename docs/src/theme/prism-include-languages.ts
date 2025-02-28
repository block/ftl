import siteConfig from '@generated/docusaurus.config';
import type * as PrismNamespace from 'prismjs';
import type {Optional} from 'utility-types';

export default function prismIncludeLanguages(
  PrismObject: typeof PrismNamespace,
): void {
  const {
    themeConfig: {prism},
  } = siteConfig;
  const {additionalLanguages} = prism as {additionalLanguages: string[]};

  // Prism components work on the Prism instance on the window, while prism-
  // react-renderer uses its own Prism instance. We temporarily mount the
  // instance onto window, import components to enhance it, then remove it to
  // avoid polluting global namespace.
  // You can mutate PrismObject: registering plugins, deleting languages... As
  // long as you don't re-assign it

  const PrismBefore = globalThis.Prism;
  globalThis.Prism = PrismObject;

  additionalLanguages.forEach((lang) => {
    if (lang === 'php') {
      // eslint-disable-next-line global-require
      require('prismjs/components/prism-markup-templating.js');
    }

    // Skip loading schema from prismjs since we have our custom implementation
    if (lang !== 'schema') {
      // eslint-disable-next-line global-require, import/no-dynamic-require
      require(`prismjs/components/prism-${lang}`);
    }
  });

  // Define custom schema language directly
  PrismObject.languages.schema = {
    'comment': {
      pattern: /(^|[^\\])\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
      lookbehind: true,
      greedy: true
    },
    'string': {
      pattern: /(["'])(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1/,
      greedy: true
    },
    'keyword': /\b(?:module|export|data|verb|topic|enum|typealias|builtin|config|database|Unit|Time|Int|String|Bool|from|calls|publish|subscribe|ingress|cron|migration|query|column|alias|json|sql|exec|many)\b/,
    'operator': /[<>]=?|[!=]=?=?|--?|\+\+?|&&?|\|\|?|[?*/~^%]/,
    'punctuation': /[{}[\];(),.:]/,
    'boolean': /\b(?:true|false)\b/,
    'number': /\b0x[\da-f]+\b|(?:\b\d+(?:\.\d*)?|\B\.\d+)(?:e[+-]?\d+)?\b/i,
    'directive': {
      pattern: /\+[a-zA-Z][a-zA-Z0-9_]*/,
      alias: 'function'
    },
    'type': {
      pattern: /\b(?:GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\b/,
      alias: 'builtin'
    },
    'special-char': {
      pattern: /\?|\*/,
      alias: 'important'
    },
    'path-pattern': {
      pattern: /\/[a-zA-Z0-9_/{}.-]+/,
      alias: 'string'
    }
  };

  // Clean up and eventually restore former globalThis.Prism object (if any)
  delete (globalThis as Optional<typeof globalThis, 'Prism'>).Prism;
  if (typeof PrismBefore !== 'undefined') {
    globalThis.Prism = PrismObject;
  }
}
