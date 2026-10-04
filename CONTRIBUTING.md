# Cómo contribuir

SIGTA lo desarrolla un equipo cerrado de cuatro integrantes. Esta guía resume la forma de trabajo.

## Flujo

1. Cada cambio parte de un ticket de Linear (acceso restringido al equipo).
2. Se trabaja en una rama con el nombre que propone Linear para el ticket.
3. Se abre un pull request contra `main` con la plantilla del repositorio completa.
4. El pull request entra con la aprobación de otro integrante y la integración continua sin errores.

## Convenciones

- **Idioma:** el código, los endpoints y los nombres técnicos van en inglés; la documentación, en español.
- **Comportamiento primero:** cada historia de usuario se especifica con escenarios Gherkin antes de implementarse.
- **Contrato primero:** un cambio en la API empieza por la especificación OpenAPI.
- **Documentación en el mismo pull request:** si un cambio modifica el comportamiento, la arquitectura o la configuración de trámites, la documentación se actualiza junto con el código.
- **Decisiones:** toda decisión que condicione la arquitectura o el alcance se registra en la sección "Decisiones de diseño" del README.
- **Datos:** nunca se suben datos personales reales; solo datos sintéticos o anonimizados.

## Commits

Los commits siguen [Conventional Commits](https://www.conventionalcommits.org/), en inglés y en una sola línea de hasta 72 caracteres:

```text
feat(engine): add step confirmation
fix(api): return 404 for unknown procedures
docs: describe the ports and adapters
```

Tipos: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore` y `revert`.

La convención la valida [commitlint](https://commitlint.js.org/) con la configuración convencional, mediante un hook de [husky](https://typicode.github.io/husky/) que se instala solo al correr `npm install` en la raíz del repositorio. En cada pull request, la integración continua valida el título con la misma configuración.

Los pull requests se integran con squash merge, así que el título del pull request es el commit que queda en `main` y sigue la misma convención.
