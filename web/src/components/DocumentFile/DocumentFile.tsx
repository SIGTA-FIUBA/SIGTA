import { Alert, Box, Button, Group, Image, Loader, Modal, Stack, Text, useMantineTheme } from '@mantine/core'
import { useEffect, useState } from 'react'
import { canPreview, mediaTypeOf, sizeText } from '../../documentFormats'

type AttachedFile = {
  id: string
  file_name: string
  // Absent where the API does not send them, as in the history.
  media_type?: string
  size?: number
}

type Preview =
  | { status: 'loading' }
  | { status: 'failed' }
  | { status: 'unsupported' }
  | { status: 'ready'; url: string; mediaType: string }

function contentUrl(procedureId: string, documentId: string) {
  return `/api/procedures/${encodeURIComponent(procedureId)}/documents/${encodeURIComponent(documentId)}/content`
}

// DocumentContent shows the file from a blob, so that the download keeps the
// headers that stop the browser from running it, and only for the types that
// the server says it read: the name of the file does not decide what is shown.
function DocumentContent({ procedureId, file }: { procedureId: string; file: AttachedFile }) {
  const theme = useMantineTheme()
  const [preview, setPreview] = useState<Preview>({ status: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    let url: string | undefined
    async function load() {
      try {
        const request = new Request(new URL(contentUrl(procedureId, file.id), window.location.origin), {
          signal: controller.signal,
        })
        const response = await globalThis.fetch(request)
        const mediaType = (response.headers.get('Content-Type') ?? '').split(';')[0]!.trim()
        if (!response.ok) {
          setPreview({ status: 'failed' })
          return
        }
        if (!canPreview(mediaType)) {
          setPreview({ status: 'unsupported' })
          return
        }
        const blob = new Blob([await response.arrayBuffer()], { type: mediaType })
        url = URL.createObjectURL(blob)
        setPreview({ status: 'ready', url, mediaType })
      } catch {
        if (!controller.signal.aborted) {
          setPreview({ status: 'failed' })
        }
      }
    }
    void load()
    return () => {
      controller.abort()
      if (url) {
        URL.revokeObjectURL(url)
      }
    }
  }, [procedureId, file.id])

  switch (preview.status) {
    case 'loading':
      return <Loader aria-label={`Cargando ${file.file_name}…`} />
    case 'failed':
      return (
        <Alert color="red" role="alert">
          No pudimos abrir el archivo. Volvé a intentarlo o descargalo.
        </Alert>
      )
    case 'unsupported':
      return <Text>Este archivo no se puede ver en el navegador. Descargalo para abrirlo.</Text>
    case 'ready':
      return preview.mediaType === 'application/pdf' ? (
        <Box
          component="iframe"
          src={preview.url}
          title={file.file_name}
          w="100%"
          h={theme.other.documentHeight}
          bd="none"
        />
      ) : (
        <Image src={preview.url} alt={file.file_name} fit="contain" mah={theme.other.documentHeight} />
      )
  }
}

// DocumentFile names an attached file and offers to see it in the page, when
// the browser can show it, and to download it.
export function DocumentFile({ procedureId, file }: { procedureId: string; file: AttachedFile }) {
  const [opened, setOpened] = useState(false)
  const previewable = canPreview(file.media_type ?? mediaTypeOf(file.file_name))
  const download = (
    <Button
      component="a"
      href={contentUrl(procedureId, file.id)}
      download={file.file_name}
      variant="subtle"
      size="compact-sm"
      aria-label={`Descargar ${file.file_name}`}
    >
      Descargar
    </Button>
  )
  return (
    <>
      <Group gap="xs" wrap="wrap">
        <Text size="sm" miw={0}>
          {file.file_name}
        </Text>
        {file.size !== undefined && (
          <Text size="sm" c="dimmed">
            {sizeText(file.size)}
          </Text>
        )}
        <Group gap={0}>
          {previewable && (
            <Button
              variant="subtle"
              size="compact-sm"
              aria-label={`Ver ${file.file_name}`}
              onClick={() => setOpened(true)}
            >
              Ver
            </Button>
          )}
          {download}
        </Group>
      </Group>
      <Modal opened={opened} onClose={() => setOpened(false)} title={file.file_name} fullScreen>
        <Stack gap="sm">
          <Group>{download}</Group>
          <DocumentContent procedureId={procedureId} file={file} />
        </Stack>
      </Modal>
    </>
  )
}
