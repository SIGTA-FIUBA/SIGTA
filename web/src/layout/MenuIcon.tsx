// MenuIcon draws the icon of a menu entry in the color of its text. The paths
// follow the stroke style of Tabler Icons (MIT).
const paths = {
  home: 'M5 12l-2 0l9 -9l9 9l-2 0 M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7 M9 21v-6a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v6',
  inbox: 'M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z M4 13h3l3 3h4l3 -3h3',
  administration: 'M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3',
  catalog: 'M3 19a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6a9 9 0 0 1 9 0a9 9 0 0 1 9 0 M3 6l0 13 M12 6l0 13 M21 6l0 13',
  procedureTypes:
    'M5 5m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M19 12m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M5 19m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0 M7 5h5a3 3 0 0 1 3 3v2 M7 19h5a3 3 0 0 0 3 -3v-2',
}

export type MenuIconName = keyof typeof paths

export function MenuIcon({ name }: { name: MenuIconName }) {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={paths[name]} />
    </svg>
  )
}
