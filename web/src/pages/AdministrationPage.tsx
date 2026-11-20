import {
  Alert,
  Badge,
  Button,
  Divider,
  Group,
  Loader,
  MultiSelect,
  Paper,
  Select,
  Stack,
  Tabs,
  Text,
  TextInput,
  Title,
} from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { useQuery } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { useRevalidator, useSearchParams } from 'react-router'
import {
  actions,
  peopleSearchMinimum,
  useAdministrationChange,
  useAdministrators,
  useAreaMemberships,
  useAreas,
  usePeopleSearch,
} from '../api/administration'
import { catalogQuery } from '../api/catalog'
import { ForbiddenError } from '../api/procedures'
import type { components } from '../api/schema'

type Area = components['schemas']['Area']
type AreaMembership = components['schemas']['AreaMembership']
type Administrator = components['schemas']['Administrator']
type Role = components['schemas']['Role']
type RegisteredPerson = components['schemas']['RegisteredPerson']

// focusOnMount moves the focus to what appears in place of what was clicked,
// so the keyboard does not lose its place. It is stable, so it runs only when
// the element mounts.
const focusOnMount = (node: HTMLElement | null) => node?.focus()

// ConfirmButton asks again before a change that cannot be undone from the
// screen, like ending a membership. The question stays while the change runs,
// so it cannot be confirmed twice, and after a refusal, next to the reason;
// onConfirm gets done, to close it when the change goes through.
function ConfirmButton({
  label,
  question,
  confirmLabel,
  ariaLabel,
  pending,
  onConfirm,
}: {
  label: string
  question: string
  confirmLabel: string
  ariaLabel: string
  pending: boolean
  onConfirm: (done: () => void) => void
}) {
  const [asking, setAsking] = useState(false)
  if (!asking) {
    return (
      <Button variant="subtle" size="xs" flex="none" aria-label={ariaLabel} onClick={() => setAsking(true)}>
        {label}
      </Button>
    )
  }
  return (
    <Group gap="xs" justify="flex-end">
      <Text size="sm">{question}</Text>
      <Button variant="default" size="xs" loading={pending} onClick={() => onConfirm(() => setAsking(false))}>
        {confirmLabel}
      </Button>
      <Button variant="subtle" size="xs" ref={focusOnMount} disabled={pending} onClick={() => setAsking(false)}>
        Cancelar
      </Button>
    </Group>
  )
}

// PersonSearch picks someone who entered SIGTA, by part of the name or email,
// to grant them a membership or the administration. The server searches as
// they type and answers a few people, never the whole list.
function PersonSearch({
  person,
  onChange,
  focused = false,
}: {
  person: RegisteredPerson | null
  onChange: (person: RegisteredPerson | null) => void
  focused?: boolean
}) {
  const [text, setText] = useState('')
  const [debounced] = useDebouncedValue(text, 300)
  const search = usePeopleSearch(debounced)
  const people = search.data ?? []
  const searched = debounced.trim().length >= peopleSearchMinimum && search.isSuccess

  if (person !== null) {
    return (
      <Group gap="sm">
        <Stack gap={0}>
          <Text fw={600}>{person.name}</Text>
          <Text size="sm" c="dimmed">
            {person.email}
          </Text>
        </Stack>
        <Button variant="subtle" size="xs" onClick={() => onChange(null)}>
          Elegir otra persona
        </Button>
      </Group>
    )
  }
  return (
    <Select
      label="Persona"
      description="Escribí parte del nombre o del correo con el que ingresó a SIGTA."
      placeholder="Nombre o correo"
      searchable
      autoComplete="off"
      ref={focused ? focusOnMount : undefined}
      searchValue={text}
      onSearchChange={setText}
      data={people.map((p) => ({ value: p.id, label: p.name || p.id }))}
      filter={({ options }) => options}
      value={null}
      onChange={(id) => onChange(people.find((p) => p.id === id) ?? null)}
      renderOption={({ option }) => {
        const found = people.find((p) => p.id === option.value)
        return (
          <Stack gap={0}>
            <Text size="sm">{option.label}</Text>
            <Text size="xs" c="dimmed">
              {found?.email}
            </Text>
          </Stack>
        )
      }}
      nothingFoundMessage={
        searched
          ? 'Nadie que haya ingresado a SIGTA coincide. Tiene que ingresar una vez con su cuenta de la UBA.'
          : undefined
      }
      rightSection={search.isFetching ? <Loader size="xs" /> : undefined}
      error={search.isError ? 'No pudimos buscar a la persona. Revisá tu conexión y volvé a intentarlo.' : undefined}
    />
  )
}

