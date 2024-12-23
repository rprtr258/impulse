<script setup lang="ts">
import {ref, onMounted, provide, inject, useSlots, Ref} from "vue";

const active = ref(0);
const tabs = ref<[string, Ref<boolean>][]>([]);
onMounted(() => {
  const slots = useSlots();
  console.log("SLOTS", slots);
});
provide("huita", (title, isActive) => {
  isActive.value = tabs.value.length === active.value;
  tabs.value.push([title, isActive]);
});

function selectTab(index: number) {
  if (index === active.value) {
    return;
  }

  tabs.value[active.value][1].value = false;
  tabs.value[index][1].value = true;
  active.value = index;
}
</script>

<template>
<div>
  <span
    v-for="(tab, i) in tabs"
    :key="i"
    v-on:click="() => selectTab(i)"
  >
    {{ tab[0] }}
  </span>
  <div
    style="height: 100%;"
  >
    <slot></slot>
  </div>
</div>
</template>

<style lang="css" scoped>
</style>
