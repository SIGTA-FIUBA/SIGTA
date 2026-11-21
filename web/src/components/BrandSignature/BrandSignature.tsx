import { Divider, Group, Image } from '@mantine/core'
import fiubaLogo from '../../assets/fiuba-logo.png'

// BrandSignature is the joint signature of SIGTA and the faculty: the SIGTA
// logo, a divider and the FIUBA logo, always on white. It goes where SIGTA
// presents itself (the landing page, its footer and "Acerca de SIGTA"); inside
// the app the SIGTA logo alone is enough. compact leaves out the FIUBA logo on
// a narrow phone, where it would push the actions of a top bar off the screen.
export function BrandSignature({ compact = false }: { compact?: boolean }) {
  const visibleFrom = compact ? 'xs' : undefined
  return (
    <Group gap="md" wrap="nowrap">
      <Image src="/brand/sigta-horizontal-color.svg" alt="SIGTA" h="var(--sigta-logo-height)" w="auto" />
      <Divider orientation="vertical" visibleFrom={visibleFrom} />
      <Image
        src={fiubaLogo}
        alt="Facultad de Ingeniería, Universidad de Buenos Aires"
        h="var(--sigta-signature-height)"
        w="auto"
        visibleFrom={visibleFrom}
      />
    </Group>
  )
}
