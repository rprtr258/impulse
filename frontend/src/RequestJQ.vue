<script setup lang="ts">
import {computed, toRefs, ref} from "vue";
import {NInput, NButton, NInputGroup, NEmpty} from "naive-ui";
import {database} from "wailsjs/go/models";
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import {use_request} from "./store";

type Request = {kind: database.Kind.JQ} & database.JQRequest;

const {id} = defineProps<{
  id: string,
}>();

const {request, response, is_loading, update_request, send} = toRefs(use_request<Request, database.JQResponse>(ref(id)));
const jqerror = ref<string | null>(null); // TODO: use
const responseText = computed(() => (response.value?.response ?? []).join("\n"));
</script>

<template>
<NEmpty
  v-if="request === null"
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
      placeholder="JQ query"
      :status='jqerror !== null ? "error" : "success"'
      :value="request.query"
      v-on:update:value="query => update_request({query: query})"
    />
    <!-- TODO: autosend -->
    <NButton
      type="primary"
      v-on:click="send()"
      :disabled="is_loading"
    >Send</NButton>
  </NInputGroup>
  <EditorJSON
    class="h100"
    :value="request.json"
    v-on:update="(value: string) => update_request({json: value})"
  />
  <div v-if="jqerror !== null" style="position: fixed; color: red; bottom: 3em;">{{jqerror}}</div>
  <template v-if="response === null">
    <NEmpty
      description="Send request or choose one from history."
      class="h100"
      style="justify-content: center;"
    />
  </template>
  <template v-else>
    <ViewJSON :value="responseText"></ViewJSON>
  </template>
</div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
