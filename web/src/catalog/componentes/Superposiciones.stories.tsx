import {
  ActionIcon,
  Affix,
  Anchor,
  Avatar,
  Badge,
  Button,
  CloseButton,
  Dialog,
  Drawer,
  FloatingWindow,
  Group,
  HoverCard,
  LoadingOverlay,
  Menu,
  Modal,
  Overlay,
  Paper,
  Popover,
  Stack,
  Text,
  Textarea,
  Tooltip,
  Tour,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { Icon } from '../../components/Icon/Icon'

// Components drawn over the page. Each story opens its overlay from a button
// and starts closed, so the docs page stays usable.

const meta: Meta = {
  title: 'Componentes/Superposiciones',
  parameters: {
    docs: {
      description: {
        component:
          'Ventanas, paneles y avisos que se dibujan sobre la página. Cada uno se abre con una acción de la persona y se cierra con `Escape` o con su botón de cerrar.',
      },
    },
  },
}

export default meta

function ModalDemo() {
  const [opened, { open, close }] = useDisclosure(false)
  return (
    <>
      <Button variant="default" onClick={open}>
        Enviar solicitud
      </Button>
      <Modal opened={opened} onClose={close} title="Enviar la solicitud">
        <Stack>
          <Text size="sm">
            Una vez enviada, la Secretaría revisa la documentación y la solicitud ya no se puede modificar.
          </Text>
          <Group justify="flex-end">
            <Button variant="default" onClick={close}>
              Cancelar
            </Button>
            <Button onClick={close}>Enviar</Button>
          </Group>
        </Stack>
      </Modal>
    </>
  )
}

export const ModalStory: StoryObj = {
  name: 'Modal',
  render: () => <ModalDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'Confirma una acción que no se puede deshacer, como enviar una solicitud. Lleva una sola acción principal.',
      },
    },
  },
}

function DrawerDemo() {
  const [opened, { open, close }] = useDisclosure(false)
  return (
    <>
      <Button variant="default" onClick={open}>
        Ver detalle
      </Button>
      <Drawer opened={opened} onClose={close} position="right" title="Trámite EQ-2026-00042">
        <Stack gap="sm">
          <Badge variant="light">En un área</Badge>
          <Text size="sm">Equivalencia entre asignaturas</Text>
          <Text size="sm" c="dimmed">
            Titular: Martina Ferreyra
          </Text>
          <Text size="sm" c="dimmed">
            Presentado el 3 de noviembre
          </Text>
        </Stack>
      </Drawer>
    </>
  )
}

export const DrawerStory: StoryObj = {
  name: 'Drawer',
  render: () => <DrawerDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Un panel lateral para ver el detalle de un elemento sin dejar la lista, como un trámite de la bandeja.',
      },
    },
  },
}

function DialogDemo() {
  const [opened, { open, close }] = useDisclosure(false)
  return (
    <>
      <Button variant="default" onClick={open}>
        Guardar borrador
      </Button>
      <Dialog opened={opened} onClose={close} withCloseButton size="lg" radius="sm">
        <Text size="sm">Se guardó el borrador de la solicitud. Podés retomarla desde el inicio.</Text>
      </Dialog>
    </>
  )
}

export const DialogStory: StoryObj = {
  name: 'Dialog',
  render: () => <DialogDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Un aviso fijo en una esquina que no bloquea la página. Para un aviso pasajero, una notificación.',
      },
    },
  },
}

export const PopoverStory: StoryObj = {
  name: 'Popover',
  render: () => (
    <Popover position="bottom-start" shadow="md">
      <Popover.Target>
        <Button variant="default">Ver requisitos</Button>
      </Popover.Target>
      <Popover.Dropdown>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            Requisitos
          </Text>
          <Text size="sm">Certificado analítico y programa de la asignatura.</Text>
        </Stack>
      </Popover.Dropdown>
    </Popover>
  ),
  parameters: {
    docs: {
      description: { story: 'Información breve que se abre con un clic, junto al control que la explica.' },
    },
  },
}

