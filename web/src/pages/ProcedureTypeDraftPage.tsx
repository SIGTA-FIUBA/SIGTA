import {
  Accordion,
  Alert,
  Anchor,
  Badge,
  Button,
  Checkbox,
  Divider,
  Group,
  Loader,
  Modal,
  MultiSelect,
  NumberInput,
  Paper,
  SegmentedControl,
  Select,
  Stack,
  TagsInput,
  Text,
  Textarea,
  TextInput,
  Title,
  useMantineTheme,
} from '@mantine/core'
import { useDisclosure, useMediaQuery } from '@mantine/hooks'
import { useState, type ReactNode } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { ForbiddenError } from '../api/procedures'
import { draftActions, useDraftChange, useEditableProcedureTypes, useProcedureTypeDraft } from '../api/procedureTypes'
import type { components } from '../api/schema'
import { StepDiagram, type DiagramStep } from '../components/StepDiagram/StepDiagram'
import { problemPlace, problemText, type Names, type ProblemSection } from '../content/definitionProblems'
import { draftStatus } from '../content/draftStatuses'
import { typeName } from '../content/documentFormats'
import {
  actions,
  addDocument,
  addField,
  addOutcome,
  addStepAfter,
  askedBy,
  defaultTarget,
  isEarlier,
  removeDocument,
  removeField,
  removeStep,
  setAction,
  setOptions,
  setRole,
  stepOf,
  updateDocument,
  updateField,
  updateStep,
  type Definition,
  type DefinitionStep,
  type StepAction,
} from '../definitionEditing'

type ProcedureTypeDraft = components['schemas']['ProcedureTypeDraft']
type DefinitionProblem = components['schemas']['DefinitionProblem']
type StepArea = components['schemas']['StepArea']
type Person = components['schemas']['Person']

const actionNames: Record<StepAction, string> = {
  enter_data: 'Cargar dato',
  upload_document: 'Cargar documento',
  confirm_external_fact: 'Confirmar hecho externo',
  decide: 'Decidir',
}

const fieldTypes = [
  { value: 'text', label: 'Texto' },
  { value: 'long_text', label: 'Texto largo' },
  { value: 'number', label: 'Número' },
  { value: 'date', label: 'Fecha' },
  { value: 'option', label: 'Opciones' },
  { value: 'degree', label: 'Carrera' },
  { value: 'subject_list', label: 'Lista de asignaturas' },
]

const acceptedTypes = ['application/pdf', 'image/jpeg', 'image/png'].map((value) => ({ value, label: typeName(value) }))
const megabyte = 1024 * 1024
// A step without next closes the procedure.
const closes = ''

const dateTimeFormat = new Intl.DateTimeFormat('es-AR', { dateStyle: 'short', timeStyle: 'short' })

// ProcedureTypeDraftPage edits the draft of a procedure type, proposes it and,
// for another person, approves or rejects it (ADR-013). The route is a
// diagram of its steps; data, documents and what describes it to the
// applicant go apart.
export function ProcedureTypeDraftPage() {
  const procedureTypeId = Number(useParams().procedureTypeId)
  const draft = useProcedureTypeDraft(procedureTypeId)
  const editable = useEditableProcedureTypes()
  const [section, setSection] = useState<ProblemSection>('steps')
  const [selected, setSelected] = useState('')

  return (
    <Stack gap="md">
      <Anchor component={Link} to="/procedure-types" size="sm">
        ← Tipos de trámite
      </Anchor>
      {draft.isPending && <Loader aria-label="Cargando el borrador" />}
      {draft.isError && (
        <Alert color="red">
          {draft.error instanceof ForbiddenError
            ? 'No tenés permiso para ver este borrador.'
            : 'No pudimos cargar el borrador: puede que ya se haya publicado o descartado. Volvé a la lista de tipos de trámite.'}
        </Alert>
      )}
      {draft.data && (
        <DraftEditor
          key={draft.data.procedure_type_id}
          draft={draft.data}
          areas={editable.data?.areas ?? []}
          section={section}
          onSection={setSection}
          selected={selected}
          onSelect={setSelected}
        />
      )}
    </Stack>
  )
}

