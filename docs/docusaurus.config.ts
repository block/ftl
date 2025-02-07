import type * as Preset from '@docusaurus/preset-classic'
import type { Config } from '@docusaurus/types'
import { themes as prismThemes } from 'prism-react-renderer'

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'FTL',
  tagline: 'Towards a 𝝺-calculus for large-scale systems',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://block.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/ftl/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'block', // Usually your GitHub org/user name.
  projectName: 'ftl', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      /** @type {import("@easyops-cn/docusaurus-search-local").PluginOptions} */
      ({
        hashed: true,
        language: ['en'],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      }),
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/block/ftl/tree/main/docs/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: '/ftl/img/docusaurus-social-card.jpg',
    metadata: [
      {
        name: 'description',
        content: 'FTL - Towards a λ-calculus for large-scale systems. A modern approach to building and deploying distributed systems.',
      },
      {
        name: 'og:title',
        content: 'FTL Documentation',
      },
      {
        name: 'og:description',
        content: 'FTL - Towards a λ-calculus for large-scale systems. A modern approach to building and deploying distributed systems.',
      },
      {
        property: 'og:image',
        content: '/ftl/img/docusaurus-social-card.jpg',
      },
      {
        name: 'twitter:card',
        content: 'summary_large_image',
      },
      {
        name: 'twitter:image',
        content: '/ftl/img/docusaurus-social-card.jpg',
      },
    ],
    navbar: {
      title: 'FTL',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/block/ftl',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Quickstart',
              to: '/docs/getting-started/quick-start',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Twitter',
              href: 'https://twitter.com/blocks',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/block/ftl',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} The FTL Team. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['java', 'kotlin', 'bash', 'hcl'],
    },
  } satisfies Preset.ThemeConfig,
}

export default config
