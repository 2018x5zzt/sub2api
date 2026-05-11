import { useEffect, useMemo, useState } from 'react'
import { ExternalLink, Image as ImageIcon, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { oauthAPI } from '@/api/oauth'

const MIKU_CALLBACK_URL = 'https://ai.mikuapi.org/auth/xlab/callback?layout=horizontal'
const MIKU_CLIENT_ID = 'miku-app'
const MIKU_LOGIN_URL = 'https://ai.mikuapi.org/login?xlab_auto=1'
const LEGACY_IMAGE_STUDIO_URL = 'https://iframe.mikuapi.org/?codexCli=true'

export default function ImageStudioPage() {
  const [mode, setMode] = useState<'new' | 'legacy'>('new')
  const [mikuUrl, setMikuUrl] = useState(MIKU_LOGIN_URL)
  const [loading, setLoading] = useState(false)

  async function refreshOAuth() {
    setLoading(true)
    try {
      const result: any = await oauthAPI.authorizeXlabOAuth({
        client_id: MIKU_CLIENT_ID,
        redirect_uri: MIKU_CALLBACK_URL,
        response_type: 'code',
        scope: 'profile email',
        state: 'image-studio'
      })
      setMikuUrl(result?.redirect_uri || result?.redirect_url || result?.url || MIKU_LOGIN_URL)
    } catch {
      setMikuUrl(MIKU_LOGIN_URL)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refreshOAuth()
  }, [])

  const activeUrl = useMemo(() => mode === 'new' ? mikuUrl : LEGACY_IMAGE_STUDIO_URL, [mode, mikuUrl])

  return (
    <>
      <PageHeader
        title="图片创作"
        description="使用你的 XlabAPI 会话打开图片创作工具。"
        actions={
          <>
            <Button variant="ghost" onClick={refreshOAuth} loading={loading}>
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <a href={activeUrl} target="_blank" rel="noreferrer" className="btn btn-primary">
              <ExternalLink className="h-4 w-4" />
              Open
            </a>
          </>
        }
      />

      <Card className="p-4 mb-4 flex flex-wrap gap-2">
        <Button variant={mode === 'new' ? 'accent' : 'ghost'} onClick={() => setMode('new')}>
          <ImageIcon className="h-4 w-4" />
          New studio
        </Button>
        <Button variant={mode === 'legacy' ? 'accent' : 'ghost'} onClick={() => setMode('legacy')}>
          旧版入口
        </Button>
      </Card>

      <Card className="overflow-hidden p-0">
        <iframe src={activeUrl} className="block h-[calc(100vh-240px)] min-h-[560px] w-full border-0 bg-bg-1" allowFullScreen />
      </Card>
    </>
  )
}
