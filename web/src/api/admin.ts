import axios from 'axios'

const TOKEN_KEY = 'admin_token'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
    }
    return Promise.reject(err)
  },
)

export interface GatewayConfig {
  address: string
  port: number
  protocol: string
  upstream_url: string
}

export interface InitPayload {
  gateway: GatewayConfig
  admin: {
    username: string
    password: string
    confirm_password: string
    email: string
  }
}

export interface LoginPayload {
  username: string
  password: string
}

export async function getStatus(): Promise<{
  initialized: boolean
  has_gateway_config: boolean
  has_admin: boolean
}> {
  const { data } = await api.get('/system/init-status')
  return data.data
}

export async function initSystem(payload: InitPayload): Promise<void> {
  await api.post('/system/init', payload)
}

export async function login(payload: LoginPayload): Promise<{
  token: string
  admin: { username: string; email: string }
}> {
  const { data } = await api.post('/admin/login', payload)
  localStorage.setItem(TOKEN_KEY, data.token)
  return data
}

export async function getDashboard(): Promise<{
  gateway: GatewayConfig
  admin: { username: string; email: string }
  version: string
}> {
  const { data } = await api.get('/admin/dashboard')
  return data.data
}

export function logout() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export interface FullConfig {
  server: {
    port: number
    tls: { cert: string; key: string }
  }
  upstream: { url: string }
  oauth: { access_token: string; refresh_token: string; expires_at: number }
  auth: { tokens: { name: string; token: string }[] }
  identity: { device_id: string; email: string }
  env: {
    platform: string
    platform_raw: string
    arch: string
    node_version: string
    terminal: string
    package_managers: string
    runtimes: string
    is_running_with_bun: boolean
    is_ci: boolean
    is_claude_ai_auth: boolean
    version: string
    version_base: string
    build_time: string
    deployment_environment: string
    vcs: string
  }
  prompt_env: {
    platform: string
    shell: string
    os_version: string
    working_dir: string
  }
  process: {
    constrained_memory: number
    rss_range: [number, number]
    heap_total_range: [number, number]
    heap_used_range: [number, number]
  }
  logging: {
    level: string
    audit: boolean
  }
}

export async function getConfig(): Promise<FullConfig> {
  const { data } = await api.get('/config')
  return data.data
}

export async function updateConfig(config: FullConfig): Promise<void> {
  await api.put('/config', { config })
}

export async function reloadConfig(): Promise<void> {
  await api.post('/config/reload')
}
