import fs from 'fs'
import path from 'path'

const readSvgSymbols = (directories) => {
  const symbols = []

  const visit = (directory, pluginName = '') => {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      const filePath = path.join(directory, entry.name)
      if (entry.isDirectory()) {
        visit(filePath, pluginName || (directory.includes('/src/plugin/') ? entry.name : ''))
      } else if (entry.name.endsWith('.svg')) {
        const source = fs.readFileSync(filePath, 'utf8').replace(/[\r\n]/g, '')
        const id = `${pluginName ? `${pluginName}-` : ''}${path.basename(entry.name, '.svg')}`
        const symbol = source
          .replace(/<svg\b([^>]*)>/i, (_, attributes) => {
            const viewBox = /\bviewBox\s*=\s*(["'])[^"']*\1/i.test(attributes)
            const width = attributes.match(/\bwidth\s*=\s*(["'])([^"']*)\1/i)?.[2]
            const height = attributes.match(/\bheight\s*=\s*(["'])([^"']*)\1/i)?.[2]
            const cleaned = attributes.replace(/\s+(?:width|height)\s*=\s*(["'])[^"']*\1/gi, '')
            const fallbackViewBox = viewBox || !width || !height ? '' : ` viewBox="0 0 ${width} ${height}"`
            return `<symbol id="${id}"${cleaned}${fallbackViewBox}>`
          })
          .replace(/<\/svg>/i, '</symbol>')
        symbols.push(symbol)
      }
    }
  }

  for (const directory of directories) {
    visit(directory)
  }
  return symbols
}

export const svgSpritePlugin = (directories) => ({
  name: 'local-svg-sprite',
  transformIndexHtml(html) {
    const symbols = readSvgSymbols(directories)
    return html.replace('<body>', `<body>\n  <svg xmlns="http://www.w3.org/2000/svg" style="position:absolute;width:0;height:0">${symbols.join('')}</svg>`)
  }
})
