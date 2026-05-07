import { Card } from '@/components/ui/Card'
import { PageHeader } from '@/components/layout/ConsoleLayout'

export function PlaceholderPage({ title, description }: { title: string; description?: string }) {
  return (
    <>
      <PageHeader title={title} description={description} />
      <Card className="p-12 text-center text-ink-3">
        This admin section is not yet migrated to v2 — see <code className="font-mono">MIGRATION_TODO.md</code>.
      </Card>
    </>
  )
}
