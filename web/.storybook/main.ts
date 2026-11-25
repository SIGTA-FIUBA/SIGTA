import type { StorybookConfig } from '@storybook/react-vite'

// The catalog of SIGTA: brand, tokens and every component, built from the same
// theme as the app. `npm run catalog` opens it; the image serves it at /design/.
const config: StorybookConfig = {
  framework: '@storybook/react-vite',
  stories: ['../src/catalog/**/*.mdx', '../src/**/*.stories.tsx'],
  addons: ['@storybook/addon-docs', '@storybook/addon-a11y'],
  staticDirs: [{ from: '../public/brand', to: '/brand' }],
  core: { disableTelemetry: true },
}

export default config
