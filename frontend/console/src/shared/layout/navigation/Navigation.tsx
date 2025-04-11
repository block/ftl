import { CellsIcon, Database01Icon, ListViewIcon, WorkflowSquare06Icon } from 'hugeicons-react'
import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { AIAgent } from '../../../features/ai-agent/AIAgent'
import { CommandPalette } from '../../../features/command-pallete/CommandPalette'
import { DarkModeSwitch } from '../../components'
import { classNames } from '../../utils'
import { AIInput } from './AIInput'
import { SearchInput } from './SearchInput'
import { Version } from './Version'

const navigation = [
  { name: 'Events', href: '/events', icon: ListViewIcon },
  { name: 'Modules', href: '/modules', icon: CellsIcon },
  { name: 'Graph', href: '/graph', icon: WorkflowSquare06Icon },
  { name: 'Infrastructure', href: '/infrastructure', icon: Database01Icon },
]

export const Navigation = ({ version }: { version?: string }) => {
  const [isCommandPalleteOpen, setIsCommandPalleteOpen] = useState(false)
  const [isAIAgentOpen, setIsAIAgentOpen] = useState(false)

  return (
    <nav className='bg-indigo-600'>
      <div className='mx-auto pl-3 pr-4'>
        <div className='flex h-16 items-center justify-between'>
          <div className='flex items-center'>
            <div>
              <div className='flex items-baseline space-x-4'>
                {navigation.map((item) => (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    className={({ isActive }) =>
                      classNames(
                        isActive ? 'bg-indigo-700 text-white' : 'text-white hover:bg-indigo-500 hover:bg-opacity-75',
                        'rounded-md px-3 py-2 text-sm font-medium flex items-center space-x-2',
                      )
                    }
                  >
                    <item.icon className='text-lg size-5' />
                    <span className='hidden md:inline'>{item.name}</span>
                  </NavLink>
                ))}
              </div>
            </div>
          </div>
          <div className='flex items-center'>
            <div className='flex items-center mr-4'>
              <SearchInput onFocus={() => setIsCommandPalleteOpen(true)} />
              <AIInput isOpen={isAIAgentOpen} onToggle={() => setIsAIAgentOpen(!isAIAgentOpen)} />
            </div>
            <div className='flex items-center space-x-4'>
              <Version version={version} />
              <DarkModeSwitch />
            </div>
          </div>
          <CommandPalette isOpen={isCommandPalleteOpen} onClose={() => setIsCommandPalleteOpen(false)} />
          <AIAgent isOpen={isAIAgentOpen} onClose={() => setIsAIAgentOpen(false)} />
        </div>
      </div>
    </nav>
  )
}
