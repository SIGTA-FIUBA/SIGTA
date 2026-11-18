import {
  Alert,
  Anchor,
  Button,
  Code,
  FileInput,
  Group,
  Loader,
  Modal,
  Select,
  Stack,
  Stepper,
  TagsInput,
  Text,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core'
import { useForm, type UseFormReturnType } from '@mantine/form'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Link, useBlocker, useParams } from 'react-router'
import { catalogQuery } from '../api/catalog'
import { api } from '../api/client'
import type { components } from '../api/schema'
import { LongValue } from '../components/LongValue/LongValue'
import { documentDescription } from '../documentFormats'
import {
  emptySubmission,
  fieldError,
  fileError,
  isEmpty,
  maxSubjects,
  problemErrors,
  submissionBody,
  totalSizeErrors,
  valueText,
  type SubmissionDraft,
} from '../submission'

type CatalogEntry = components['schemas']['CatalogEntry']
type FieldRequirement = components['schemas']['FieldRequirement']
type DocumentRequirement = components['schemas']['DocumentRequirement']
type StartedProcedure = components['schemas']['StartedProcedure']
type Form = UseFormReturnType<SubmissionDraft>

const stageData = 0
const stageDocuments = 1
const stageSummary = 2

// Beyond this many options, a select lets the applicant type to find one.
const searchableOptions = 7

// failureText says what to do when the server rejects the submission for
// something other than its content.
function failureText(status: number) {
  switch (status) {
    case 401:
      return 'Tu sesión venció. Volvé a ingresar y enviá el trámite de nuevo.'
    case 404:
      return 'Este trámite ya no está disponible. Revisá el catálogo de trámites.'
    case 400:
      return 'No pudimos leer el envío. Revisá los archivos y volvé a intentarlo.'
    default:
      return 'No pudimos enviar el trámite. Volvé a intentarlo más tarde.'
  }
}

function FieldInput({ form, field }: { form: Form; field: FieldRequirement }) {
  const path = `values.${field.id}`
  const common = {
    label: field.name,
    name: field.id,
    withAsterisk: field.required,
    key: form.key(path),
    ...form.getInputProps(path),
  }
  switch (field.type) {
    case 'long_text':
      return <Textarea {...common} autosize minRows={3} maxRows={12} />
    case 'number':
      return <TextInput {...common} inputMode="decimal" description="Con coma o punto para los decimales." />
    case 'date':
      return <TextInput {...common} type="date" />
    case 'option':
      return (
        <Select
          {...common}
          data={field.options.map((option) => ({ value: option.id, label: option.name }))}
          placeholder="Elegí una opción…"
          searchable={field.options.length > searchableOptions}
          nothingFoundMessage="No hay opciones con ese texto."
        />
      )
    case 'subject_list':
      return (
        <TagsInput
          {...common}
          description={`Escribí cada asignatura y apretá Enter. Hasta ${maxSubjects}.`}
          splitChars={[]}
          maxTags={maxSubjects}
          clearable
        />
      )
    default:
      return <TextInput {...common} />
  }
}

function DocumentInput({ form, document }: { form: Form; document: DocumentRequirement }) {
  const path = `files.${document.id}`
  return (
    <FileInput
      label={document.name}
      name={document.id}
      description={documentDescription(document)}
      withAsterisk={document.required}
      accept={document.accepted_types.join(',')}
      placeholder="Elegí un archivo…"
      clearable
      clearButtonProps={{ 'aria-label': `Quitar el archivo de ${document.name}` }}
      key={form.key(path)}
      {...form.getInputProps(path)}
    />
  )
}

// SummaryItem puts the name above its value, so that a long name and a long
// value do not share a line.
function SummaryItem({ name, children }: { name: string; children: ReactNode }) {
  return (
    <Stack gap={0}>
      <Text size="sm" c="dimmed">
        {name}
      </Text>
      {children}
    </Stack>
  )
}

function Summary({ form, entry }: { form: Form; entry: CatalogEntry }) {
  const submission = form.getValues()
  const { fields, documents } = entry.submission
  if (fields.length === 0 && documents.length === 0) {
    return <Text c="dimmed">Este trámite no pide datos ni documentos.</Text>
  }
  return (
    <Stack gap="lg">
      {fields.length > 0 && (
        <Stack gap="sm">
          <Title order={2}>Datos</Title>
          {fields.map((field) => {
            const value = submission.values[field.id]
            return (
              <SummaryItem key={field.id} name={field.name}>
                {field.type === 'long_text' && !isEmpty(value) ? (
                  <LongValue text={valueText(field, value)} />
                ) : (
                  <Text>{valueText(field, value)}</Text>
                )}
              </SummaryItem>
            )
          })}
        </Stack>
      )}
      {documents.length > 0 && (
        <Stack gap="sm">
          <Title order={2}>Documentos</Title>
          {documents.map((document) => (
            <SummaryItem key={document.id} name={document.name}>
              <Text>{submission.files[document.id]?.name ?? 'Sin adjuntar'}</Text>
            </SummaryItem>
          ))}
        </Stack>
      )}
    </Stack>
  )
}