export const HoverCardStory: StoryObj = {
  name: 'HoverCard',
  render: () => (
    <HoverCard position="bottom-start" shadow="md">
      <HoverCard.Target>
        <Anchor href="#" size="sm">
          Martina Ferreyra
        </Anchor>
      </HoverCard.Target>
      <HoverCard.Dropdown>
        <Group gap="sm" wrap="nowrap">
          <Avatar name="Martina Ferreyra" color="fiuba" aria-hidden />
          <Stack gap={0}>
            <Text size="sm" fw={600}>
              Martina Ferreyra
            </Text>
            <Text size="xs" c="dimmed">
              Ingeniería en Informática
            </Text>
          </Stack>
        </Group>
      </HoverCard.Dropdown>
    </HoverCard>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Una vista previa al pasar el mouse, como los datos del titular en la bandeja. No guarda nada que no esté también en la página.',
      },
    },
  },
}

export const TooltipStory: StoryObj = {
  name: 'Tooltip',
  render: () => (
    <Tooltip label="Ir al inicio">
      <ActionIcon variant="default" size="lg" aria-label="Ir al inicio">
        <Icon name="home" />
      </ActionIcon>
    </Tooltip>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Nombra un botón que solo tiene un ícono. No reemplaza al `aria-label`, que el botón lleva igual.',
      },
    },
  },
}

export const MenuStory: StoryObj = {
  name: 'Menu',
  render: () => (
    <Menu position="bottom-start">
      <Menu.Target>
        <Avatar component="button" name="Laura Méndez" aria-label="Cuenta de Laura Méndez" />
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>
          <Text size="sm" c="bright">
            Laura Méndez
          </Text>
          <Text size="xs" c="dimmed">
            laura.mendez@ejemplo.edu.ar
          </Text>
        </Menu.Label>
        <Menu.Divider />
        <Menu.Item>Cerrar sesión</Menu.Item>
      </Menu.Dropdown>
    </Menu>
  ),
  parameters: {
    docs: {
      description: {
        story: 'El menú de la cuenta al pie del menú lateral (`UserMenu`): quién tiene la sesión y cómo cerrarla.',
      },
    },
  },
}

function OverlayDemo() {
  const [visible, { toggle }] = useDisclosure(false)
  return (
    <Stack align="flex-start">
      <Button variant="default" onClick={toggle}>
        {visible ? 'Quitar la capa' : 'Mostrar la capa'}
      </Button>
      <Paper withBorder p="lg" pos="relative">
        <Text size="sm">Reconocimiento de créditos, presentado el 28 de octubre.</Text>
        {visible && <Overlay backgroundOpacity={0.35} radius="sm" />}
      </Paper>
    </Stack>
  )
}

export const OverlayStory: StoryObj = {
  name: 'Overlay',
  render: () => <OverlayDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'Cubre una parte de la página para que no se pueda usar. Las pantallas usan `LoadingOverlay`, que ya lo incluye.',
      },
    },
  },
}

function LoadingOverlayDemo() {
  const [visible, { toggle }] = useDisclosure(false)
  return (
    <Stack align="flex-start">
      <Button variant="default" onClick={toggle}>
        {visible ? 'Terminar la carga' : 'Simular la carga'}
      </Button>
      <Paper withBorder p="lg" pos="relative">
        <LoadingOverlay visible={visible} />
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            Pase de carrera
          </Text>
          <Text size="sm" c="dimmed">
            Pendiente del titular
          </Text>
        </Stack>
      </Paper>
    </Stack>
  )
}

export const LoadingOverlayStory: StoryObj = {
  name: 'LoadingOverlay',
  render: () => <LoadingOverlayDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Bloquea un panel mientras se guarda o se vuelve a cargar, sin quitar lo que ya se mostraba.',
      },
    },
  },
}

