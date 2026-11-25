import { Alert, Badge, Button, Flex, Group, Loader, Modal, Stack, Text, TextInput } from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router'
import { draftActions, useDraftChange, useEditableProcedureTypes } from '../api/procedureTypes'
import type { components } from '../api/schema'
import { draftStatus } from '../content/draftStatuses'
import { PageHeader } from '../components/PageHeader/PageHeader'

type EditableProcedureType = components['schemas']['EditableProcedureType']

const dateFormat = new Intl.DateTimeFormat('es-AR', { dateStyle: 'short' })

// ProcedureTypesPage lists the procedure types the person can edit or
// approve, with the draft of each one, and starts a new procedure type or the
// draft of a new version.
export function ProcedureTypesPage() {
  const editable = useEditableProcedureTypes()
  const [searchParams] = useSearchParams()
  const published = searchParams.get('publicada')

  return (
    <Stack gap="lg">
      <PageHeader title="Tipos de trámite" action={editable.data?.can_create && <NewProcedureType />} />
      {published && (
        <Alert color="fiuba" title="Se publicó la versión nueva">
          Se publicó la versión {published}. Los trámites nuevos empiezan con ella; los que están en curso siguen con la
          suya.
        </Alert>
      )}
      {editable.isPending && <Loader aria-label="Cargando los tipos de trámite" />}
      {editable.isError && (
        <Alert color="red">No pudimos cargar los tipos de trámite. Volvé a intentarlo en unos minutos.</Alert>
      )}
      {editable.data && editable.data.procedure_types.length === 0 && (
        <Text c="dimmed">
          {editable.data.can_create
            ? 'Todavía no hay tipos de trámite. Creá el primero.'
            : 'No hay tipos de trámite que puedas editar o aprobar.'}
        </Text>
      )}
      {editable.data && editable.data.procedure_types.length > 0 && (
        <Stack component="ul" variant="facts" gap={0} p={0} m={0} aria-label="Tipos de trámite">
          {editable.data.procedure_types.map((type) => (
            <ProcedureTypeRow key={type.procedure_type_id} type={type} />
          ))}
        </Stack>
      )}
    </Stack>
  )
}

// ProcedureTypeRow is one procedure type, in a row like the ones of the
// administration: its name, its published version and its draft, with what
// the person can do on the right, or under them on a phone.
function ProcedureTypeRow({ type }: { type: EditableProcedureType }) {
  const navigate = useNavigate()
  const open = useDraftChange(draftActions.open)
  const draft = type.draft
  const status = draft && draftStatus(draft)
  const awaitsReview = draft?.status === 'proposed' && draft.can_review
  const draftPath = `/procedure-types/${type.procedure_type_id}/draft`

  return (
    <Stack component="li" gap="xs" py="md">
      <Flex direction={{ base: 'column', sm: 'row' }} justify="space-between" align="flex-start" gap="sm">
        <Stack gap="xxs" flex={{ sm: 1 }} miw={0}>
          <Text fw={600}>{type.name}</Text>
          <Flex wrap="wrap" align="center" rowGap="xxs" columnGap="md">
            <Text size="sm" c="gray.7">
              {type.version > 0 && type.published_at
                ? `Versión ${type.version}, publicada el ${dateFormat.format(new Date(type.published_at))}`
                : 'Sin publicar'}
            </Text>
            {draft && status ? (
              <Group gap="xs">
                <Text size="sm" c="gray.7">
                  Borrador
                </Text>
                <Badge variant="light" color={status.color}>
                  {status.label}
                </Badge>
              </Group>
            ) : (
              <Text size="sm" c="gray.7">
                Sin borrador
              </Text>
            )}
          </Flex>
          {awaitsReview && (
            <Text size="sm" fw={600} c="attention.7">
              Pendiente de aprobación
            </Text>
          )}
        </Stack>
        {draft && (
          <Button component={Link} to={draftPath} variant={awaitsReview ? 'light' : 'default'}>
            {awaitsReview ? 'Revisar propuesta' : draft.can_edit ? 'Abrir borrador' : 'Ver borrador'}
          </Button>
        )}
        {!draft && type.can_open_draft && (
          <Button
            variant="default"
            loading={open.isPending}
            onClick={() => open.mutate(type.procedure_type_id, { onSuccess: () => void navigate(draftPath) })}
          >
            Editar versión nueva
          </Button>
        )}
      </Flex>
      {open.isError && (
        <Alert color="red" role="alert">
          {open.error.message}
        </Alert>
      )}
    </Stack>
  )
}

// NewProcedureType asks for the name and creates the draft of a new procedure
// type, which opens for editing.
function NewProcedureType() {
  const [opened, { open, close }] = useDisclosure(false)
  const [name, setName] = useState('')
  const navigate = useNavigate()
  const create = useDraftChange(draftActions.create)

  function submit() {
    create.mutate(
      { name: name.trim(), steps: [] },
      {
        onSuccess: (draft) => {
          if (draft) {
            void navigate(`/procedure-types/${draft.procedure_type_id}/draft`)
          }
        },
      },
    )
  }

  return (
    <>
      <Button onClick={open}>Nuevo tipo de trámite</Button>
      <Modal opened={opened} onClose={close} title="Nuevo tipo de trámite">
        <form
          onSubmit={(event) => {
            event.preventDefault()
            submit()
          }}
        >
          <Stack>
            <TextInput
              label="Nombre"
              value={name}
              onChange={(event) => setName(event.currentTarget.value)}
              data-autofocus
            />
            {create.isError && <Alert color="red">{create.error.message}</Alert>}
            <Group justify="flex-end">
              <Button variant="default" onClick={close}>
                Cancelar
              </Button>
              <Button type="submit" disabled={name.trim() === ''} loading={create.isPending}>
                Crear el borrador
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </>
  )
}
