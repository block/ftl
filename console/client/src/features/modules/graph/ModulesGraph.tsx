import React from 'react'
import { modulesContext } from '../../../providers/modules-provider'
import { generateDot } from './generate-dot'
import { dotToSVG } from './dot-to-svg'
import { formatSVG } from './format-svg'
import './graph.css'

export const ModulesGraph: React.FC<{ className: string}> = ({className}) => {
  const modules = React.useContext(modulesContext)
  const viewportRef = React.useRef<HTMLDivElement>(null)
  const [viewport, setViewPort] = React.useState<HTMLDivElement>()

  React.useEffect(() => {
    const viewCur = viewportRef.current
    viewCur && setViewPort(viewCur)
  }, [])

  React.useEffect(() => {
    const renderSvg = async () => {
      const dot = generateDot(modules)
      const unformattedSVG = await dotToSVG(dot)
      if (unformattedSVG) {
        const formattedSVG = formatSVG(unformattedSVG)
        viewport?.replaceChildren(formattedSVG)
      }
    }
    viewport && void renderSvg()
  }, [modules, viewport])

  
  return <div ref={viewportRef} className={`viewport flex-1 overflow-hidden ${className}`} />
}
