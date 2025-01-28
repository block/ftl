import type { StoryObj } from '@storybook/react/*'
import { Pill } from '../shared/components/Pill'

const meta = {
  title: 'Components/Pill',
  component: Pill,
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  args: {
    text: 'name',
  },
}
