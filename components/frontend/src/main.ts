import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createVuestic } from 'vuestic-ui'
import 'vuestic-ui/css'

import App from './App.vue'
import router from './router'
import { shortcutsPlugin } from './plugins/shortcuts'

import './styles/global.scss'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(createVuestic({
  config: {
    colors: {
      variables: {
        primary: '#2c82e0',
        secondary: '#767C88',
        success: '#40e583',
        info: '#2c82e0',
        danger: '#e34b4a',
        warning: '#ffc200',
        gray: '#8396a5',
        dark: '#34495e',
      },
    },
  },
}))

app.use(shortcutsPlugin)

app.mount('#app')
