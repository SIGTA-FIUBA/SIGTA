# Componentes de terceros

Componentes de código abierto copiados de su fuente (Magic UI, Aceternity UI, los ejemplos de motion.dev), para no rehacer a mano lo que ya existe (ADR-028).

- Cada componente va en su carpeta, con su `LICENSE` y un `SOURCE.md` que dice de qué repositorio, ruta y commit viene, como las rutinas de terceros de `.claude/skills/`.
- Se copia tal como viene: estructura, comportamiento, animación y nombres de archivo, así se compara con la fuente cuando cambia. Solo cambian las clases de color y de medida, que pasan a las escalas del tema definidas en `src/theme/tailwind.css` (por ejemplo, `text-neutral-500` pasa a `text-dimmed` y `text-base` a `text-md`); `SOURCE.md` anota cada cambio.
- Es la única carpeta donde se escribe `className` y `style`. El lint sigue rechazando colores y medidas sueltos (`bg-[#fff]`, `w-[13px]`), y una clase fuera del tema (`bg-violet-500`) no genera estilo.
- Las pantallas lo usan como a cualquier componente, y entra con su historia en el catálogo.
