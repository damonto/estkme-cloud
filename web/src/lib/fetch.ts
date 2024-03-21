import { createFetch } from '@vueuse/core'
import { hasPinCode, retrievePinCode } from './pincode'

export default createFetch({
  baseUrl: import.meta.env.VITE_API_URL,
  fetchOptions: {
    mode: 'no-cors'
  },
  options: {
    async beforeFetch({ options }) {
      if (hasPinCode()) {
        options.headers = {
          ...options.headers,
          Authorization: `Bearer ${retrievePinCode()}`
        }
      }
      return { options }
    }
  }
})
