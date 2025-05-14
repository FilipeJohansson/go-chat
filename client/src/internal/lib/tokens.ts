
export type Tokens = {
  accessToken: string
  refreshToken: string
}

export const getTokens = (): Tokens | undefined => {
  const value: string | null  = window.sessionStorage.getItem("tokens")
  if (!value) {
    console.error("Error getting tokens")
    return undefined
  }
  return JSON.parse(value) as Tokens
}

export const saveTokens = (tokens: Tokens): void => {
  const value: string = JSON.stringify(tokens)
  window.sessionStorage.setItem("tokens", value)
}

export const clearTokens = () => {
  window.sessionStorage.removeItem("tokens")
}
