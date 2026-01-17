<template>
  <va-modal v-model="isOpen" title="Keyboard Shortcuts" size="large" okText="Close" @ok="close">
    <div class="shortcuts-help">
      <div class="grid">
        <div class="item"><span class="key">⌘K / Ctrl+K</span><span class="desc">Open Global Search</span></div>
        <div class="item"><span class="key">?</span><span class="desc">Open Shortcuts Help</span></div>
        <div class="item"><span class="key">Esc</span><span class="desc">Close dialogs / modals</span></div>
        <div class="item"><span class="key">g then d</span><span class="desc">Go to Dashboard</span></div>
        <div class="item"><span class="key">g then w</span><span class="desc">Go to Workspaces</span></div>
        <div class="item"><span class="key">g then t</span><span class="desc">Go to Templates</span></div>
        <div class="item"><span class="key">g then o</span><span class="desc">Go to Organizations</span></div>
        <div class="item"><span class="key">g then s</span><span class="desc">Go to Settings</span></div>
      </div>
    </div>
  </va-modal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useShortcutsStore } from '../stores/shortcuts'

const shortcutsStore = useShortcutsStore()
const isOpen = computed({
  get: () => shortcutsStore.showHelp,
  set: (v: boolean) => (v ? shortcutsStore.openHelp() : shortcutsStore.closeHelp()),
})

function close() { shortcutsStore.closeHelp() }
</script>

<style scoped lang="scss">
.shortcuts-help {
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem 1rem;

    .item {
      display: flex;
      align-items: center;
      gap: 0.75rem;

      .key {
        font-family: 'Monaco', 'Menlo', monospace;
        background: var(--va-background-shade);
        border: 1px solid var(--va-background-border);
        border-radius: 6px;
        padding: 0.25rem 0.5rem;
        min-width: 120px;
        text-align: center;
      }

      .desc {
        color: var(--va-textColorSecondary);
      }
    }
  }
}
</style>
