const LOCAL_STORAGE_PIN_CODE_KEY: string = 'pinCode'

export function hasPinCode() {
  return !!localStorage.getItem(LOCAL_STORAGE_PIN_CODE_KEY)
}

export function retrievePinCode() {
  return localStorage.getItem(LOCAL_STORAGE_PIN_CODE_KEY)
}
