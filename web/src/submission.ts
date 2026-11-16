import type { components } from './api/schema'
import { sizeText, typeName } from './documentFormats'

type FieldRequirement = components['schemas']['FieldRequirement']
type DocumentRequirement = components['schemas']['DocumentRequirement']
type SubmissionProblem = components['schemas']['SubmissionProblem']

// SubmissionValues holds what the applicant fills in before submitting: the
// text of each field, or its items for a list of subjects, and the file of
// each document. Nothing reaches the server until the submission is sent.
export type SubmissionValues = {
  values: Record<string, string | string[]>
  files: Record<string, File | null>
}

// The same bounds as the server, so that what passes here passes there.
const maxText = 200
const maxLongText = 5000
export const maxItems = 100
// A decimal comma is accepted, as people write it here, and sent as a point,
// the only separator the server takes.
const number = /^-?[0-9]+([.,][0-9]+)?$/
const date = /^\d{4}-\d{2}-\d{2}$/
// maxRequestBytes is the most a request can carry (api.MaxRequestBytes in the
// backend), whatever each document allows.
export const maxRequestBytes = 25 * 1024 * 1024

export function initialValues(fields: FieldRequirement[], documents: DocumentRequirement[]): SubmissionValues {
  return {
    values: Object.fromEntries(fields.map((field) => [field.id, field.type === 'subject_list' ? [] : ''])),
    files: Object.fromEntries(documents.map((document) => [document.id, null])),
  }
}

export function isEmpty(value: string | string[] | undefined) {
  return Array.isArray(value) ? value.length === 0 : (value ?? '').trim() === ''
}

function validDate(text: string) {
  if (!date.test(text)) {
    return false
  }
  const parsed = new Date(`${text}T00:00:00Z`)
  return !Number.isNaN(parsed.getTime()) && parsed.toISOString().startsWith(text)
}

// fieldError returns the message for a value that the server would reject,
// or null. A field left empty is only wrong if it is mandatory.
export function fieldError(field: FieldRequirement, value: string | string[] | undefined): string | null {
  if (isEmpty(value)) {
    return field.required ? 'Completá este dato.' : null
  }
  if (Array.isArray(value)) {
    if (value.length > maxItems) {
      return `Cargá hasta ${maxItems} asignaturas.`
    }
    if (value.some((item) => item.trim() === '' || item.trim().length > maxText)) {
      return `Cada asignatura tiene que tener entre 1 y ${maxText} caracteres.`
    }
    return null
  }
  const text = (value ?? '').trim()
  switch (field.type) {
    case 'long_text':
      return text.length > maxLongText ? `Escribí hasta ${maxLongText} caracteres.` : null
    case 'number':
      return number.test(text) ? null : 'Escribí un número, con coma o punto para los decimales.'
    case 'date':
      return validDate(text) ? null : 'Elegí una fecha válida.'
    case 'option':
      return field.options.some((option) => option.id === text) ? null : 'Elegí una de las opciones.'
    default:
      return text.length > maxText ? `Escribí hasta ${maxText} caracteres.` : null
  }
}

// fileError checks what the browser can know of a file. The server reads the
// type from the content, so a file that passes here can still be rejected.
export function fileError(document: DocumentRequirement, file: File | null | undefined): string | null {
  if (!file) {
    return document.required ? 'Adjuntá este documento.' : null
  }
  if (file.size === 0) {
    return 'El archivo está vacío.'
  }
  if (document.max_size > 0 && file.size > document.max_size) {
    return `El archivo supera el tamaño máximo de ${sizeText(document.max_size)}.`
  }
  if (file.type !== '' && document.accepted_types.length > 0 && !document.accepted_types.includes(file.type)) {
    return `El archivo tiene que ser ${document.accepted_types.map(typeName).join(' o ')}.`
  }
  return null
}

// totalErrors marks the first file with which the files of the submission
// add up to more than maxRequestBytes, even if each one fits its document.
export function totalErrors(documents: DocumentRequirement[], files: Record<string, File | null>) {
  let total = 0
  for (const document of documents) {
    total += files[document.id]?.size ?? 0
    if (total > maxRequestBytes) {
      return {
        [`files.${document.id}`]: `Los archivos suman más de ${sizeText(maxRequestBytes)}, lo máximo que se puede enviar. Achicá alguno, por ejemplo escaneándolo con menos resolución.`,
      }
    }
  }
  return {}
}

// valueText shows a value as the applicant chose it: an option by its name.
export function valueText(field: FieldRequirement, value: string | string[] | undefined) {
  if (isEmpty(value)) {
    return 'Sin completar'
  }
  if (Array.isArray(value)) {
    return value.map((item) => item.trim()).join(', ')
  }
  const text = (value ?? '').trim()
  if (field.type === 'option') {
    return field.options.find((option) => option.id === text)?.name ?? text
  }
  return text
}

// formData puts the parts in the order the server reads them: the procedure
// type, the values and then each document followed by its file. A field left
// empty is not sent, and a number goes with a decimal point.
export function formData(procedureType: number, fields: FieldRequirement[], submission: SubmissionValues) {
  const values = Object.fromEntries(
    fields
      .filter((field) => !isEmpty(submission.values[field.id]))
      .map((field) => {
        const value = submission.values[field.id]
        return [field.id, field.type === 'number' && typeof value === 'string' ? value.trim().replace(',', '.') : value]
      }),
  )
  const body = new FormData()
  body.append('procedure_type', String(procedureType))
  body.append('values', JSON.stringify(values))
  for (const [document, file] of Object.entries(submission.files)) {
    if (file) {
      body.append('document', document)
      body.append('file', file, file.name)
    }
  }
  return body
}

const problemTexts: Record<SubmissionProblem['code'], string> = {
  missing: 'Falta completarlo.',
  invalid: 'El valor no es válido. Revisalo.',
  not_asked: 'El trámite no lo pide.',
  repeated: 'Se envió más de un archivo.',
  type_not_accepted: 'El archivo no es de un formato que acepte este documento.',
  too_large: 'El archivo supera el tamaño máximo.',
  empty: 'El archivo está vacío.',
}

// problemErrors places each problem the server found on its field or
// document, with the path the form uses.
export function problemErrors(problems: SubmissionProblem[]) {
  return Object.fromEntries(
    problems.map((problem) => [
      `${problem.kind === 'field' ? 'values' : 'files'}.${problem.id}`,
      problemTexts[problem.code],
    ]),
  )
}
