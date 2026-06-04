import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Check, Clipboard, Info, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { toast } from '@/components/ui/Toast'
import type { GroupPlatform } from '@/types'
import {
  buildClientTabs,
  buildShellTabs,
  buildUsageFiles,
  defaultClientTab,
  platformDescription,
  platformNote,
  type ClientTabId,
  type ShellTabId
} from './keyUsageConfig'

interface UseKeyModalProps {
  open: boolean
  apiKey: string
  baseUrl: string
  platform?: GroupPlatform | null
  allowMessagesDispatch?: boolean
  onClose: () => void
}

export function UseKeyModal({ open, apiKey, baseUrl, platform, allowMessagesDispatch, onClose }: UseKeyModalProps) {
  const { t } = useTranslation()
  const [clientTab, setClientTab] = useState<ClientTabId>(defaultClientTab(platform))
  const [shellTab, setShellTab] = useState<ShellTabId>('unix')
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  useEffect(() => {
    if (!open) return
    setClientTab(defaultClientTab(platform))
    setShellTab('unix')
    setCopiedIndex(null)
  }, [open, platform])

  useEffect(() => {
    setShellTab('unix')
  }, [clientTab])

  const translate = t as (key: string) => string
  const clientTabs = useMemo(
    () => buildClientTabs(platform, allowMessagesDispatch, translate),
    [allowMessagesDispatch, platform, translate]
  )
  const shellTabs = useMemo(() => buildShellTabs(clientTab), [clientTab])
  const files = useMemo(
    () => buildUsageFiles({ platform, clientTab, shellTab, baseUrl, apiKey, t: translate }),
    [apiKey, baseUrl, clientTab, platform, shellTab, translate]
  )
  const note = platformNote(platform, clientTab, shellTab, translate)

  async function copyContent(content: string, index: number) {
    try {
      await navigator.clipboard.writeText(content)
      setCopiedIndex(index)
      toast.success(t('common.copiedToClipboard') as string)
      window.setTimeout(() => setCopiedIndex(null), 1500)
    } catch {
      toast.error(t('common.copyFailed') as string)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={t('keys.useKeyModal.title')} size="lg">
      <div className="space-y-4">
        {!platform ? (
          <div className="flex items-start gap-3 rounded-xl border border-yellow-500/30 bg-yellow-500/10 p-4">
            <AlertTriangle className="mt-0.5 h-5 w-5 flex-shrink-0 text-yellow-500" />
            <div>
              <p className="text-sm font-medium text-ink-1">{t('keys.useKeyModal.noGroupTitle')}</p>
              <p className="mt-1 text-sm text-ink-2">{t('keys.useKeyModal.noGroupDescription')}</p>
            </div>
          </div>
        ) : (
          <>
            <p className="text-sm text-ink-2">{platformDescription(platform, clientTab, translate)}</p>

            {clientTabs.length > 0 && (
              <div className="flex flex-wrap gap-2 border-b border-line-2 pb-2">
                {clientTabs.map((tab) => (
                  <button
                    key={tab.id}
                    type="button"
                    className={tab.id === clientTab ? 'btn btn-primary btn-sm' : 'btn btn-ghost btn-sm'}
                    onClick={() => setClientTab(tab.id as ClientTabId)}
                  >
                    <Terminal className="h-4 w-4" />
                    {tab.label}
                  </button>
                ))}
              </div>
            )}

            {shellTabs.length > 0 && (
              <div className="flex flex-wrap gap-2 border-b border-line-2 pb-2">
                {shellTabs.map((tab) => (
                  <button
                    key={tab.id}
                    type="button"
                    className={tab.id === shellTab ? 'btn btn-accent btn-sm' : 'btn btn-ghost btn-sm'}
                    onClick={() => setShellTab(tab.id as ShellTabId)}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
            )}

            <div className="space-y-4">
              {files.map((file, index) => (
                <div key={`${file.path}-${index}`}>
                  {file.hint && <p className="mb-1.5 text-xs text-yellow-500">{file.hint}</p>}
                  <div className="overflow-hidden rounded-xl border border-line-2 bg-slate-950">
                    <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900 px-4 py-2">
                      <span className="font-mono text-xs text-slate-300">{file.path}</span>
                      <button
                        type="button"
                        className="btn btn-ghost btn-sm text-slate-200"
                        onClick={() => copyContent(file.content, index)}
                      >
                        {copiedIndex === index ? <Check className="h-3.5 w-3.5" /> : <Clipboard className="h-3.5 w-3.5" />}
                        {copiedIndex === index ? t('keys.useKeyModal.copied') : t('keys.useKeyModal.copy')}
                      </button>
                    </div>
                    <pre className="max-h-80 overflow-auto p-4 text-sm text-slate-100"><code>{file.content}</code></pre>
                  </div>
                </div>
              ))}
            </div>

            {note && (
              <div className="flex items-start gap-3 rounded-xl border border-blue-500/30 bg-blue-500/10 p-3">
                <Info className="mt-0.5 h-4 w-4 flex-shrink-0 text-blue-400" />
                <p className="text-sm text-ink-2">{note}</p>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  )
}
