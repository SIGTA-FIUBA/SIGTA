import { SimpleGrid, Stack, Text, Title } from '@mantine/core'
import { BrandSignature } from '../components/BrandSignature/BrandSignature'
import { PageHeader } from '../components/PageHeader/PageHeader'
import { type ContentItem, pillars, principles } from '../content/platform'

// Entries is a section of the page: its title and its entries in three
// columns, each under a hairline, with no box. There are three of each, so
// the row closes.
function Entries({ id, title, entries }: { id: string; title: string; entries: ContentItem[] }) {
  return (
    <Stack component="section" aria-labelledby={id} gap="lg">
      <Title order={2} id={id}>
        {title}
      </Title>
      <SimpleGrid component="ul" cols={{ base: 1, sm: entries.length }} spacing="xl" verticalSpacing="lg" p={0} m={0}>
        {entries.map((entry) => (
          <Stack key={entry.title} component="li" variant="part" gap="xs">
            <Text fw={600}>{entry.title}</Text>
            <Text size="sm" c="gray.7">
              {entry.text}
            </Text>
          </Stack>
        ))}
      </SimpleGrid>
    </Stack>
  )
}

// AboutPage says what SIGTA is and the criteria it follows, under the header
// of every page, and closes with the signature of SIGTA and the FIUBA. It moved here
// from the home page, which now shows what to do.
export function AboutPage() {
  return (
    <Stack gap="xxl">
      <PageHeader
        title="Acerca de SIGTA"
        description="SIGTA, el Sistema Integral de Gestión de Trámites Académicos, es la plataforma de la Secretaría de Gestión Académica para iniciar, gestionar y seguir los trámites académicos de la facultad, del inicio a la resolución."
      />
      <Entries id="pillars" title="Qué ofrece" entries={pillars} />
      <Entries id="principles" title="Criterios de la plataforma" entries={principles} />
      <BrandSignature />
    </Stack>
  )
}
