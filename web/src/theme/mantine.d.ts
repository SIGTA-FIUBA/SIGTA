// The spacing scale of tokens.json has eight steps, three more than Mantine's
// xs to xl, so the props that take a spacing accept them too.
type Spacing = 'xxs' | 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl' | 'xxxl'

declare module '@mantine/core' {
  export interface MantineThemeSizesOverride {
    spacing: Record<Spacing, string>
  }
}

export {}
