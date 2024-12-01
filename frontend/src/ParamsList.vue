<script setup lang="ts">
import {computed} from 'vue';
import {NDynamicInput} from 'naive-ui';
import {Parameter} from './api';

const {value} = defineProps<{
  value?: Parameter[] | null | undefined, // TODO: optional govno is ignored in typing
}>();
const emit = defineEmits<{
  update: [value: Parameter[]],
}>();

const valueNonNull = computed(() => value ?? []);
</script>

<template>
<NDynamicInput
  :value='valueNonNull.concat([{key: "", value: ""}])'
  @update:value='value => emit("update", value as Parameter[])'
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
