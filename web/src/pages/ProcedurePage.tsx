import {
  Alert,
  Anchor,
  Badge,
  Button,
  Code,
  FileInput,
  Group,
  Loader,
  Paper,
  Select,
  SimpleGrid,
  Stack,
  Text,
  Textarea,
  TextInput,
  Timeline,
  Title,
} from '@mantine/core'
import { isNotEmpty, useForm } from '@mantine/form'
import { useQueryClient } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { Link, useLocation, useParams } from 'react-router'
import { api } from '../api/client'
import { NotFoundError, useProcedure } from '../api/procedures'
import type { components } from '../api/schema'
import { LongValue } from '../components/LongValue/LongValue'
import { DocumentFile } from '../components/DocumentFile/DocumentFile'
import { documentDescription } from '../content/documentFormats'
import { stateColors, stateNames } from '../content/procedureStates'
import { fileError, valueText } from '../submission'

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

function BackLink() {
  const state: unknown = useLocation().state
  if (typeof state === 'object' && state !== null && 'from' in state && typeof state.from === 'string') {
    return (
      <Anchor component={Link} to={state.from}>
        Volver a la bandeja de trámites
      </Anchor>
    )
  }
  return (
    <Anchor component={Link} to="/">
      Volver al inicio
    </Anchor>
  )
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <Stack gap={0}>
      <Text size="sm" c="dimmed">
        {label}
      </Text>
      {children}
    </Stack>
  )
}

function responsible(step: Step) {
  return step.area?.name ?? 'Titular'
}

function personName(person: { id: string; name: string }) {
  return person.name || person.id
}

function OpenStep({ step, canAct }: { step: Step | undefined; canAct: boolean }) {
  if (step === undefined) {
    return <Text>Sin paso abierto</Text>
  }
  return (
    <Group gap="xs">
      <Text>
        {step.name} · {responsible(step)}
      </Text>
      {canAct && <Badge variant="filled">Acción requerida</Badge>}
    </Group>
  )
}

// expected says, for the applicant, what the procedure waits for now and who
// has it.
function expected(detail: Detail, step: Step | undefined) {
  if (step === undefined) {
    return 'Nada: el trámite está cerrado.'
  }
  if (step.role === 'applicant') {
    return `${step.name}, a cargo del titular.`
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
    return <Text c="dimmed">El trámite se cerró sin un documento resolutivo adjunto.</Text>
  }
  return (
    <Paper withBorder p="md">
      <Stack gap="sm">
        {files.map((file) => (
          <Stack key={file.id} gap={0}>
            <Text size="sm" fw={500}>
              {documentName(detail, file.document)}
            </Text>
            <DocumentFile procedureId={detail.id} file={file} />
          </Stack>
        ))}
      </Stack>
    </Paper>
  )
}

function referralText(referral: NonNullable<Detail['referral']>) {
  const days = referral.days_without_return === 1 ? '1 día' : `${referral.days_without_return} días`
  return `En ${referral.area.name}, desde ${referral.from.name}. ${days} sin retorno.`
}

function StepCard({ procedureId, step, open }: { procedureId: string; step: Step; open: boolean }) {
  return (
    <Paper withBorder p="md">
      <Stack gap="xs">
        <Group justify="space-between" gap="xs">
          <Text fw={600}>{step.name}</Text>
          {open && <Badge variant="light">Paso actual</Badge>}
        </Group>
        <Text size="sm" c="dimmed">
          {responsible(step)}
        </Text>
        {step.documents.map((requirement) => {
          const file = step.files.find((candidate) => candidate.document === requirement.id)
          return (
            <Stack key={requirement.id} gap={0}>
              <Text size="sm" fw={500}>
                {requirement.name}
              </Text>
              {file ? (
                <DocumentFile procedureId={procedureId} file={file} />
              ) : (
                <Text size="sm" c="dimmed">
                  Sin archivo
                </Text>
              )}
            </Stack>
          )
        })}
      </Stack>
    </Paper>
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
    <Paper withBorder p="md">
      <Stack gap="sm">
        {entered.map(({ field, value }) => (
          <Fact key={field.id} label={field.name}>
            {value.items ? (
              <LongValue text={value.items.map((item) => item.trim()).join('\n')} />
            ) : field.type === 'long_text' ? (
              <LongValue text={valueText(field, value.text)} />
            ) : (
              <Text>{valueText(field, value.text)}</Text>
            )}
          </Fact>
        ))}
      </Stack>
    </Paper>
  )
}

