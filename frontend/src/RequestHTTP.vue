<script setup lang="ts">
import {NTag, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NEmpty} from "naive-ui";
import {Method as Methods, ResponseHTTP} from "./api";
import {database} from '../wailsjs/go/models';
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import ParamsList from "./ParamsList.vue";

type HTTPRequest = Omit<database.HTTPRequest, "createFrom">;

const {request, response} = defineProps<{
  request: HTTPRequest,
  response?: ResponseHTTP,
}>();
const emit = defineEmits<{
  send: [],
  update: [request: HTTPRequest],
}>();
function updateRequest(patch: Partial<HTTPRequest>) {
  emit("update", {...request, ...patch});
}

function responseBodyLanguage(contentType: string): string {
  for (const [key, value] of Object.entries({
    "application/json;": "json",
    "text/html;": "html",
  })) {
    if (contentType.startsWith(key)) {
      return value;
    }
  }
  return "text";
};

function updateHeaders(value: database.KV[]){
  updateRequest({
    headers: value.filter(({key, value}) => key!=="" || value!==""),
  })
}
</script>

<template>
<div class="h100" style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;">
  <NInputGroup style="grid-column: span 2;">
    <NSelect
      :options="Object.keys(Methods).map(method => ({label: method, value: method}))"
      :value="request.method"
      v-on:update:value="method => updateRequest({method: method})"
      style="width: 10%; min-width: 8em;"
    />
    <NInput
      placeholder="URL"
      :value="request.url"
      v-on:update:value="url => updateRequest({url: url})"
    />
    <NButton type="primary" v-on:click='emit("send")'>Send</NButton>
  </NInputGroup>
  <NTabs
    type="card"
    size="small"
    class="h100"
  >
    <NTabPane
      name="tab-req-request"
      tab='Request'
      class="h100"
      display-directive="show"
    >
      <EditorJSON
        class="h100"
        :value="request.body ?? null"
        v-on:update="(value: string) => updateRequest({body: value})"
      />
    </NTabPane>
    <NTabPane
      name="tab-req-headers"
      tab='Headers'
      style="display: flex; flex-direction: column; flex: 1;"
      display-directive="show"
    >
      <ParamsList
        :value="request.headers"
        v-on:update="(value: database.KV[]) => updateHeaders(value)"
      />
    </NTabPane>
  </NTabs>
  <template v-if="response === undefined">
    <NEmpty
      description="Send request or choose one from history."
      class="h100"
      style="justify-content: center;"
    />
  </template>
  <template v-else>
    <NTabs
      type="card"
      size="small"
      default-value="tab-resp-body"
      style="overflow-y: auto;"
    >
      <NTabPane
        name="tab-resp-code"
        disabled
        display-directive="show"
      ><template>
        <NTag
          :type='response.code < 300 ? "success" : response.code < 500 ? "warning" : "error"'
          size="small"
          round
        >{{response.code ?? "N/A"}}</Ntag>
      </template></NTabPane>
      <NTabPane
        name="tab-resp-body"
        tab="Body"
        style="overflow-y: auto;"
        display-directive="show"
      >
        <ViewJSON :value="response.body" />
      </NTabPane>
      <NTabPane
        name="tab-resp-headers"
        tab="Headers"
        style="flex: 1;"
        display-directive="show"
      >
        <NTable striped size="small" single-column :single-line="false">
          <colgroup>
            <col style="width: 50%" />
            <col style="width: 50%" />
          </colgroup>
          <thead>
            <tr>
              <th>NAME</th>
              <th>VALUE</th>
            </tr>
          </thead>
          <tr v-for="header in response.headers" :key="header.key">
            <td>{{header.key}}</td>
            <td>{{header.value}}</td>
          </tr>
        </NTable>
      </NTabPane>
    </NTabs>
  </template>
</div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
