import { Avatar, Menu, NavLink, Text, Tooltip } from '@mantine/core'
import type { components } from '../api/schema'
import { Icon } from '../components/Icon/Icon'

type User = components['schemas']['User']

type UserMenuProps = {
  user: User
  onLogout: () => void
  // compact shows only the initials, as the menu does on a tablet.
  compact?: boolean
}

const profileNames: Record<User['role'], string> = {
  applicant: 'Titular',
  staff_operative: 'Perfil operativo',
  staff_approver: 'Perfil de aprobación',
}

// UserMenu closes the side navigation: the initials, the name and the
// profile of whoever has the session. Unfolded, it shows the email and the
// way to close the session.
export function UserMenu({ user, onLogout, compact = false }: UserMenuProps) {
  const label = `Cuenta de ${user.name}`
  const avatar = <Avatar name={user.name} aria-hidden />
  return (
    <Menu position={compact ? 'right-end' : 'top-start'} width={compact ? undefined : 'target'}>
      <Menu.Target>
        {compact ? (
          <Tooltip label={user.name} position="right">
            <NavLink component="button" variant="menu-icon" aria-label={label} leftSection={avatar} />
          </Tooltip>
        ) : (
          <NavLink
            component="button"
            variant="menu"
            aria-label={label}
            leftSection={avatar}
            label={user.name}
            description={profileNames[user.role]}
            rightSection={<Icon name="selector" size={16} />}
          />
        )}
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>
          <Text size="sm" c="bright" truncate="end">
            {user.name}
          </Text>
          <Text size="xs" c="dimmed" truncate="end">
            {user.email}
          </Text>
        </Menu.Label>
        <Menu.Divider />
        <Menu.Item onClick={onLogout}>Cerrar sesión</Menu.Item>
      </Menu.Dropdown>
    </Menu>
  )
}
