import type { components } from './api/schema'

// The changes that the editor of procedure types makes to a definition. Each
// returns a new definition and leaves the one it gets as it was. The server
// validates the result; these only keep the pieces consistent with each
// other, such as a removed step not being named anymore.

export type Definition = components['schemas']['ProcedureTypeDefinition']
export type DefinitionStep = components['schemas']['DefinitionStep']
export type DefinitionField = components['schemas']['DefinitionField']
export type DefinitionDocument = components['schemas']['DefinitionDocument']

export const actions = ['enter_data', 'upload_document', 'confirm_external_fact', 'decide'] as const
export type StepAction = (typeof actions)[number]

// newId is the first `prefix_n` that no element of the same kind takes. The
// server asks for lowercase letters, digits and underscores, from a letter.
function newId(prefix: string, taken: string[]) {
  for (let n = taken.length + 1; ; n++) {
    const id = `${prefix}_${n}`
    if (!taken.includes(id)) {
      return id
    }
  }
}

export function stepOf(definition: Definition, stepId: string) {
  return definition.steps.find((step) => step.id === stepId)
}

export function updateStep(definition: Definition, stepId: string, change: Partial<DefinitionStep>): Definition {
  return { ...definition, steps: definition.steps.map((step) => (step.id === stepId ? { ...step, ...change } : step)) }
}

// addStepAfter puts a new step of the staff after stepId, where stepId went
// on, and makes stepId go on to it. After a decision the new step goes
// nowhere: the decision chooses through its outcomes.
export function addStepAfter(definition: Definition, stepId: string): { definition: Definition; id: string } {
  const id = newId(
    'paso',
    definition.steps.map((step) => step.id),
  )
  const index = definition.steps.findIndex((step) => step.id === stepId)
  const before = definition.steps[index]
  const decides = before?.action === 'decide'
  const added: DefinitionStep = {
    id,
    name: 'Paso nuevo',
    role: 'staff',
    action: 'confirm_external_fact',
    permissions: [],
    next: decides ? undefined : before?.next,
  }
  const steps = [...definition.steps]
  if (before && !decides) {
    steps[index] = { ...before, next: id }
  }
  steps.splice(index + 1, 0, added)
  return { definition: { ...definition, steps }, id }
}

// removeStep takes a step out. What led to it leads now where it went: the
// next step of a step that does not decide, or else the first outcome of the
// decision. A send back that lands on another step stops being one if that
// step is not earlier anymore.
export function removeStep(definition: Definition, stepId: string): Definition {
  const removed = stepOf(definition, stepId)
  if (!removed) {
    return definition
  }
  const after = (removed.action === 'decide' ? removed.outcomes?.[0]?.next : removed.next) ?? ''
  const steps = definition.steps
    .filter((step) => step.id !== stepId)
    .map((step) => {
      const goesTo = (target: string | undefined) => (target === stepId ? (after === step.id ? '' : after) : target)
      return {
        ...step,
        next: step.next === undefined ? undefined : goesTo(step.next) || undefined,
        outcomes: step.outcomes?.map((outcome) =>
          outcome.next === stepId ? { ...outcome, next: goesTo(outcome.next) ?? '' } : outcome,
        ),
      }
    })
  const kept = { ...definition, steps }
  return {
    ...kept,
    steps: steps.map((step) => ({
      ...step,
      outcomes: step.outcomes?.map((outcome) =>
        outcome.send_back && !isEarlier(kept, step.id, outcome.next) ? { ...outcome, send_back: false } : outcome,
      ),
    })),
  }
}

// isEarlier tells whether target comes before stepId in the definition,
// which is where a send back can lead.
export function isEarlier(definition: Definition, stepId: string, target: string) {
  const index = definition.steps.findIndex((step) => step.id === stepId)
  const targetIndex = definition.steps.findIndex((step) => step.id === target)
  return targetIndex >= 0 && targetIndex < index
}

// setRole changes who does a step. A step of the applicant has no area nor
// permissions, and only enters data or uploads documents.
export function setRole(step: DefinitionStep, role: 'applicant' | 'staff'): Partial<DefinitionStep> {
  if (role === 'staff') {
    return { role, permissions: step.permissions ?? [] }
  }
  const action = step.action === 'upload_document' ? 'upload_document' : 'enter_data'
  return { role, area: undefined, permissions: undefined, ...setAction(step, action) }
}

