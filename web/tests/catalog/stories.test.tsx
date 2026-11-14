import { composeStories, setProjectAnnotations } from '@storybook/react-vite'
import { render } from '@testing-library/react'
import type { ComponentType } from 'react'
import { describe, expect, test } from 'vitest'
import preview from '../../.storybook/preview'

// Every story of the catalog renders with the app's theme and without errors,
// so a change to the theme or to a component cannot break the catalog unseen.
setProjectAnnotations(preview)

const modules = import.meta.glob<Parameters<typeof composeStories>[0]>('../../src/**/*.stories.tsx', { eager: true })

describe.each(Object.entries(modules))('%s', (_path, module) => {
  test.each(Object.entries(composeStories(module)))('%s', (_name, story) => {
    const Story = story as ComponentType
    const { container } = render(<Story />)
    expect(container).not.toBeEmptyDOMElement()
  })
})
