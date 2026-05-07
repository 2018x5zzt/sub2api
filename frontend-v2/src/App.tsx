import { useEffect } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppRouter } from '@/router'
import { useAuthStore } from '@/stores/auth'
import { ToastViewport } from '@/components/ui/Toast'
import { AnnouncementPopup } from '@/components/AnnouncementPopup'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
      staleTime: 30_000
    }
  }
})

export default function App() {
  const init = useAuthStore((s) => s.init)
  const loadPublicSettings = useAuthStore((s) => s.loadPublicSettings)

  useEffect(() => {
    init()
    loadPublicSettings()
  }, [init, loadPublicSettings])

  return (
    <QueryClientProvider client={queryClient}>
      <AppRouter />
      <ToastViewport />
      <AnnouncementPopup />
    </QueryClientProvider>
  )
}
