import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { components } from '../api/schema'

type CatalogEntry = components['schemas']['CatalogEntry']
type DocumentRequirement = components['schemas']['DocumentRequirement']

const typeNames: Record<string, string> = {
  'application/pdf': 'PDF',
  'image/jpeg': 'JPG',
  'image/png': 'PNG',
}

function typeName(mediaType: string) {
  return typeNames[mediaType] ?? mediaType.split('/').pop()!.toUpperCase()
}

function sizeText(bytes: number) {
  const megabytes = bytes / (1024 * 1024)
  if (megabytes >= 1) {
    return `${Number(megabytes.toFixed(1))} MB`
  }
  return `${Math.ceil(bytes / 1024)} KB`
}

function requiredText(required: boolean) {
  return required ? 'Obligatorio' : 'Opcional'
}

// A version of an older format has no accepted types or size, so each part is
// shown only when it is known.
function documentText(document: DocumentRequirement) {
  const parts = [`${document.name} (${requiredText(document.required)}).`]
  if (document.accepted_types.length > 0) {
    parts.push(`Formatos: ${document.accepted_types.map(typeName).join(', ')}.`)
  }
  if (document.max_size > 0) {
    parts.push(`Tamaño máximo: ${sizeText(document.max_size)}.`)
  }
  return parts.join(' ')
}

function Entry({ entry }: { entry: CatalogEntry }) {
  return (
    <article>
      <h2>{entry.name}</h2>
      {entry.description !== '' && <p>{entry.description}</p>}
      {entry.fields.length > 0 && (
        <section>
          <h3>Datos</h3>
          <ul>
            {entry.fields.map((field) => (
              <li key={field.id}>
                {field.name} ({requiredText(field.required)})
              </li>
            ))}
          </ul>
        </section>
      )}
      {entry.documents.length > 0 && (
        <section>
          <h3>Documentos</h3>
          <ul>
            {entry.documents.map((document) => (
              <li key={document.id}>{documentText(document)}</li>
            ))}
          </ul>
        </section>
      )}
      {entry.deadlines.length > 0 && (
        <section>
          <h3>Plazos</h3>
          <ul>
            {entry.deadlines.map((deadline, index) => (
              <li key={index}>{deadline}</li>
            ))}
          </ul>
        </section>
      )}
      {entry.admissibility.length > 0 && (
        <section>
          <h3>Condiciones de admisibilidad</h3>
          <ul>
            {entry.admissibility.map((condition, index) => (
              <li key={index}>{condition}</li>
            ))}
          </ul>
        </section>
      )}
    </article>
  )
}

export function CatalogPage() {
  const catalog = useQuery({
    queryKey: ['procedure-types'],
    queryFn: async () => {
      const { data, error } = await api.GET('/procedure-types')
      if (error !== undefined || data === undefined) {
        throw new Error('The catalog could not be loaded')
      }
      return data
    },
  })

  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <main>
      <h1>Trámites</h1>
      {catalog.isPending && <p>Cargando…</p>}
      {catalog.isError && <p role="alert">No pudimos cargar los trámites. Probá de nuevo más tarde.</p>}
      {catalog.isSuccess && entries.length === 0 && <p>Todavía no hay trámites disponibles.</p>}
      {entries.map((entry) => (
        <Entry key={entry.id} entry={entry} />
      ))}
    </main>
  )
}
