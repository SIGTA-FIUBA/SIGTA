import { List, Paper, Stack, Text, Title } from '@mantine/core'
import { BrandSignature } from '../components/BrandSignature/BrandSignature'
import { PageHeader } from '../components/PageHeader/PageHeader'
import { type Entry, pillars, principles } from '../content/platform'

function Entries({ id, title, entries }: { id: string; title: string; entries: Entry[] }) {
  return (
    <Stack component="section" aria-labelledby={id} gap="md">
      <Title order={2} id={id}>
        {title}
      </Title>
      <List listStyleType="none" spacing="md" p={0}>
        {entries.map((entry) => (
          <List.Item key={entry.title}>
            <Text fw={600}>{entry.title}</Text>
            <Text size="sm" c="dimmed" maw="var(--sigta-measure)">
              {entry.text}
            </Text>
          </List.Item>
        ))}
      </List>
    </Stack>
  )
}

// AboutPage says what SIGTA is and the criteria it follows. It moved here
// from the home page, which now shows what to do.
export function AboutPage() {
  return (
    <Stack gap="xxl">
      {/* The joint signature goes on white; the page behind is gray. */}
      <Paper withBorder p="lg" w="fit-content">
        <BrandSignature />
      </Paper>
      <PageHeader
        title="Acerca de SIGTA"
        description="SIGTA, el Sistema Integral de Gestión de Trámites Académicos, es la plataforma de la Secretaría de Gestión Académica para iniciar, gestionar y seguir los trámites académicos de la facultad, del inicio a la resolución."
      />
      <Entries id="pillars" title="Qué ofrece" entries={pillars} />
      <Entries id="principles" title="Criterios de la plataforma" entries={principles} />
    </Stack>
  )
}
