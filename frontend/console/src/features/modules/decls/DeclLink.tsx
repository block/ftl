import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Link } from 'react-router-dom'
import { classNames } from '../../../shared/utils'
import { useStreamModules } from '../hooks/use-stream-modules'
import { Schema } from '../schema/Schema'
import { type DeclSchema, declSchemaFromModules } from '../schema/schema.utils'

const maxWidth = 400

const SnippetContainer = ({
  decl,
  moduleName,
  visible,
  linkRect,
  containerRect,
}: { decl: DeclSchema; moduleName: string; visible: boolean; linkRect?: DOMRect; containerRect?: DOMRect }) => {
  const ref = useRef<HTMLDivElement>(null)
  const [dimensions, setDimensions] = useState<{ width: number; height: number }>()

  // After first render, get the popup dimensions
  useEffect(() => {
    if (ref.current) {
      setDimensions({
        width: ref.current.offsetWidth,
        height: ref.current.offsetHeight,
      })
    }
  }, [ref.current])

  if (!linkRect) return null

  // Calculate position to keep popup within window bounds
  const left = Math.min(
    linkRect.left,
    window.innerWidth - (dimensions?.width || maxWidth) - 10, // 10px safety margin
  )

  // Check if popup would go off bottom of screen
  const wouldGoOffBottom = linkRect.bottom + 4 + (dimensions?.height || 300) > window.innerHeight
  const top = wouldGoOffBottom
    ? Math.max(10, linkRect.top - (dimensions?.height || 300) - 4) // Show above, with 10px minimum from top
    : linkRect.bottom + 4 // Show below

  const content = (
    <div
      ref={ref}
      key={`${moduleName}.${decl.declType}`}
      style={{
        position: 'fixed',
        top: `${top}px`,
        left: `${left}px`,
        maxWidth: `${maxWidth}px`,
        maxHeight: '80vh',
        overflow: 'auto',
      }}
      className={classNames(
        visible ? '' : 'invisible',
        'rounded-md border-solid border border-gray-400 bg-gray-200 dark:border-gray-800 dark:bg-gray-700',
        'text-gray-700 dark:text-white text-xs font-normal z-10 drop-shadow-xl cursor-default p-2',
      )}
    >
      <Schema schema={decl.schema} moduleName={moduleName} containerRect={containerRect} />
    </div>
  )

  return createPortal(content, document.body)
}

// When `slim` is true, print only the decl name, not the module name, and show nothing on hover.
export const DeclLink = ({
  moduleName,
  declName,
  slim,
  textColors = 'text-indigo-600 dark:text-indigo-400',
  containerRect,
}: { moduleName?: string; declName: string; slim?: boolean; textColors?: string; containerRect?: DOMRect }) => {
  if (!moduleName || !declName) {
    return
  }
  const modules = useStreamModules()
  const decl = useMemo(
    () => (moduleName && !!modules?.data ? declSchemaFromModules(moduleName, declName, modules?.data.modules) : undefined),
    [moduleName, declName, modules?.data],
  )
  const [isHovering, setIsHovering] = useState(false)
  const linkRef = useRef<HTMLAnchorElement>(null)

  const str = moduleName && slim !== true ? `${moduleName}.${declName}` : declName

  if (!decl) {
    return str
  }

  return (
    <span
      className='inline-block rounded-md cursor-pointer hover:bg-gray-400/30 hover:dark:bg-gray-900/30 p-1 -m-1 relative'
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
    >
      <Link ref={linkRef} className={textColors} to={`/modules/${moduleName}/${decl.declType}/${declName}`}>
        {str}
      </Link>
      {!slim && (
        <SnippetContainer
          decl={decl}
          moduleName={moduleName}
          visible={isHovering}
          linkRect={linkRef?.current?.getBoundingClientRect()}
          containerRect={containerRect}
        />
      )}
    </span>
  )
}
