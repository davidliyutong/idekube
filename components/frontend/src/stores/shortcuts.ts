import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useShortcutsStore = defineStore('shortcuts', () => {
  const showHelp = ref(false)

  function openHelp() { showHelp.value = true }
  function closeHelp() { showHelp.value = false }

  return { showHelp, openHelp, closeHelp }
})