// DraftEditor keeps the changes until they are saved, also when the draft is
// read again: a revision that someone else saved meanwhile is shown as a
// conflict, with the changes still there, and never replaces them.
function DraftEditor({
  draft,
  areas,
  section,
  onSection,
  selected,
  onSelect,
}: {
  draft: ProcedureTypeDraft
  areas: StepArea[]
  section: ProblemSection
  onSection: (section: ProblemSection) => void
  selected: string
  onSelect: (stepId: string) => void
}) {
  const navigate = useNavigate()
  const theme = useMantineTheme()
  // On a phone the parts go one under the other: their names do not fit side by side.
  const phone = useMediaQuery(`(max-width: ${theme.breakpoints.sm})`)
  const id = draft.procedure_type_id
  // edit holds the definition being edited, whether it has changes to save,
  // and the revision they were made on, which saving names: if the draft
  // moved on, the server refuses them.
  const [edit, setEdit] = useState({ definition: draft.definition, dirty: false, revision: draft.revision })
  const { definition, dirty } = edit
  const proposed = draft.status === 'proposed'
  const readOnly = !draft.can_edit
  const change = (next: Definition) => setEdit((current) => ({ ...current, definition: next, dirty: true }))
  const takeLatest = (latest: ProcedureTypeDraft) =>
    setEdit({ definition: latest.definition, dirty: false, revision: latest.revision })
  // A draft read again without changes pending is shown as it came; with
  // changes, it is kept for the conflict below.
  if (!dirty && draft.revision !== edit.revision) {
    takeLatest(draft)
  }
  const conflict = dirty && draft.revision !== edit.revision

  // What was saved takes the place of what was sent before the draft is read
  // again, so the new revision is not taken for someone else's. What changed
  // while saving stays, to be saved on the new revision.
  const save = useDraftChange((sent: Definition) => draftActions.edit(id, edit.revision, sent), {
    onSuccess: (saved, sent) =>
      saved &&
      setEdit((current) =>
        current.definition === sent
          ? { definition: saved.definition, dirty: false, revision: saved.revision }
          : { ...current, revision: saved.revision },
      ),
  })
  const propose = useDraftChange(() => draftActions.propose(id, draft.revision))
  const discard = useDraftChange(() => draftActions.discard(id, draft.revision))
  const approve = useDraftChange(() => draftActions.approve(id, draft.revision))
  const reject = useDraftChange((reason: string) => draftActions.reject(id, draft.revision, reason))
  const failure = [save, propose, discard, approve, reject].find((action) => action.isError)?.error

  const names = namesOf(definition, areas)
  const problems = draft.problems
  const count = (of: ProblemSection) => problems.filter((problem) => problemPlace(problem).section === of).length
  const current = stepOf(definition, selected) ?? definition.steps[0]
  const status = draftStatus(draft)

  function goTo(problem: DefinitionProblem) {
    const place = problemPlace(problem)
    onSection(place.section)
    if (place.step) {
      onSelect(place.step)
    }
  }

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <Stack gap={0}>
          <Text size="sm" c="dimmed">
            {draft.base > 0
              ? `Borrador de la versión ${draft.base + 1}, sobre la ${draft.base}`
              : 'Tipo de trámite nuevo'}
          </Text>
          <Group gap="sm">
            <Title order={1}>{definition.name || 'Tipo de trámite sin nombre'}</Title>
            <Badge variant="light" color={status.color}>
              {status.label}
            </Badge>
          </Group>
        </Stack>
        {!readOnly && (
          <Group gap="sm">
            <ConfirmAction
              label="Descartar"
              title="¿Descartar el borrador?"
              text="Se pierden todos sus cambios. La versión publicada, si la hay, no cambia."
              confirmLabel="Descartar el borrador"
              pending={discard.isPending}
              onConfirm={(done) =>
                discard.mutate(undefined, { onSettled: done, onSuccess: () => void navigate('/procedure-types') })
              }
            />
            <Button
              variant="default"
              disabled={!dirty || conflict}
              loading={save.isPending}
              onClick={() => save.mutate(definition)}
            >
              Guardar
            </Button>
            <ConfirmAction
              label="Proponer"
              primary
              disabled={dirty || problems.length > 0}
              title="¿Proponer el borrador?"
              text="Mientras esté propuesto no se puede editar. Lo aprueba o lo rechaza otra persona con permiso de aprobar; quien lo editó o lo propuso no puede."
              confirmLabel="Proponer"
              pending={propose.isPending}
              onConfirm={(done) => propose.mutate(undefined, { onSettled: done })}
            />
          </Group>
        )}
      </Group>

      {failure && !conflict && <Alert color="red">{failure.message}</Alert>}
      {conflict && (
        <Alert color="orange" title="Otra persona guardó este borrador mientras lo editabas">
          <Stack gap="sm">
            <Text size="sm">
              Tus cambios siguen acá, pero no se pueden guardar sobre la versión anterior. Anotá lo que necesites y
              cargá la última versión para seguir desde ella.
            </Text>
            <Group>
              <Button variant="default" onClick={() => takeLatest(draft)}>
                Descartar mis cambios y cargar la última versión
              </Button>
            </Group>
          </Stack>
        </Alert>
      )}
      <StatusNotice
        draft={draft}
        approvePending={approve.isPending}
        rejectPending={reject.isPending}
        onApprove={(done) =>
          approve.mutate(undefined, {
            onSettled: done,
            onSuccess: (version) => void navigate(`/procedure-types?publicada=${version?.number ?? ''}`),
          })
        }
        onReject={(reason, done) => reject.mutate(reason, { onSettled: done })}
      />
      {!readOnly && dirty && !conflict && (
        <Text size="sm" c="dimmed">
          Hay cambios sin guardar. Los problemas que se marcan son los de la última vez que se guardó.
        </Text>
      )}
      {!readOnly && !dirty && problems.length === 0 && (
        <Text size="sm" c="dimmed">
          La definición no tiene problemas con las áreas actuales: se puede proponer.
        </Text>
      )}
      {problems.length > 0 && (
        <Stack gap="xs">
          <Text size="sm" fw={600} c="red.8">
            {proposed
              ? `${problems.length === 1 ? 'La propuesta tiene un problema' : `La propuesta tiene ${problems.length} problemas`} con las áreas de hoy: así no se puede publicar.`
              : `${problems.length === 1 ? 'Hay un problema' : `Hay ${problems.length} problemas`} antes de proponer:`}
          </Text>
          <Stack gap="xs" align="flex-start">
            {problems.map((problem, index) => (
              <Anchor
                key={index}
                component="button"
                type="button"
                size="sm"
                c="red.8"
                ta="left"
                onClick={() => goTo(problem)}
              >
                {problemLabel(problem, definition, names)}
              </Anchor>
            ))}
          </Stack>
        </Stack>
      )}

      <SegmentedControl
        orientation={phone ? 'vertical' : 'horizontal'}
        fullWidth={phone}
        value={section}
        onChange={(value) => onSection(value as ProblemSection)}
        data={[
          { value: 'steps', label: withCount('Recorrido', count('steps')) },
          { value: 'requirements', label: withCount('Datos y documentos', count('requirements')) },
          { value: 'general', label: withCount('General', count('general')) },
        ]}
        aria-label="Partes del tipo de trámite"
      />

      {section === 'steps' && (
        <StepDiagram
          steps={definition.steps.map((step) => diagramStep(step, names, problems))}
          selected={current?.id ?? ''}
          onSelect={onSelect}
          toolbar={
            !readOnly &&
            current && (
              <Button
                variant="default"
                onClick={() => {
                  const added = addStepAfter(definition, current.id)
                  change(added.definition)
                  onSelect(added.id)
                }}
              >
                + Paso después de «{current.name}»
              </Button>
            )
          }
          panel={
            current ? (
              <StepPanel
                definition={definition}
                step={current}
                areas={areas}
                readOnly={readOnly}
                problems={problems.filter((problem) => problemPlace(problem).step === current.id)}
                names={names}
                onChange={change}
                onRemoved={() => onSelect('')}
              />
            ) : (
              <Stack>
                <Text>Todavía no hay pasos. El trámite empieza en el primero.</Text>
                {!readOnly && (
                  <Button
                    onClick={() => {
                      const first: DefinitionStep = {
                        id: 'paso_1',
                        name: 'Paso nuevo',
                        role: 'applicant',
                        action: 'enter_data',
                      }
                      change({ ...definition, steps: [first] })
                      onSelect(first.id)
                    }}
                  >
                    Agregar el primer paso
                  </Button>
                )}
              </Stack>
            )
          }
        />
      )}
      {section === 'requirements' && (
        <RequirementsSection
          definition={definition}
          readOnly={readOnly}
          problems={problems}
          names={names}
          onChange={change}
        />
      )}
      {section === 'general' && (
        <GeneralSection
          definition={definition}
          readOnly={readOnly}
          problems={problems}
          names={names}
          onChange={change}
        />
      )}

      <Text size="sm" c="dimmed">
        Lo editaron: {(draft.editors.length > 0 ? draft.editors : [draft.edited_by]).map(personName).join(', ')}. Última
        edición: {dateTimeFormat.format(new Date(draft.edited_at))}
      </Text>
    </Stack>
  )
}

