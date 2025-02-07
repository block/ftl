import Link from '@docusaurus/Link'
import useDocusaurusContext from '@docusaurus/useDocusaurusContext'
import Heading from '@theme/Heading'
import Layout from '@theme/Layout'
import clsx from 'clsx'
import type { ReactNode } from 'react'
import { BiCodeBlock } from 'react-icons/bi'
import { BsCodeSlash, BsGearWideConnected, BsRobot } from 'react-icons/bs'
import { TbVersions } from 'react-icons/tb'

import styles from './index.module.css'

const features = [
  {
    title: 'Infrastructure as code',
    Icon: BiCodeBlock,
    description:
      "Not YAML. Declare your infrastructure in the same language you're writing in as type-safe values, rather than in separate configuration files disassociated from their point of use.",
  },
  {
    title: 'Language agnostic',
    Icon: BsCodeSlash,
    description:
      'FTL makes it possible to write backend code in your language of choice. You write normal code, and FTL extracts a service interface from your code directly, making your functions and types automatically available to all supported languages.',
  },
  {
    title: 'Fearless development against production',
    Icon: BsGearWideConnected,
    description: 'There is no substitute for production data. FTL plans to support forking of production infrastructure and code during development.',
  },
  {
    title: 'Fearlessly modify types',
    Icon: TbVersions,
    description: (
      <>
        Multiple versions of a single verb with different signatures can be live concurrently. See <Link to='https://www.unison-lang.org/'>Unison</Link> for
        inspiration. We can statically detect changes that would violate runtime and persistent data constraints.
      </>
    ),
  },
  {
    title: 'AI from the ground up',
    Icon: BsRobot,
    description:
      'We plan to integrate AI sensibly and deeply into the FTL platform. Automated AI-driven tuning suggestions, automated third-party API integration, and so on.',
  },
]

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext()
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className='container'>
        <h1 className='hero__title'>{siteConfig.title}</h1>
        <p className='hero__subtitle'>{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link className='button button--secondary button--lg' to='docs/getting-started/quick-start'>
            Get Started
          </Link>
        </div>
        <div className={styles.version}>
          <p>
            <a href='https://github.com/block/ftl'>Open-source Apache License</a>
          </p>
        </div>
      </div>
    </header>
  )
}

function Feature({ title, Icon, description }: { title: string; Icon: React.ComponentType<React.ComponentProps<'svg'>>; description: ReactNode }) {
  return (
    <div className={clsx('col col--4')}>
      <div className='text--center'>
        <Icon className={styles.featureIcon} />
      </div>
      <div className='text--center padding-horiz--md'>
        <Heading as='h3'>{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  )
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext()
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <section className={styles.features}>
          <div className='container'>
            <div className='row'>
              {features.map((props, idx) => (
                <Feature key={idx} {...props} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  )
}
