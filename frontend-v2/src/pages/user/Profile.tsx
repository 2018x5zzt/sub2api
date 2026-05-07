import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation } from '@tanstack/react-query'
import { Lock, User as UserIcon } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/api/user'
import { toast } from '@/components/ui/Toast'

export default function ProfilePage() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const refreshUser = useAuthStore((s) => s.refreshUser)

  const [username, setUsername] = useState(user?.username || '')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  const updateMut = useMutation({
    mutationFn: (payload: { username: string }) => userAPI.updateProfile(payload),
    onSuccess: async () => {
      await refreshUser()
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const passwordMut = useMutation({
    mutationFn: ({ oldP, newP }: { oldP: string; newP: string }) => userAPI.changePassword(oldP, newP),
    onSuccess: () => {
      toast.success(t('common.success') as string)
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function onSubmitProfile(e: FormEvent) {
    e.preventDefault()
    updateMut.mutate({ username })
  }

  function onSubmitPassword(e: FormEvent) {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error(t('auth.passwordsDoNotMatch') as string)
      return
    }
    passwordMut.mutate({ oldP: oldPassword, newP: newPassword })
  }

  return (
    <>
      <PageHeader title={t('profile.title')} description={t('profile.description') as string} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card className="p-6">
          <h2 className="font-medium text-ink-1 mb-4">{t('common.name')}</h2>
          <form onSubmit={onSubmitProfile} className="space-y-4">
            <Input
              name="email"
              label={t('common.email') as string}
              value={user?.email || ''}
              leftIcon={<UserIcon className="h-4 w-4" />}
              disabled
            />
            <Input
              name="username"
              label={t('common.name') as string}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              leftIcon={<UserIcon className="h-4 w-4" />}
            />
            <Button type="submit" loading={updateMut.isPending}>
              {t('common.save')}
            </Button>
          </form>
        </Card>

        <Card className="p-6">
          <h2 className="font-medium text-ink-1 mb-4">{t('auth.passwordLabel')}</h2>
          <form onSubmit={onSubmitPassword} className="space-y-4">
            <Input
              name="old_password"
              type="password"
              label={t('auth.passwordLabel') as string}
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              leftIcon={<Lock className="h-4 w-4" />}
              required
            />
            <Input
              name="new_password"
              type="password"
              label={t('auth.newPassword') as string}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              leftIcon={<Lock className="h-4 w-4" />}
              required
              minLength={6}
            />
            <Input
              name="confirm_password"
              type="password"
              label={t('auth.confirmPassword') as string}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              leftIcon={<Lock className="h-4 w-4" />}
              required
            />
            <Button type="submit" loading={passwordMut.isPending}>
              {t('auth.resetPassword')}
            </Button>
          </form>
        </Card>
      </div>
    </>
  )
}
