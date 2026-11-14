import { addons } from 'storybook/manager-api'
import { create } from 'storybook/theming'
import tokens from '../src/theme/tokens.json'

// The catalog's own frame carries the SIGTA logo and colors.
addons.setConfig({
  theme: create({
    base: 'light',
    brandTitle: 'SIGTA, catálogo de diseño',
    brandImage: 'brand/sigta-horizontal-color.svg',
    brandTarget: '_self',
    colorPrimary: tokens.color.brand.navy.$value,
    colorSecondary: tokens.color.fiuba['7'].$value,
    textColor: tokens.color.text.$value,
  }),
})
