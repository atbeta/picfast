import { useEffect, useState } from 'react'

/**
 * Reactively track a CSS media query. Safe in the browser-only SPA: on the
 * first render it resolves synchronously so layout does not flash.
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() =>
    typeof window === 'undefined' ? false : window.matchMedia(query).matches,
  )

  useEffect(() => {
    const mql = window.matchMedia(query)
    const onChange = () => setMatches(mql.matches)
    onChange()
    mql.addEventListener('change', onChange)
    return () => mql.removeEventListener('change', onChange)
  }, [query])

  return matches
}

/** True on touch-first devices (phones, tablets) where drag/paste is unavailable. */
export function useIsCoarsePointer(): boolean {
  return useMediaQuery('(pointer: coarse)')
}
