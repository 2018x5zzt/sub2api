import { create } from 'zustand'
import { announcementsAPI } from '@/api/announcements'
import type { UserAnnouncement } from '@/types'

const THROTTLE_MS = 20 * 60 * 1000 // 20 minutes

interface State {
  announcements: UserAnnouncement[]
  loading: boolean
  popupQueue: UserAnnouncement[]
  currentPopup: UserAnnouncement | null
  shownPopupIds: Set<number>
  lastFetchTime: number
}

interface Actions {
  unreadCount: () => number
  fetch: (force?: boolean) => Promise<void>
  dismissPopup: () => void
  markAsRead: (id: number) => Promise<void>
  reset: () => void
}

const initial: State = {
  announcements: [],
  loading: false,
  popupQueue: [],
  currentPopup: null,
  shownPopupIds: new Set<number>(),
  lastFetchTime: 0
}

export const useAnnouncementStore = create<State & Actions>((set, get) => ({
  ...initial,

  unreadCount: () => get().announcements.filter((a) => !a.read_at).length,

  fetch: async (force = false) => {
    const now = Date.now()
    const last = get().lastFetchTime
    if (!force && last > 0 && now - last < THROTTLE_MS) return

    set({ lastFetchTime: now, loading: true })
    try {
      const all = await announcementsAPI.list(false)
      const announcements = all.slice(0, 20)
      const shown = get().shownPopupIds
      const newPopups = announcements.filter(
        (a) => a.notify_mode === 'popup' && !a.read_at && !shown.has(a.id)
      )

      const queue = [...get().popupQueue]
      for (const p of newPopups) {
        if (!queue.some((q) => q.id === p.id)) queue.push(p)
      }

      // If nothing currently shown and we have something queued, promote one
      let current = get().currentPopup
      if (!current && queue.length > 0) {
        current = queue.shift()!
        shown.add(current.id)
      }

      set({
        announcements,
        popupQueue: queue,
        currentPopup: current,
        shownPopupIds: new Set(shown)
      })
    } catch (err) {
      // Revert throttle timestamp on failure so retry is allowed
      set({ lastFetchTime: 0 })
      console.error('Failed to fetch announcements:', err)
    } finally {
      set({ loading: false })
    }
  },

  dismissPopup: () => {
    const cur = get().currentPopup
    if (!cur) return
    const id = cur.id
    const queue = [...get().popupQueue]
    const next = queue.shift() || null
    if (next) get().shownPopupIds.add(next.id)
    set({
      currentPopup: next,
      popupQueue: queue,
      announcements: get().announcements.map((a) =>
        a.id === id ? { ...a, read_at: a.read_at || new Date().toISOString() } : a
      )
    })
    // Fire-and-forget mark-as-read
    get().markAsRead(id)
  },

  markAsRead: async (id) => {
    try {
      await announcementsAPI.markRead(id)
    } catch (e) {
      console.error('markRead failed:', e)
    }
  },

  reset: () => set({ ...initial, shownPopupIds: new Set<number>() })
}))
