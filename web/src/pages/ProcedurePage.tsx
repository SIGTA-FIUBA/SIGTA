import {
  Alert,
  Anchor,
  Button,
  FileInput,
  Grid,
  Group,
  List,
  Loader,
  Paper,
  MultiSelect,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
  ThemeIcon,
  Title,
  VisuallyHidden,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { useQueryClient } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { Link, useLocation, useParams } from 'react-router'
import { api } from '../api/client'
import { NotFoundError, useProcedure } from '../api/procedures'
import type { components } from '../api/schema'
import { FieldInput } from '../components/FieldInput/FieldInput'
import { LongValue } from '../components/LongValue/LongValue'
import { DocumentFile } from '../components/DocumentFile/DocumentFile'
import { Fact } from '../components/Fact/Fact'
import { Facts } from '../components/Facts/Facts'
import { Icon } from '../components/Icon/Icon'
import { Part } from '../components/Part/Part'
import { documentDescription } from '../content/documentFormats'
import { closedWithoutResolution, stateText, requestedNames as namesOf, requirementsOf } from '../content/procedures'
import { emptySubmission, fieldError, fileError, sentValue, valueText, type SubmissionDraft } from '../submission'

type Detail = components['schemas']['ProcedureDetail']
type Step = components['schemas']['ProcedureStep']
type Event = components['schemas']['ProcedureEvent']
type FieldRequirement = components['schemas']['FieldRequirement']
type DocumentRequirement = components['schemas']['DocumentRequirement']
type Value = components['schemas']['ProcedureValue']

const dateTime = new Intl.DateTimeFormat('es-AR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
  timeZone: 'America/Argentina/Buenos_Aires',
})

// backTo names the page a procedure was opened from, for the link back to it.
// Without one, the link goes to the home page.
function backTo(from: string | undefined) {
  if (from?.startsWith('/inbox')) {
    return { to: from, label: 'Volver a la bandeja de trámites' }
  }
  if (from?.startsWith('/my-procedures')) {
    return { to: from, label: 'Volver a Mis trámites' }
  }
  return { to: '/', label: 'Volver al inicio' }
}

function BackLink() {
  const state: unknown = useLocation().state
  const from =
    typeof state === 'object' && state !== null && 'from' in state && typeof state.from === 'string'
      ? state.from
      : undefined
  const back = backTo(from)
  return (
    <Anchor component={Link} to={back.to} size="sm" c="gray.7">
      <Group component="span" gap="xxs" wrap="nowrap">
        <Icon name="chevron-left" size={16} />
        {back.label}
      </Group>
    </Anchor>
  )
}

function responsible(step: Step) {
  return step.area?.name ?? 'Titular'
}

function personName(person: { id: string; name: string }) {
  return person.name || person.id
}

// expected says, for the applicant, what the procedure waits for now and who
// has it.
function expected(detail: Detail, step: Step) {
  if (step.role === 'applicant') {
    return `${step.name}, a cargo del titular.`
  }
  if (detail.correction_requests.length > 0) {
    return 'Lo que pide la subsanación, a cargo del titular.'
  }
  if (detail.referral) {
    return `${step.name}, en evaluación de ${detail.referral.area.name}, fuera de ${detail.referral.from.name}.`
  }
  return `${step.name}, a cargo de ${responsible(step)}.`
}

function resolutionFiles(detail: Detail) {
  return detail.steps.flatMap((step) =>
    step.files.filter((file) => step.documents.some((d) => d.id === file.document && d.resolution)),
  )
}

function Resolution({ detail }: { detail: Detail }) {
  const files = resolutionFiles(detail)
  if (files.length === 0) {
    return <Text c="dimmed">{closedWithoutResolution}</Text>
  }
  return (
    <Stack gap="md">
      {files.map((file) => (
        <Stack key={file.id} gap="xxs">
          <Text size="sm" fw={600}>
            {documentName(detail, file.document)}
          </Text>
          <DocumentFile procedureId={detail.id} file={file} />
        </Stack>
      ))}
    </Stack>
  )
}

function referralText(referral: NonNullable<Detail['referral']>) {
  const days = referral.days_without_return === 1 ? '1 día' : `${referral.days_without_return} días`
  return `En ${referral.area.name}, desde ${referral.from.name}. ${days} sin retorno.`
}

// confirmed lists the steps that someone confirmed: the events that are not
// a document, a value or a correction.
function confirmed(detail: Detail) {
  return new Set(
    detail.history
      .filter((event) => !event.document && !event.value && !event.correction_request && !event.correction_answer)
      .map((event) => event.step),
  )
}