const dateFormat = new Intl.DateTimeFormat('es-AR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  timeZone: 'America/Argentina/Buenos_Aires',
})

// A membership is valid until the end of the day picked, in Buenos Aires.
function endOfDay(date: string) {
  return new Date(`${date}T23:59:59-03:00`).toISOString()
}

// today is the date in Buenos Aires as YYYY-MM-DD, the first day a membership
// can end on.
function today() {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(new Date())
}

function names(ids: string[], options: { id: string; name: string }[]) {
  return ids.map((id) => options.find((option) => option.id === id)?.name ?? id).join(', ')
}

function failureOf(error: Error | null) {
  return error?.message ?? null
}

function useProcedureTypes() {
  const catalog = useQuery(catalogQuery)
  const types = catalog.data ?? []
  return {
    options: types.map((type) => ({ value: String(type.id), label: type.name })),
    name: (id: number) => types.find((type) => type.id === id)?.name ?? `Tipo de trámite ${id}`,
  }
}

function PermissionsFields({
  area,
  roles,
  permissions,
  onRoles,
  onPermissions,
}: {
  area: Area
  roles: string[]
  permissions: string[]
  onRoles: (roles: string[]) => void
  onPermissions: (permissions: string[]) => void
}) {
  return (
    <>
      <MultiSelect
        label="Roles"
        placeholder={area.roles.length === 0 ? 'El área no tiene roles' : 'Elegí uno o más…'}
        data={area.roles.map((role) => ({ value: role.id, label: role.name }))}
        value={roles}
        onChange={onRoles}
        disabled={area.roles.length === 0}
      />
      <MultiSelect
        label="Permisos sueltos"
        description="Además de los que dan los roles."
        placeholder="Elegí uno o más…"
        data={area.permissions.map((permission) => ({ value: permission.id, label: permission.name }))}
        value={permissions}
        onChange={onPermissions}
      />
    </>
  )
}

function MembershipCard({ area, membership }: { area: Area; membership: AreaMembership }) {
  const procedureTypes = useProcedureTypes()
  const [editing, setEditing] = useState(false)
  const [roles, setRoles] = useState(membership.roles)
  const [permissions, setPermissions] = useState(membership.permissions)
  const change = useAdministrationChange(() => actions.changeMembership(membership.id, { roles, permissions }))
  const end = useAdministrationChange(() => actions.endMembership(membership.id))
  const person = membership.person.name || membership.person.id

  return (
    <Paper withBorder p="sm">
      <Stack gap="xs">
        <Stack gap={0}>
          <Text fw={600}>{person}</Text>
          <Text size="sm" c="dimmed">
            {membership.person.email}
          </Text>
        </Stack>
        {editing ? (
          <Stack gap="xs">
            <PermissionsFields
              area={area}
              roles={roles}
              permissions={permissions}
              onRoles={setRoles}
              onPermissions={setPermissions}
            />
            <Text size="sm" c="dimmed">
              La membresía actual termina y empieza otra con los permisos nuevos, para que quede registrado qué tuvo en
              cada momento.
            </Text>
            <Group gap="xs">
              <Button
                variant="default"
                loading={change.isPending}
                onClick={() => change.mutate(undefined, { onSuccess: () => setEditing(false) })}
              >
                Guardar permisos
              </Button>
              <Button
                variant="subtle"
                onClick={() => {
                  setEditing(false)
                  setRoles(membership.roles)
                  setPermissions(membership.permissions)
                  change.reset()
                }}
              >
                Cancelar
              </Button>
            </Group>
          </Stack>
        ) : (
          <Stack gap={0}>
            {membership.roles.length > 0 && <Text size="sm">Roles: {names(membership.roles, area.roles)}</Text>}
            {membership.permissions.length > 0 && (
              <Text size="sm">Permisos sueltos: {names(membership.permissions, area.permissions)}</Text>
            )}
            <Text size="sm">
              Alcance:{' '}
              {membership.scope.length === 0
                ? 'todos los tipos de trámite'
                : membership.scope.map(procedureTypes.name).join(', ')}
            </Text>
            <Text size="sm" c="dimmed">
              Desde el {dateFormat.format(new Date(membership.from))}
              {membership.until ? ` hasta el ${dateFormat.format(new Date(membership.until))}` : ', sin fecha de fin'}.
              Otorgada por {membership.granted_by.name || 'una persona que ya no está registrada'}.
            </Text>
          </Stack>
        )}
        {!editing && (
          <Group gap="xs">
            <Button
              variant="subtle"
              size="xs"
              aria-label={`Cambiar los permisos de ${person}`}
              onClick={() => setEditing(true)}
            >
              Cambiar permisos
            </Button>
            <ConfirmButton
              label="Terminar"
              ariaLabel={`Terminar la membresía de ${person}`}
              question="¿Terminar la membresía ahora?"
              confirmLabel="Terminar membresía"
              pending={end.isPending}
              onConfirm={(done) => end.mutate(undefined, { onSuccess: done })}
            />
          </Group>
        )}
        {(change.error || end.error) && (
          <Alert color="red" role="alert">
            {failureOf(change.error) ?? failureOf(end.error)}
          </Alert>
        )}
      </Stack>
    </Paper>
  )
}

