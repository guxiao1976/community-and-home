// 调试日志工具

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

class Logger {
  private isDevelopment = import.meta.env.DEV

  private log(level: LogLevel, message: string, data?: any) {
    if (!this.isDevelopment && level === 'DEBUG') {
      return
    }

    const timestamp = new Date().toISOString()
    const prefix = `[${timestamp}] [${level}]`

    switch (level) {
      case 'DEBUG':
        console.log(`${prefix} ${message}`, data || '')
        break
      case 'INFO':
        console.info(`${prefix} ${message}`, data || '')
        break
      case 'WARN':
        console.warn(`${prefix} ${message}`, data || '')
        break
      case 'ERROR':
        console.error(`${prefix} ${message}`, data || '')
        break
    }
  }

  debug(message: string, data?: any) {
    this.log('DEBUG', message, data)
  }

  info(message: string, data?: any) {
    this.log('INFO', message, data)
  }

  warn(message: string, data?: any) {
    this.log('WARN', message, data)
  }

  error(message: string, data?: any) {
    this.log('ERROR', message, data)
  }

  // API请求日志
  apiRequest(method: string, url: string, data?: any) {
    this.debug(`API Request: ${method} ${url}`, data)
  }

  apiResponse(method: string, url: string, status: number, data?: any) {
    this.debug(`API Response: ${method} ${url} [${status}]`, data)
  }

  apiError(method: string, url: string, error: any) {
    this.error(`API Error: ${method} ${url}`, {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data
    })
  }

  // 路由日志
  routeChange(from: string, to: string) {
    this.info(`Route Change: ${from} → ${to}`)
  }

  // 组件生命周期日志
  componentMounted(componentName: string) {
    this.debug(`Component Mounted: ${componentName}`)
  }

  componentUnmounted(componentName: string) {
    this.debug(`Component Unmounted: ${componentName}`)
  }
}

export const logger = new Logger()
