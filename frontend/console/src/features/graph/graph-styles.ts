import type { StylesheetCSS } from 'cytoscape'
import colors from 'tailwindcss/colors'

export const createGraphStyles = (isDarkMode: boolean): StylesheetCSS[] => {
  const theme = {
    primary: isDarkMode ? colors.indigo[400] : colors.indigo[200],
    background: isDarkMode ? colors.slate[700] : colors.slate[200],
    border: isDarkMode ? colors.gray[800] : colors.slate[600],
    arrow: isDarkMode ? colors.slate[300] : colors.slate[700],
    text: isDarkMode ? colors.slate[100] : colors.slate[900],
    selected: {
      bg: isDarkMode ? colors.blue[400] : colors.blue[500],
      border: isDarkMode ? colors.blue[300] : colors.blue[400],
    },
  }

  return [
    {
      selector: 'node',
      css: {
        'background-color': theme.primary,
        label: 'data(label)',
        'text-valign': 'center',
        'text-halign': 'center',
        color: theme.text,
        shape: 'round-rectangle',
        width: '120px',
        height: '40px',
        'text-wrap': 'wrap',
        'text-max-width': '100px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '12px',
        'border-width': '1px',
        'border-color': theme.border,
      },
    },
    {
      selector: 'edge',
      css: {
        width: 2,
        'line-color': theme.arrow,
        'curve-style': 'bezier',
        'target-arrow-shape': 'triangle',
        'target-arrow-color': theme.arrow,
        'arrow-scale': 1,
      },
    },
    {
      selector: '$node > node',
      css: {
        padding: '10px',
        'text-valign': 'top',
        'text-halign': 'center',
        'background-color': theme.background,
      },
    },
    {
      selector: 'node[type="groupNode"]',
      css: {
        'background-color': theme.primary,
        shape: 'round-rectangle',
        width: '180px',
        height: '120px',
        'text-valign': 'top',
        'text-halign': 'center',
        'text-wrap': 'wrap',
        'text-max-width': '120px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '14px',
        'border-width': '1px',
        'border-color': theme.border,
      },
    },
    {
      selector: ':parent',
      css: {
        'text-valign': 'top',
        'text-halign': 'center',
        'background-opacity': 1,
      },
    },
    {
      selector: '.selected',
      css: {
        'background-color': theme.selected.bg,
        'border-width': 2,
        'border-color': theme.selected.border,
      },
    },
    {
      selector: 'node[type="node"]',
      css: {
        'background-color': 'data(backgroundColor)',
        color: theme.text,
        shape: 'round-rectangle',
        width: '100px',
        height: '30px',
        'border-width': '1px',
        'border-color': theme.border,
        'text-wrap': 'wrap',
        'text-max-width': '80px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '11px',
      },
    },
  ]
}

export const nodeColors = {
  light: {
    verb: colors.indigo[500],
    config: colors.sky[400],
    data: colors.gray[400],
    database: colors.blue[400],
    secret: colors.blue[400],
    subscription: colors.violet[400],
    topic: colors.violet[400],
    default: colors.gray[400],
  },
  dark: {
    verb: colors.indigo[600],
    config: colors.blue[500],
    data: colors.gray[400],
    database: colors.blue[600],
    secret: colors.blue[500],
    subscription: colors.violet[600],
    topic: colors.violet[600],
    default: colors.gray[700],
  },
}

export const getNodeBackgroundColor = (isDarkMode: boolean, nodeType: string): string => {
  const theme = isDarkMode ? nodeColors.dark : nodeColors.light
  return theme[nodeType as keyof typeof theme] || theme.default
}
