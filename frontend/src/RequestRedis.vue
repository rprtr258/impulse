<script setup lang="ts">
import {ref} from "vue";
import {NButton, NInputGroup, NInput, NEmpty} from "naive-ui";
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import {database} from "../wailsjs/go/models";
import {use_request} from "./store";

type Request = {kind: database.Kind.REDIS} & database.RedisRequest;

const {id} = defineProps<{
  id: string,
}>();
const request = use_request<Request, database.RedisResponse>(ref(id));
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
      v-on:update:value="dsn => request.value!.update_request({dsn: dsn})"
    />
    <NButton
      type="primary"
      v-on:click="request.value.send()"
      :disabled="request.value.is_loading"
    >Send</NButton>
  </NInputGroup>
  <EditorJSON
    class="h100"
    :value="request.value.request.query ?? null"
    v-on:update="(value: string) => request.value!.update_request({query: value})"
  />
  <template v-if="request.value.response === null">
    <NEmpty
      description="Send request or choose one from history."
      class="h100"
      style="justify-content: center;"
    />
  </template>
  <template v-else>
    <ViewJSON :value="request.value.response.response" />
  </template>
</div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
