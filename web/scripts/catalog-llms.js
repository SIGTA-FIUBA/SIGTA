// Completes the built catalog for agents: copies the tokens, the design rules
// and the logo next to it, and writes llms.txt, a plain text index of all of
// it built from Storybook's index.json, so it lists every story there is.
// Usage: node scripts/catalog-llms.js <catalog output folder>
import { copyFileSync, mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const out = process.argv[2]
const web = join(import.meta.dirname, '..')

copyFileSync(join(web, 'src/theme/tokens.json'), join(out, 'tokens.json'))
copyFileSync(join(web, 'DESIGN.md'), join(out, 'DESIGN.md'))
mkdirSync(join(out, 'brand'), { recursive: true })
const logos = readdirSync(join(web, 'public/brand')).filter((file) => file.endsWith('.svg'))
for (const file of logos) {
  copyFileSync(join(web, 'public/brand', file), join(out, 'brand', file))
}

const { entries } = JSON.parse(readFileSync(join(out, 'index.json'), 'utf8'))
const sections = new Map()
for (const entry of Object.values(entries)) {
  const [section, ...rest] = entry.title.split('/')
  const page = rest.join('/') || section
  if (!sections.has(section)) {
    sections.set(section, new Map())
  }
  const pages = sections.get(section)
  if (!pages.has(page)) {
    pages.set(page, [])
  }
  pages.get(page).push(entry)
}

const link = (entry) => `iframe.html?id=${entry.id}&viewMode=${entry.type === 'docs' ? 'docs' : 'story'}`
const lines = [
  '# Catálogo de SIGTA',
  '',
  '> Identidad visual y componentes de la interfaz de SIGTA (Sistema Integral de Gestión de Trámites Académicos, FIUBA). La interfaz usa Mantine 9 con el tema de SIGTA; todo el aspecto sale de tokens.json.',
  '',
  'Reglas para construir pantallas: no escribir colores ni medidas sueltas, no usar style ni className, usar los componentes de este catálogo con valores del tema (tamaños xs a xl, colores por nombre). Detalle en DESIGN.md.',
  '',
  '## Fuentes',
  '',
  '- [tokens.json](tokens.json): colores, tipografías, tamaños, espaciado, radios, anchos de pantalla y decisiones por componente, en formato de design tokens del W3C. En el repositorio: web/src/theme/tokens.json.',
  '- [DESIGN.md](DESIGN.md): reglas de la interfaz y su motivo.',
  '- Componentes de base: Mantine 9, documentación para agentes en https://mantine.dev/llms.txt',
  '',
  '## Logo',
  '',
  'Archivos sigta-<pieza>-<versión>.svg. Piezas: isotipo, logotipo, horizontal, horizontal-descriptor, vertical, vertical-descriptor. Versiones: color (fondo claro), color-dark (fondo oscuro), black, white.',
  '',
  ...logos.map((file) => `- [${file}](brand/${file})`),
  '',
]
for (const [section, pages] of sections) {
  lines.push(`## ${section}`, '')
  for (const [page, pageEntries] of pages) {
    const stories = pageEntries.filter((entry) => entry.type === 'story')
    const docs = pageEntries.find((entry) => entry.type === 'docs')
    const target = docs ?? stories[0]
    const names = stories.map((entry) => entry.name).join(', ')
    lines.push(`- [${page}](${link(target)})${names ? `: ${names}` : ''}`)
  }
  lines.push('')
}
writeFileSync(join(out, 'llms.txt'), lines.join('\n'))
console.log(`llms.txt: ${Object.keys(entries).length} entries`)
