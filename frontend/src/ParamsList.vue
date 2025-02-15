<script setup lang="ts">
import {computed} from 'vue';
import {NDynamicInput} from 'naive-ui';
import {database} from '../wailsjs/go/models';

const {value} = defineProps<{
  value?: database.KV[],
}>();
const emit = defineEmits<{
  update: [value: database.KV[]],
}>();

const valueNonNull = computed(() => value ?? []);
</script>

<template>
<NDynamicInput
  :value='valueNonNull.concat([{key: "", value: ""}])'
  @update:value='value => emit("update", value as database.KV[])'
  preset="pair"
  key-placeholder="Header"
  value-placeholder="Value"
/>
</template>

<style lang="css">
div.n-dynamic-input-preset-pair > div:nth-child(1) {
  margin-right: 4px !important;
}
</style>
