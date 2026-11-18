import { describe, expect, it } from 'vitest'
import { maxRequestBytes, submissionBody, totalSizeErrors } from '../src/submission'

const megabyte = 1024 * 1024

const document = (id: string) => ({
  id,
  name: id,
  required: true,
  accepted_types: ['application/pdf'],
  max_size: 20 * megabyte,
})

const file = (size: number) => new File([new Uint8Array(size)], 'archivo.pdf', { type: 'application/pdf' })

describe('totalSizeErrors', () => {
  it('marks the file with which the submission goes past the most a request carries', () => {
    const errors = totalSizeErrors([document('transcript'), document('programs')], {
      transcript: file(15 * megabyte),
      programs: file(15 * megabyte),
    })

    expect(Object.keys(errors)).toEqual(['files.programs'])
    expect(errors['files.programs']).toContain('25 MB')
  })

  it('lets files through that fit together', () => {
    expect(totalSizeErrors([document('transcript')], { transcript: file(maxRequestBytes - megabyte) })).toEqual({})
  })
})

describe('submissionBody', () => {
  it('sends a number with a decimal point and leaves empty fields out', () => {
    const fields = [
      { id: 'credits', name: 'Créditos', required: true, type: 'number' as const, options: [] },
      { id: 'comment', name: 'Comentario', required: false, type: 'text' as const, options: [] },
    ]
    const body = submissionBody(7, fields, { values: { credits: ' 8,5 ', comment: '  ' }, files: {} })

    expect(JSON.parse(String(body.get('values')))).toEqual({ credits: '8.5' })
  })
})
