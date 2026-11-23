import { Select, TagsInput, Textarea, TextInput } from '@mantine/core'
import type { UseFormReturnType } from '@mantine/form'
import type { components } from '../../api/schema'
import { maxSubjects, type SubmissionDraft } from '../../submission'

type FieldRequirement = components['schemas']['FieldRequirement']

// Beyond this many options, a select lets the applicant type to find one.
const searchableOptions = 7

// FieldInput is the input of a field of a procedure type, by its type, bound
// to the value of the field in a form of a submission.
export function FieldInput({ form, field }: { form: UseFormReturnType<SubmissionDraft>; field: FieldRequirement }) {
  const path = `values.${field.id}`
  const common = {
    label: field.name,
    name: field.id,
    withAsterisk: field.required,
    key: form.key(path),
    ...form.getInputProps(path),
  }
  switch (field.type) {
    case 'long_text':
      return <Textarea {...common} autosize minRows={3} maxRows={12} />
    case 'number':
      return <TextInput {...common} inputMode="decimal" description="Con coma o punto para los decimales." />
    case 'date':
      return <TextInput {...common} type="date" />
    case 'option':
      return (
        <Select
          {...common}
          data={field.options.map((option) => ({ value: option.id, label: option.name }))}
          placeholder="Elegí una opción…"
          searchable={field.options.length > searchableOptions}
          nothingFoundMessage="No hay opciones con ese texto."
        />
      )
    case 'subject_list':
      return (
        <TagsInput
          {...common}
          description={`Escribí cada asignatura y apretá Enter. Hasta ${maxSubjects}.`}
          splitChars={[]}
          maxTags={maxSubjects}
          clearable
        />
      )
    default:
      return <TextInput {...common} />
  }
}