// StartedNotice takes the focus, since the button that had it is gone.
function StartedNotice({ procedure }: { procedure: StartedProcedure }) {
  const alert = useRef<HTMLDivElement>(null)
  useEffect(() => {
    alert.current?.focus()
  }, [])
  return (
    <Stack gap="md">
      <Alert title="Tu trámite quedó iniciado" tabIndex={-1} ref={alert}>
        <Text>
          Número de trámite: <Code translate="no">{procedure.id}</Code>
        </Text>
        <Text>Guardalo para seguirlo. Ya quedó para que el personal lo revise.</Text>
      </Alert>
      <Group>
        <Button component={Link} to={`/procedures/${encodeURIComponent(procedure.id)}`}>
          Ver el trámite
        </Button>
        <Button component={Link} to="/" variant="default">
          Volver al inicio
        </Button>
      </Group>
    </Stack>
  )
}

// LeaveWarning asks before leaving with something filled in, since nothing is
// kept until the submission is sent.
function LeaveWarning({ dirty }: { dirty: boolean }) {
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) => dirty && currentLocation.pathname !== nextLocation.pathname,
  )
  useEffect(() => {
    if (!dirty) {
      return
    }
    const warn = (event: BeforeUnloadEvent) => event.preventDefault()
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [dirty])

  return (
    <Modal opened={blocker.state === 'blocked'} onClose={() => blocker.reset?.()} title="¿Salir sin enviar el trámite?">
      <Stack gap="md">
        <Text>Se pierde lo que cargaste: los datos y los archivos.</Text>
        <Group justify="flex-end">
          <Button variant="default" onClick={() => blocker.proceed?.()}>
            Salir sin enviar
          </Button>
          <Button onClick={() => blocker.reset?.()}>Seguir completando</Button>
        </Group>
      </Stack>
    </Modal>
  )
}

