import { useEffect } from 'react'
import { Megaphone } from 'lucide-react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { useAnnouncementStore } from '@/stores/announcements'
import { useAuthStore } from '@/stores/auth'
import { useTranslation } from 'react-i18next'

export function AnnouncementPopup() {
  const { t } = useTranslation()
  const current = useAnnouncementStore((s) => s.currentPopup)
  const dismiss = useAnnouncementStore((s) => s.dismissPopup)
  const fetchAnnouncements = useAnnouncementStore((s) => s.fetch)
  const authed = useAuthStore((s) => s.isAuthenticated())

  useEffect(() => {
    if (authed) fetchAnnouncements()
  }, [authed, fetchAnnouncements])

  if (!current) return null

  return (
    <Modal
      open
      onClose={dismiss}
      title={
        <span className="inline-flex items-center gap-2">
          <Megaphone className="h-4 w-4 text-orange" />
          {current.title}
        </span>
      }
      footer={
        <Button variant="accent" onClick={dismiss}>
          {t('common.confirm')}
        </Button>
      }
    >
      <div
        className="prose prose-sm prose-invert max-w-none text-ink-2 whitespace-pre-wrap break-words"
        // Announcement content is admin-authored markdown/html and trusted.
        dangerouslySetInnerHTML={{ __html: current.content }}
      />
      {current.starts_at && (
        <p className="mt-4 text-xs text-ink-3 font-mono">
          {new Date(current.starts_at).toLocaleString()}
        </p>
      )}
    </Modal>
  )
}
