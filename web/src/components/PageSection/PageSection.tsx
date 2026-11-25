import { Grid, Stack, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'

// PageSection is a part of a page on the grid of the page: on a wide screen,
// its title and what it says on the left quarter and its content on the rest,
// so every section starts on the same line; on a narrow one, the title above.
export function PageSection({
  id,
  title,
  lead,
  action,
  children,
}: {
  id: string
  title: string
  lead?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <Grid component="section" aria-labelledby={id} gap={{ base: 'md', lg: 'xl' }}>
      <Grid.Col span={{ base: 12, lg: 3 }}>
        <Stack gap="xs" align="flex-start">
          <Title order={2} id={id}>
            {title}
          </Title>
          {lead && (
            <Text size="sm" c="gray.7">
              {lead}
            </Text>
          )}
          {action}
        </Stack>
      </Grid.Col>
      <Grid.Col span={{ base: 12, lg: 9 }}>{children}</Grid.Col>
    </Grid>
  )
}
