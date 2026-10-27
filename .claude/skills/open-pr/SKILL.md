---
name: open-pr
description: Prepara la rama, los commits y el pull request de un cambio de SIGTA con la plantilla completa y asociado a su ticket de Linear. Usar cuando se pida commitear, abrir o preparar un PR.
---

# Pull request con su descripción

La autoría del PR es humana: se abre desde la cuenta de quien pidió el cambio y lo revisa otro integrante antes de entrar.

## Pasos

1. **Identificar el ticket** de Linear (`SIG-n`) y leer sus criterios con `get_issue`. Si no hay ticket, preguntar antes de seguir: todo cambio parte de uno.
2. **Rama** desde `develop` actualizado, con el formato del equipo `<tipo>/sig-<n>-<descripción>`, por ejemplo `feat/sig-17-start-procedure`. No usar el nombre que propone Linear.
3. **Documentación en el mismo PR** (ADR-007). Revisar si el cambio exige actualizar:
   - `docs/` y el README del módulo, si cambia el comportamiento;
   - un ADR en `docs/04-decisiones-adr.md` y el README, si hay una decisión de arquitectura o alcance;
   - `AGENTS.md`, si cambia una convención;
   - el instructivo de configuración, si cambia el modelo de trámites;
   - `features/` y su matriz, si cambia el comportamiento esperado.
4. **Contrato:** si cambió `contracts/openapi.yaml`, correr `make contracts` y commitear el código generado.
5. **Verificar localmente** lo que toca el cambio:
   - backend: `cd services/backend && go test -race ./...` y `golangci-lint run`;
   - frontend: `cd web && npm run lint && npm run format:check && npm run typecheck && npm test`;
   - si cambia la integración entre servicios: `docker compose up --build`.
6. **Commits:** Conventional Commits, en inglés, una sola línea de hasta 72 caracteres.
7. **PR contra `develop`.** El título sigue Conventional Commits, porque con el squash merge es el commit que queda.

## Descripción del PR

Completar la plantilla de `.github/pull_request_template.md`, en español:

- **Descripción:** qué cambia y por qué, agrupado por tema.
- **Issue / historia relacionada:** título e identificador, por ejemplo `"Iniciar una solicitud" (SIG-17, US14)`. Si el PR no completa todos los criterios, decir cuáles faltan.
- **Cómo probarlo:** pasos concretos que otra persona pueda seguir.
- **Definición de terminado:** marcar solo lo que se cumple de verdad. Si algo no aplica, dejarlo sin marcar y aclararlo entre paréntesis.
- **Dudas abiertas**, si las hay: criterios ambiguos o decisiones que tiene que tomar el equipo.

## Detalles

- `gh pr edit` puede fallar por una función vieja de GitHub (Projects classic). Para cambiar la descripción, usar `gh api -X PATCH repos/SIGTA-FIUBA/SIGTA/pulls/<n> -F "body=@<archivo>"`.
- No cambiar el estado del ticket en Linear sin que lo pidan.
