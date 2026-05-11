import { apiClient } from './client'
import { listAvailableChannels } from './channels'
import type { Group, GroupModelCatalog } from '@/types'

/** List groups visible to the current user. */
export async function getUserGroups(): Promise<Group[]> {
  const { data } = await apiClient.get<Group[]>('/groups/available')
  return data
}

/** Model catalog for a single group. */
export async function getGroupModelCatalog(groupId: number): Promise<GroupModelCatalog> {
  const [groups, channels] = await Promise.all([
    getUserGroups(),
    listAvailableChannels()
  ])
  const group = groups.find((item) => item.id === groupId)
  if (!group) {
    throw new Error(`Unknown group ${groupId}`)
  }

  const modelMap = new Map<string, GroupModelCatalog['models'][number]>()
  for (const channel of channels) {
    for (const section of channel.platforms || []) {
      const inGroup = (section.groups || []).some((g) => g.id === groupId)
      if (!inGroup) continue
      if (group.platform && section.platform && section.platform !== group.platform) continue
      for (const model of section.supported_models || []) {
        if (!model?.name) continue
        if (!modelMap.has(model.name)) {
          modelMap.set(model.name, {
            id: model.name,
            display_name: model.name,
            input_price_per_mtoken: model.pricing?.input_price ?? undefined,
            output_price_per_mtoken: model.pricing?.output_price ?? undefined
          })
        }
      }
    }
  }

  return {
    group,
    models: Array.from(modelMap.values()).sort((a, b) => a.id.localeCompare(b.id)),
    source: 'mixed'
  }
}

export const modelsAPI = { getUserGroups, getGroupModelCatalog }
export default modelsAPI
