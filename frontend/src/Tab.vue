<script setup lang="ts">
import {inject, onBeforeMount, onMounted, ref} from 'vue';

const {title} = defineProps<{
  title: string,
}>();

const isActive = ref(false);
onBeforeMount(() => {
  inject("huita")(title, isActive);
});
</script>

<template>
<div v-show="isActive">
  <slot></slot>
</div>
</template>

<style lang="css" scoped>
</style>