function AffixDemo() {
  const [visible, { toggle, close }] = useDisclosure(false)
  return (
    <>
      <Button variant="default" onClick={toggle}>
        {visible ? 'Ocultar el botón fijo' : 'Mostrar el botón fijo'}
      </Button>
      {visible && (
        <Affix position={{ bottom: 'xl', right: 'xl' }}>
          <Button onClick={close}>Volver arriba</Button>
        </Affix>
      )}
    </>
  )
}

export const AffixStory: StoryObj = {
  name: 'Affix',
  render: () => <AffixDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Fija un elemento a una esquina de la pantalla, como un botón para volver arriba en una lista larga.',
      },
    },
  },
}

const tourLabels = {
  skip: 'Omitir',
  back: 'Anterior',
  next: 'Siguiente',
  close: 'Cerrar',
  stepCounter: (current: number, total: number) => `${current} de ${total}`,
  beacon: 'Iniciar el recorrido',
}

function TourDemo() {
  const [active, setActive] = useState(false)
  return (
    <Stack align="flex-start">
      <Button variant="default" onClick={() => setActive(true)}>
        Recorrer la pantalla
      </Button>
      <Group>
        <Button variant="light" leftSection={<Icon name="home" size={16} />} data-tour-step="home">
          Inicio
        </Button>
        <Button variant="light" leftSection={<Icon name="inbox" size={16} />} data-tour-step="inbox">
          Bandeja de trámites
        </Button>
        <Button variant="light" leftSection={<Icon name="catalog" size={16} />} data-tour-step="catalog">
          Catálogo de trámites
        </Button>
      </Group>
      <Tour active={active} onClose={() => setActive(false)} labels={tourLabels}>
        <Tour.Step target='[data-tour-step="home"]' title="Inicio">
          El resumen de los trámites en curso.
        </Tour.Step>
        <Tour.Step target='[data-tour-step="inbox"]' title="Bandeja de trámites">
          Los trámites a cargo de la Secretaría, con los que tienen acción requerida primero.
        </Tour.Step>
        <Tour.Step target='[data-tour-step="catalog"]' title="Catálogo de trámites">
          Los tipos de trámite que se pueden iniciar y sus requisitos.
        </Tour.Step>
      </Tour>
    </Stack>
  )
}

export const TourStory: StoryObj = {
  name: 'Tour',
  render: () => <TourDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Un recorrido guiado por las partes de una pantalla, para presentar la plataforma la primera vez.',
      },
    },
  },
}

function FloatingWindowDemo() {
  const [opened, { toggle, close }] = useDisclosure(false)
  return (
    <>
      <Button variant="default" onClick={toggle}>
        {opened ? 'Cerrar las notas' : 'Abrir las notas'}
      </Button>
      {opened && (
        <FloatingWindow
          initialPosition={{ top: 96, right: 32 }}
          dimensions={{ initialWidth: 320 }}
          dragHandleSelector="[data-drag-handle]"
          excludeDragHandleSelector="button"
          withBorder
          shadow="md"
          p="md"
        >
          <Stack gap="xs">
            <Group justify="space-between" wrap="nowrap" data-drag-handle>
              <Text size="sm" fw={600}>
                Notas de la revisión
              </Text>
              <CloseButton aria-label="Cerrar las notas" onClick={close} />
            </Group>
            <Text size="xs" c="dimmed">
              Arrastrá la ventana desde el título.
            </Text>
            <Textarea label="Nota" placeholder="Falta el programa de Física I" autosize minRows={3} />
          </Stack>
        </FloatingWindow>
      )}
    </>
  )
}

export const FloatingWindowStory: StoryObj = {
  name: 'FloatingWindow',
  render: () => <FloatingWindowDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'Una ventana que se puede mover y queda sobre la página, como notas que se toman mientras se revisa un trámite.',
      },
    },
  },
}