// setAction changes what a step does and drops what that action does not
// take: documents from a step that enters data or confirms a fact, fields
// from one that uploads documents, and the outcomes of a step that stops
// deciding, which goes on where its first outcome went. A step that starts to
// decide gets its next step as its first outcome.
export function setAction(step: DefinitionStep, action: StepAction): Partial<DefinitionStep> {
  const change: Partial<DefinitionStep> = { action }
  if (action === 'upload_document') {
    change.fields = undefined
    change.optional_fields = undefined
  }
  if (action === 'enter_data' || action === 'confirm_external_fact') {
    change.documents = undefined
    change.optional_documents = undefined
  }
  if (action === 'decide' && step.action !== 'decide') {
    change.outcomes = [{ id: 'continuar', name: 'Continuar', next: step.next ?? '', send_back: false }]
    change.next = undefined
  }
  if (action !== 'decide' && step.action === 'decide') {
    change.next = step.outcomes?.[0]?.next || undefined
    change.outcomes = undefined
  }
  return change
}

export function addOutcome(step: DefinitionStep): Partial<DefinitionStep> {
  const outcomes = step.outcomes ?? []
  const id = newId(
    'salida',
    outcomes.map((outcome) => outcome.id),
  )
  return { outcomes: [...outcomes, { id, name: `Salida ${outcomes.length + 1}`, next: '', send_back: false }] }
}

export function addField(definition: Definition): { definition: Definition; id: string } {
  const fields = definition.fields ?? []
  const id = newId(
    'dato',
    fields.map((field) => field.id),
  )
  return { definition: { ...definition, fields: [...fields, { id, name: 'Dato nuevo', type: 'text' }] }, id }
}

export function updateField(definition: Definition, fieldId: string, change: Partial<DefinitionField>): Definition {
  return {
    ...definition,
    fields: (definition.fields ?? []).map((field) => (field.id === fieldId ? { ...field, ...change } : field)),
  }
}

// removeField also takes the field out of every step that asks for it.
export function removeField(definition: Definition, fieldId: string): Definition {
  const without = (ids?: string[]) => ids?.filter((id) => id !== fieldId)
  return {
    ...definition,
    fields: (definition.fields ?? []).filter((field) => field.id !== fieldId),
    steps: definition.steps.map((step) => ({
      ...step,
      fields: without(step.fields),
      optional_fields: without(step.optional_fields),
    })),
  }
}

export function addDocument(definition: Definition): { definition: Definition; id: string } {
  const documents = definition.documents ?? []
  const id = newId(
    'documento',
    documents.map((document) => document.id),
  )
  const added: DefinitionDocument = {
    id,
    name: 'Documento nuevo',
    accepted_types: ['application/pdf'],
    max_size: 5 * 1024 * 1024,
  }
  return { definition: { ...definition, documents: [...documents, added] }, id }
}

export function updateDocument(
  definition: Definition,
  documentId: string,
  change: Partial<DefinitionDocument>,
): Definition {
  return {
    ...definition,
    documents: (definition.documents ?? []).map((document) =>
      document.id === documentId ? { ...document, ...change } : document,
    ),
  }
}

// removeDocument also takes the document out of every step that asks for it.
export function removeDocument(definition: Definition, documentId: string): Definition {
  const without = (ids?: string[]) => ids?.filter((id) => id !== documentId)
  return {
    ...definition,
    documents: (definition.documents ?? []).filter((document) => document.id !== documentId),
    steps: definition.steps.map((step) => ({
      ...step,
      documents: without(step.documents),
      optional_documents: without(step.optional_documents),
    })),
  }
}

// askedBy names the steps that ask for a field or a document.
export function askedBy(definition: Definition, requirementId: string) {
  return definition.steps
    .filter((step) =>
      [step.fields, step.optional_fields, step.documents, step.optional_documents].some((ids) =>
        ids?.includes(requirementId),
      ),
    )
    .map((step) => step.name)
}
