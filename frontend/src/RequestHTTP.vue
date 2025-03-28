<script setup lang="ts">
import {h, ref, toRefs, VNodeChild} from "vue";
import {
  NTag, NTabs, NTabPane,
  NInput, NButton, NInputGroup, NSelect,
  NTable, NEmpty,
} from "naive-ui";
import {Method as Methods} from "./api";
import {database} from '../wailsjs/go/models';
import ViewJSON from "./ViewJSON.vue";
import EditorJSON from "./EditorJSON.vue";
import ParamsList from "./ParamsList.vue";
import {use_request} from "./store";

type Request = {kind: database.Kind.HTTP} & database.HTTPRequest;

const {id} = defineProps<{
  id: string,
}>();
const {request, response, is_loading, update_request, send} = toRefs(use_request<Request, database.HTTPResponse>(ref(id)));

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

function responseBadge(): VNodeChild {
  const code = response.value!.code;
  return h(NTag, {
    type: code < 300 ? "success"
        : code < 500 ? "warning"
        : "error",
    size: "small",
    round: true,
  }, () => code ?? "N/A");
}
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
    <NSelect
      :options="Object.keys(Methods).map(method => ({label: method, value: method}))"
      :value="request.method"
      v-on:update:value="method => update_request({method: method})"
      style="width: 10%; min-width: 8em;"
    />
    <NInput
      placeholder="URL"
      :value="request.url"
      v-on:update:value="url => update_request({url: url})"
    />
    <NButton
      type="primary"
      v-on:click='send()'
      :disabled="is_loading"
    >Send</NButton>
  </NInputGroup>
  <NTabs
    type="line"
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
        v-on:update="(value: string) => update_request({body: value})"
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
        v-on:update='(value: database.KV[]) => update_request({headers: value.filter(({key, value}) => key!=="" || value!=="")})'
      />
    </NTabPane>
  </NTabs>
  <template v-if="response === null">
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
        :tab="responseBadge"
        disabled
        display-directive="show"
      ></NTabPane>
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
