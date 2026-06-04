import { useTranslation } from 'react-i18next'
import { Bot, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import type { CcsClientType } from './ccswitch'

interface CcsClientSelectModalProps {
  open: boolean
  onClose: () => void
  onSelect: (clientType: CcsClientType) => void
}

export function CcsClientSelectModal({ open, onClose, onSelect }: CcsClientSelectModalProps) {
  const { t } = useTranslation()

  return (
    <Modal open={open} onClose={onClose} title={t('keys.ccsClientSelect.title')}>
      <div className="space-y-4">
        <p className="text-sm text-ink-2">{t('keys.ccsClientSelect.description')}</p>
        <div className="grid gap-3 sm:grid-cols-2">
          <Button variant="secondary" className="h-auto justify-start p-4 text-left" onClick={() => onSelect('claude')}>
            <Bot className="h-5 w-5 text-orange" />
            <span>
              <span className="block font-medium text-ink-1">{t('keys.ccsClientSelect.claudeCode')}</span>
              <span className="block text-xs text-ink-3">{t('keys.ccsClientSelect.claudeCodeDesc')}</span>
            </span>
          </Button>
          <Button variant="secondary" className="h-auto justify-start p-4 text-left" onClick={() => onSelect('gemini')}>
            <Sparkles className="h-5 w-5 text-orange" />
            <span>
              <span className="block font-medium text-ink-1">{t('keys.ccsClientSelect.geminiCli')}</span>
              <span className="block text-xs text-ink-3">{t('keys.ccsClientSelect.geminiCliDesc')}</span>
            </span>
          </Button>
        </div>
      </div>
    </Modal>
  )
}
