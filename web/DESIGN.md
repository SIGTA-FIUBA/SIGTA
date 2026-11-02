# Diseño de la interfaz

Reglas para construir pantallas de SIGTA. Cada regla dice por qué existe, para resolver los casos que no están escritos. La decisión de fondo está en ADR-015.

## Dónde está cada cosa

| Qué                                                                          | Dónde                                                                                        |
| ---------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Componentes                                                                  | [Mantine 9](https://mantine.dev), documentación para agentes en https://mantine.dev/llms.txt |
| Aspecto: colores, tipografía, radios, valores por defecto de cada componente | `src/theme/theme.ts`, y en ningún otro lugar                                                 |
| Componentes propios de SIGTA                                                 | `src/components/`                                                                            |
| Tests, con la misma estructura de carpetas que `src/`                        | `tests/`                                                                                     |
| Reglas de lint que hacen cumplir este documento                              | `.oxlintrc.json` y `lint/design-rules.js`                                                    |

## Reglas

| Regla                                                                                                                                                                                                 | Motivo                                                                                                               |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Toda la interfaz se arma con componentes de Mantine o de `src/components/`. No se escribe HTML con estilo propio.                                                                                     | Revisar una pantalla tiene que ser revisar qué componentes usa y con qué propiedades.                                |
| Colores y medidas salen del tema: tamaños `xs` a `xl` y colores por nombre (`dimmed`, `bright`, `red`, el primario). Nunca `#1f1f1f`, `13px`, `rgb(...)` ni números en propiedades de espaciado.      | El aspecto se cambia en un solo archivo. Un valor suelto es una decisión de diseño que nadie tomó.                   |
| Las pantallas no usan `style`, `className`, `styles`, `classNames` ni `vars`. Si un componente necesita otro aspecto, se agrega al tema con `defaultProps` o con la Styles API en `theme.components`. | Lo que se repite en una pantalla termina repitiéndose en todas. En el tema se decide una vez.                        |
| No hay CSS propio. La única excepción es `GoogleSignInButton`, que Google exige con sus colores y tipografía.                                                                                         | Cada excepción tiene que tener un motivo externo y quedar aislada.                                                   |
| El rojo se usa solo para errores (`Alert color="red"`).                                                                                                                                               | Si el color decora, deja de avisar.                                                                                  |
| Una pantalla tiene un solo botón principal (`variant="filled"`). El resto usa `default`, `light` o `subtle`.                                                                                          | Si todo es importante, nada lo es.                                                                                   |
| Cada página tiene un solo título de nivel 1 (`Title order={1}`) y los demás títulos bajan de nivel sin saltear.                                                                                       | Los lectores de pantalla navegan por títulos.                                                                        |
| Toda pantalla que muestra datos contempla cuatro estados: cargando, vacío, error y con datos.                                                                                                         | El estado vacío y el de error son los que se ven la primera semana.                                                  |
| Los textos están en español rioplatense, con voseo y mayúscula solo al principio ("Cerrar sesión", no "Cerrar Sesión"). Un error dice qué hacer ("Volvé a intentarlo").                               | Es el idioma de quienes usan SIGTA. Un error que no dice qué hacer deja a la persona sin salida.                     |
| Los nombres de personas se muestran como vienen, sin cambiar mayúsculas. Las iniciales salen de `Avatar` con `name`.                                                                                  | El sistema no corrige el nombre de nadie.                                                                            |
| Los botones que solo tienen un ícono o un avatar llevan `aria-label`.                                                                                                                                 | Sin texto visible, el lector de pantalla no tiene qué leer.                                                          |
| La interfaz es solo clara (`forceColorScheme="light"`), sin modo oscuro.                                                                                                                              | Es un sistema institucional de uso administrativo. Un solo modo es la mitad de combinaciones para diseñar y revisar. |

## A confirmar

- La paleta institucional y la tipografía. Hasta que se defina la identidad visual, el color primario es el azul de Mantine y la tipografía es la del sistema.
