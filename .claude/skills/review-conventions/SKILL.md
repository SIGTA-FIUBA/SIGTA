---
name: review-conventions
description: Revisa un cambio de SIGTA contra las convenciones del repositorio (nombres, ubicación, reutilización, tests) antes de abrir el PR, para que no se acumulen diferencias que después piden un refactor. Usar antes de cada PR (es un paso de open-pr) o cuando se pida revisar si un cambio respeta las convenciones.
---

# Revisar las convenciones de un cambio

El lint y la CI cubren lo mecánico. Esta revisión cubre lo que ninguna herramienta ve: que lo nuevo no repita algo que ya existe, que se llame como el resto y que esté donde se lo va a buscar. Se revisa el diff completo contra `develop` (`git diff develop...`), no solo el último commit.

## Pasos

1. **Correr los chequeos mecánicos** y arreglar lo que marquen antes de seguir:
   - backend: `cd services/backend && golangci-lint run` y `go run golang.org/x/tools/cmd/deadcode@v0.52.0 -test ./...`;
   - frontend: `cd web && npm run lint && npm run unused`;
   - los dos: `scripts/check-test-layout.sh`.
2. **Buscar antes de crear.** Por cada función, componente, tipo, texto, consulta o ayuda de test nueva, buscar con `grep` si ya hay una que haga lo mismo o casi:
   - web: `src/components/`, `src/content/`, `src/api/` y `tests/test-utils/`;
   - backend: el paquete del concepto en `domain/`, los puertos en `ports/` y los archivos de test del mismo paquete.
   Si existe, usarla o extenderla. Si hay dos que hacen lo mismo, dejar una sola.
3. **Nombres.** Cada nombre nuevo sale del vocabulario de AGENTS.md y sigue "Nombres" de CONTRIBUTING.md. Un concepto tiene un solo nombre del contrato a la web: el esquema de `contracts/openapi.yaml`, el tipo de Go y el tipo de TypeScript se llaman igual. Una función se nombra por lo que hace, no por dónde se usa.
4. **Ubicación.** Cada archivo nuevo va donde la tabla "Estructura del repositorio" de AGENTS.md dice que se busca: un texto que usan varias pantallas, en `content/`; una consulta compartida, en `api/`; un componente que usan varias páginas, en `components/`; en el backend, un concepto por archivo y un puerto por archivo.
5. **Firmas.** Sin parámetros que siempre reciben el mismo valor, resultados que nadie usa ni funciones que solo pasan la llamada a otra. Si una función tiene que cambiar de firma para el cambio, actualizar todas las llamadas en vez de sumar una variante.
6. **Restos de un renombre.** Si el cambio renombra o mueve algo, buscar el nombre viejo en todo el repositorio, incluidos `docs/`, `features/`, comentarios y textos de tests.
7. **Tests.** Cada test está donde dice CONTRIBUTING.md y usa las ayudas que ya existen (`stubApi`, `render<Algo>`, los fakes del paquete) en vez de copiarlas.

## Resultado

Una lista de hallazgos, cada uno con `archivo:línea`, qué convención no cumple y cómo se arregla. Arreglarlos antes de abrir el PR. Si un hallazgo pide cambiar algo que no es parte del cambio, anotarlo como duda abierta del PR en vez de tocarlo.