// personName shows who someone is; the API has no name for a person who
// never entered SIGTA.
function personName(person: Person) {
  return person.name || 'una persona sin registrar'
}

function withCount(label: string, problems: number) {
  return problems > 0 ? `${label} · ${problems}` : label
}

// namesOf finds the name of a step, field, document or area by its
// identifier, for the texts of the problems.
function namesOf(definition: Definition, areas: StepArea[]): Names {
  const named = new Map<string, string>()
  for (const item of [...definition.steps, ...(definition.fields ?? []), ...(definition.documents ?? []), ...areas]) {
    named.set(item.id, item.name)
  }
  return (id) => named.get(id) || id
}

// problemLabel says what a problem is about and what is wrong with it. A
// step, a field and a document can share an identifier, so the name is found
// among the elements of the kind the problem is about.
function problemLabel(problem: DefinitionProblem, definition: Definition, names: Names) {
  const about = problem.path.findLast((element) => element.id && ['step', 'field', 'document'].includes(element.kind))
  const text = problemText(problem, names)
  if (!about?.id) {
    return text
  }
  const elements = { step: definition.steps, field: definition.fields ?? [], document: definition.documents ?? [] }[
    about.kind as 'step' | 'field' | 'document'
  ]
  return `«${elements.find((element) => element.id === about.id)?.name || about.id}»: ${text}`
}