type RouteState = 'done' | 'current' | 'pending'

// Route draws the steps of the procedure on a vertical line, in the order the
// type defines them: the confirmed ones filled, the open one ringed in the
// action color, the rest in gray.
function Route({ detail, canAct }: { detail: Detail; canAct: boolean }) {
  const done = confirmed(detail)
  return (
    <List component="ol" variant="route" aria-label="Pasos del trámite">
      {detail.steps.map((step) => {
        const state: RouteState = step.id === detail.open_step ? 'current' : done.has(step.id) ? 'done' : 'pending'
        return (
          <List.Item key={step.id} data-state={state} aria-current={state === 'current' ? 'step' : undefined}>
            <Stack gap={0}>
              <Text fw={600} c={state === 'pending' ? 'gray.7' : undefined}>
                {step.name}
                {state === 'done' && <VisuallyHidden>, hecho</VisuallyHidden>}
              </Text>
              <Text size="sm" c="gray.7">
                {responsible(step)}
              </Text>
              {state === 'current' && canAct && (
                <Text size="sm" fw={600} c="fiuba.7">
                  Acción requerida
                </Text>
              )}
            </Stack>
          </List.Item>
        )
      })}
    </List>
  )
}

// stepFiles lists each document of the procedure that has a file, in the
// order of the steps.
function stepFiles(detail: Detail) {
  return detail.steps.flatMap((step) =>
    step.documents.flatMap((requirement) => {
      const file = step.files.find((candidate) => candidate.document === requirement.id)
      return file ? [{ requirement, file }] : []
    }),
  )
}

function Documents({ detail }: { detail: Detail }) {
  const files = stepFiles(detail)
  if (files.length === 0) {
    return <Text c="dimmed">El trámite no tiene documentos.</Text>
  }
  return (
    <Stack gap="md">
      {files.map(({ requirement, file }) => (
        <Stack key={file.id} gap="xxs">
          <Text size="sm" fw={600}>
            {requirement.name}
          </Text>
          <DocumentFile procedureId={detail.id} file={file} />
        </Stack>
      ))}
    </Stack>
  )
}

// enteredValues lists each field that has a value once, in the order of the
// steps, even if more than one step asks for it.
function enteredValues(detail: Detail) {
  const entered = new Map<string, { field: FieldRequirement; value: Value }>()
  for (const step of detail.steps) {
    for (const value of step.values) {
      const field = step.fields.find((candidate) => candidate.id === value.field)
      if (field && !entered.has(field.id)) {
        entered.set(field.id, { field, value })
      }
    }
  }
  return [...entered.values()]
}

function Values({ detail }: { detail: Detail }) {
  const entered = enteredValues(detail)
  if (entered.length === 0) {
    return <Text c="dimmed">El trámite no tiene datos cargados.</Text>
  }
  return (
    <Facts label="Datos cargados">
      {entered.map(({ field, value }) => (
        <Fact key={field.id} label={field.name} split>
          {value.items ? (
            <LongValue text={value.items.map((item) => item.trim()).join('\n')} />
          ) : field.type === 'long_text' ? (
            <LongValue text={valueText(field, value.text)} />
          ) : (
            <Text>{valueText(field, value.text)}</Text>
          )}
        </Fact>
      ))}
    </Facts>
  )
}

function areaName(detail: Detail, id: string) {
  return detail.steps.find((step) => step.area?.id === id)?.area?.name ?? id
}

function documentName(detail: Detail, id: string) {
  return requirementsOf(detail.steps).documents.get(id) ?? id
}

function requestedNames(detail: Detail, requested: { fields: string[]; documents: string[] }) {
  return namesOf(detail.steps, requested)
}

// doneBy names who did a step or, for the applicant, who has the steps of
// staff: the area, since the applicant does not see the person.
// doneBy names the person or, when the applicant does not see them, the area
// of the step.
function doneBy(by: Event['by'], step: Step | undefined) {
  if (by) {
    return personName(by)
  }
  return step ? responsible(step) : 'Personal'
}

function stepName(detail: Detail, id: string | undefined) {
  return detail.steps.find((step) => step.id === id)?.name ?? id
}

// WhoWhen says who did something and when: the name, and the date after it in
// gray.
function WhoWhen({ who, at }: { who: string; at: string }) {
  return (
    <Group gap="xs" wrap="wrap">
      <Text size="sm" fw={500} data-who>
        {who}
      </Text>
      <Text component="time" size="sm" c="gray.7" dateTime={at}>
        {dateTime.format(new Date(at))}
      </Text>
    </Group>
  )
}

