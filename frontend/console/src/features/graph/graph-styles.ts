import colors from 'tailwindcss/colors'

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