function GrantMembership({ area, onDone }: { area: Area; onDone: () => void }) {
  const procedureTypes = useProcedureTypes()
  const [person, setPerson] = useState<RegisteredPerson | null>(null)
  const [roles, setRoles] = useState<string[]>([])
  const [permissions, setPermissions] = useState<string[]>([])
  const [scope, setScope] = useState<string[]>([])
  const [until, setUntil] = useState('')
  const grant = useAdministrationChange((userId: string) =>
    actions.grantMembership({
      user_id: userId,
      area: area.id,
      roles,
      permissions,
      scope: scope.map(Number),
      until: until === '' ? undefined : endOfDay(until),
    }),
  )

  return (
    <Stack gap="xs">
      <Title order={4}>Otorgar una membresía</Title>
      <PersonSearch person={person} onChange={setPerson} focused />
      {person !== null && (
        <>
          <PermissionsFields
            area={area}
            roles={roles}
            permissions={permissions}
            onRoles={setRoles}
            onPermissions={setPermissions}
          />
          <MultiSelect
            label="Alcance"
            description="Vacío vale para todos los tipos de trámite, incluidos los que se publiquen después."
            placeholder="Todos los tipos de trámite"
            data={procedureTypes.options}
            value={scope}
            onChange={setScope}
          />
          <TextInput
            label="Hasta"
            description="Opcional. Sin fecha, vale hasta que se la termine."
            type="date"
            min={today()}
            value={until}
            onChange={(event) => setUntil(event.currentTarget.value)}
          />
          {grant.error && (
            <Alert color="red" role="alert">
              {grant.error.message}
            </Alert>
          )}
        </>
      )}
      <Group>
        {person !== null && (
          <Button
            variant="default"
            loading={grant.isPending}
            onClick={() => grant.mutate(person.id, { onSuccess: onDone })}
          >
            Otorgar membresía
          </Button>
        )}
        <Button variant="subtle" onClick={onDone}>
          Cancelar
        </Button>
      </Group>
    </Stack>
  )
}

