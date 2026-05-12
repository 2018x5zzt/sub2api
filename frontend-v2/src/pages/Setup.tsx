import { useMemo, useState } from 'react'
import { Check, ChevronLeft, ChevronRight, Database, KeyRound, ServerCog, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { toast } from '@/components/ui/Toast'
import { setupAPI, type InstallRequest } from '@/api/setup'

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

type StepId = 'database' | 'redis' | 'admin' | 'ready'

function currentPort() {
  const port = Number(window.location.port)
  if (Number.isFinite(port) && port > 0) return port
  return window.location.protocol === 'https:' ? 443 : 80
}

function emptyConfig(): InstallRequest {
  return {
    database: {
      host: 'localhost',
      port: 5432,
      user: 'postgres',
      password: '',
      dbname: 'sub2api',
      sslmode: 'disable'
    },
    redis: {
      host: 'localhost',
      port: 6379,
      password: '',
      db: 0,
      enable_tls: false
    },
    admin: {
      email: '',
      password: ''
    },
    server: {
      host: '0.0.0.0',
      port: currentPort(),
      mode: 'release'
    }
  }
}

function fieldError(error: unknown, fallback: string) {
  const e = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
  return e.response?.data?.detail || e.response?.data?.message || e.message || fallback
}

export default function SetupPage() {
  const { t } = useTranslation()
  const [step, setStep] = useState(0)
  const [config, setConfig] = useState<InstallRequest>(() => emptyConfig())
  const [confirmPassword, setConfirmPassword] = useState('')
  const [dbConnected, setDbConnected] = useState(false)
  const [redisConnected, setRedisConnected] = useState(false)
  const [testingDb, setTestingDb] = useState(false)
  const [testingRedis, setTestingRedis] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [installSuccess, setInstallSuccess] = useState(false)
  const [serviceReady, setServiceReady] = useState(false)
  const [message, setMessage] = useState('')

  const steps = useMemo(
    () => [
      { id: 'database' as StepId, title: t('setup.database.title'), icon: Database },
      { id: 'redis' as StepId, title: t('setup.redis.title'), icon: ServerCog },
      { id: 'admin' as StepId, title: t('setup.admin.title'), icon: KeyRound },
      { id: 'ready' as StepId, title: t('setup.ready.title'), icon: ShieldCheck }
    ],
    [t]
  )

  const canProceed = useMemo(() => {
    if (step === 0) return dbConnected
    if (step === 1) return redisConnected
    if (step === 2) {
      return !!config.admin.email && config.admin.password.length >= 8 && config.admin.password === confirmPassword
    }
    return true
  }, [config.admin.email, config.admin.password, confirmPassword, dbConnected, redisConnected, step])

  const updateDatabase = (patch: Partial<InstallRequest['database']>) => {
    setDbConnected(false)
    setConfig((current) => ({ ...current, database: { ...current.database, ...patch } }))
  }

  const updateRedis = (patch: Partial<InstallRequest['redis']>) => {
    setRedisConnected(false)
    setConfig((current) => ({ ...current, redis: { ...current.redis, ...patch } }))
  }

  const updateAdmin = (patch: Partial<InstallRequest['admin']>) => {
    setConfig((current) => ({ ...current, admin: { ...current.admin, ...patch } }))
  }

  const testDatabase = async () => {
    setTestingDb(true)
    setMessage('')
    try {
      await setupAPI.testDatabase(config.database)
      setDbConnected(true)
      toast.success(t('setup.status.success') as string)
    } catch (error) {
      setDbConnected(false)
      setMessage(fieldError(error, 'Database connection failed'))
    } finally {
      setTestingDb(false)
    }
  }

  const testRedis = async () => {
    setTestingRedis(true)
    setMessage('')
    try {
      await setupAPI.testRedis(config.redis)
      setRedisConnected(true)
      toast.success(t('setup.status.success') as string)
    } catch (error) {
      setRedisConnected(false)
      setMessage(fieldError(error, 'Redis connection failed'))
    } finally {
      setTestingRedis(false)
    }
  }

  const waitForRestart = async () => {
    await new Promise((resolve) => setTimeout(resolve, 3000))
    for (let attempt = 0; attempt < 60; attempt += 1) {
      try {
        const status = await setupAPI.getSetupStatus()
        if (!status.needs_setup) {
          setServiceReady(true)
          setTimeout(() => {
            window.location.href = '/login'
          }, 1500)
          return
        }
      } catch {
        // Service may be restarting; keep polling.
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
    }
    setMessage(t('setup.status.timeout') as string)
  }

  const performInstall = async () => {
    setInstalling(true)
    setMessage('')
    try {
      await setupAPI.install(config)
      setInstallSuccess(true)
      void waitForRestart()
    } catch (error) {
      setMessage(fieldError(error, 'Installation failed'))
    } finally {
      setInstalling(false)
    }
  }

  return (
    <div className="min-h-screen bg-bg-0 px-4 py-10 text-ink-1 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <div className="text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-orange text-white shadow-elevated">
            <ServerCog className="h-8 w-8" />
          </div>
          <h1 className="mt-5 font-display text-4xl tracking-tight">{t('setup.title')}</h1>
          <p className="mt-2 text-ink-3">{t('setup.description')}</p>
        </div>

        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          {steps.map((s, index) => {
            const Icon = s.icon
            const done = step > index
            const active = step === index
            return (
              <div key={s.id} className="flex items-center gap-2">
                <div
                  className={`flex h-10 w-10 items-center justify-center rounded-full border text-sm font-semibold ${
                    done || active ? 'border-orange bg-orange text-white' : 'border-line-2 bg-bg-1 text-ink-3'
                  }`}
                >
                  {done ? <Check className="h-4 w-4" /> : active ? <Icon className="h-4 w-4" /> : index + 1}
                </div>
                <span className={active || done ? 'text-sm font-medium text-ink-1' : 'text-sm text-ink-3'}>{s.title}</span>
              </div>
            )
          })}
        </div>

        <Card className="mt-8 p-6 sm:p-8">
          {step === 0 && (
            <div className="space-y-5">
              <StepTitle title={t('setup.database.title')} description={t('setup.database.description')} />
              <div className="grid gap-4 sm:grid-cols-2">
                <Input label={t('setup.database.host') as string} value={config.database.host} onChange={(e) => updateDatabase({ host: e.target.value })} />
                <Input label={t('setup.database.port') as string} type="number" value={config.database.port} onChange={(e) => updateDatabase({ port: Number(e.target.value) })} />
                <Input label={t('setup.database.username') as string} value={config.database.user} onChange={(e) => updateDatabase({ user: e.target.value })} />
                <Input label={t('setup.database.password') as string} type="password" value={config.database.password} onChange={(e) => updateDatabase({ password: e.target.value })} placeholder={t('setup.database.passwordPlaceholder') as string} />
                <Input label={t('setup.database.databaseName') as string} value={config.database.dbname} onChange={(e) => updateDatabase({ dbname: e.target.value })} />
                <div>
                  <label className="input-label">{t('setup.database.sslMode')}</label>
                  <select className={selectClass} value={config.database.sslmode} onChange={(e) => updateDatabase({ sslmode: e.target.value })}>
                    <option value="disable">{t('setup.database.ssl.disable')}</option>
                    <option value="require">{t('setup.database.ssl.require')}</option>
                    <option value="verify-ca">{t('setup.database.ssl.verifyCa')}</option>
                    <option value="verify-full">{t('setup.database.ssl.verifyFull')}</option>
                  </select>
                </div>
              </div>
              <Button variant={dbConnected ? 'primary' : 'secondary'} loading={testingDb} onClick={() => void testDatabase()} className="w-full">
                {dbConnected && <Check className="h-4 w-4" />}
                {testingDb ? t('setup.status.testing') : dbConnected ? t('setup.status.success') : t('setup.status.testConnection')}
              </Button>
            </div>
          )}

          {step === 1 && (
            <div className="space-y-5">
              <StepTitle title={t('setup.redis.title')} description={t('setup.redis.description')} />
              <div className="grid gap-4 sm:grid-cols-2">
                <Input label={t('setup.redis.host') as string} value={config.redis.host} onChange={(e) => updateRedis({ host: e.target.value })} />
                <Input label={t('setup.redis.port') as string} type="number" value={config.redis.port} onChange={(e) => updateRedis({ port: Number(e.target.value) })} />
                <Input label={t('setup.redis.password') as string} type="password" value={config.redis.password} onChange={(e) => updateRedis({ password: e.target.value })} placeholder={t('setup.redis.passwordPlaceholder') as string} />
                <Input label={t('setup.redis.database') as string} type="number" value={config.redis.db} onChange={(e) => updateRedis({ db: Number(e.target.value) })} />
              </div>
              <label className="flex items-center justify-between gap-4 rounded-xl border border-line-2 bg-bg-2 p-4">
                <span>
                  <span className="block text-sm font-medium">{t('setup.redis.enableTls')}</span>
                  <span className="mt-1 block text-xs text-ink-3">{t('setup.redis.enableTlsHint')}</span>
                </span>
                <input type="checkbox" checked={config.redis.enable_tls} onChange={(e) => updateRedis({ enable_tls: e.target.checked })} />
              </label>
              <Button variant={redisConnected ? 'primary' : 'secondary'} loading={testingRedis} onClick={() => void testRedis()} className="w-full">
                {redisConnected && <Check className="h-4 w-4" />}
                {testingRedis ? t('setup.status.testing') : redisConnected ? t('setup.status.success') : t('setup.status.testConnection')}
              </Button>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-5">
              <StepTitle title={t('setup.admin.title')} description={t('setup.admin.description')} />
              <Input label={t('setup.admin.email') as string} type="email" value={config.admin.email} onChange={(e) => updateAdmin({ email: e.target.value })} placeholder="admin@example.com" />
              <Input label={t('setup.admin.password') as string} type="password" value={config.admin.password} onChange={(e) => updateAdmin({ password: e.target.value })} placeholder={t('setup.admin.passwordPlaceholder') as string} />
              <Input
                label={t('setup.admin.confirmPassword') as string}
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder={t('setup.admin.confirmPasswordPlaceholder') as string}
                error={confirmPassword && config.admin.password !== confirmPassword ? (t('setup.admin.passwordMismatch') as string) : undefined}
              />
            </div>
          )}

          {step === 3 && (
            <div className="space-y-5">
              <StepTitle title={t('setup.ready.title')} description={t('setup.ready.description')} />
              <ReviewItem label={t('setup.ready.database') as string} value={`${config.database.user}@${config.database.host}:${config.database.port}/${config.database.dbname}`} />
              <ReviewItem label={t('setup.ready.redis') as string} value={`${config.redis.host}:${config.redis.port} / db ${config.redis.db}${config.redis.enable_tls ? ' / TLS' : ''}`} />
              <ReviewItem label={t('setup.ready.adminEmail') as string} value={config.admin.email} />
            </div>
          )}

          {message && <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{message}</div>}

          {installSuccess && (
            <div className="mt-6 flex items-start gap-3 rounded-xl border border-green-200 bg-green-50 p-4 text-sm text-green-700">
              {serviceReady ? <Check className="mt-0.5 h-4 w-4" /> : <Spinner className="mt-0.5 h-4 w-4 text-green-700" />}
              <div>
                <div className="font-medium">{t('setup.status.completed')}</div>
                <div className="mt-1">{serviceReady ? t('setup.status.redirecting') : t('setup.status.restarting')}</div>
              </div>
            </div>
          )}

          <div className="mt-8 flex justify-between gap-3">
            {step > 0 && !installSuccess ? (
              <Button variant="secondary" onClick={() => setStep((current) => Math.max(0, current - 1))}>
                <ChevronLeft className="h-4 w-4" />
                {t('common.back')}
              </Button>
            ) : <span />}

            {step < 3 ? (
              <Button disabled={!canProceed} onClick={() => setStep((current) => current + 1)}>
                {t('common.next')}
                <ChevronRight className="h-4 w-4" />
              </Button>
            ) : !installSuccess ? (
              <Button variant="accent" loading={installing} disabled={!canProceed} onClick={() => void performInstall()}>
                {installing ? t('setup.status.installing') : t('setup.status.completeInstallation')}
              </Button>
            ) : null}
          </div>
        </Card>
      </div>
    </div>
  )
}

function StepTitle({ title, description }: { title: React.ReactNode; description: React.ReactNode }) {
  return (
    <div className="text-center">
      <h2 className="font-display text-2xl tracking-tight">{title}</h2>
      <p className="mt-1 text-sm text-ink-3">{description}</p>
    </div>
  )
}

function ReviewItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-line-2 bg-bg-2 p-4">
      <div className="text-xs uppercase tracking-[0.14em] text-ink-3 font-mono">{label}</div>
      <div className="mt-2 break-all font-mono text-sm text-ink-1">{value}</div>
    </div>
  )
}