function diagramStep(step: DefinitionStep, names: Names, problems: DefinitionProblem[]): DiagramStep {
  const action = (step.action as StepAction) in actionNames ? actionNames[step.action as StepAction] : step.action
  const asks = [step.fields, step.optional_fields, step.documents, step.optional_documents].flatMap((ids) => ids ?? [])
  const detail =
    step.action === 'decide'
      ? `${action} · ${step.outcomes?.length ?? 0} salidas`
      : [
          action,
          asks.length > 0 && `${asks.length} ${asks.length === 1 ? 'requisito' : 'requisitos'}`,
          !step.next && 'cierra',
        ]
          .filter(Boolean)
          .join(' · ')
  const problem = problems.some((p) => problemPlace(p).step === step.id)
  return {
    id: step.id,
    next: step.next,
    outcomes: step.action === 'decide' ? step.outcomes : undefined,
    name: step.name || 'Paso sin nombre',
    applicant: step.role === 'applicant',
    problem,
    kicker: problem
      ? 'Con problemas'
      : step.role === 'applicant'
        ? 'Titular'
        : step.area
          ? names(step.area)
          : 'Sin área',
    detail,
  }
}

// StepPanel edits the selected step: who does it and what, what it asks
// for, folded, and where it goes on.
function StepPanel({
  definition,
  step,
  areas,
  readOnly,
  problems,
  names,
  onChange,
  onRemoved,
}: {
  definition: Definition
  step: DefinitionStep
  areas: StepArea[]
  readOnly: boolean
  problems: DefinitionProblem[]
  names: Names
  onChange: (definition: Definition) => void
  onRemoved: () => void
}) {
  const index = definition.steps.findIndex((s) => s.id === step.id)
  const set = (change: Partial<DefinitionStep>) => onChange(updateStep(definition, step.id, change))
  const staff = step.role !== 'applicant'
  const area = areas.find((a) => a.id === step.area)
  const action = step.action as StepAction
  const takesFields = action !== 'upload_document'
  const takesDocuments = action === 'upload_document' || action === 'decide'
  const fieldOptions = (definition.fields ?? []).map((field) => ({ value: field.id, label: field.name || field.id }))
  const documentOptions = (definition.documents ?? []).map((document) => ({
    value: document.id,
    label: document.name || document.id,
  }))
  // An outcome leads to another step; only a step without next closes the
  // procedure.
  const steps = definition.steps
    .filter((s) => s.id !== step.id)
    .map((s) => ({ value: s.id, label: s.name || 'Paso sin nombre' }))
  const asks = [step.fields, step.optional_fields, step.documents, step.optional_documents]
    .flatMap((ids) => ids ?? [])
    .map(names)

  return (
    <Stack gap="md">
      <Text size="sm" c="dimmed" fw={600}>
        Paso {index + 1}
        {index === 0 && ' · el trámite empieza acá'}
      </Text>
      {problems.length > 0 && (
        <Alert color="red" p="sm">
          <Stack gap="xs">
            {problems.map((problem, i) => (
              <Text key={i} size="sm">
                {problemText(problem, names)}
              </Text>
            ))}
          </Stack>
        </Alert>
      )}
      <TextInput
        label="Nombre"
        value={step.name}
        disabled={readOnly}
        onChange={(event) => set({ name: event.currentTarget.value })}
      />
      <Stack gap="xs">
        <Text size="sm" fw={600}>
          Quién lo hace
        </Text>
        <SegmentedControl
          value={staff ? 'staff' : 'applicant'}
          disabled={readOnly}
          onChange={(value) => set(setRole(step, value as 'applicant' | 'staff'))}
          data={[
            { value: 'applicant', label: 'Titular' },
            { value: 'staff', label: 'Personal' },
          ]}
          aria-label="Quién hace el paso"
        />
      </Stack>
      {staff && (
        <>
          <Select
            label="Área"
            placeholder="Elegí un área"
            data={areas.map((a) => ({ value: a.id, label: a.name }))}
            value={step.area ?? null}
            disabled={readOnly}
            onChange={(value) => set({ area: value ?? undefined, permissions: [] })}
          />
          <MultiSelect
            label="Permisos que exige"
            description="Los del área elegida. Para hacer el paso hay que tenerlos todos."
            data={(area?.permissions ?? []).map((p) => ({ value: p.id, label: p.name }))}
            value={step.permissions ?? []}
            disabled={readOnly || !area}
            onChange={(permissions) => set({ permissions })}
          />
        </>
      )}
      <Select
        label="Acción"
        data={actions
          .filter((a) => staff || a === 'enter_data' || a === 'upload_document')
          .map((a) => ({ value: a, label: actionNames[a] }))}
        value={step.action}
        allowDeselect={false}
        disabled={readOnly}
        onChange={(value) => value && set(setAction(step, value as StepAction, defaultTarget(definition, step.id)))}
      />
      <Accordion variant="contained" chevronPosition="right">
        <Accordion.Item value="asks">
          <Accordion.Control>
            <Text size="sm" fw={600}>
              Pide
            </Text>
            <Text size="xs" c="dimmed" lineClamp={1}>
              {asks.length > 0 ? asks.join(', ') : 'Nada todavía'}
            </Text>
          </Accordion.Control>
          <Accordion.Panel>
            <Stack gap="sm">
              <Text size="xs" c="dimmed">
                Arma el formulario del titular, y el paso no se confirma sin sus documentos obligatorios.
              </Text>
              {takesFields && (
                <>
                  <MultiSelect
                    label="Datos obligatorios"
                    data={fieldOptions}
                    value={step.fields ?? []}
                    disabled={readOnly}
                    onChange={(fields) => set({ fields })}
                  />
                  <MultiSelect
                    label="Datos opcionales"
                    data={fieldOptions}
                    value={step.optional_fields ?? []}
                    disabled={readOnly}
                    onChange={(optional_fields) => set({ optional_fields })}
                  />
                </>
              )}
              {takesDocuments && (
                <>
                  <MultiSelect
                    label="Documentos obligatorios"
                    data={documentOptions}
                    value={step.documents ?? []}
                    disabled={readOnly}
                    onChange={(documents) => set({ documents })}
                  />
                  <MultiSelect
                    label="Documentos opcionales"
                    data={documentOptions}
                    value={step.optional_documents ?? []}
                    disabled={readOnly}
                    onChange={(optional_documents) => set({ optional_documents })}
                  />
                </>
              )}
              {fieldOptions.length + documentOptions.length === 0 && (
                <Text size="sm">Los datos y documentos se crean en «Datos y documentos».</Text>
              )}
            </Stack>
          </Accordion.Panel>
        </Accordion.Item>
      </Accordion>
      <Divider />
      {action === 'decide' ? (
        <Stack gap="sm">
          <Text size="sm" fw={600}>
            Salidas
          </Text>
          {(step.outcomes ?? []).map((outcome, i) => {
            const setOutcome = (patch: Partial<typeof outcome>) =>
              set({ outcomes: (step.outcomes ?? []).map((o, j) => (j === i ? { ...o, ...patch } : o)) })
            return (
              <Paper key={outcome.id} withBorder p="sm" bg={outcome.send_back ? 'orange.0' : 'gray.0'}>
                <Stack gap="xs">
                  <TextInput
                    label="Nombre de la salida"
                    value={outcome.name}
                    disabled={readOnly}
                    onChange={(event) => setOutcome({ name: event.currentTarget.value })}
                  />
                  <Select
                    label="Lleva a"
                    data={steps}
                    value={outcome.next || null}
                    placeholder="Elegí un paso"
                    allowDeselect={false}
                    disabled={readOnly}
                    onChange={(next) =>
                      next && setOutcome({ next, send_back: outcome.send_back && isEarlier(definition, step.id, next) })
                    }
                  />
                  {isEarlier(definition, step.id, outcome.next) && (
                    <Checkbox
                      label="Es una devolución: pide el motivo"
                      checked={outcome.send_back}
                      disabled={readOnly}
                      onChange={(event) => setOutcome({ send_back: event.currentTarget.checked })}
                    />
                  )}
                  {!readOnly && (
                    <Button
                      variant="subtle"
                      size="compact-sm"
                      color="gray"
                      onClick={() => set({ outcomes: (step.outcomes ?? []).filter((_, j) => j !== i) })}
                    >
                      Quitar la salida
                    </Button>
                  )}
                </Stack>
              </Paper>
            )
          })}
          {!readOnly && (
            <Button variant="light" onClick={() => set(addOutcome(definition, step))}>
              + Salida
            </Button>
          )}
        </Stack>
      ) : (
        <Select
          label="Sigue en"
          data={[...steps, { value: closes, label: 'Cierra el trámite' }]}
          value={step.next ?? closes}
          allowDeselect={false}
          disabled={readOnly}
          onChange={(next) => set({ next: next || undefined })}
        />
      )}
      {!readOnly && (
        <ConfirmAction
          label="Quitar este paso"
          subtle
          disabled={definition.steps.length <= 1}
          title={`¿Quitar «${step.name || 'Paso sin nombre'}»?`}
          text={`Lo que llevaba a este paso pasa a llevar a «${
            names((action === 'decide' ? step.outcomes?.[0]?.next : step.next) ?? '') || 'Cierre'
          }». Se aplica al guardar.`}
          confirmLabel="Quitar el paso"
          onConfirm={(done) => {
            onChange(removeStep(definition, step.id))
            onRemoved()
            done()
          }}
        />
      )}
    </Stack>
  )
}

