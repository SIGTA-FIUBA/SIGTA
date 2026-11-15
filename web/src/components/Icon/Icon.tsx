import { tokens } from '../../theme/tokens'

// Icon draws an outline icon in the color of its text. The paths follow the
// stroke style of Tabler Icons (MIT), on a 24 grid; the stroke is scaled so it
// measures layout.iconStroke at any size.
const paths = {
  home: 'M5 12l-2 0l9 -9l9 9l-2 0 M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7 M9 21v-6a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v6',
  inbox: 'M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z M4 13h3l3 3h4l3 -3h3',
  catalog: 'M3 19a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6l0 13 M12 6l0 13 M21 6l0 13',
  'chevron-right': 'M9 6l6 6l-6 6',
  'chevron-down': 'M6 9l6 6l6 -6',
}

export type IconName = keyof typeof paths

// 20 in the navigation, 16 next to text.
export type IconSize = 16 | 20

export function Icon({ name, size = 20 }: { name: IconName; size?: IconSize }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={(tokens.layout.iconStroke * 24) / size}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={paths[name]} />
    </svg>
  )
}
