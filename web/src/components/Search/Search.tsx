import { Group, Kbd, Text } from '@mantine/core'
import { Spotlight } from '@mantine/spotlight'
import { useState } from 'react'
import { useNavigate } from 'react-router'
import { normalize, searchWords } from '../../content/search'
import { Icon, type IconName } from '../Icon/Icon'

export type SearchItem = {
  // id is unique in the whole search.
  id: string
  label: string
  // description is the second line, such as the procedure an item belongs to.
  description?: string
  // keywords match too, without being shown, such as the name of an applicant.
  keywords?: string[]
  to: string
}

export type SearchGroup = {
  name: string
  icon: IconName
  items: SearchItem[]
  // idle is how many items the group shows before anything is written; 0
  // hides the group until then.
  idle?: number
}

// The most items a group shows once something is written.
const matchLimit = 6

function matches(item: SearchItem, words: string[]) {
  const text = normalize([item.label, item.description ?? '', ...(item.keywords ?? [])].join(' '))
  return words.every((word) => text.includes(word))
}

// Search is the search of the whole platform, as a spotlight: Ctrl K (Cmd K
// on a Mac) or "/" opens it from any page, the arrows choose and Enter opens.
// It finds in what the person already has at hand (pending procedures, their
// own procedures, the catalog and the pages), grouped, and before anything is
// written it shows what comes first in each group.
export function Search({ groups }: { groups: SearchGroup[] }) {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const words = searchWords(query)
  const shown = groups
    .map((group) => ({
      ...group,
      items:
        words.length === 0
          ? group.items.slice(0, group.idle ?? matchLimit)
          : group.items.filter((item) => matches(item, words)).slice(0, matchLimit),
    }))
    .filter((group) => group.items.length > 0)

  return (
    <Spotlight.Root query={query} onQueryChange={setQuery} shortcut={['mod + K', '/']} scrollable maxHeight="60dvh">
      <Spotlight.Search
        placeholder="Buscar trámites, personas o páginas"
        aria-label="Buscar"
        size="lg"
        leftSection={<Icon name="search" />}
      />
      <Spotlight.ActionsList>
        {shown.length === 0 ? (
          <Spotlight.Empty>Sin resultados para “{query}”.</Spotlight.Empty>
        ) : (
          shown.map((group) => (
            <Spotlight.ActionsGroup key={group.name} label={group.name}>
              {group.items.map((item) => (
                <Spotlight.Action
                  key={item.id}
                  onClick={() => void navigate(item.to)}
                  leftSection={<Icon name={group.icon} />}
                  label={item.label}
                  description={item.description}
                  highlightQuery
                />
              ))}
            </Spotlight.ActionsGroup>
          ))
        )}
      </Spotlight.ActionsList>
      <Spotlight.Footer>
        <Group gap="md" wrap="nowrap">
          <Text size="xs" c="dimmed">
            <Kbd size="xs">↑</Kbd> <Kbd size="xs">↓</Kbd> elegir
          </Text>
          <Text size="xs" c="dimmed">
            <Kbd size="xs">Enter</Kbd> abrir
          </Text>
          <Text size="xs" c="dimmed">
            <Kbd size="xs">Esc</Kbd> cerrar
          </Text>
        </Group>
      </Spotlight.Footer>
    </Spotlight.Root>
  )
}
