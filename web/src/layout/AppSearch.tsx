import { useQuery } from '@tanstack/react-query'
import { catalogQuery } from '../api/catalog'
import { useApplicantTasks, useInboxProcedures, useMyProcedures } from '../api/procedures'
import { Search, type SearchGroup, type SearchItem } from '../components/Search/Search'
import { applicantName, procedurePath, procedureStates } from '../content/procedures'

// AppSearch fills the search with what the person has at hand: what they have
// to do, their procedures (or, for the staff, the procedures in progress, also
// found by their applicant), the catalog and the pages of the menu.
export function AppSearch({ staff, pages }: { staff: boolean; pages: SearchItem[] }) {
  const tasks = useApplicantTasks()
  const mine = useMyProcedures()
  const inbox = useInboxProcedures(staff)
  const catalog = useQuery(catalogQuery)

  const actionable = (inbox.data?.items ?? []).filter((item) => item.actionable)
  const pending: SearchItem[] = staff
    ? actionable.map((item) => ({
        id: `pending-${item.id}`,
        label: item.step_name,
        description: `${item.procedure_type_name}, ${applicantName(item)}`,
        to: procedurePath(item.id),
      }))
    : tasks.tasks.map((task) => ({
        id: `pending-${task.id}`,
        label: task.task,
        description: task.context,
        to: procedurePath(task.procedureId),
      }))
  const procedures: SearchItem[] = staff
    ? (inbox.data?.items ?? []).map((item) => ({
        id: `procedure-${item.id}`,
        label: item.procedure_type_name,
        description: `${applicantName(item)}, ${item.step_name}`,
        to: procedurePath(item.id),
      }))
    : (mine.data ?? []).map((procedure) => ({
        id: `procedure-${procedure.id}`,
        label: procedure.procedure_type_name,
        description: procedure.step_name ?? procedureStates[procedure.state].label,
        to: procedurePath(procedure.id),
      }))
  const start: SearchItem[] = (catalog.data ?? []).map((entry) => ({
    id: `start-${entry.id}`,
    label: entry.name,
    description: entry.description,
    to: `/catalog/${entry.id}/start`,
  }))

  const groups: SearchGroup[] = [
    { name: 'Pendientes', icon: 'file-pencil', items: pending, idle: 3 },
    { name: staff ? 'Trámites' : 'Mis trámites', icon: 'files', items: procedures, idle: 0 },
    { name: 'Iniciar un trámite', icon: 'plus', items: start, idle: staff ? 0 : 4 },
    { name: 'Ir a', icon: 'arrow-right', items: pages, idle: 0 },
  ]
  return <Search groups={groups} />
}
