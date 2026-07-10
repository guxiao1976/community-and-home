# Templates Directory

This directory previously contained template files that have been **replaced by the new modular architecture**.

## ⚠️ Deprecated Files

The following files are **no longer used** and kept only for reference:

- `harness-pipeline.template.js` - Replaced by modular system
- `harness-gate-check.template.sh` - Integrated into workflow
- `ci-cd.template.yml` - Static file in `.github/workflows/`
- `quick-deploy.sh` - Deprecated

## ✅ New Architecture

### Problem 1: Pipeline Template Modularization

**Old**: Monolithic template with hardcoded prompts
**New**: Modular system with Markdown templates

```
.harness/agents/prompts/templates/
├── generator.md      - Markdown template with syntax highlighting
├── qa.md
├── review.md
└── debug.md

Build script: .harness/scripts/build-pipeline-new.sh
Output: .harness/workflows/harness-pipeline.js
```

### Problem 3: Service Metadata

**Old**: Hardcoded service lists
**New**: Auto-generated from `.service.json`

```
services/*/.service.json  → .harness/registry/services.json
```

### Problem 8: Circuit Breaker

**Old**: `circuit_breaker.py` (Python)
**New**: `circuit_breaker.sh` (Shell)

## 📝 Migration Status

- ✅ Pipeline templates → Markdown
- ✅ Service mappings → Registry
- ✅ Circuit breaker → Shell
- ⏸️ CI/CD templates (still in use)

## 🗑️ Cleanup Recommendation

These deprecated files can be safely deleted after verification:
- `harness-pipeline.template.js`
- `harness-gate-check.template.sh`
- `quick-deploy.sh`

Keep:
- `ci-cd.template.yml` (still used by quick-deploy)
- This README for documentation
