import { Avatar, Menu, Text } from '@mantine/core'
import type { components } from '../api/schema'

type User = components['schemas']['User']

type UserMenuProps = {
  user: User
  onLogout: () => void
}

// UserMenu shows who has the session, by their initials, and lets them close it.
export function UserMenu({ user, onLogout }: UserMenuProps) {
  return (
    <Menu position="bottom-end">
      <Menu.Target>
        <Avatar component="button" name={user.name} aria-label={`Cuenta de ${user.name}`} />
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>
          <Text size="sm" c="bright">
            {user.name}
          </Text>
          <Text size="xs" c="dimmed">
            {user.email}
          </Text>
        </Menu.Label>
        <Menu.Divider />
        <Menu.Item onClick={onLogout}>Cerrar sesión</Menu.Item>
      </Menu.Dropdown>
    </Menu>
  )
}
