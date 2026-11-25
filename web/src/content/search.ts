// normalize lets "tramite" find "Trámite": no accents, no case.
export function normalize(text: string) {
  return text
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase()
}

// searchWords are the words of a search, normalized, so that each one has to
// be found.
export function searchWords(search: string) {
  return normalize(search).split(/\s+/).filter(Boolean)
}
