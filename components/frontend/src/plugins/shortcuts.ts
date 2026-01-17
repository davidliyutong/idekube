import type { App } from 'vue'
import router from '@/router'
import { useSearchStore } from '@/stores/search'
import { useShortcutsStore } from '@/stores/shortcuts'

export const shortcutsPlugin = {
  install(app: App) {
    const searchStore = useSearchStore()
    const shortcutsStore = useShortcutsStore()

    let sequence: string[] = []
    let seqTimer: number | null = null

    function resetSeq() {
      sequence = []
      if (seqTimer) window.clearTimeout(seqTimer)
      seqTimer = null
    }

    function pushSeq(key: string) {
      sequence.push(key)
      if (seqTimer) window.clearTimeout(seqTimer)
      seqTimer = window.setTimeout(resetSeq, 800)
    }

    function onKeydown(e: KeyboardEvent) {
      const isMetaK = (e.key.toLowerCase() === 'k' && (e.metaKey || e.ctrlKey))
      if (isMetaK) {
        e.preventDefault()
        searchStore.open()
        return
      }

      // '?' help
      if (e.key === '?' || (e.shiftKey && e.key === '/')) {
        e.preventDefault()
        shortcutsStore.openHelp()
        return
      }

      // Esc closes
      if (e.key === 'Escape') {
        if (searchStore.isOpen) searchStore.close()
        if (shortcutsStore.showHelp) shortcutsStore.closeHelp()
        return
      }

      // Go-to sequences: g + letter
      if (e.key.toLowerCase() === 'g') {
        pushSeq('g')
        return
      }

      if (sequence[0] === 'g') {
        const k = e.key.toLowerCase()
        if (['d','w','t','o','s'].includes(k)) {
          e.preventDefault()
          resetSeq()
          const nameMap: Record<string, string> = {
            d: 'dashboard',
            w: 'workspaces',
            t: 'templates',
            o: 'organizations',
            s: 'settings',
          }
          const routeName = nameMap[k]
          if (routeName) router.push({ name: routeName })
        }
      }
    }

    window.addEventListener('keydown', onKeydown)

    // Cleanup on app unmount (optional)
    app.mixin({
      unmounted() {
        window.removeEventListener('keydown', onKeydown)
      }
    })
  }
}
