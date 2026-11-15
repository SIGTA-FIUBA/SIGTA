import { Avatar, Group, Menu, Stack, Text, UnstyledButton } from '@mantine/core'
import type { components } from '../api/schema'
import { Icon } from '../components/Icon/Icon'

type User = components['schemas']['User']

type UserMenuProps = {
  user: User
  onLogout: () => void
}

const profileNames: Record<User['role'], string> = {
  applicant: 'Titular',
  staff_operative: 'Perfil operativo',
  staff_approver: 'Perfil de aprobación',
}

// UserMenu is the only place that says who has the session: the initials, the
// name and the profile, and, unfolded, the full name, the email and the way to
// close the session. On a phone only the initials show.
export function UserMenu({ user, onLogout }: UserMenuProps) {
  return (
    <Menu position="bottom-end">
      <Menu.Target>
        <UnstyledButton aria-label={`Cuenta de ${user.name}`}>
          <Group gap="xs" wrap="nowrap">
            <Avatar name={user.name} aria-hidden />
            <Stack gap={0} miw={0} visibleFrom="sm">
              <Text size="sm" truncate="end">
                {user.name}
              </Text>
              <Text size="xs" c="gray.6">
                {profileNames[user.role]}
              </Text>
            </Stack>
            <Text component="span" c="gray.6" display="flex" visibleFrom="sm">
              <Icon name="chevron-down" size={16} />
            </Text>
          </Group>
        </UnstyledButton>
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
