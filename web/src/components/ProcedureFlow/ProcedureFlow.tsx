import { motion, useReducedMotion } from 'motion/react'
import { tokens } from '../../theme/tokens'

// The route of a procedure on the landing page, drawn as on a plan: from the
// request of the applicant to the Secretariat, the department that rules,
// back to the Secretariat and to the resolution. It is the same route for
// every procedure that is referred, so it says what SIGTA follows without
// inventing one. The corners of the route, in the units of the drawing.
const corners = [
  [40, 64],
  [240, 64],
  [240, 180],
  [440, 180],
  [440, 296],
  [240, 296],
  [40, 296],
] as const

// The stops, with the corner where each one is and where its text goes.
const stops = [
  { corner: 0, label: 'Solicitud', note: 'El titular la presenta', x: 28, y: 96, anchor: 'start' },
  { corner: 1, label: 'Secretaría', note: 'Revisa y deriva', x: 256, y: 92, anchor: 'start' },
  { corner: 3, label: 'Departamento', note: 'Dictamina', x: 424, y: 208, anchor: 'end' },
  { corner: 5, label: 'Secretaría', note: 'Recibe el dictamen', x: 240, y: 258, anchor: 'middle' },
  { corner: 6, label: 'Resolución', note: 'Se comunica al titular', x: 28, y: 328, anchor: 'start' },
] as const

const route = `M${corners.map(([x, y]) => `${x} ${y}`).join(' L')}`

// Where along the route each corner is, from 0 at the request to 1 at the
// resolution: the dot keeps the same speed on every stretch.
const lengths = corners.map((corner, index) => {
  const before = corners[index - 1]
  return before ? Math.abs(corner[0] - before[0]) + Math.abs(corner[1] - before[1]) : 0
})
const total = lengths.reduce((sum, length) => sum + length, 0)
const along = lengths.map((_, index) => lengths.slice(0, index + 1).reduce((sum, length) => sum + length, 0) / total)

const seconds = (ms: number) => ms / 1000
const draw = seconds(tokens.motion.draw)
const flow = seconds(tokens.motion.flow)
const rest = seconds(tokens.motion.rest)
const cycle = flow + rest
// pulse is how long, as a part of the cycle, the ring of a stop takes to open.
const pulse = seconds(tokens.motion.slow) / cycle

// ring writes the keyframes of the ring of a stop: closed until the dot
// arrives, then it opens and fades.
function ring(at: number) {
  const arrival = (at * flow) / cycle
  const times = [0, Math.max(arrival - 0.001, 0), arrival, Math.min(arrival + pulse, 1), 1]
  return {
    animate: { opacity: [0, 0, 0.8, 0, 0], r: [6, 6, 6, 22, 22] },
    transition: { duration: cycle, times, repeat: Infinity, delay: draw, ease: 'easeOut' as const },
  }
}

// ProcedureFlow draws the route once when the page arrives, then a dot
// travels it from the request to the resolution, again and again, and each
// stop lights up when the dot reaches it. It only illustrates what the text
// of the page says, so screen readers skip it. With reduced motion the route
// is drawn at once and nothing travels.
export function ProcedureFlow() {
  const reduce = useReducedMotion() === true
  return (
    <svg viewBox="0 0 480 360" data-flow aria-hidden focusable="false">
      <text x={40} y={20} data-caption>
        Recorrido de un trámite derivado
      </text>
      <path d={route} data-line="base" />
      <motion.path
        d={route}
        data-line="drawn"
        initial={{ pathLength: reduce ? 1 : 0 }}
        animate={{ pathLength: 1 }}
        transition={{ duration: draw, ease: tokens.motion.easingOut }}
      />
      {stops.map((stop, index) => {
        const [cx, cy] = corners[stop.corner]
        const at = along[stop.corner]
        const last = index === stops.length - 1
        return (
          <motion.g
            key={stop.corner}
            initial={{ opacity: reduce ? 1 : 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: seconds(tokens.motion.slow), delay: reduce ? 0 : at * draw }}
          >
            {!reduce && <motion.circle cx={cx} cy={cy} r={6} data-ring initial={{ opacity: 0 }} {...ring(at)} />}
            <circle cx={cx} cy={cy} r={6} data-node={last ? 'end' : 'stop'} />
            <text x={stop.x} y={stop.y} textAnchor={stop.anchor} data-label>
              {stop.label}
            </text>
            <text x={stop.x} y={stop.y + 18} textAnchor={stop.anchor} data-note>
              {stop.note}
            </text>
          </motion.g>
        )
      })}
      {!reduce && (
        <motion.circle
          r={4}
          data-dot
          initial={{ opacity: 0, cx: corners[0][0], cy: corners[0][1] }}
          animate={{
            opacity: 1,
            cx: corners.map(([x]) => x),
            cy: corners.map(([, y]) => y),
          }}
          transition={{
            opacity: { delay: draw, duration: seconds(tokens.motion.base) },
            default: {
              duration: flow,
              times: along,
              ease: 'linear',
              repeat: Infinity,
              repeatDelay: rest,
              delay: draw,
            },
          }}
        />
      )}
    </svg>
  )
}
