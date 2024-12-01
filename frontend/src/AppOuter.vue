<script setup lang="ts">
import {
  NConfigProvider, darkTheme,
  NNotificationProvider,
} from "naive-ui";
import App from "./App.vue";
</script>

<template>
<NConfigProvider :theme='darkTheme' class="h100">
<NNotificationProvider :max="1" placement="bottom-right">
  <App />
</NNotificationProvider>
</NConfigProvider>
</template>
