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

// documentDescription says the formats and the maximum size a document
// accepts, leaving out what it does not limit.
export function documentDescription(document: Pick<DocumentRequirement, 'accepted_types' | 'max_size'>) {
  const parts = []
  if (document.accepted_types.length > 0) {
    parts.push(`Formatos: ${document.accepted_types.map(typeName).join(', ')}.`)
  }
  if (document.max_size > 0) {
    parts.push(`Tamaño máximo: ${sizeText(document.max_size)}.`)
  }
  return parts.join(' ')
}

const typesByExtension: Record<string, string> = {
  pdf: 'application/pdf',
  png: 'image/png',
  jpg: 'image/jpeg',
  jpeg: 'image/jpeg',
}

// mediaTypeOf reads the type of a file that comes without it, as in the
// history, from its name: the server ends every name in an extension of the
// type it read from the content.
export function mediaTypeOf(fileName: string) {
  return typesByExtension[fileName.split('.').pop()!.toLowerCase()] ?? ''
}

// canPreview says whether the browser shows the file by itself: an image, or
// a PDF where the browser has a viewer, which Chrome for Android does not.
export function canPreview(mediaType: string) {
  if (mediaType === 'application/pdf') {
    return navigator.pdfViewerEnabled !== false
  }
  return mediaType === 'image/png' || mediaType === 'image/jpeg'
}
