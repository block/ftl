import { type ReactNode, useEffect, useRef, useState } from 'react'

interface HoverPopupProps {
  children: ReactNode
  popupContent: ReactNode
  className?: string
  position?: 'top' | 'bottom'
}

export const HoverPopup = ({ children, popupContent, className = '', position = 'top' }: HoverPopupProps) => {
  const [isHovering, setIsHovering] = useState(false)
  const elementRef = useRef<HTMLDivElement>(null)
  const [popupStyle, setPopupStyle] = useState<{ top: number; left: number; transform?: string }>({ top: 0, left: 0 })

  useEffect(() => {
    if (isHovering && elementRef.current) {
      const rect = elementRef.current.getBoundingClientRect()
      setPopupStyle({
        top: position === 'top' ? rect.top - 40 : rect.bottom + 8,
        left: rect.left + rect.width / 2,
        transform: 'translateX(-50%)',
      })
    }
  }, [isHovering, position])

  return (
    <div ref={elementRef} className={`relative ${className}`} onMouseEnter={() => setIsHovering(true)} onMouseLeave={() => setIsHovering(false)}>
      {children}
      {isHovering && (
        <div className='fixed bg-gray-100 dark:bg-gray-700 text-xs p-2 rounded shadow-lg z-[100] w-max whitespace-nowrap' style={popupStyle}>
          {popupContent}
        </div>
      )}
    </div>
  )
}