// RequirementsSection edits the data and documents a procedure type can ask
// for. Which step asks for each one is chosen in the route.
function RequirementsSection({
  definition,
  readOnly,
  problems,
  names,
  onChange,
}: {
  definition: Definition
  readOnly: boolean
  problems: DefinitionProblem[]
  names: Names
  onChange: (definition: Definition) => void
}) {
  const problemsOf = (key: 'field' | 'document', id: string) =>
    problems.filter((problem) => problemPlace(problem)[key] === id).map((problem) => problemText(problem, names))
  const asked = (id: string) => {
    const steps = askedBy(definition, id)
    return steps.length > 0 ? `Lo pide: ${steps.join(', ')}` : 'Ningún paso lo pide: se elige en el recorrido.'
  }

  return (
    <Stack gap="lg">
      <Stack gap="sm">
        <Title order={2} size="h4">
          Datos
        </Title>
        {(definition.fields ?? []).length === 0 && <Text c="dimmed">Todavía no hay datos.</Text>}
        {(definition.fields ?? []).map((field) => (
          <RequirementCard key={field.id} problems={problemsOf('field', field.id)} asked={asked(field.id)}>
            <Group align="flex-end" wrap="wrap">
              <TextInput
                label="Nombre"
                value={field.name}
                disabled={readOnly}
                onChange={(event) => onChange(updateField(definition, field.id, { name: event.currentTarget.value }))}
              />
              <Select
                label="Tipo"
                data={fieldTypes}
                value={field.type}
                allowDeselect={false}
                disabled={readOnly}
                onChange={(type) =>
                  type &&
                  onChange(
                    updateField(definition, field.id, {
                      type,
                      options: type === 'option' ? (field.options ?? []) : undefined,
                    }),
                  )
                }
              />
              {!readOnly && (
                <Button variant="subtle" color="gray" onClick={() => onChange(removeField(definition, field.id))}>
                  Quitar
                </Button>
              )}
            </Group>
            {field.type === 'option' && (
              <TagsInput
                label="Opciones"
                description="Escribí cada opción y apretá Enter."
                value={(field.options ?? []).map((option) => option.name)}
                disabled={readOnly}
                onChange={(options) => onChange(updateField(definition, field.id, setOptions(field, options)))}
              />
            )}
          </RequirementCard>
        ))}
        {!readOnly && (
          <Button variant="light" onClick={() => onChange(addField(definition).definition)}>
            + Dato
          </Button>
        )}
      </Stack>
      <Stack gap="sm">
        <Title order={2} size="h4">
          Documentos
        </Title>
        {(definition.documents ?? []).length === 0 && <Text c="dimmed">Todavía no hay documentos.</Text>}
        {(definition.documents ?? []).map((document) => (
          <RequirementCard key={document.id} problems={problemsOf('document', document.id)} asked={asked(document.id)}>
            <Group align="flex-end" wrap="wrap">
              <TextInput
                label="Nombre"
                value={document.name}
                disabled={readOnly}
                onChange={(event) =>
                  onChange(updateDocument(definition, document.id, { name: event.currentTarget.value }))
                }
              />
              <MultiSelect
                label="Formatos"
                data={acceptedTypes}
                value={document.accepted_types ?? []}
                disabled={readOnly}
                onChange={(accepted_types) => onChange(updateDocument(definition, document.id, { accepted_types }))}
              />
              <NumberInput
                label="Tamaño máximo (MB)"
                min={1}
                value={document.max_size ? document.max_size / megabyte : ''}
                disabled={readOnly}
                onChange={(value) =>
                  onChange(updateDocument(definition, document.id, { max_size: Math.round(Number(value) * megabyte) }))
                }
              />
              {!readOnly && (
                <Button variant="subtle" color="gray" onClick={() => onChange(removeDocument(definition, document.id))}>
                  Quitar
                </Button>
              )}
            </Group>
            <Checkbox
              label="Es la resolución del trámite: la descarga el titular cuando se cierra"
              checked={document.resolution ?? false}
              disabled={readOnly}
              onChange={(event) =>
                onChange(updateDocument(definition, document.id, { resolution: event.currentTarget.checked }))
              }
            />
          </RequirementCard>
        ))}
        {!readOnly && (
          <Button variant="light" onClick={() => onChange(addDocument(definition).definition)}>
            + Documento
          </Button>
        )}
      </Stack>
    </Stack>
  )
}

