import {
  Alert,
  Anchor,
  Badge,
  Code,
  Group,
  Loader,
  Paper,
  SimpleGrid,
  Stack,
  Text,
  Timeline,
  Title,
} from '@mantine/core'
import type { ReactNode } from 'react'
import { Link, useLocation, useParams } from 'react-router'
import { NotFoundError, useProcedure } from '../api/procedures'
import type { components } from '../api/schema'
import { sizeText } from '../documentFormats'
import { stateColors, stateNames } from '../procedureState'

type Detail = components['schemas']['ProcedureDetail']
type Step = components['schemas']['ProcedureStep']
type Event = components['schemas']['ProcedureEvent']

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

function ProcedureDetail({ detail }: { detail: Detail }) {
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
