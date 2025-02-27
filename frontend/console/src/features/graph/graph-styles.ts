import colors from 'tailwindcss/colors'

// Define the base colors for each node type
const nodeColorMap = {
  verb: colors.indigo,
  ingress: colors.green,
  cronjob: colors.blue,
  subscriber: colors.violet,
  sqlquery: colors.teal,
  config: colors.sky,
  data: colors.green,
  database: colors.teal,
  secret: colors.sky,
  subscription: colors.violet,
  topic: colors.violet,
  enum: colors.zinc,
  default: colors.gray,
}

// Define the shade for light and dark modes
const LIGHT_SHADE = '500'
const DARK_SHADE = '600'

// Generate the nodeColors object
export const nodeColors = {
  light: Object.fromEntries(Object.entries(nodeColorMap).map(([key, colorObj]) => [key, colorObj[LIGHT_SHADE]])),
  dark: Object.fromEntries(Object.entries(nodeColorMap).map(([key, colorObj]) => [key, colorObj[DARK_SHADE]])),
}

export const getNodeBackgroundColor = (isDarkMode: boolean, nodeType: string): string => {
  const theme = isDarkMode ? nodeColors.dark : nodeColors.light
  return theme[nodeType as keyof typeof theme] || theme.default
}