function RequirementCard({ problems, asked, children }: { problems: string[]; asked: string; children: ReactNode }) {
  return (
    <Paper withBorder shadow="xs" p="md" bg={problems.length > 0 ? 'red.0' : undefined}>
      <Stack gap="sm">
        {children}
        <Text size="sm" c="dimmed">
          {asked}
        </Text>
        {problems.map((problem, i) => (
          <Text key={i} size="sm" c="red.8">
            {problem}
          </Text>
        ))}
      </Stack>
    </Paper>
  )
}

// GeneralSection edits what describes the procedure type to the applicant
// in the catalog: its name, description, deadlines and conditions.
function GeneralSection({
  definition,
  readOnly,
  problems,
  names,
  onChange,
}: {
  definition: Definition
  readOnly: boolean
  problems: DefinitionProblem[]
  names: Names
  onChange: (definition: Definition) => void
}) {
  const general = problems.filter((problem) => problemPlace(problem).section === 'general')
  const errorOf = (kind: string, position?: number) =>
    general
      .filter((problem) => {
        const last = problem.path.at(-1)
        return (last?.kind ?? 'definition') === kind && (position === undefined || last?.position === position)
      })
      .map((problem) => problemText(problem, names))
      .join(' ') || undefined

  return (
    <Paper withBorder shadow="xs" p="lg">
      <Stack gap="md">
        <TextInput
          label="Nombre"
          value={definition.name}
          disabled={readOnly}
          error={
            general.some((p) => ['no_name', 'name_with_spaces'].includes(p.code)) ? errorOf('definition') : undefined
          }
          onChange={(event) => onChange({ ...definition, name: event.currentTarget.value })}
        />
        <Textarea
          label="Descripción"
          description="La muestra el catálogo al titular, antes de iniciar el trámite."
          autosize
          minRows={3}
          value={definition.description ?? ''}
          disabled={readOnly}
          error={
            general.some((p) => p.code === 'no_description') ? 'Falta la descripción del tipo de trámite.' : undefined
          }
          onChange={(event) => onChange({ ...definition, description: event.currentTarget.value })}
        />
        <TextList
          label="Plazos"
          addLabel="+ Plazo"
          items={definition.deadlines ?? []}
          readOnly={readOnly}
          errorAt={(position) => errorOf('deadline', position)}
          onChange={(deadlines) => onChange({ ...definition, deadlines })}
        />
        <TextList
          label="Condiciones de admisibilidad"
          addLabel="+ Condición"
          items={definition.admissibility ?? []}
          readOnly={readOnly}
          errorAt={(position) => errorOf('admissibility', position)}
          onChange={(admissibility) => onChange({ ...definition, admissibility })}
        />
      </Stack>
    </Paper>
  )
}