// MembershipsSection lists who is in the area now or later, and grants,
// changes and ends their memberships.
function MembershipsSection({ area }: { area: Area }) {
  const memberships = useAreaMemberships(area.id)
  return (
    <Stack gap="sm">
      <Title order={3}>Membresías</Title>
      {memberships.isPending && <Loader aria-label={`Cargando las membresías de ${area.name}…`} />}
      {memberships.isError && (
        <Alert color="red" role="alert">
          No pudimos cargar las membresías. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {memberships.isSuccess && memberships.data.length === 0 && (
        <Text c="dimmed">Nadie tiene una membresía en esta área.</Text>
      )}
      {memberships.data?.map((membership) => (
        <MembershipCard key={membership.id} area={area} membership={membership} />
      ))}
      {/* GrantMembership holds its own change, so closing it forgets a refusal. */}
      <Reveal label="Otorgar una membresía">{(close) => <GrantMembership area={area} onDone={close} />}</Reveal>
    </Stack>
  )
}

const idDescription = 'Fijo. En minúsculas, dígitos y guiones bajos, empezando con una letra.'

// Reveal keeps a form folded behind a button, so the area reads first as what
// it has. The form gets close, for its Cancelar and for when it is done, and
// onClose forgets what a failed attempt said.
function Reveal({
  label,
  onClose,
  children,
}: {
  label: string
  onClose?: () => void
  children: (close: () => void) => ReactNode
}) {
  const [open, setOpen] = useState(false)
  const [closed, setClosed] = useState(false)
  const close = () => {
    setOpen(false)
    setClosed(true)
    onClose?.()
  }
  if (!open) {
    return (
      <Group>
        <Button variant="light" ref={closed ? focusOnMount : undefined} onClick={() => setOpen(true)}>
          {label}
        </Button>
      </Group>
    )
  }
  return (
    <Paper withBorder p="sm">
      {children(close)}
    </Paper>
  )
}

// EntryRow shows a permission or a role: its name, its fixed identifier and
// what else it says, with the action to the right.
function EntryRow({
  name,
  id,
  children,
  action,
}: {
  name: string
  id: string
  children?: ReactNode
  action?: ReactNode
}) {
  return (
    <Group justify="space-between" align="flex-start" gap="xs" wrap="nowrap">
      <Stack gap={0} miw={0}>
        <Text>{name}</Text>
        <Text size="xs" c="dimmed" translate="no">
          {id}
        </Text>
        {children}
      </Stack>
      {action}
    </Group>
  )
}

// NewEntry is the form for something with an identifier and a name: an area,
// a permission or a role.
function NewEntry({
  title,
  titleOrder = 4,
  submitLabel,
  primary = false,
  extra,
  onSubmit,
  onCancel,
  pending,
  failure,
}: {
  title: string
  titleOrder?: 2 | 4
  submitLabel: string
  primary?: boolean
  extra?: ReactNode
  onSubmit: (entry: { id: string; name: string }) => void
  onCancel: () => void
  pending: boolean
  failure: string | null
}) {
  const [id, setId] = useState('')
  const [name, setName] = useState('')
  return (
    <form
      noValidate
      onSubmit={(event) => {
        event.preventDefault()
        onSubmit({ id: id.trim(), name: name.trim() })
      }}
    >
      <Stack gap="xs">
        <Title order={titleOrder}>{title}</Title>
        <TextInput
          label="Nombre"
          autoComplete="off"
          ref={focusOnMount}
          value={name}
          onChange={(event) => setName(event.currentTarget.value)}
        />
        <TextInput
          label="Identificador"
          description={idDescription}
          autoComplete="off"
          spellCheck={false}
          value={id}
          onChange={(event) => setId(event.currentTarget.value)}
        />
        {extra}
        {failure && (
          <Alert color="red" role="alert">
            {failure}
          </Alert>
        )}
        <Group>
          <Button type="submit" variant={primary ? 'filled' : 'default'} loading={pending}>
            {submitLabel}
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Cancelar
          </Button>
        </Group>
      </Stack>
    </form>
  )
}

function Permissions({ area }: { area: Area }) {
  const create = useAdministrationChange((permission: { id: string; name: string }) =>
    actions.createPermission(area.id, permission),
  )
  const remove = useAdministrationChange((permissionId: string) => actions.deletePermission(area.id, permissionId))
  return (
    <Stack gap="sm">
      <Title order={3}>Permisos</Title>
      {area.permissions.length === 0 && <Text c="dimmed">El área todavía no tiene permisos.</Text>}
      {area.permissions.map((permission) => (
        <EntryRow
          key={permission.id}
          name={permission.name}
          id={permission.id}
          action={
            !permission.system && (
              <ConfirmButton
                label="Dar de baja"
                ariaLabel={`Dar de baja el permiso ${permission.name}`}
                question="¿Dar de baja el permiso?"
                confirmLabel="Dar de baja"
                pending={remove.isPending}
                onConfirm={(done) => remove.mutate(permission.id, { onSuccess: done })}
              />
            )
          }
        >
          {permission.system && (
            <Group>
              <Badge variant="light">De plataforma</Badge>
            </Group>
          )}
        </EntryRow>
      ))}
      {remove.error && (
        <Alert color="red" role="alert">
          {remove.error.message}
        </Alert>
      )}
      <Reveal label="Agregar un permiso" onClose={create.reset}>
        {(close) => (
          <NewEntry
            title="Agregar un permiso"
            submitLabel="Agregar permiso"
            pending={create.isPending}
            failure={create.error?.message ?? null}
            onSubmit={(permission) => create.mutate(permission, { onSuccess: close })}
            onCancel={close}
          />
        )}
      </Reveal>
    </Stack>
  )
}

// RoleRow shows a role with what it grants, and changes its permissions in
// place, like MembershipCard.
function RoleRow({
  area,
  role,
  removing,
  onRemove,
}: {
  area: Area
  role: Role
  removing: boolean
  onRemove: (done: () => void) => void
}) {
  const [editing, setEditing] = useState(false)
  // A permission that was dropped from the area is not offered, so saving
  // leaves it out.
  const active = (ids: string[]) => ids.filter((id) => area.permissions.some((permission) => permission.id === id))
  const [permissions, setPermissions] = useState(() => active(role.permissions))
  const change = useAdministrationChange(() => actions.changeRole(area.id, role.id, { permissions }))
  const cancel = () => {
    setEditing(false)
    setPermissions(active(role.permissions))
    change.reset()
  }
  return (
    <Stack gap="xs">
      <EntryRow
        name={role.name}
        id={role.id}
        action={
          !editing && (
            <Group gap="xs" justify="flex-end">
              <Button
                variant="subtle"
                size="xs"
                aria-label={`Cambiar los permisos del rol ${role.name}`}
                onClick={() => {
                  setPermissions(active(role.permissions))
                  setEditing(true)
                }}
              >
                Cambiar permisos
              </Button>
              <ConfirmButton
                label="Dar de baja"
                ariaLabel={`Dar de baja el rol ${role.name}`}
                question="¿Dar de baja el rol?"
                confirmLabel="Dar de baja"
                pending={removing}
                onConfirm={onRemove}
              />
            </Group>
          )
        }
      >
        {!editing && (
          <Text size="sm" c="dimmed">
            Permisos: {names(role.permissions, area.permissions)}
          </Text>
        )}
      </EntryRow>
      {editing && (
        <Stack gap="xs">
          <MultiSelect
            label="Permisos del rol"
            placeholder="Elegí uno o más…"
            data={area.permissions.map((permission) => ({ value: permission.id, label: permission.name }))}
            value={permissions}
            onChange={setPermissions}
            ref={focusOnMount}
          />
          <Text size="sm" c="dimmed">
            Cambia al instante lo que puede hacer cada persona con este rol.
          </Text>
          {change.error && (
            <Alert color="red" role="alert">
              {change.error.message}
            </Alert>
          )}
          <Group gap="xs">
            <Button
              variant="default"
              loading={change.isPending}
              onClick={() => change.mutate(undefined, { onSuccess: () => setEditing(false) })}
            >
              Guardar permisos
            </Button>
            <Button variant="subtle" disabled={change.isPending} onClick={cancel}>
              Cancelar
            </Button>
          </Group>
        </Stack>
      )}
    </Stack>
  )
}

function Roles({ area }: { area: Area }) {
  const [permissions, setPermissions] = useState<string[]>([])
  const create = useAdministrationChange((role: { id: string; name: string }) =>
    actions.createRole(area.id, { ...role, permissions }),
  )
  const remove = useAdministrationChange((roleId: string) => actions.deleteRole(area.id, roleId))
  return (
    <Stack gap="sm">
      <Title order={3}>Roles</Title>
      {area.roles.length === 0 && <Text c="dimmed">El área todavía no tiene roles.</Text>}
      {area.roles.map((role) => (
        <RoleRow
          key={role.id}
          area={area}
          role={role}
          removing={remove.isPending}
          onRemove={(done) => remove.mutate(role.id, { onSuccess: done })}
        />
      ))}
      {remove.error && (
        <Alert color="red" role="alert">
          {remove.error.message}
        </Alert>
      )}
      <Reveal label="Agregar un rol" onClose={create.reset}>
        {(close) => (
          <NewEntry
            title="Agregar un rol"
            submitLabel="Agregar rol"
            pending={create.isPending}
            failure={create.error?.message ?? null}
            extra={
              <MultiSelect
                label="Permisos del rol"
                placeholder="Elegí uno o más…"
                data={area.permissions.map((permission) => ({ value: permission.id, label: permission.name }))}
                value={permissions}
                onChange={setPermissions}
              />
            }
            onSubmit={(role) =>
              create.mutate(role, {
                onSuccess: () => {
                  close()
                  setPermissions([])
                },
              })
            }
            onCancel={() => {
              close()
              setPermissions([])
            }}
          />
        )}
      </Reveal>
    </Stack>
  )
}

// AreaPanel shows an area with its permissions, roles and memberships.
function AreaPanel({ area }: { area: Area }) {
  const remove = useAdministrationChange(() => actions.deleteArea(area.id))
  return (
    <Paper withBorder p="md" component="section" aria-label={area.name}>
      <Stack gap="md">
        <Stack gap={0}>
          <Group gap="xs">
            <Title order={2}>{area.name}</Title>
            {area.system && <Badge variant="light">Viene con SIGTA</Badge>}
          </Group>
          <Text size="sm" c="dimmed" translate="no">
            {area.id}
          </Text>
        </Stack>
        {!area.system && (
          <Group>
            <ConfirmButton
              label="Dar de baja el área"
              ariaLabel={`Dar de baja el área ${area.name}`}
              question="¿Dar de baja el área?"
              confirmLabel="Dar de baja"
              pending={remove.isPending}
              onConfirm={(done) => remove.mutate(undefined, { onSuccess: done })}
            />
          </Group>
        )}
        {remove.error && (
          <Alert color="red" role="alert">
            {remove.error.message}
          </Alert>
        )}
        <Permissions area={area} />
        <Divider />
        <Roles area={area} />
        <Divider />
        <MembershipsSection area={area} />
      </Stack>
    </Paper>
  )
}

const notAdministrator = 'Solo una persona administradora puede ver esta página.'

function Areas() {
  const areas = useAreas()
  const create = useAdministrationChange(actions.createArea)
  if (areas.isPending) {
    return <Loader aria-label="Cargando las áreas…" />
  }
  if (areas.isError) {
    return (
      <Alert color="red" role="alert">
        {areas.error instanceof ForbiddenError
          ? notAdministrator
          : 'No pudimos cargar las áreas. Volvé a intentarlo más tarde.'}
      </Alert>
    )
  }
  return (
    <Stack gap="lg">
      {areas.data.map((area) => (
        <AreaPanel key={area.id} area={area} />
      ))}
      <Reveal label="Crear un área" onClose={create.reset}>
        {(close) => (
          <NewEntry
            title="Crear un área"
            titleOrder={2}
            submitLabel="Crear área"
            primary
            pending={create.isPending}
            failure={create.error?.message ?? null}
            onSubmit={(area) => create.mutate(area, { onSuccess: close })}
            onCancel={close}
          />
        )}
      </Reveal>
    </Stack>
  )
}

function AdministratorRow({ administrator, last }: { administrator: Administrator; last: boolean }) {
  const end = useAdministrationChange(() => actions.endAdministrator(administrator.id))
  // Whoever removes themselves stops being an administrator: reading /me
  // again, through the loaders, takes Administración off the menu.
  const revalidator = useRevalidator()
  const person = administrator.person.name || administrator.person.id
  return (
    <Paper withBorder p="sm">
      <Stack gap="xs">
        <Group justify="space-between" align="flex-start" gap="xs" wrap="nowrap">
          <Stack gap={0} miw={0}>
            <Text fw={600}>{person}</Text>
            <Text size="sm" c="dimmed">
              {administrator.person.email}
            </Text>
            <Text size="sm" c="dimmed">
              Desde el {dateFormat.format(new Date(administrator.from))}.{' '}
              {administrator.granted_by
                ? `Otorgada por ${administrator.granted_by.name || 'una persona que ya no está registrada'}.`
                : 'Primera persona administradora, dada al instalar SIGTA.'}
            </Text>
          </Stack>
          {last ? (
            <Badge variant="light">Única persona administradora</Badge>
          ) : (
            <ConfirmButton
              label="Quitar"
              ariaLabel={`Quitarle la administración a ${person}`}
              question="¿Quitarle la administración?"
              confirmLabel="Quitar administración"
              pending={end.isPending}
              onConfirm={(done) =>
                end.mutate(undefined, {
                  onSuccess: () => {
                    done()
                    void revalidator.revalidate()
                  },
                })
              }
            />
          )}
        </Group>
        {end.error && (
          <Alert color="red" role="alert">
            {end.error.message}
          </Alert>
        )}
      </Stack>
    </Paper>
  )
}

function Administrators() {
  const administrators = useAdministrators()
  const [person, setPerson] = useState<RegisteredPerson | null>(null)
  const grant = useAdministrationChange(actions.grantAdministrator)
  if (administrators.isPending) {
    return <Loader aria-label="Cargando las personas administradoras…" />
  }
  if (administrators.isError) {
    return (
      <Alert color="red" role="alert">
        {administrators.error instanceof ForbiddenError
          ? notAdministrator
          : 'No pudimos cargar las personas administradoras. Volvé a intentarlo más tarde.'}
      </Alert>
    )
  }
  return (
    <Stack gap="lg">
      <Text c="dimmed">Administrar SIGTA no da acceso a los trámites: para eso hace falta una membresía.</Text>
      <Stack gap="sm">
        {administrators.data.map((administrator) => (
          <AdministratorRow
            key={administrator.id}
            administrator={administrator}
            last={administrators.data.length === 1}
          />
        ))}
      </Stack>
      <Paper withBorder p="md">
        <Stack gap="xs">
          <Title order={2}>Agregar una persona administradora</Title>
          <PersonSearch person={person} onChange={setPerson} />
          {grant.error && (
            <Alert color="red" role="alert">
              {grant.error.message}
            </Alert>
          )}
          {person !== null && (
            <Group>
              <Button
                loading={grant.isPending}
                onClick={() => grant.mutate(person.id, { onSuccess: () => setPerson(null) })}
              >
                Hacer administradora
              </Button>
            </Group>
          )}
        </Stack>
      </Paper>
    </Stack>
  )
}

// AdministrationPage is where an administrator manages the areas of the staff,
// their permissions, roles and memberships, and who administers SIGTA. The
// API checks each change; the page only shows what it answers.
export function AdministrationPage() {
  const [params, setParams] = useSearchParams()
  const tab = params.get('tab') === 'administrators' ? 'administrators' : 'areas'
  return (
    <Stack gap="lg">
      <Stack gap={0}>
        <Title order={1}>Administración</Title>
        <Text c="dimmed">Áreas del personal, sus permisos, roles y membresías, y quiénes administran SIGTA.</Text>
      </Stack>
      <Tabs
        value={tab}
        keepMounted={false}
        onChange={(value) => {
          const next = new URLSearchParams(params)
          if (value === 'administrators') {
            next.set('tab', value)
          } else {
            next.delete('tab')
          }
          setParams(next, { replace: true })
        }}
      >
        <Tabs.List>
          <Tabs.Tab value="areas">Áreas y membresías</Tabs.Tab>
          <Tabs.Tab value="administrators">Personas administradoras</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="areas" pt="lg">
          <Areas />
        </Tabs.Panel>
        <Tabs.Panel value="administrators" pt="lg">
          <Administrators />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
