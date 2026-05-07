import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { useTranslation } from 'react-i18next'

export default function RedeemPage() {
  const { t } = useTranslation()
  return (
    <>
      <PageHeader title={t('redeem.title')} description={t('redeem.description') as string} />
      <Card className="p-12 text-center text-ink-3">
        Redemption flow not yet migrated to v2 — see MIGRATION_TODO.md
      </Card>
    </>
  )
}
