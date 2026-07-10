// Simple Mustache-like template renderer for Prompt templates
// Supports: {{var}}, {{#condition}}...{{/condition}}, {{^condition}}...{{/condition}}

import fs from 'fs'
import path from 'path'

/**
 * Render a template with given context
 * @param {string} template - Template string with {{...}} placeholders
 * @param {object} context - Variables to substitute
 * @returns {string} Rendered template
 */
export function render(template, context) {
  let result = template

  // Handle {{#condition}}...{{/condition}} (if truthy)
  result = result.replace(/\{\{#(\w+)\}\}([\s\S]*?)\{\{\/\1\}\}/g, (match, key, content) => {
    return context[key] ? render(content, context) : ''
  })

  // Handle {{^condition}}...{{/condition}} (if falsy)
  result = result.replace(/\{\{\^(\w+)\}\}([\s\S]*?)\{\{\/\1\}\}/g, (match, key, content) => {
    return !context[key] ? render(content, context) : ''
  })

  // Handle {{variable}} (simple substitution)
  result = result.replace(/\{\{(\w+)\}\}/g, (match, key) => {
    return context[key] !== undefined ? String(context[key]) : match
  })

  return result
}

/**
 * Load and render a template file
 * @param {string} templateName - Template filename (without .md extension)
 * @param {object} context - Variables to substitute
 * @returns {string} Rendered template
 */
export function renderFile(templateName, context) {
  const templatesDir = path.join(process.cwd(), '.harness/agents/prompts/templates')
  const templatePath = path.join(templatesDir, `${templateName}.md`)

  if (!fs.existsSync(templatePath)) {
    throw new Error(`Template not found: ${templatePath}`)
  }

  const template = fs.readFileSync(templatePath, 'utf-8')
  return render(template, context)
}

// For use in workflow scripts (Node.js style imports not supported)
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { render, renderFile }
}
