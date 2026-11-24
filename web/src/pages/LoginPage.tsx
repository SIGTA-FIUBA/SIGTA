import { Alert, Button, Image, Paper, SimpleGrid, Stack, Text, Title } from '@mantine/core'
import { Fragment, useId, useRef, type MouseEvent } from 'react'
import { Link, useSearchParams } from 'react-router'
import type { components } from '../api/schema'
import { loginUrl } from '../api/session'
import { Icon, type IconName } from '../components/Icon/Icon'
import { principles, staffHowTo } from '../content/platform'
import { glide, isPlainClick } from '../layout/glide'

type LoginError = components['schemas']['LoginError']

// The callback only says why a login failed; the text lives here, so the
// language of the interface does not depend on the backend.
const loginErrorTexts: Record<LoginError, string> = {
  state_mismatch: 'El ingreso no coincide con el que se inició. Volvé a intentarlo.',
  rejected: 'El ingreso no se pudo completar. Volvé a intentarlo.',
  domain_not_allowed: 'Ingresá con tu cuenta institucional.',
}

function loginErrorText(error: string | null): string | undefined {
  if (error === null) {
    return undefined
  }
  return Object.hasOwn(loginErrorTexts, error) ? loginErrorTexts[error as LoginError] : loginErrorTexts.rejected
}

// lines is where the design breaks the title: the number of words of its
// first line.
type Offer = {
  icon: IconName
  title: string
  lines: number
  text: string
  tone: 'outline' | 'cyan' | 'navy'
  to: string
}

// What a visitor can do with SIGTA, each card with its way in: the inbox, for
// the staff, asks for the login, the catalog is public and the administration
// is the page about SIGTA.
const offers: Offer[] = [
  {
    icon: 'file-pencil',
    title: 'Bandeja de trámites',
    lines: 1,
    text: 'Para el personal: revisá los trámites en curso.',
    tone: 'outline',
    to: loginUrl,
  },
  {
    icon: 'search',
    title: 'Catálogo de trámites',
    lines: 1,
    text: 'Requisitos y plazos de cada trámite.',
    tone: 'cyan',
    to: '/catalog',
  },
  {
    icon: 'settings-check',
    title: 'Administración de SIGTA',
    lines: 1,
    text: 'A cargo de la Secretaría de Gestión Académica.',
    tone: 'navy',
    to: '/about',
  },
]

// Lines writes a title in the lines of the design, which break it where
// they end; between them a space keeps the words apart for a screen reader.
function Lines({ lines }: { lines: string[] }) {
  return lines.map((line, index) => (
    <Fragment key={line}>
      {index > 0 && ' '}
      <Text component="span" inherit variant="landing-line">
        {line}
      </Text>
    </Fragment>
  ))
}

// The title in two lines, the first one of words words.
const twoLines = (text: string, words: number) => {
  const all = text.split(' ')
  return [all.slice(0, words).join(' '), all.slice(words).join(' ')]
}

// The first lines of the titles of the steps, as the design breaks them.
const stepLines = [1, 2, 1]

// The whole card is its way in, as it looks: "Ver más", in its notch, only
// says so. The link is named by its title and described by its text, as they
// read.
function OfferCard({ offer }: { offer: Offer }) {
  const id = useId()
  return (
    <Paper
      component={Link}
      to={offer.to}
      reloadDocument={offer.to === loginUrl}
      variant="landing-card"
      mod={{ tone: offer.tone }}
      aria-labelledby={`${id}-title`}
      aria-describedby={`${id}-text`}
    >
      <Paper variant="landing-card-icon" aria-hidden>
        <Icon name={offer.icon} size="fill" weight={1.6} />
      </Paper>
      <Title order={3} variant="landing-card-title" id={`${id}-title`}>
        <Lines lines={twoLines(offer.title, offer.lines)} />
      </Title>
      <Text variant="landing-card-text" id={`${id}-text`}>
        {offer.text}
      </Text>
      <Text component="span" variant="landing-more" aria-hidden>
        Ver más
      </Text>
    </Paper>
  )
}

// The photographs of the hero, among the shapes of the brand. The headline
// says what the page is about, so a screen reader skips it; the three round
// shapes look like buttons, and for the pointer they are: the plus opens the
// catalog, where a procedure starts, the arrow the login and the cap the page
// about SIGTA. The keyboard has those links already, so they are not tab
// stops.
function Collage() {
  return (
    <Paper variant="landing-collage" aria-hidden>
      <Image variant="landing-hall" src="/photos/portada-pasillo.webp" alt="" />
      <Image variant="landing-facade" src="/photos/portada-fachada.webp" alt="" />
      <Paper component={Link} to="/about" variant="landing-cap" tabIndex={-1} title="Sobre SIGTA">
        <Icon name="school" size="fill" weight={1.75} />
      </Paper>
      <Paper component={Link} to="/catalog" variant="landing-plus" tabIndex={-1} title="Catálogo de trámites">
        <Icon name="plus" size="fill" weight={2.5} />
      </Paper>
      <Image variant="landing-mate" src="/photos/portada-mate.webp" alt="" />
      <Paper component="a" href={loginUrl} variant="landing-arrow" tabIndex={-1} title="Iniciar sesión">
        <Icon name="arrow-right" size="fill" weight={2} />
      </Paper>
    </Paper>
  )
}

