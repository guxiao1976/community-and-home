// ============================================================
// Service Registry Loader (replaces hardcoded VALID_SERVICES)
// ============================================================

const fs = require('fs')
const path = require('path')

function loadServiceRegistry() {
  const registryPath = path.join(process.cwd(), '.harness/registry/services.json')

  if (!fs.existsSync(registryPath)) {
    throw new Error(
      `Service registry not found at ${registryPath}\n` +
      `Run: bash .harness/scripts/build-service-registry.sh`
    )
  }

  const registry = JSON.parse(fs.readFileSync(registryPath, 'utf-8'))

  return {
    services: registry.services.map(s => s.name),
    web: registry.web.map(w => w.name),
    getService: (name) => registry.services.find(s => s.name === name),
    getServiceModule: (name) => {
      const svc = registry.services.find(s => s.name === name)
      return svc ? svc.module : null
    },
  }
}

const ServiceRegistry = loadServiceRegistry()
const VALID_SERVICES = ServiceRegistry.services
const VALID_WEB = ServiceRegistry.web
const ALL_VALID = [...VALID_SERVICES, ...VALID_WEB]

// Export for CommonJS
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { ServiceRegistry, VALID_SERVICES, VALID_WEB, ALL_VALID }
}
