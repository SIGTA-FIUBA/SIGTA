import { describe, expect, it } from 'vitest'
import {
  addOutcome,
  addStepAfter,
  isEarlier,
  setOptions,
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

    expect(deciding.outcomes).toEqual([{ id: 'salida_1', name: 'Salida 1', next: 'review', send_back: false }])
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

  it('keeps the identifier of an option while it keeps its name', () => {
    const field = {
      id: 'level',
      name: 'Nivel',
      type: 'option',
      options: [
        { id: 'grado', name: 'Grado' },
        { id: 'posgrado', name: 'Posgrado' },
      ],
    }

    const { options } = setOptions(field, ['Grado', 'Posgrado', 'Doctorado'])

    expect(options).toEqual([
      { id: 'grado', name: 'Grado' },
      { id: 'posgrado', name: 'Posgrado' },
      { id: 'opcion_3', name: 'Doctorado' },
    ])
    expect(setOptions(field, ['Posgrado']).options).toEqual([{ id: 'posgrado', name: 'Posgrado' }])
  })

  it('leads a new outcome to another step, never to the close', () => {
    const review = definition().steps[1]

    expect(addOutcome(definition(), review).outcomes?.at(-1)).toMatchObject({ next: 'resolution', send_back: false })
  })

  it('tells an earlier step by the route, not by the order of the steps', () => {
    // In the definition a, b, c, d; on the route a, c, b, d: b can send back to c.
    const route: Definition = {
      name: 'Ruta',
      steps: [
        { id: 'a', name: 'A', action: 'confirm_external_fact', next: 'c' },
        {
          id: 'b',
          name: 'B',
          action: 'decide',
          outcomes: [
            { id: 'on', name: 'Seguir', next: 'd', send_back: false },
            { id: 'back', name: 'Volver', next: 'c', send_back: true },
          ],
        },
        { id: 'c', name: 'C', action: 'confirm_external_fact', next: 'b' },
        { id: 'd', name: 'D', action: 'confirm_external_fact' },
      ],
    }

    expect(isEarlier(route, 'b', 'c')).toBe(true)
    expect(isEarlier(route, 'c', 'b')).toBe(false)
    expect(removeStep(route, 'd').steps.find((step) => step.id === 'b')?.outcomes?.[1].send_back).toBe(true)
  })
})
