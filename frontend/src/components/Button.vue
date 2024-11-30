<script setup lang="ts">
import { NButton } from "naive-ui";


defineProps({
    type: "default" | "primary",
})

const emit = defineEmits({
    click: null
})
</script>

<template>
    <NButton :type="type" @click="emit('click')">
        <slot></slot>
    </NButton>
</template>

<style lang="css" scoped></style>
