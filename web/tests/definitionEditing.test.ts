import { describe, expect, it } from 'vitest'
import {
  addStepAfter,
  askedBy,
  removeDocument,
  removeStep,
  setAction,
  setRole,
  type Definition,
} from '../src/definitionEditing'

function definition(): Definition {
  return {
    name: 'Equivalencia',
    documents: [{ id: 'transcript', name: 'Analítico' }],
    steps: [
      {
        id: 'request',
        name: 'Solicitud',
        role: 'applicant',
        action: 'upload_document',
        documents: ['transcript'],
        next: 'review',
      },
      {
        id: 'review',
        name: 'Revisión',
        role: 'staff',
        action: 'decide',
        outcomes: [
          { id: 'go', name: 'Continuar', next: 'resolution', send_back: false },
          { id: 'back', name: 'Pedir información', next: 'request', send_back: true },
        ],
      },
      { id: 'resolution', name: 'Resolución', role: 'staff', action: 'confirm_external_fact' },
    ],
  }
}

describe('definitionEditing', () => {
  it('adds a step where the previous one went on', () => {
    const { definition: added, id } = addStepAfter(definition(), 'request')

    expect(added.steps.map((step) => step.id)).toEqual(['request', id, 'review', 'resolution'])
    expect(added.steps[0].next).toBe(id)
    expect(added.steps[1]).toMatchObject({ role: 'staff', next: 'review' })
    expect(id).toMatch(/^[a-z][a-z0-9_]*$/)
  })

  it('removes a step and leads what led to it where it went', () => {
    const removed = removeStep(definition(), 'request')

    expect(removed.steps.map((step) => step.id)).toEqual(['review', 'resolution'])
    // The send back led to the removed step, now to review itself: it closes,
    // and is not a send back anymore.
    expect(removed.steps[0].outcomes?.[1]).toMatchObject({ next: '', send_back: false })
  })

  it('removes a decision and leads to where its first outcome went', () => {
    const removed = removeStep(definition(), 'review')

    expect(removed.steps[0].next).toBe('resolution')
  })

  it('turns a step into a decision and back without losing where it goes', () => {
    const request = definition().steps[0]
    const deciding = { ...request, ...setAction(request, 'decide') }

    expect(deciding.outcomes).toEqual([{ id: 'continuar', name: 'Continuar', next: 'review', send_back: false }])
    expect(deciding.next).toBeUndefined()
    expect({ ...deciding, ...setAction(deciding, 'confirm_external_fact') }).toMatchObject({
      next: 'review',
      outcomes: undefined,
      documents: undefined,
    })
  })

  it('gives a step of the applicant no area nor permissions, and an action of theirs', () => {
    const review = definition().steps[1]

    expect(setRole({ ...review, area: 'office', permissions: ['review'] }, 'applicant')).toMatchObject({
      role: 'applicant',
      area: undefined,
      permissions: undefined,
      action: 'enter_data',
    })
  })

  it('takes a removed document out of the steps that asked for it', () => {
    const removed = removeDocument(definition(), 'transcript')

    expect(removed.documents).toEqual([])
    expect(removed.steps[0].documents).toEqual([])
    expect(askedBy(definition(), 'transcript')).toEqual(['Solicitud'])
  })
})
