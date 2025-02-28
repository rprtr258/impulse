import pluginVue from 'eslint-plugin-vue'
import {defineConfigWithVueTs, vueTsConfigs} from '@vue/eslint-config-typescript'

export default defineConfigWithVueTs(
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
  },
  {
    name: 'app/files-to-ignore',
    ignores: ['**/dist/**', '**/dist-ssr/**', '**/coverage/**', '**/wailsjs/**'],
  },
  {
    name: 'vue/multi-word-component-names',
    ignores: ['**/*.vue'],
  },
  ...pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
);
