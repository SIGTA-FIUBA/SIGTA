import type { components } from '../../api/schema'
import { movedText, procedurePath, procedureStates } from '../../content/procedures'
import { ProcedureRow } from '../ProcedureRow/ProcedureRow'
import { ProcedureRows } from '../ProcedureRows/ProcedureRows'

type MyProcedure = components['schemas']['MyProcedure']

const columns = ['Trámite', 'Paso', 'Estado', 'Último movimiento'] as const

const closeFormat = new Intl.DateTimeFormat('es-AR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  timeZone: 'America/Argentina/Buenos_Aires',
})

// MyProcedureRows lists the person's own procedures: no applicant, and the
// state in the third column.
export function MyProcedureRows({
  label,
  procedures,
  from,
}: {
  label: string
  procedures: MyProcedure[]
  from: string
}) {
  return (
    <ProcedureRows label={label} columns={columns}>
      {procedures.map((procedure) => (
        <ProcedureRow
          key={procedure.id}
          to={procedurePath(procedure.id)}
          state={{ from }}
          type={procedure.procedure_type_name}
          note={procedure.correction_pending ? 'Subsanación pendiente del titular' : undefined}
          step={{ name: procedure.step_name ?? 'Trámite cerrado' }}
          standing={{ text: procedureStates[procedure.state].label }}
          date={
            procedure.state === 'closed'
              ? `Cierre: ${closeFormat.format(new Date(procedure.last_moved_at))}`
              : movedText(procedure.last_moved_at)
          }
        />
      ))}
    </ProcedureRows>
  )
}
