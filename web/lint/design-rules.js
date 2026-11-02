// Lint rules that keep the look of the interface in the theme
// (src/theme/theme.ts). They run inside oxlint as a JS plugin.

// A color or a fixed measure written by hand: #1f1f1f, rgb(...), 13px, 1.5rem.
const RAW_VALUE = /#[0-9a-f]{3,8}\b|\b(rgba?|hsla?|oklch|color-mix)\(|\d(px|rem|em)\b/i

// Mantine style props that take a theme size ("md") or a number of pixels. A
// number skips the theme's scale, so it is not allowed.
const SPACING_PROPS = new Set([
  ...['m', 'mt', 'mb', 'ml', 'mr', 'ms', 'me', 'mx', 'my'],
  ...['p', 'pt', 'pb', 'pl', 'pr', 'ps', 'pe', 'px', 'py'],
  ...['w', 'h', 'miw', 'maw', 'mih', 'mah', 'fz', 'lh', 'lts'],
  ...['gap', 'rowGap', 'columnGap', 'top', 'left', 'right', 'bottom', 'inset'],
])

const noRawDesignValues = {
  meta: {
    type: 'problem',
    docs: { description: 'Take colors and measures from the theme, not from literals' },
    messages: {
      raw: '"{{value}}" is a raw design value. Use a theme value (a size such as "md", a color such as "dimmed") or add it to src/theme/theme.ts.',
    },
  },
  create(context) {
    function check(node, value, shown = value) {
      if (typeof value === 'string' && RAW_VALUE.test(value)) {
        context.report({ node, messageId: 'raw', data: { value: shown } })
      }
    }
    return {
      JSXAttribute(node) {
        const value = node.value
        if (!value) {
          return
        }
        if (value.type === 'Literal') {
          check(node, value.value)
          return
        }
        if (value.type !== 'JSXExpressionContainer') {
          return
        }
        const expression = value.expression
        if (expression.type === 'Literal') {
          if (typeof expression.value === 'number' && expression.value !== 0 && SPACING_PROPS.has(node.name.name)) {
            context.report({ node, messageId: 'raw', data: { value: String(expression.value) } })
          }
          check(node, expression.value)
        }
        if (expression.type === 'TemplateLiteral') {
          // In `${size}px` the unit follows an expression, so it is checked as
          // if the expression were a number.
          expression.quasis.forEach((quasi) => check(node, `0${quasi.value.raw}`, quasi.value.raw))
        }
      },
    }
  },
}

export default {
  meta: { name: 'sigta-design' },
  rules: { 'no-raw-design-values': noRawDesignValues },
}
