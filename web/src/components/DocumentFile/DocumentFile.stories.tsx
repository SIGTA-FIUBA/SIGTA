import { Stack } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { DocumentFile } from './DocumentFile'

const meta: Meta<typeof DocumentFile> = {
  title: 'SIGTA/DocumentFile',
  component: DocumentFile,
  parameters: {
    docs: {
      description: {
        component:
          'Un archivo adjunto: su nombre, su tamaño y un icono para descargarlo. Si el navegador lo puede mostrar, el nombre lo abre en pantalla completa.',
      },
    },
  },
}

export default meta

export const Formats: StoryObj<typeof DocumentFile> = {
  name: 'Por formato',
  render: () => (
    <Stack gap="xs">
      <DocumentFile
        procedureId="1"
        file={{ id: '1', file_name: 'programa-analisis-ii.pdf', media_type: 'application/pdf', size: 482_000 }}
      />
      <DocumentFile
        procedureId="1"
        file={{ id: '2', file_name: 'certificado.png', media_type: 'image/png', size: 96_000 }}
      />
      <DocumentFile procedureId="1" file={{ id: '3', file_name: 'planilla.xlsx', size: 31_000 }} />
      <DocumentFile procedureId="1" file={{ id: '4', file_name: 'historial.pdf' }} />
    </Stack>
  ),
}
