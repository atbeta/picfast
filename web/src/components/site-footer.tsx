import { useQuery } from '@tanstack/react-query'

import { getVersionInfo, type SiteConfig } from '@/lib/site-config'

const GITHUB_URL = 'https://github.com/atbeta/picfast'
const ICP_URL = 'https://beian.miit.gov.cn/'
const PSB_URL = 'https://www.beian.gov.cn/'

export function SiteFooter({ config }: { config: SiteConfig }) {
  const { data: version } = useQuery({
    queryKey: ['version-info'],
    queryFn: getVersionInfo,
    staleTime: 5 * 60_000,
  })

  const githubURL = version?.github_url || GITHUB_URL

  return (
    <footer className="relative z-10 mx-auto w-full max-w-[1400px] px-6 pb-8 text-xs text-muted-foreground">
      <div className="flex flex-col gap-2 border-t border-border/40 pt-4 sm:flex-row sm:flex-wrap sm:items-center sm:justify-center">
        {config.icp_number && (
          <a
            href={ICP_URL}
            target="_blank"
            rel="noreferrer"
            className="transition-colors hover:text-foreground"
          >
            {config.icp_number}
          </a>
        )}
        {config.psb_number && (
          <a
            href={PSB_URL}
            target="_blank"
            rel="noreferrer"
            className="transition-colors hover:text-foreground"
          >
            {config.psb_number}
          </a>
        )}
        <span>
          Powered by{' '}
          <a href={githubURL} target="_blank" rel="noreferrer" className="transition-colors hover:text-foreground">
            PicFast
          </a>
          {version?.version ? ` ${version.version}` : ''}
        </span>
      </div>
    </footer>
  )
}
