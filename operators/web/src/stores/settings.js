import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useSettingsStore = defineStore('settings-store', {
  state: () => ({
    straightLinks: false,
    darkTheme: false,
    selectedFileType: 'YAML',
    selectedInterval: '1m',
    maxLogLines: '1000',
  }),
  persist: true,
})