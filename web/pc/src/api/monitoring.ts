import request from '@/utils/request'

export interface ServiceHealth {
  name: string
  display_name: string
  api_port: number
  api_status: 'healthy' | 'unhealthy' | 'unknown'
  api_error?: string
  rpc_port: number
  rpc_status: 'healthy' | 'unhealthy' | 'unknown'
  rpc_error?: string
  health_status: 'healthy' | 'unhealthy' | 'unknown'
  health_error?: string
  health_endpoint?: string
}

export interface ContainerHealth {
  name: string
  display_name: string
  image: string
  status: 'healthy' | 'unhealthy'
  running_for: string
  error?: string
}

export interface AiModelHealth {
  id: string
  name: string
  display_name: string
  provider: string
  status: 'healthy' | 'unhealthy'
  error?: string
}

export interface HealthResponse {
  timestamp: string
  overall_status: 'healthy' | 'unhealthy'
  services: ServiceHealth[]
  containers: ContainerHealth[]
  ai_models: AiModelHealth[]
}

export const getHealthStatus = () => {
  return request.get<HealthResponse>('/api/monitoring/health')
}