function HistoryItem({ detail, event, step }: { detail: Detail; event: Event; step: Step | undefined }) {
  const outcome = step?.outcomes.find((candidate) => candidate.id === event.outcome)
  return (
    <Stack gap="xs">
      <WhoWhen who={doneBy(event.by, step)} at={event.at} />
      {event.document && (
        <Stack gap={0}>
          <Text size="sm" fw={500}>
            {documentName(detail, event.document.document)}
          </Text>
          <DocumentFile procedureId={detail.id} file={event.document} />
        </Stack>
      )}
      {event.correction_request && <Text size="sm">Se pide: {requestedNames(detail, event.correction_request)}</Text>}
      {event.value && (
        <Text size="sm" fw={500}>
          {requestedNames(detail, { fields: [event.value.field], documents: [] })}
        </Text>
      )}
      {event.correction_answer && (
        <Text size="sm">Se respondió: {requestedNames(detail, event.correction_answer)}</Text>
      )}
      {event.send_back ? (
        <Text size="sm" fw={500}>
          Devuelto a {stepName(detail, event.next)}
        </Text>
      ) : (
        outcome && <Text size="sm">Salida: {outcome.name}</Text>
      )}
      {event.observation && (
        <Text size="sm">
          {event.send_back || event.correction_request ? `Motivo: ${event.observation}` : event.observation}
        </Text>
      )}
      {event.referral && (
        <Text size="sm">
          Derivado a {areaName(detail, event.referral.to)}
          {event.referral.external_file_number && `, expediente ${event.referral.external_file_number}`}
        </Text>
      )}
      {event.returned && <Text size="sm">Volvió al área de la que se derivó</Text>}
      {event.declared && (
        <Text size="xs" c="dimmed">
          Hecho fuera de SIGTA, declarado por quien lo confirmó
        </Text>
      )}
    </Stack>
  )
}

function historyTitle(event: Event, step: Step | undefined) {
  const name = step?.name ?? event.step
  if (event.document) {
    return `Documento adjuntado en ${name}`
  }
  if (event.value) {
    return `Dato cargado en ${name}`
  }
  if (event.correction_answer) {
    return `Respuesta a la subsanación en ${name}`
  }
  return event.correction_request ? `Subsanación pedida en ${name}` : name
}

function History({ detail }: { detail: Detail }) {
  if (detail.history.length === 0) {
    return <Text c="dimmed">Todavía no hay pasos hechos.</Text>
  }
  return (
    <List variant="timeline" aria-label="Historial">
      {detail.history.map((event, i) => {
        const step = detail.steps.find((candidate) => candidate.id === event.step)
        return (
          <List.Item key={i}>
            <Stack gap="xs">
              <Text fw={600}>{historyTitle(event, step)}</Text>
              <HistoryItem detail={detail} event={event} step={step} />
            </Stack>
          </List.Item>
        )
      })}
    </List>
  )
}

// SentBack tells the one who has the procedure again why it came back.
function SentBack({ detail }: { detail: Detail }) {
  const open = detail.steps.find((step) => step.id === detail.open_step)
  const last = detail.history.filter((event) => !event.document).at(-1)
  if (!last?.send_back || open?.role !== 'applicant' || last.next !== open.id) {
    return null
  }
  return (
    <Alert title="Trámite devuelto">
      <Text size="sm">Volvió al paso {open.name}. Motivo:</Text>
      {last.observation.split('\n').map((line, i) => (
        <Text key={i}>{line === '' ? ' ' : line}</Text>
      ))}
    </Alert>
  )
}

// PendingCorrections shows what the applicant has been asked for and has not
// answered yet, which holds the open step.
function PendingCorrections({ detail }: { detail: Detail }) {
  const requests = detail.correction_requests
  if (requests.length === 0) {
    return null
  }
  return (
    <Alert title={requests.length === 1 ? 'Subsanación pendiente' : `${requests.length} subsanaciones pendientes`}>
      <Stack gap="sm">
        {requests.map((request, i) => {
          const step = detail.steps.find((candidate) => candidate.id === request.step)
          return (
            <Stack key={i} gap={0}>
              <Text size="sm" fw={500}>
                Se pide: {requestedNames(detail, request)}
              </Text>
              <WhoWhen who={doneBy(request.by, step)} at={request.at} />
              <Text size="sm">Motivo:</Text>
              <LongValue text={request.reason} />
            </Stack>
          )
        })}
      </Stack>
    </Alert>
  )
}

