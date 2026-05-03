import { useQuery } from '@tanstack/react-query'

import { getVersionInfo, type SiteConfig } from '@/lib/site-config'

const GITHUB_URL = 'https://github.com/atbeta/picfast'

export function SiteFooter({ config }: { config: SiteConfig }) {
  const { data: version } = useQuery({
    queryKey: ['version-info'],
    queryFn: getVersionInfo,
    staleTime: 5 * 60_000,
  })

  const githubURL = version?.github_url || GITHUB_URL

  const footerItems = [
    { text: config.footer_text_1, link: config.footer_link_1 },
    { text: config.footer_text_2, link: config.footer_link_2 },
  ].filter(item => item.text && item.text.trim() !== '')

  return (
    <footer className="relative z-10 mx-auto w-full max-w-[1400px] px-6 pb-8 text-xs text-muted-foreground">
      <div className="flex flex-col gap-2 border-t border-border/40 pt-4">
        {footerItems.length > 0 && (
          <div className="flex flex-wrap items-center justify-center gap-3">
            {footerItems.map((item, index) =>
              item.link && item.link.trim() !== '' ? (
                <a
                  key={index}
                  href={item.link}
                  target="_blank"
                  rel="noreferrer"
                  className="transition-colors hover:text-foreground"
                >
                  {item.text}
                </a>
              ) : (
                <span key={index}>{item.text}</span>
              ),
            )}
          </div>
        )}
        <span className="text-center">
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
