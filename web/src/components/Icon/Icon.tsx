import { tokens } from '../../theme/tokens'

// Icon draws an outline icon in the color of its text. The paths follow the
// stroke style of Tabler Icons (MIT), on a 24 grid; the stroke is scaled so it
// measures layout.iconStroke at any size.
const paths = {
  home: 'M5 12l-2 0l9 -9l9 9l-2 0 M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7 M9 21v-6a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v6',
  inbox: 'M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z M4 13h3l3 3h4l3 -3h3',
  administration: 'M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3',
  catalog: 'M3 19a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6l0 13 M12 6l0 13 M21 6l0 13',
  procedureTypes:
    'M5 5m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M19 12m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M5 19m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M7 5h5a3 3 0 0 1 3 3v2 M7 19h5a3 3 0 0 0 3 -3v-2',
  'chevron-right': 'M9 6l6 6l-6 6',
  'chevron-left': 'M15 6l-6 6l6 6',
  'chevron-down': 'M6 9l6 6l6 -6',
  selector: 'M8 9l4 -4l4 4 M16 15l-4 4l-4 -4',
  info: 'M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0 M12 9h.01 M11 12h1v4h1',
  // The icons of the landing page.
  school: 'M22 9l-10 -4l-10 4l10 4l10 -4v6 M6 10.6v5.4a6 3 0 0 0 12 0v-5.4',
  plus: 'M12 5l0 14 M5 12l14 0',
  'arrow-right': 'M5 12l14 0 M13 18l6 -6 M13 6l6 6',
  files:
    'M15 3v4a1 1 0 0 0 1 1h4 M18 17h-7a2 2 0 0 1 -2 -2v-10a2 2 0 0 1 2 -2h4l5 5v7a2 2 0 0 1 -2 2z M16 17v2a2 2 0 0 1 -2 2h-7a2 2 0 0 1 -2 -2v-10a2 2 0 0 1 2 -2h2',
  download: 'M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2 -2v-2 M7 11l5 5l5 -5 M12 4l0 12',
  history: 'M12 8l0 4l2 2 M3.05 11a9 9 0 1 1 .5 4m-.5 5v-5h5',
  'file-text':
    'M14 3v4a1 1 0 0 0 1 1h4 M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z M9 9l1 0 M9 13l6 0 M9 17l6 0',
  search: 'M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0 M21 21l-6 -6',
  'file-pencil':
    'M14 3v4a1 1 0 0 0 1 1h4 M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z M8 11h5 M8 14h3 M10 18l5 -5a1.414 1.414 0 0 0 -2 -2l-5 5v2h2z',
  'settings-check':
    'M10.325 4.317c.426 -1.756 2.924 -1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066c1.543 -.94 3.31 .826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756 .426 1.756 2.924 0 3.35a1.724 1.724 0 0 0 -1.066 2.573c.94 1.543 -.826 3.31 -2.37 2.37a1.724 1.724 0 0 0 -2.572 1.065c-.426 1.756 -2.924 1.756 -3.35 0a1.724 1.724 0 0 0 -2.573 -1.066c-1.543 .94 -3.31 -.826 -2.37 -2.37a1.724 1.724 0 0 0 -1.065 -2.572c-1.756 -.426 -1.756 -2.924 0 -3.35a1.724 1.724 0 0 0 1.066 -2.573c-.94 -1.543 .826 -3.31 2.37 -2.37c1 .608 2.296 .07 2.572 -1.065z M8 12a4 4 0 1 0 8 0a4 4 0 1 0 -8 0 M10.5 12l1 1l2 -2',
}

export type IconName = keyof typeof paths

// 20 in the navigation, 16 next to text. "fill" takes the size of its box,
// which the theme sets, as the landing page does; its stroke is then weight,
// on the 24 grid, since there is no size to scale it to.
export type IconSize = 16 | 20 | 'fill'

export function Icon({ name, size = 20, weight = 2 }: { name: IconName; size?: IconSize; weight?: number }) {
  return (
    <svg
      width={size === 'fill' ? '100%' : size}
      height={size === 'fill' ? '100%' : size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={size === 'fill' ? weight : (tokens.layout.iconStroke * 24) / size}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={paths[name]} />
    </svg>
  )
}
