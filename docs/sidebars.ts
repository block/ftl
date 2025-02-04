import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      link: {
        type: 'generated-index',
      },
      collapsed: false,
      items: [
        'getting-started/introduction',
        'getting-started/quick-start',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      link: {
        type: 'generated-index',
      },
      collapsed: false,
      items: [
        'reference/verbs',
        'reference/types',
        'reference/visibility',
        'reference/databases',
        'reference/cron',
        'reference/ingress',
        'reference/secretsconfig',
        'reference/pubsub',
        'reference/retries',
        'reference/externaltypes',
        'reference/unittests',
        'reference/matrix',
      ],
    },
  ],
};

export default sidebars;
