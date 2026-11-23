import {
  Alert,
  Badge,
  Button,
  Group,
  Loader,
  Modal,
  Paper,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  VisuallyHidden,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router'
import { draftActions, useDraftChange, useEditableProcedureTypes } from '../api/procedureTypes'
import type { components } from '../api/schema'
import { draftStatus } from '../content/draftStatuses'

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
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <Title order={1}>Tipos de trámite</Title>
        {editable.data?.can_create && <NewProcedureType />}
      </Group>
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
        <Paper withBorder shadow="xs">
          <Table.ScrollContainer minWidth={640}>
            <Table verticalSpacing="sm" horizontalSpacing="md">
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Tipo de trámite</Table.Th>
                  <Table.Th>Versión publicada</Table.Th>
                  <Table.Th>Borrador</Table.Th>
                  <Table.Th>
                    <VisuallyHidden>Acción</VisuallyHidden>
                  </Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {editable.data.procedure_types.map((type) => (
                  <ProcedureTypeRow key={type.procedure_type_id} type={type} />
                ))}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>
        </Paper>
      )}
    </Stack>
  )
}

function ProcedureTypeRow({ type }: { type: EditableProcedureType }) {
  const navigate = useNavigate()
  const open = useDraftChange(draftActions.open)
  const draft = type.draft
  const status = draft && draftStatus(draft)
  const awaitsReview = draft?.status === 'proposed' && draft.can_review
  const draftPath = `/procedure-types/${type.procedure_type_id}/draft`

  return (
    <Table.Tr>
      <Table.Td>
        <Text fw={600}>{type.name}</Text>
      </Table.Td>
      <Table.Td>
        {type.version > 0 && type.published_at ? (
          <Text>
            Versión {type.version} · {dateFormat.format(new Date(type.published_at))}
          </Text>
        ) : (
          <Text c="dimmed">Sin publicar</Text>
        )}
      </Table.Td>
      <Table.Td>
        {draft && status ? (
          <Group gap="xs">
            <Badge variant="light" color={status.color}>
              {status.label}
            </Badge>
            {awaitsReview && (
              <Text size="sm" fw={600} c="fiuba.7">
                Pendiente de aprobación
              </Text>
            )}
          </Group>
        ) : (
          <Text c="dimmed">Sin borrador</Text>
        )}
      </Table.Td>
      <Table.Td ta="right">
        {draft && (
          <Button component={Link} to={draftPath} variant={awaitsReview ? 'light' : 'default'}>
            {awaitsReview ? 'Revisar propuesta' : draft.can_edit ? 'Abrir borrador' : 'Ver borrador'}
          </Button>
        )}
        {!draft && type.can_open_draft && (
          <Stack gap="xs" align="flex-end">
            <Button
              variant="default"
              loading={open.isPending}
              onClick={() => open.mutate(type.procedure_type_id, { onSuccess: () => void navigate(draftPath) })}
            >
              Editar versión nueva
            </Button>
            {open.isError && (
              <Text size="sm" c="red">
                {open.error.message}
              </Text>
            )}
          </Stack>
        )}
      </Table.Td>
    </Table.Tr>
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
