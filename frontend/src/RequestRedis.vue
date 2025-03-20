<script setup lang="ts">
import {NButton, NInputGroup, NInput, NEmpty} from "naive-ui";
import {database} from '../wailsjs/go/models';
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import {useResponse} from "./store";

type Request = Omit<database.RedisRequest, "createFrom">;

const {id, request} = defineProps<{
  id: string,
  request: Request,
}>();
const emit = defineEmits<{
  send: [],
  update: [request: Request],
}>();
function updateRequest(patch: Partial<Request>) {
  emit("update", {...request, ...patch});
}
const response = useResponse<database.RedisResponse>(id);
</script>

<template>
<div class="h100" style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;">
  <NInputGroup style="grid-column: span 2;">
    <NInput
      placeholder="DSN"
      :value="request.dsn"
      v-on:update:value="dsn => updateRequest({dsn: dsn})"
    />
    <NButton type="primary" v-on:click='emit("send")'>Send</NButton>
  </NInputGroup>
  <EditorJSON
    class="h100"
    :value="request.query ?? null"
    v-on:update="(value: string) => updateRequest({query: value})"
  />
  <template v-if="response === null">
    <NEmpty
      description="Send request or choose one from history."
      class="h100"
      style="justify-content: center;"
    />
  </template>
  <template v-else>
    <ViewJSON :value="response.response" />
  </template>
</div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