// LoginPage is the public landing page, drawn after its design (SIGTA.ai),
// under the header of PublicLayout: the headline
// beside a collage of the faculty, what SIGTA offers in three cards, how the
// staff works in it and the criteria it follows. Its look is its own
// (landing.module.css, the landing-* variants), within the FIUBA manual:
// Helvetica Neue, its blue and cyan, flat, and rounded blocks.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = loginErrorText(params.get('error'))
  // What SIGTA offers: "Conocé SIGTA" and the band glide to it, the band
  // under the header.
  const offersSection = useRef<HTMLDivElement>(null)
  const toOffers = (event: MouseEvent) => {
    if (isPlainClick(event)) {
      event.preventDefault()
      glide(offersSection.current)
    }
  }

  return (
    <>
      <Stack component="section" variant="landing-hero" aria-labelledby="hero">
        <Stack variant="landing-hero-text">
          <Title order={1} id="hero" variant="landing-headline">
            Una nueva forma <br />
            de hacer{' '}
            <Text component="span" variant="landing-pill">
              trámites
            </Text>
          </Title>
          <Text variant="landing-lead">
            <strong>SIGTA</strong> nace para que hacer trámites académicos <br />
            sea más sencillo, desde el inicio hasta su finalización.
          </Text>
          {message && (
            <Stack variant="landing-notice">
              <Alert color="red">{message}</Alert>
            </Stack>
          )}
          <Button component="a" href="#que-podes-hacer" variant="landing-discover" onClick={toOffers}>
            Conocé SIGTA
          </Button>
        </Stack>
        <Collage />
      </Stack>

      <Stack ref={offersSection} component="section" gap={0} variant="landing-offers" aria-labelledby="que-podes-hacer">
        {/* The band looks like it leads below, and does: its arrow is a link to
            the cards, and the click anywhere on the band is the arrow's. */}
        <Paper variant="landing-banner">
          <Title order={2} id="que-podes-hacer" variant="landing-banner-title">
            ¿Qué podés hacer con SIGTA?
          </Title>
          <Paper
            component="a"
            href="#que-podes-hacer"
            variant="landing-chevron"
            aria-label="Ver qué podés hacer con SIGTA"
            onClick={toOffers}
          >
            <Icon name="chevron-down" size="fill" weight={2.5} />
          </Paper>
        </Paper>
        <SimpleGrid variant="landing-cards" cols={{ base: 1, md: 3 }}>
          {offers.map((offer) => (
            <OfferCard key={offer.title} offer={offer} />
          ))}
        </SimpleGrid>
      </Stack>

      <Stack component="section" variant="landing-how" aria-labelledby="como-se-trabaja">
        <Paper variant="landing-how-photo">
          <Title order={2} id="como-se-trabaja" variant="landing-how-title">
            <Lines lines={['¿Cómo', 'se trabaja', 'en SIGTA?']} />
          </Title>
        </Paper>
        <Paper variant="landing-how-steps">
          <Stack component="ol" variant="landing-steps">
            {staffHowTo.map((step, index) => (
              <Stack key={step.title} component="li" variant="landing-step">
                <Text component="span" variant="landing-step-number" aria-hidden>
                  {index + 1}
                </Text>
                <Title order={3} variant="landing-step-title">
                  <Lines lines={twoLines(step.title, stepLines[index] ?? 1)} />
                </Title>
                <Text variant="landing-step-text">{step.text}</Text>
              </Stack>
            ))}
          </Stack>
        </Paper>
      </Stack>

      <Stack component="section" gap={0} aria-labelledby="criterios">
        <Title order={2} id="criterios" variant="landing-criteria-title">
          Criterios de la plataforma
        </Title>
        <Stack component="ul" variant="landing-criteria">
          {principles.map((principle) => (
            <Paper key={principle.title} component="li" variant="landing-criterion">
              <Title order={3} variant="landing-criterion-title">
                {principle.title}
              </Title>
              <Text variant="landing-criterion-text">{principle.text}</Text>
            </Paper>
          ))}
        </Stack>
      </Stack>
    </>
  )
}