// loadedSinceRequest says whether the applicant loaded the field or document
// again after the last request that asks for it, as the server checks it when
// the answer is sent.
function loadedSinceRequest(detail: Detail, kind: 'field' | 'document', id: string) {
  let loaded = false
  for (const event of detail.history) {
    const asked = event.correction_request
    if (asked && (kind === 'field' ? asked.fields : asked.documents).includes(id)) {
      loaded = false
    } else if (
      event.by?.id === detail.applicant.id &&
      (kind === 'field' ? event.value?.field === id : event.document?.document === id)
    ) {
      loaded = true
    }
  }
  return loaded
}

// requestedOfApplicant is everything the pending corrections ask for, once
// each, with how the steps of the applicant ask for it.
function requestedOfApplicant(detail: Detail) {
  const steps = detail.steps.filter((step) => step.role === 'applicant')
  const fields = new Map<string, FieldRequirement>()
  const documents = new Map<string, DocumentRequirement>()
  for (const request of detail.correction_requests) {
    for (const id of request.fields) {
      const field = steps.flatMap((step) => step.fields).find((candidate) => candidate.id === id)
      if (field && !fields.has(id)) {
        fields.set(id, field)
      }
    }
    for (const id of request.documents) {
      const document = steps.flatMap((step) => step.documents).find((candidate) => candidate.id === id)
      if (document && !documents.has(id)) {
        documents.set(id, document)
      }
    }
  }
  return { fields: [...fields.values()], documents: [...documents.values()] }
}

