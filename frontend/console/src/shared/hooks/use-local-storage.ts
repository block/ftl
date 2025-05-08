import { useEffect, useState } from 'react'

export const useLocalStorage = <T>(key: string, initialValue: T) => {
  const [value, setValue] = useState<T>(() => {
    try {
      const jsonValue = localStorage.getItem(key)
      if (jsonValue === null || jsonValue === 'undefined') return initialValue
      return JSON.parse(jsonValue) as T
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error)
      return initialValue
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch (error) {
      console.warn(`Error saving to localStorage key "${key}":`, error)
    }
  }, [key, value])

  return [value, setValue] as const
}