function StartProcedureForm({ entry }: { entry: CatalogEntry }) {
  const { fields, documents } = entry.submission
  const [active, setActive] = useState(stageData)
  const [sending, setSending] = useState(false)
  const [failure, setFailure] = useState<string | null>(null)
  const [started, setStarted] = useState<StartedProcedure | null>(null)
  const form = useForm<SubmissionDraft>({
    mode: 'uncontrolled',
    initialValues: emptySubmission(fields, documents),
    validate: (submission) => ({
      ...Object.fromEntries(
        fields.map((field) => [`values.${field.id}`, fieldError(field, submission.values[field.id])]),
      ),
      ...Object.fromEntries(
        documents.map((document) => [`files.${document.id}`, fileError(document, submission.files[document.id])]),
      ),
    }),
  })

  // After the stage changes, the focus goes to the field to fix, if there is
  // one, or else to the content of the new stage, so that a screen reader
  // reads it. Only the active stage is mounted, so the three share the ref.
  const content = useRef<HTMLDivElement>(null)
  const pendingFocus = useRef<string | null | undefined>(undefined)
  useEffect(() => {
    if (pendingFocus.current === undefined) {
      return
    }
    const path = pendingFocus.current
    pendingFocus.current = undefined
    const node = path ? form.getInputNode(path) : null
    ;(node ?? content.current)?.focus()
  })

  function goToStage(stage: number, path: string | null = null) {
    pendingFocus.current = path
    setActive(stage)
  }

  const pathsOf = [fields.map((field) => `values.${field.id}`), documents.map((document) => `files.${document.id}`), []]

  function firstError(errors: Record<string, unknown>) {
    for (const [stage, paths] of pathsOf.entries()) {
      const path = paths.find((candidate) => errors[candidate])
      if (path) {
        return { stage, path }
      }
    }
    return null
  }

  // validateStage checks the fields or the files of a stage, and the files
  // together against the most a request can carry. It returns the first
  // path to fix.
  function validateStage(stage: number) {
    // Every path is validated, so that all that is wrong shows at once.
    const results = pathsOf[stage].map((path) => ({ path, hasError: form.validateField(path).hasError }))
    const invalid = results.find((result) => result.hasError)
    if (invalid) {
      return invalid.path
    }
    if (stage === stageDocuments) {
      const total = totalSizeErrors(documents, form.getValues().files)
      const [path] = Object.keys(total)
      if (path) {
        form.setErrors(total)
        return path
      }
    }
    return null
  }

  function next() {
    const invalid = validateStage(active)
    if (invalid) {
      form.getInputNode(invalid)?.focus()
      return
    }
    goToStage(active + 1)
  }

  async function submit() {
    const validation = form.validate()
    const errors = { ...validation.errors, ...totalSizeErrors(documents, form.getValues().files) }
    const invalid = firstError(errors)
    if (invalid) {
      form.setErrors(errors)
      goToStage(invalid.stage, invalid.path)
      return
    }
    setFailure(null)
    setSending(true)
    try {
      const { data, error, response } = await api.POST('/procedures', {
        body: { procedure_type: entry.id, values: '' },
        bodySerializer: () => submissionBody(entry.id, fields, form.getValues()),
      })
      if (data) {
        setStarted(data)
        return
      }
      if (response.status === 400 && error && 'problems' in error && error.problems.length > 0) {
        const rejected = problemErrors(error.problems)
        form.setErrors(rejected)
        const first = firstError(rejected)
        goToStage(first?.stage ?? active, first?.path ?? null)
        return
      }
      // A request the server could not read comes with a message that says
      // why, such as one past the most it takes.
      const message = response.status === 400 && error && 'message' in error ? error.message : ''
      setFailure(message || failureText(response.status))
    } catch {
      setFailure('No pudimos enviar el trámite. Revisá tu conexión y volvé a intentarlo.')
    } finally {
      setSending(false)
    }
  }

  if (started) {
    return <StartedNotice procedure={started} />
  }

  return (
    <Stack gap="lg">
      <LeaveWarning dirty={form.isDirty()} />
      <Text>Completá lo que pide el trámite. No se envía nada hasta que confirmes el resumen.</Text>
      <Stepper active={active} onStepClick={(stage) => goToStage(stage)} allowNextStepsSelect={false}>
        <Stepper.Step label="Datos">
          <Stack gap="md" mt="md" ref={content} tabIndex={-1}>
            {fields.length === 0 && <Text c="dimmed">Este trámite no pide datos.</Text>}
            {fields.map((field) => (
              <FieldInput key={field.id} form={form} field={field} />
            ))}
          </Stack>
        </Stepper.Step>
        <Stepper.Step label="Documentos">
          <Stack gap="md" mt="md" ref={content} tabIndex={-1}>
            {documents.length === 0 && <Text c="dimmed">Este trámite no pide documentos.</Text>}
            {documents.map((document) => (
              <DocumentInput key={document.id} form={form} document={document} />
            ))}
          </Stack>
        </Stepper.Step>
        <Stepper.Step label="Resumen">
          <Stack gap="md" mt="md" ref={content} tabIndex={-1}>
            <Summary form={form} entry={entry} />
          </Stack>
        </Stepper.Step>
      </Stepper>
      {failure && (
        <Alert color="red" role="alert">
          {failure}
        </Alert>
      )}
      <Group justify="space-between">
        {active > stageData ? (
          <Button variant="default" onClick={() => goToStage(active - 1)} disabled={sending}>
            Volver
          </Button>
        ) : (
          <Button component={Link} to="/catalog" variant="subtle">
            Cancelar
          </Button>
        )}
        {active < stageSummary ? (
          <Button onClick={next}>{active === stageDocuments ? 'Ver el resumen' : 'Continuar'}</Button>
        ) : (
          <Button onClick={() => void submit()} loading={sending}>
            Enviar el trámite
          </Button>
        )}
      </Group>
    </Stack>
  )
}

export function StartProcedurePage() {
  const { procedureTypeId } = useParams()
  const catalog = useQuery(catalogQuery)
  const entry = catalog.data?.find((candidate) => String(candidate.id) === procedureTypeId)

  return (
    <Stack gap="lg">
      <Title order={1}>{entry ? entry.name : 'Iniciar un trámite'}</Title>
      {catalog.isPending && <Loader aria-label="Cargando el trámite…" />}
      {catalog.isError && (
        <Alert color="red" role="alert">
          No pudimos cargar el trámite. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {catalog.isSuccess && !entry && (
        <Text>
          No existe ese trámite. Elegí uno del{' '}
          <Anchor component={Link} to="/catalog">
            catálogo de trámites
          </Anchor>
          .
        </Text>
      )}
      {entry && <StartProcedureForm entry={entry} />}
    </Stack>
  )
}
