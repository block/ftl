import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

const features = [
  {
    title: 'Infrastructure as code',
    content: 'Not YAML. Declare your infrastructure in the same language you\'re writing in as type-safe values, rather than in separate configuration files disassociated from their point of use.',
  },
  {
    title: 'Language agnostic',
    content: 'FTL makes it possible to write backend code in your language of choice. You write normal code, and FTL extracts a service interface from your code directly, making your functions and types automatically available to all supported languages.',
  },
  {
    title: 'Fearless development against production',
    content: 'There is no substitute for production data. FTL plans to support forking of production infrastructure and code during development.',
  },
  {
    title: 'Fearlessly modify types',
    content: 'Multiple versions of a single verb with different signatures can be live concurrently. See <a href="https://www.unison-lang.org/">Unison</a> for inspiration. We can statically detect changes that would violate runtime and persistent data constraints.',
  },
  {
    title: 'AI from the ground up',
    content: 'We plan to integrate AI sensibly and deeply into the FTL platform. Automated AI-driven tuning suggestions, automated third-party API integration, and so on.',
  },
];

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/getting-started/introduction">
            Get started
          </Link>
        </div>
        <div className={styles.version}>
          <p><a href="https://github.com/block/ftl">Open-source Apache License</a></p>
        </div>
      </div>
    </header>
  );
}

function Feature({title, content}: {title: string; content: string}) {
  return (
    <div className={styles.feature}>
      <Heading as="h3">{title}</Heading>
      <p dangerouslySetInnerHTML={{__html: content}} />
    </div>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <div className="container">
          <div className={styles.features}>
            {features.map((props, idx) => (
              <Feature key={idx} {...props} />
            ))}
          </div>
        </div>
      </main>
    </Layout>
  );
}
