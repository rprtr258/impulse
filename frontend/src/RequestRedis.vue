<script setup lang="ts">
import {NButton, NInputGroup, NInput, NEmpty} from "naive-ui";
import {database} from '../wailsjs/go/models';
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import {use_request, use_response, useStore} from "./store";

type Request = {kind: database.Kind.REDIS} & Omit<database.RedisRequest, "createFrom">;
const store = useStore();

const {id} = defineProps<{
  id: string,
}>();
const emit = defineEmits<{
  update: [request: Request],
}>();
const request = use_request<Request>(id);
const response = use_response<database.RedisResponse>(() => id);
function updateRequest(patch: Partial<Request>) {
  emit("update", {...request.value!.request, ...patch});
}
</script>

<template>
<NEmpty
  v-if="request.value === null"
  description="Loading request..."
  class="h100"
  style="justify-content: center;"
/>
<div
  v-else
  class="h100"
  style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;"
>
  <NInputGroup style="grid-column: span 2;">
    <NInput
      placeholder="DSN"
      :value="request.value.request.dsn"
      v-on:update:value="dsn => updateRequest({dsn: dsn})"
    />
    <NButton type="primary" v-on:click='store.send(id)'>Send</NButton>
  </NInputGroup>
  <EditorJSON
    class="h100"
    :value="request.value.request.query ?? null"
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