function TextList({
  label,
  addLabel,
  items,
  readOnly,
  errorAt,
  onChange,
}: {
  label: string
  addLabel: string
  items: string[]
  readOnly: boolean
  errorAt: (position: number) => string | undefined
  onChange: (items: string[]) => void
}) {
  return (
    <Stack gap="xs">
      <Text size="sm" fw={600}>
        {label}
      </Text>
      {items.length === 0 && (
        <Text size="sm" c="dimmed">
          Ninguno.
        </Text>
      )}
      {items.map((item, i) => (
        <Group key={i} align="flex-start" wrap="nowrap">
          <TextInput
            flex={1}
            aria-label={`${label} ${i + 1}`}
            value={item}
            disabled={readOnly}
            error={errorAt(i + 1)}
            onChange={(event) => onChange(items.map((text, j) => (j === i ? event.currentTarget.value : text)))}
          />
          {!readOnly && (
            <Button variant="subtle" color="gray" onClick={() => onChange(items.filter((_, j) => j !== i))}>
              Quitar
            </Button>
          )}
        </Group>
      ))}
      {!readOnly && (
        <Button variant="light" onClick={() => onChange([...items, ''])}>
          {addLabel}
        </Button>
      )}
    </Stack>
  )
}

// StatusNotice says where the draft stands for whoever looks at it, and lets
// another person with the permission approve or reject a proposed one.
function StatusNotice({
  draft,
  approvePending,
  rejectPending,
  onApprove,
  onReject,
}: {
  draft: ProcedureTypeDraft
  approvePending: boolean
  rejectPending: boolean
  onApprove: (done: () => void) => void
  onReject: (reason: string, done: () => void) => void
}) {
  const proposed = draft.status === 'proposed'
  if (!proposed && draft.rejection) {
    return (
      <Alert
        color="orange"
        title={`Rechazado por ${personName(draft.rejection.by)} el ${dateTimeFormat.format(new Date(draft.rejection.at))}`}
      >
        <Stack gap="xs">
          <Text size="sm">Motivo: «{draft.rejection.reason.trim()}»</Text>
          {draft.can_edit && <Text size="sm">Se puede editar y volver a proponer.</Text>}
        </Stack>
      </Alert>
    )
  }
  if (!proposed) {
    return draft.can_edit ? null : (
      <Alert color="gray">
        Lo está editando {personName(draft.edited_by)}. Cuando lo proponga, se aprueba o se rechaza desde acá.
      </Alert>
    )
  }
  const by = personName(draft.proposed_by ?? draft.edited_by)
  const at = draft.proposed_at ? ` el ${dateTimeFormat.format(new Date(draft.proposed_at))}` : ''
  if (!draft.can_review) {
    return (
      <Alert color="fiuba" title={`Propuesto por ${by}${at}`}>
        No se edita mientras espera. Lo aprueba o lo rechaza otra persona con permiso de aprobar: quien lo editó o lo
        propuso no puede.
      </Alert>
    )
  }
  return (
    <Alert color="fiuba" title={`Propuesto por ${by}${at}`}>
      <Stack gap="sm">
        <Text size="sm">
          Revisá el recorrido, los datos y los documentos. Al aprobarlo se publica como versión nueva; los trámites en
          curso siguen con la suya.
        </Text>
        <Group gap="sm">
          <RejectAction pending={rejectPending} onReject={onReject} />
          <ConfirmAction
            label="Aprobar y publicar"
            primary
            title="¿Aprobar y publicar la versión nueva?"
            text="Los trámites nuevos van a empezar con ella. Los que están en curso siguen con la suya. Una versión publicada no cambia."
            confirmLabel="Aprobar y publicar"
            pending={approvePending}
            onConfirm={onApprove}
          />
        </Group>
      </Stack>
    </Alert>
  )
}

