export interface ResolvedDocLink {
  href: string
  external: boolean
  target?: '_blank'
  rel?: string
}

export const DEFAULT_DOC_URL =
  'https://www.yuque.com/zztzz-caqqk/ns9iik/gzhhg0teo5bsgp55?singleDoc#'

export function resolveDocLink(docUrl?: string | null): ResolvedDocLink {
  const href = docUrl?.trim() || DEFAULT_DOC_URL
  const external = /^https?:\/\//i.test(href)

  return external
    ? {
        href,
        external,
        target: '_blank',
        rel: 'noreferrer'
      }
    : {
        href,
        external
      }
}
