// Lint rules that keep file names as CONTRIBUTING.md asks ("Nombres"), so a
// component is found by its name. They run inside oxlint as a JS plugin.

const PASCAL_CASE = /^[A-Z]/

// The names a file exports as values, leaving types out.
function exportedValues(program) {
  const names = []
  for (const statement of program.body) {
    if (statement.type === 'ExportDefaultDeclaration' && statement.declaration.id) {
      names.push(statement.declaration.id.name)
    }
    if (statement.type !== 'ExportNamedDeclaration' || !statement.declaration) {
      continue
    }
    const declaration = statement.declaration
    if (declaration.type === 'FunctionDeclaration' || declaration.type === 'ClassDeclaration') {
      names.push(declaration.id.name)
    }
    if (declaration.type === 'VariableDeclaration') {
      declaration.declarations.forEach((d) => d.id.type === 'Identifier' && names.push(d.id.name))
    }
  }
  return names
}

const fileName = {
  meta: {
    type: 'suggestion',
    docs: { description: 'Name a file after the component it exports' },
    messages: {
      component: 'This file exports {{component}}, so it is {{component}}.tsx.',
      oneComponent: 'One component per file: {{components}} go in files of their own.',
      page: 'A component in pages/ is a page, so its name ends in Page: {{component}}.',
      module: 'A module without JSX is named in camelCase: {{expected}}.ts.',
    },
  },
  create(context) {
    const path = context.filename
    const match = /([^/\\]+)\.(tsx?)$/.exec(path)
    if (!match || path.endsWith('.d.ts')) {
      return {}
    }
    const [, base, extension] = match
    return {
      Program(program) {
        if (extension === 'ts') {
          if (PASCAL_CASE.test(base)) {
            const expected = base[0].toLowerCase() + base.slice(1)
            context.report({ node: program, messageId: 'module', data: { expected } })
          }
          return
        }
        const components = exportedValues(program).filter((name) => PASCAL_CASE.test(name))
        if (components.length > 1) {
          context.report({ node: program, messageId: 'oneComponent', data: { components: components.join(', ') } })
          return
        }
        const [component] = components
        if (!component) {
          return
        }
        if (component !== base) {
          context.report({ node: program, messageId: 'component', data: { component } })
        }
        if (/[/\\]pages[/\\]/.test(path) && !component.endsWith('Page')) {
          context.report({ node: program, messageId: 'page', data: { component } })
        }
      },
    }
  },
}

export default {
  meta: { name: 'sigta-naming' },
  rules: { 'file-name': fileName },
}
