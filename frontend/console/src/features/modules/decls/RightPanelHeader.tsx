import { LinkSquare02Icon } from 'hugeicons-react'
import { Link } from 'react-router-dom'
import { HoverPopup } from '../../../shared/components/HoverPopup'

export const RightPanelHeader = ({ Icon, title, url, urlHoverText }: { Icon?: React.ElementType; title?: string; url?: string; urlHoverText?: string }) => {
  return (
    <div className='flex items-center justify-between gap-2 px-2 py-2'>
      <div className='flex items-center gap-2'>
        {Icon && <Icon className='h-5 w-5 text-indigo-500' />}
        {title && <div className='flex flex-col min-w-0'>{title}</div>}
      </div>
      {url && (
        <div className='flex items-center gap-2'>
          <HoverPopup popupContent={urlHoverText || url} position='top'>
            <Link to={url} className='text-indigo-500 hover:text-indigo-700' aria-label='View in modules'>
              <LinkSquare02Icon className='w-5 h-5' />
            </Link>
          </HoverPopup>
        </div>
      )}
    </div>
  )
}
