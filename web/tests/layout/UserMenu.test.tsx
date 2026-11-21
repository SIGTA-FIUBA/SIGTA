import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { UserMenu } from '../../src/layout/UserMenu'
import { render } from '../test-utils/render'

const user = { id: 'U-1', name: 'Ana Pérez', email: 'aperez@fi.uba.ar', staff: true, administrator: false }

describe('UserMenu', () => {
  it.each([false, true])('says that the account opens a menu, compact: %s', async (compact) => {
    render(<UserMenu user={user} onLogout={() => {}} compact={compact} />)

    const account = screen.getByRole('button', { name: 'Cuenta de Ana Pérez' })
    expect(account.getAttribute('aria-haspopup')).toBe('menu')
    expect(account.getAttribute('aria-expanded')).toBe('false')

    await userEvent.click(account)

    expect(account.getAttribute('aria-expanded')).toBe('true')
    expect(screen.getByText('aperez@fi.uba.ar')).toBeDefined()
  })

  it('closes the session from the menu', async () => {
    const onLogout = vi.fn()
    render(<UserMenu user={user} onLogout={onLogout} />)

    await userEvent.click(screen.getByRole('button', { name: 'Cuenta de Ana Pérez' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Cerrar sesión' }))

    expect(onLogout).toHaveBeenCalledOnce()
  })
})