function areaName(detail: Detail, id: string) {
  return detail.steps.find((step) => step.area?.id === id)?.area?.name ?? id
}

function documentName(detail: Detail, id: string) {
  for (const step of detail.steps) {
    const document = step.documents.find((candidate) => candidate.id === id)
    if (document) {
      return document.name
    }
  }
  return id
}

function HistoryItem({ detail, event, step }: { detail: Detail; event: Event; step: Step | undefined }) {
  const outcome = step?.outcomes.find((candidate) => candidate.id === event.outcome)
  return (
    <Stack gap="xs">
      <Text size="sm" c="dimmed">
        {personName(event.by)} · {dateTime.format(new Date(event.at))}
      </Text>
      {event.document && (
        <Stack gap={0}>
          <Text size="sm" fw={500}>
            {documentName(detail, event.document.document)}
          </Text>
          <DocumentFile procedureId={detail.id} file={event.document} />
        </Stack>
      )}
      {outcome && <Text size="sm">Salida: {outcome.name}</Text>}
      {event.observation && <Text size="sm">{event.observation}</Text>}
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
  return event.document ? `Documento adjuntado en ${name}` : name
}

function History({ detail }: { detail: Detail }) {
  if (detail.history.length === 0) {
    return <Text c="dimmed">Todavía no hay pasos hechos.</Text>
  }
  return (
    <Timeline active={detail.history.length - 1} bulletSize={20} lineWidth={2}>
      {detail.history.map((event, i) => {
        const step = detail.steps.find((candidate) => candidate.id === event.step)
        return (
          <Timeline.Item key={i} title={historyTitle(event, step)}>
            <HistoryItem detail={detail} event={event} step={step} />
          </Timeline.Item>
        )
      })}
    </Timeline>
  )
}

// leadsElsewhere says whether moving from the step to the one that follows it
// refers the procedure: only then can the confirmation carry a file number.
function leadsElsewhere(detail: Detail, step: Step, outcome: string | null) {
  const next = step.action === 'decide' ? step.outcomes.find((o) => o.id === outcome)?.next : step.next
  const area = detail.steps.find((candidate) => candidate.id === next)?.area
  return area !== undefined && area.id !== step.area?.id
}

function AttachDocument({
  procedureId,
  document,
  onAttached,
  onCancel,
}: {
  procedureId: string
  document: DocumentRequirement
  onAttached: () => void
  onCancel?: () => void
}) {
  const queryClient = useQueryClient()
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const form = useForm({
    mode: 'uncontrolled',
    initialValues: { file: null as File | null },
    validate: { file: (file) => fileError({ ...document, required: true }, file) },
  })

  async function attach({ file }: { file: File | null }) {
    if (!file) {
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
    <form onSubmit={form.onSubmit((values) => void attach(values))} noValidate>
      <Stack gap="xs">
        <FileInput
          label={document.name}
          description={documentDescription(document)}
          withAsterisk={document.required}
          accept={document.accepted_types.join(',')}
          placeholder="Elegí un archivo…"
          clearable
          clearButtonProps={{ 'aria-label': `Quitar el archivo de ${document.name}` }}
          key={form.key('file')}
          {...form.getInputProps('file')}
        />
        {failure && (
          <Alert color="red" role="alert">
            {failure}
          </Alert>
        )}
        <Group>
          <Button type="submit" variant="default" loading={sending} aria-label={`Adjuntar ${document.name}`}>
            Adjuntar
          </Button>
          {onCancel && (
            <Button variant="subtle" onClick={onCancel} aria-label={`No reemplazar ${document.name}`}>
              Cancelar
            </Button>
          )}
        </Group>
      </Stack>
    </form>
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
    <Paper withBorder p="md">
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
    </Paper>
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
      observation: isNotEmpty('Escribí una observación.'),
      outcome: (value) => (decide && !value ? 'Elegí una salida.' : null),
    },
  })
  const outcome = form.useWatchValue('outcome')
  const refers = leadsElsewhere(detail, step, outcome)

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
    <Paper withBorder p="md">
      <form onSubmit={form.onSubmit((values) => void confirm(values))} noValidate>
        <Stack gap="md">
          <Title order={2}>Confirmar {step.name}</Title>
          {step.action === 'confirm_external_fact' && (
            <Text size="sm" c="dimmed">
              Se registra que declarás que el paso se hizo fuera de SIGTA.
            </Text>
          )}
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
            label="Observación"
            withAsterisk
            autosize
            minRows={3}
            key={form.key('observation')}
            {...form.getInputProps('observation')}
          />
          {refers && (
            <TextInput
              label="Número de expediente externo"
              description="Si el trámite tiene expediente en otro sistema."
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
    </Paper>
  )
}

function ProcedureDetail({ detail }: { detail: Detail }) {
  const [notice, setNotice] = useState<string | null>(null)
  const open = detail.steps.find((step) => step.id === detail.open_step)
  return (
    <Stack gap="lg">
      <Stack gap="xs">
        <Title order={1}>{detail.name}</Title>
        <Text c="dimmed">
          Trámite <Code translate="no">{detail.id}</Code>
        </Text>
      </Stack>
      <Paper withBorder p="md">
        <SimpleGrid cols={{ base: 1, sm: 2 }}>
          <Fact label="Estado">
            <Group>
              <Badge variant="light" color={stateColors[detail.state]}>
                {stateNames[detail.state]}
              </Badge>
            </Group>
          </Fact>
          <Fact label="Titular">
            <Text>{personName(detail.applicant)}</Text>
          </Fact>
          <Fact label="Paso actual">
            <OpenStep step={open} canAct={detail.can_act} />
          </Fact>
          {detail.referral && (
            <Fact label="Derivación">
              <Text>{referralText(detail.referral)}</Text>
            </Fact>
          )}
          {detail.as_applicant && (
            <Fact label="Qué se espera ahora">
              <Text>{expected(detail, open)}</Text>
            </Fact>
          )}
        </SimpleGrid>
      </Paper>
      {detail.state === 'closed' && (
        <Stack gap="sm">
          <Title order={2}>Resolución</Title>
          <Resolution detail={detail} />
        </Stack>
      )}
      {notice && (
        <Alert withCloseButton onClose={() => setNotice(null)} closeButtonLabel="Cerrar el aviso">
          {notice}
        </Alert>
      )}
      <Stack gap="sm">
        <Title order={2}>Datos del trámite</Title>
        <Values detail={detail} />
      </Stack>
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
          <ConfirmStep
            key={open.id}
            detail={detail}
            step={open}
            onConfirmed={() => setNotice('Paso confirmado. Quedó en el historial.')}
          />
        </>
      )}
      <Stack gap="sm">
        <Title order={2}>Historial</Title>
        <History detail={detail} />
      </Stack>
      <Stack gap="sm">
        <Title order={2}>Pasos</Title>
        <Text size="sm" c="dimmed">
          En el orden en que los define el trámite. No es necesariamente el recorrido.
        </Text>
        {detail.steps.map((step) => (
          <StepCard key={step.id} procedureId={detail.id} step={step} open={step.id === detail.open_step} />
        ))}
      </Stack>
    </Stack>
  )
}

export function ProcedurePage() {
  const { id = '' } = useParams()
  const procedure = useProcedure(id)

  return (
    <Stack gap="md">
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
