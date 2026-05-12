import axios from 'axios'

const setupClient = axios.create({
  baseURL: '',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

setupClient.interceptors.response.use((response) => {
  const body = response.data
  if (body && typeof body === 'object' && 'code' in body) {
    if (body.code === 0) {
      response.data = body.data
      return response
    }
    return Promise.reject({ message: body.message || body.detail || 'Setup request failed' })
  }
  return response
})

export interface SetupStatus {
  needs_setup: boolean
  step: string
}

export interface DatabaseConfig {
  host: string
  port: number
  user: string
  password: string
  dbname: string
  sslmode: string
}

export interface RedisConfig {
  host: string
  port: number
  password: string
  db: number
  enable_tls: boolean
}

export interface AdminConfig {
  email: string
  password: string
}

export interface ServerConfig {
  host: string
  port: number
  mode: string
}

export interface InstallRequest {
  database: DatabaseConfig
  redis: RedisConfig
  admin: AdminConfig
  server: ServerConfig
}

export interface InstallResponse {
  message: string
  restart: boolean
}

export async function getSetupStatus() {
  const { data } = await setupClient.get<SetupStatus>('/setup/status', { headers: { 'Cache-Control': 'no-store' } })
  return data
}

export async function testDatabase(config: DatabaseConfig) {
  await setupClient.post('/setup/test-db', config)
}

export async function testRedis(config: RedisConfig) {
  await setupClient.post('/setup/test-redis', config)
}

export async function install(config: InstallRequest) {
  const { data } = await setupClient.post<InstallResponse>('/setup/install', config)
  return data
}

export const setupAPI = { getSetupStatus, testDatabase, testRedis, install }
