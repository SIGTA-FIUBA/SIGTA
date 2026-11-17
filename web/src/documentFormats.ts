import type { components } from './api/schema'

type DocumentRequirement = components['schemas']['DocumentRequirement']

const typeNames: Record<string, string> = {
  'application/pdf': 'PDF',
  'image/jpeg': 'JPG',
  'image/png': 'PNG',
}

// typeName shows a media type as the format people know, such as PDF.
export function typeName(mediaType: string) {
  return typeNames[mediaType] ?? mediaType.split('/').pop()!.toUpperCase()
}

// sizeText shows a size in bytes in MB, or in KB below one MB, with a
// non-breaking space so that the unit never starts a line.
export function sizeText(bytes: number) {
  const megabytes = bytes / (1024 * 1024)
  if (megabytes >= 1) {
    return `${Number(megabytes.toFixed(1))}\u00a0MB`
  }
  return `${Math.ceil(bytes / 1024)}\u00a0KB`
}

// documentDescription says which formats and size a document accepts.
export function documentDescription(document: DocumentRequirement) {
  const parts = []
  if (document.accepted_types.length > 0) {
    parts.push(`Formatos: ${document.accepted_types.map(typeName).join(', ')}.`)
  }
  if (document.max_size > 0) {
    parts.push(`Tamaño máximo: ${sizeText(document.max_size)}.`)
  }
  return parts.join(' ')
}