// ConfirmAction asks again before a change that has consequences, and keeps
// the question open while it runs.
function ConfirmAction({
  label,
  title,
  text,
  confirmLabel,
  pending = false,
  disabled = false,
  primary = false,
  subtle = false,
  onConfirm,
}: {
  label: string
  title: string
  text: string
  confirmLabel: string
  pending?: boolean
  disabled?: boolean
  primary?: boolean
  subtle?: boolean
  onConfirm: (done: () => void) => void
}) {
  const [opened, { open, close }] = useDisclosure(false)
  return (
    <>
      <Button
        variant={primary ? 'filled' : subtle ? 'subtle' : 'default'}
        color={subtle ? 'gray' : undefined}
        disabled={disabled}
        onClick={open}
      >
        {label}
      </Button>
      <Modal opened={opened} onClose={close} title={title}>
        <Stack>
          <Text size="sm">{text}</Text>
          <Group justify="flex-end">
            <Button variant="default" onClick={close}>
              Cancelar
            </Button>
            <Button loading={pending} onClick={() => onConfirm(close)}>
              {confirmLabel}
            </Button>
          </Group>
        </Stack>
      </Modal>
    </>
  )
}

function RejectAction({
  pending,
  onReject,
}: {
  pending: boolean
  onReject: (reason: string, done: () => void) => void
}) {
  const [opened, { open, close }] = useDisclosure(false)
  const [reason, setReason] = useState('')
  return (
    <>
      <Button variant="default" onClick={open}>
        Rechazar
      </Button>
      <Modal opened={opened} onClose={close} title="Rechazar la propuesta">
        <Stack>
          <Text size="sm">El borrador vuelve a edición con el motivo, para que se corrija y se vuelva a proponer.</Text>
          <Textarea
            label="Motivo"
            description="Lo ve quien editó el borrador, hasta que lo vuelva a proponer."
            autosize
            minRows={3}
            value={reason}
            onChange={(event) => setReason(event.currentTarget.value)}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={close}>
              Cancelar
            </Button>
            <Button disabled={reason.trim() === ''} loading={pending} onClick={() => onReject(reason.trim(), close)}>
              Rechazar
            </Button>
          </Group>
        </Stack>
      </Modal>
    </>
  )
}