function EnterValue({ procedureId, field }: { procedureId: string; field: FieldRequirement }) {
  const queryClient = useQueryClient()
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const required = { ...field, required: true }
  const form = useForm<SubmissionDraft>({
    mode: 'uncontrolled',
    initialValues: emptySubmission([field], []),
    validate: { values: { [field.id]: (value) => fieldError(required, value) } },
  })

  async function enter(values: SubmissionDraft) {
    const value = sentValue(field, values.values[field.id] ?? '')
    setFailure(null)
    setSending(true)
    try {
      const { response, error } = await api.POST('/procedures/{procedureId}/values', {
        params: { path: { procedureId } },
        body: Array.isArray(value) ? { field: field.id, items: value } : { field: field.id, text: value },
      })
      if (response.ok) {
        form.reset()
      } else {
        setFailure(error && 'message' in error ? error.message : 'No pudimos cargar el dato. Volvé a intentarlo.')
      }
      await queryClient.invalidateQueries({ queryKey: ['procedure', procedureId] })
    } catch {
      setFailure('No pudimos cargar el dato. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  return (
    <form onSubmit={form.onSubmit((values) => void enter(values))} noValidate>
      <Stack gap="xs">
        <FieldInput form={form} field={required} labelled={false} />
        {failure && (
          <Alert color="red" role="alert">
            {failure}
          </Alert>
        )}
        <Group>
          <Button type="submit" variant="default" loading={sending} aria-label={`Cargar ${field.name}`}>
            Cargar
          </Button>
        </Group>
      </Stack>
    </form>
  )
}

// RequestedItem is one thing a correction asks for: its name, once, whether
// it was loaded again, and the input to load it.
function RequestedItem({ name, loaded, children }: { name: string; loaded: boolean; children: ReactNode }) {
  return (
    <Stack gap="xs" py="md">
      <Group justify="space-between" gap="xs" wrap="nowrap">
        <Text fw={600}>{name}</Text>
        <Text size="sm" fw={600} c={loaded ? 'green.8' : 'orange.8'}>
          {loaded ? 'Cargado' : 'Pendiente'}
        </Text>
      </Group>
      {children}
    </Stack>
  )
}

// CorrectionAnswer lets the applicant load again what the pending corrections
// ask for, one by one, and then send the answer, which takes the procedure
// back to staff. What they load lands on the step of theirs that asks for it.
function CorrectionAnswer({ detail, onAnswered }: { detail: Detail; onAnswered: () => void }) {
  const queryClient = useQueryClient()
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const { fields, documents } = requestedOfApplicant(detail)

  async function answer() {
    setFailure(null)
    setSending(true)
    try {
      const { response, error } = await api.POST('/procedures/{procedureId}/correction-answer', {
        params: { path: { procedureId: detail.id } },
      })
      if (response.ok) {
        onAnswered()
      } else {
        setFailure(error && 'message' in error ? error.message : 'No pudimos enviar la respuesta. Volvé a intentarlo.')
      }
      await queryClient.invalidateQueries({ queryKey: ['procedure', detail.id] })
      await queryClient.invalidateQueries({ queryKey: ['me', 'procedures'] })
    } catch {
      setFailure('No pudimos enviar la respuesta. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  return (
    <Paper component="section" withBorder p="xl" aria-label="Subsanación pendiente">
      <Stack gap="lg">
        <Group wrap="nowrap" gap="md" align="flex-start">
          <ThemeIcon variant="light" color="orange" radius="xl" size="lg" aria-hidden>
            <Icon name="file-pencil" size={20} />
          </ThemeIcon>
          <Stack gap="xs" miw={0}>
            <Title order={2}>Subsanación pendiente</Title>
            {detail.correction_requests.map((request, i) => {
              const step = detail.steps.find((candidate) => candidate.id === request.step)
              return (
                <Stack key={i} gap="xxs">
                  <LongValue text={request.reason} />
                  <WhoWhen who={doneBy(request.by, step)} at={request.at} />
                </Stack>
              )
            })}
          </Stack>
        </Group>
        <Stack variant="facts" gap={0}>
          {fields.map((field) => (
            <RequestedItem
              key={`field-${field.id}`}
              name={field.name}
              loaded={loadedSinceRequest(detail, 'field', field.id)}
            >
              <EnterValue procedureId={detail.id} field={field} />
            </RequestedItem>
          ))}
          {documents.map((document) => (
            <RequestedItem
              key={`document-${document.id}`}
              name={document.name}
              loaded={loadedSinceRequest(detail, 'document', document.id)}
            >
              <AttachDocument procedureId={detail.id} document={document} labelled={false} onAttached={() => {}} />
            </RequestedItem>
          ))}
        </Stack>
        {failure && (
          <Alert color="red" role="alert">
            {failure}
          </Alert>
        )}
        <Group>
          <Button onClick={() => void answer()} loading={sending}>
            Enviar la respuesta
          </Button>
        </Group>
      </Stack>
    </Paper>
  )
}

// leadsElsewhere says whether moving from the step to the one that follows it
// refers the procedure: only then can the confirmation carry a file number.
function leadsElsewhere(detail: Detail, step: Step, outcome: string | null) {
  const next = step.action === 'decide' ? step.outcomes.find((o) => o.id === outcome)?.next : step.next
  const area = detail.steps.find((candidate) => candidate.id === next)?.area
  return area !== undefined && area.id !== step.area?.id
}

// AttachDocument attaches the file as soon as it is chosen: the input is the
// button. Without labelled, the name goes only to assistive technology, since
// the item that holds it already shows it.
function AttachDocument({
  procedureId,
  document,
  labelled = true,
  onAttached,
  onCancel,
}: {
  procedureId: string
  document: DocumentRequirement
  labelled?: boolean
  onAttached: () => void
  onCancel?: () => void
}) {
  const queryClient = useQueryClient()
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)

  async function attach(file: File | null) {
    if (!file) {
      return
    }
    const invalid = fileError({ ...document, required: true }, file)
    if (invalid) {
      setFailure(invalid)
      return
    }
    setFailure(null)
    setSending(true)
    try {
      const body = new FormData()
      body.append('document', document.id)
      body.append('file', file)
      const { response, error } = await api.POST('/procedures/{procedureId}/documents', {
        params: { path: { procedureId } },
        body: { document: document.id, file: '' },
        bodySerializer: () => body,
      })
      if (response.ok) {
        onAttached()
      } else {
        setFailure(error && 'message' in error ? error.message : 'No pudimos adjuntar el archivo. Volvé a intentarlo.')
      }
      await queryClient.invalidateQueries({ queryKey: ['procedure', procedureId] })
    } catch {
      setFailure('No pudimos adjuntar el archivo. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  return (
    <Stack gap="xs">
      <FileInput
        label={labelled ? document.name : undefined}
        aria-label={labelled ? undefined : document.name}
        description={documentDescription(document)}
        accept={document.accepted_types.join(',')}
        placeholder="Adjuntar un archivo…"
        value={null}
        onChange={(file) => void attach(file)}
        disabled={sending}
        rightSection={sending ? <Loader size="xs" aria-label={`Adjuntando ${document.name}`} /> : undefined}
        error={failure}
      />
      {onCancel && (
        <Group>
          <Button variant="subtle" onClick={onCancel} aria-label={`No reemplazar ${document.name}`}>
            Cancelar
          </Button>
        </Group>
      )}
    </Stack>
  )
}

// StepDocument shows the file attached for a document in the open step, which
// can be replaced, or the form to attach one if the step has none.
function StepDocument({
  procedureId,
  step,
  document,
  onAttached,
}: {
  procedureId: string
  step: Step
  document: DocumentRequirement
  onAttached: (replaced: boolean) => void
}) {
  const [replacing, setReplacing] = useState(false)
  const file = step.files.find((candidate) => candidate.document === document.id && candidate.step === step.id)
  if (!file) {
    return <AttachDocument procedureId={procedureId} document={document} onAttached={() => onAttached(false)} />
  }
  if (replacing) {
    return (
      <AttachDocument
        procedureId={procedureId}
        document={document}
        onCancel={() => setReplacing(false)}
        onAttached={() => {
          setReplacing(false)
          onAttached(true)
        }}
      />
    )
  }
  return (
    <Stack gap="xs">
      <Text size="sm" fw={500}>
        {document.name}
      </Text>
      <DocumentFile procedureId={procedureId} file={file} />
      <Group>
        <Button variant="default" onClick={() => setReplacing(true)} aria-label={`Reemplazar ${document.name}`}>
          Reemplazar
        </Button>
      </Group>
    </Stack>
  )
}

// StepDocuments lists the documents of the open step. A document counts as
// attached only with a file from that same step, the only one that lets the
// step be confirmed.
function StepDocuments({
  procedureId,
  step,
  onAttached,
}: {
  procedureId: string
  step: Step
  onAttached: (replaced: boolean) => void
}) {
  if (step.documents.length === 0) {
    return null
  }
  return (
    <Stack component="section" variant="part" gap={0}>
      <Stack gap="md">
        <Title order={2}>Documentos del paso</Title>
        {step.documents.map((document) => (
          <StepDocument
            key={document.id}
            procedureId={procedureId}
            step={step}
            document={document}
            onAttached={onAttached}
          />
        ))}
      </Stack>
    </Stack>
  )
}

function ConfirmStep({ detail, step, onConfirmed }: { detail: Detail; step: Step; onConfirmed: () => void }) {
  const queryClient = useQueryClient()
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const decide = step.action === 'decide'
  const form = useForm({
    mode: 'uncontrolled',
    initialValues: { observation: '', outcome: null as string | null, externalFileNumber: '' },
    validate: {
      observation: (value, values) =>
        value.trim()
          ? null
          : step.outcomes.find((o) => o.id === values.outcome)?.send_back
            ? 'Escribí el motivo de la devolución.'
            : 'Escribí una observación.',
      outcome: (value) => (decide && !value ? 'Elegí una salida.' : null),
    },
  })
  const outcome = form.useWatchValue('outcome')
  const refers = leadsElsewhere(detail, step, outcome)
  const sendBackOutcome = step.outcomes.find((o) => o.id === outcome && o.send_back)

  async function confirm(values: typeof form.values) {
    setFailure(null)
    setSending(true)
    try {
      const { response, error } = await api.POST('/procedures/{procedureId}/steps/{stepId}/confirmation', {
        params: { path: { procedureId: detail.id, stepId: step.id } },
        body: {
          observation: values.observation.trim(),
          outcome: decide ? (values.outcome ?? undefined) : undefined,
          external_file_number:
            refers && values.externalFileNumber.trim() ? values.externalFileNumber.trim() : undefined,
        },
      })
      if (response.ok) {
        onConfirmed()
      } else {
        setFailure(error && 'message' in error ? error.message : 'No pudimos confirmar el paso. Volvé a intentarlo.')
      }
      await queryClient.invalidateQueries({ queryKey: ['procedure', detail.id] })
      await queryClient.invalidateQueries({ queryKey: ['procedures'] })
    } catch {
      setFailure('No pudimos confirmar el paso. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  return (
    <Stack component="section" variant="part" gap={0}>
      <form onSubmit={form.onSubmit((values) => void confirm(values))} noValidate>
        <Stack gap="md">
          <Title order={2}>Confirmar {step.name}</Title>
          {decide && (
            <Select
              label="Salida"
              withAsterisk
              placeholder="Elegí una salida…"
              data={step.outcomes.map((o) => ({ value: o.id, label: o.name }))}
              key={form.key('outcome')}
              {...form.getInputProps('outcome')}
            />
          )}
          <Textarea
            label={sendBackOutcome ? 'Motivo de la devolución' : 'Observación'}
            withAsterisk
            autosize
            minRows={3}
            key={form.key('observation')}
            {...form.getInputProps('observation')}
          />
          {refers && (
            <TextInput
              label="Número de expediente externo"
              key={form.key('externalFileNumber')}
              {...form.getInputProps('externalFileNumber')}
            />
          )}
          {failure && (
            <Alert color="red" role="alert">
              {failure}
            </Alert>
          )}
          <Group>
            <Button type="submit" loading={sending}>
              Confirmar el paso
            </Button>
          </Group>
        </Stack>
      </form>
    </Stack>
  )
}

const fieldPrefix = 'field:'
const documentPrefix = 'document:'

function RequestCorrection({ detail, step, onRequested }: { detail: Detail; step: Step; onRequested: () => void }) {
  const queryClient = useQueryClient()
  const [opened, setOpened] = useState(false)
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const form = useForm({
    mode: 'uncontrolled',
    initialValues: { requested: [] as string[], reason: '' },
    validate: {
      requested: (value) => (value.length === 0 ? 'Elegí qué falta.' : null),
      reason: (value) => (value.trim() ? null : 'Escribí el motivo.'),
    },
  })
  // Only what the steps of the applicant ask for, so the answer lands on one of them.
  const { fields, documents } = requirementsOf(detail.steps.filter((candidate) => candidate.role === 'applicant'))
  const options = [
    { group: 'Datos', items: [...fields].map(([id, name]) => ({ value: fieldPrefix + id, label: name })) },
    { group: 'Documentos', items: [...documents].map(([id, name]) => ({ value: documentPrefix + id, label: name })) },
  ].filter((group) => group.items.length > 0)

  if (!opened) {
    return (
      <Group>
        <Button variant="default" onClick={() => setOpened(true)}>
          Pedir una subsanación
        </Button>
      </Group>
    )
  }

  async function request(values: typeof form.values) {
    setFailure(null)
    setSending(true)
    try {
      const { response, error } = await api.POST('/procedures/{procedureId}/steps/{stepId}/correction-requests', {
        params: { path: { procedureId: detail.id, stepId: step.id } },
        body: {
          reason: values.reason.trim(),
          fields: values.requested.filter((v) => v.startsWith(fieldPrefix)).map((v) => v.slice(fieldPrefix.length)),
          documents: values.requested
            .filter((v) => v.startsWith(documentPrefix))
            .map((v) => v.slice(documentPrefix.length)),
        },
      })
      if (response.ok) {
        setOpened(false)
        form.reset()
        onRequested()
      } else {
        setFailure(error && 'message' in error ? error.message : 'No pudimos pedir la subsanación. Volvé a intentarlo.')
      }
      await queryClient.invalidateQueries({ queryKey: ['procedure', detail.id] })
      await queryClient.invalidateQueries({ queryKey: ['procedures'] })
    } catch {
      setFailure('No pudimos pedir la subsanación. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  return (
    <Stack component="section" variant="part" gap={0}>
      <form onSubmit={form.onSubmit((values) => void request(values))} noValidate>
        <Stack gap="md">
          <Title order={2}>Pedir una subsanación</Title>
          <MultiSelect
            label="Qué falta"
            withAsterisk
            searchable
            placeholder="Elegí uno o más…"
            nothingFoundMessage="No hay datos ni documentos con ese nombre."
            data={options}
            key={form.key('requested')}
            {...form.getInputProps('requested')}
          />
          <Textarea
            label="Motivo"
            withAsterisk
            autosize
            minRows={3}
            key={form.key('reason')}
            {...form.getInputProps('reason')}
          />
          {failure && (
            <Alert color="red" role="alert">
              {failure}
            </Alert>
          )}
          <Group>
            <Button type="submit" variant="default" loading={sending}>
              Pedir la subsanación
            </Button>
            <Button variant="subtle" onClick={() => setOpened(false)}>
              Cancelar
            </Button>
          </Group>
        </Stack>
      </form>
    </Stack>
  )
}

// receiptUrl is the receipt that the applicant keeps of their procedure, in
// progress or closed, with what they see of it here.
function receiptUrl(procedureId: string) {
  return `/api/procedures/${encodeURIComponent(procedureId)}/receipt`
}

function ProcedureDetail({ detail }: { detail: Detail }) {
  const [notice, setNotice] = useState<string | null>(null)
  const open = detail.steps.find((step) => step.id === detail.open_step)
  const awaiting = detail.correction_requests.length > 0
  const state = stateText(detail.state, {
    area: detail.steps.find((step) => step.id === detail.open_step)?.area,
    applicant: detail.as_applicant,
  })
  return (
    <Stack gap="xl">
      {/* The type of the procedure in the row that opened it becomes this
          title (a view transition, global.css). */}
      <Title order={1} variant="display" c="fiuba.9" data-transition-title>
        {detail.name}
      </Title>
      {/* On a wide screen what the procedure is and its steps go in a narrow
          column on the right; the work and the record, on the left. On a
          phone the facts come first. */}
      <Grid gap={{ base: 'xl', lg: 'xxl' }}>
        <Grid.Col span={{ base: 12, lg: 4 }} order={{ base: 1, lg: 2 }}>
          <Stack gap="xl">
            <Stack gap="md">
              <Facts label="Resumen del trámite">
                <Fact label="Estado">
                  <Text fw={600} c={state.attention ? 'orange.8' : undefined}>
                    {state.text}
                  </Text>
                </Fact>
                <Fact label="Titular">
                  <Text>{personName(detail.applicant)}</Text>
                </Fact>
                {detail.referral && (
                  <Fact label="Derivación">
                    <Text>{referralText(detail.referral)}</Text>
                  </Fact>
                )}
                {detail.as_applicant && open && (
                  <Fact label="Qué se espera ahora">
                    <Text>{expected(detail, open)}</Text>
                  </Fact>
                )}
              </Facts>
              {detail.as_applicant && (
                <Group>
                  <Button
                    component="a"
                    href={receiptUrl(detail.id)}
                    download={`Comprobante ${detail.name}.pdf`}
                    variant="default"
                  >
                    Descargar comprobante
                  </Button>
                </Group>
              )}
            </Stack>
            <Stack component="section" gap="md" aria-label="Pasos">
              <Title order={2}>Pasos</Title>
              <Route detail={detail} canAct={detail.can_act && !awaiting} />
            </Stack>
          </Stack>
        </Grid.Col>
        <Grid.Col span={{ base: 12, lg: 8 }} order={{ base: 2, lg: 1 }}>
          <Stack gap="xxl">
            {detail.as_applicant && awaiting && (
              <CorrectionAnswer
                detail={detail}
                onAnswered={() => setNotice('Respuesta enviada. El trámite volvió a quien pidió la subsanación.')}
              />
            )}
            <SentBack detail={detail} />
            {!detail.as_applicant && <PendingCorrections detail={detail} />}
            {detail.state === 'closed' && (
              <Part title="Resolución">
                <Resolution detail={detail} />
              </Part>
            )}
            {notice && (
              <Alert withCloseButton onClose={() => setNotice(null)} closeButtonLabel="Cerrar el aviso">
                {notice}
              </Alert>
            )}
            {detail.can_act && open?.role === 'staff' && (
              <>
                <StepDocuments
                  key={`documents-${open.id}`}
                  procedureId={detail.id}
                  step={open}
                  onAttached={(replaced) =>
                    setNotice(`Documento ${replaced ? 'reemplazado' : 'adjuntado'}. Quedó en el historial.`)
                  }
                />
                {!awaiting && (
                  <ConfirmStep
                    key={open.id}
                    detail={detail}
                    step={open}
                    onConfirmed={() => setNotice('Paso confirmado. Quedó en el historial.')}
                  />
                )}
                <RequestCorrection
                  key={`correction-${open.id}`}
                  detail={detail}
                  step={open}
                  onRequested={() => setNotice('Subsanación pedida. Quedó en el historial.')}
                />
              </>
            )}
            <Part title="Datos del trámite">
              <Values detail={detail} />
            </Part>
            <Part title="Documentos">
              <Documents detail={detail} />
            </Part>
            <Part title="Historial">
              <History detail={detail} />
            </Part>
          </Stack>
        </Grid.Col>
      </Grid>
    </Stack>
  )
}

export function ProcedurePage() {
  const { id = '' } = useParams()
  const procedure = useProcedure(id)

  return (
    <Stack gap="lg">
      <Group>
        <BackLink />
      </Group>
      {procedure.isPending && <Loader aria-label="Cargando el trámite…" />}
      {procedure.error instanceof NotFoundError && (
        <>
          <Title order={1}>Trámite</Title>
          <Text>No existe el trámite, o no forma parte de tus trámites ni de los de tus áreas.</Text>
        </>
      )}
      {procedure.isError && !(procedure.error instanceof NotFoundError) && (
        <>
          <Title order={1}>Trámite</Title>
          <Alert color="red" role="alert">
            No pudimos cargar el trámite. Volvé a intentarlo más tarde.
          </Alert>
        </>
      )}
      {procedure.data && <ProcedureDetail detail={procedure.data} />}
    </Stack>
  )
}
