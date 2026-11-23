import { describe, expect, it } from 'vitest'
import { stateNames, stateText } from '../../src/content/procedures'

const area = { name: 'Secretaría' }

describe('stateText', () => {
  it('tells the applicant they have something to do', () => {
    expect(stateText('with_applicant', { area, applicant: true })).toEqual({
      text: 'Acción requerida',
      attention: true,
    })
  })

  it('names the state to someone else when it waits on the applicant', () => {
    expect(stateText('with_applicant', { applicant: false })).toEqual({
      text: stateNames.with_applicant,
      attention: false,
    })
  })

  it('names the area the procedure is in', () => {
    expect(stateText('in_office', { area, applicant: true })).toEqual({ text: 'En Secretaría', attention: false })
  })

  it('names the state when the area is unknown or the procedure closed', () => {
    expect(stateText('in_office', { applicant: true }).text).toBe(stateNames.in_office)
    expect(stateText('closed', { area, applicant: true }).text).toBe(stateNames.closed)
  })
})
