import {
  Alert,
  Anchor,
  Badge,
  Button,
  Code,
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
import { sizeText } from '../documentFormats'
import { stateColors, stateNames } from '../procedureState'
import { valueText } from '../submission'

type Detail = components['schemas']['ProcedureDetail']
type Step = components['schemas']['ProcedureStep']
type Event = components['schemas']['ProcedureEvent']
type FieldRequirement = components['schemas']['FieldRequirement']
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

function OpenStep({ detail }: { detail: Detail }) {
  const step = detail.steps.find((candidate) => candidate.id === detail.open_step)
  if (step === undefined) {
    return <Text>Sin paso abierto</Text>
  }
  return (
    <Group gap="xs">
      <Text>
        {step.name} · {responsible(step)}
      </Text>
      {detail.can_act && <Badge variant="filled">Acción requerida</Badge>}
    </Group>
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
            <Text key={requirement.id} size="sm">
              {requirement.name}:{' '}
              {file ? (
                <Anchor
                  href={`/api/procedures/${encodeURIComponent(procedureId)}/documents/${encodeURIComponent(file.id)}/content`}
                >
                  {file.file_name}
                </Anchor>
              ) : (
                'Sin archivo'
              )}
              {file && ` (${sizeText(file.size)})`}
            </Text>
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
        {entered.map(({ field, value }) => {
          const text = valueText(field, value.items ?? value.text)
          return (
            <Fact key={field.id} label={field.name}>
              {field.type === 'long_text' ? <LongValue text={text} /> : <Text>{text}</Text>}
            </Fact>
          )
        })}
      </Stack>
    </Paper>
  )
}

function areaName(detail: Detail, id: string) {
  return detail.steps.find((step) => step.area?.id === id)?.area?.name ?? id
}

function HistoryItem({ detail, event }: { detail: Detail; event: Event }) {
  const step = detail.steps.find((candidate) => candidate.id === event.step)
  const outcome = step?.outcomes.find((candidate) => candidate.id === event.outcome)
  return (
    <Stack gap="xs">
      <Text size="sm" c="dimmed">
        {personName(event.by)} · {dateTime.format(new Date(event.at))}
      </Text>
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

function History({ detail }: { detail: Detail }) {
  if (detail.history.length === 0) {
    return <Text c="dimmed">Todavía no hay pasos hechos.</Text>
  }
  return (
    <Timeline active={detail.history.length - 1} bulletSize={20} lineWidth={2}>
      {detail.history.map((event, i) => (
        <Timeline.Item key={i} title={detail.steps.find((step) => step.id === event.step)?.name ?? event.step}>
          <HistoryItem detail={detail} event={event} />
        </Timeline.Item>
      ))}
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
  const [confirmed, setConfirmed] = useState(false)
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
            <OpenStep detail={detail} />
          </Fact>
          {detail.referral && (
            <Fact label="Derivación">
              <Text>{referralText(detail.referral)}</Text>
            </Fact>
          )}
        </SimpleGrid>
      </Paper>
      {confirmed && (
        <Alert withCloseButton onClose={() => setConfirmed(false)} closeButtonLabel="Cerrar el aviso">
          Paso confirmado. Quedó en el historial.
        </Alert>
      )}
      <Stack gap="sm">
        <Title order={2}>Datos del trámite</Title>
        <Values detail={detail} />
      </Stack>
      {detail.can_act && open?.role === 'staff' && (
        <ConfirmStep key={open.id} detail={detail} step={open} onConfirmed={() => setConfirmed(true)} />
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
