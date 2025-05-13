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
  const popupRef = useRef<HTMLDivElement>(null)
  const [popupStyle, setPopupStyle] = useState<{ top: number; left: number }>({ top: 0, left: 0 })

  useEffect(() => {
    if (isHovering && elementRef.current && popupRef.current) {
      const rect = elementRef.current.getBoundingClientRect()
      const popupRect = popupRef.current.getBoundingClientRect()
      let top = position === 'top' ? rect.top - popupRect.height - 8 : rect.bottom + 8
      let left = rect.left + rect.width / 2 - popupRect.width / 2

      // Clamp to viewport with 8px margin
      top = Math.max(8, Math.min(top, window.innerHeight - popupRect.height - 8))
      left = Math.max(8, Math.min(left, window.innerWidth - popupRect.width - 8))

      setPopupStyle({
        top,
        left,
      })
    }
  }, [isHovering, position])

  return (
    <div ref={elementRef} className={`relative ${className}`} onMouseEnter={() => setIsHovering(true)} onMouseLeave={() => setIsHovering(false)}>
      {children}
      {isHovering && (
        <div ref={popupRef} className='fixed bg-gray-100 dark:bg-gray-700 text-xs p-2 rounded shadow-lg z-[100] w-max whitespace-nowrap' style={popupStyle}>
          {popupContent}
        </div>
      )}
    </div>
  )
}
