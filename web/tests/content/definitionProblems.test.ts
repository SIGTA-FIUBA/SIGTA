import { describe, expect, it } from 'vitest'
import { problemPlace, problemText } from '../../src/content/definitionProblems'

const names = (id: string) => ({ commission: 'Comisión Curricular', request: 'Solicitud' })[id] ?? id

describe('definitionProblems', () => {
  it('places a problem on its step, field or document, or else in the general part', () => {
    expect(problemPlace({ code: 'no_permissions', path: [{ kind: 'step', id: 'review' }], text: '' })).toEqual({
      section: 'steps',
      step: 'review',
    })
    expect(problemPlace({ code: 'unused_field', path: [{ kind: 'field', id: 'plan' }], text: '' })).toEqual({
      section: 'requirements',
      field: 'plan',
    })
    expect(problemPlace({ code: 'no_description', path: [{ kind: 'definition' }], text: '' })).toEqual({
      section: 'general',
    })
    expect(problemPlace({ code: 'no_steps', path: [{ kind: 'definition' }], text: '' })).toEqual({ section: 'steps' })
  })

  it('says in Spanish what is wrong, with the names people gave', () => {
    expect(
      problemText(
        {
          code: 'unknown_permission',
          path: [{ kind: 'step', id: 'review' }],
          value: 'sign',
          other: 'commission',
          text: '',
        },
        names,
      ),
    ).toBe('El área «Comisión Curricular» no tiene el permiso «sign».')
    expect(
      problemText({ code: 'unreachable_step', path: [{ kind: 'step', id: 'x' }], value: 'request', text: '' }, names),
    ).toBe('No se llega a este paso desde «Solicitud».')
    expect(problemText({ code: 'something_new', path: [], text: 'x' }, names)).toBe('Hay un problema en la definición.')
  })
})
