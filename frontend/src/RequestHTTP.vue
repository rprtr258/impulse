<script setup lang="ts">
import {} from "vue";
import {NTag, NCode, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NDynamicInput, NEmpty} from "naive-ui";
import {api, Method as Methods, RequestHTTP, ResponseHTTP} from "./api";

import hljs from "highlight.js/lib/core";
import json from "highlight.js/lib/languages/json";
import xml from "highlight.js/lib/languages/xml";
hljs.registerLanguage("json", json);
hljs.registerLanguage("html", xml);
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

const {id} = defineProps<{
  id: string,
}>();

let request = defineModel<RequestHTTP>("request");
let response = defineModel<ResponseHTTP | null>("response");

function send() {
  api
    .requestPerform(id)
    .then(v => {
      if (v.kind === "http") {
        response.value = {
          code: v.code,
          body: v.body,
          headers: v.headers,
        };
      } else {
        // TODO: fail
        throw new Error("Unexpected response kind: " + v.kind);
      }
    })
    .catch((err) => alert(`Could not perform request: ${err}`));
}
</script>

<template>
<div class="h100" style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;">
  <NInputGroup style="grid-column: span 2;">
    <NSelect
      :options="Object.keys(Methods).map(method => ({label: method, value: method}))"
      v-model:value="request.method"
      style="width: 10%; min-width: 8em;"
    />
    <NInput v-model:value="request.url"/>
    <NButton type="primary" v-on:click="send()">Send</NButton>
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
    >
      <NInput
        type="textarea"
        class="h100"
        :value='request.body'
        @update:value="value => request.body = value"
      />
    </NTabPane>
    <NTabPane
      name="tab-req-headers"
      tab='Headers'
      style="display: flex; flex-direction: column; flex: 1;"
    >
      <NDynamicInput
        :value='request.headers'
        @update:value='value => request.headers=value.filter(({key, value}) => key!=="" || value!=="").concat([{key: "", value: ""}])'
        preset="pair"
        key-placeholder="Header"
        value-placeholder="Value"
      />
      <!-- <div
        style="display: flex; flex-direction: row;"
        v-for="(obj, i) in request.headers"
        :key="i"
      >
        <NInput type="text" :value="obj.key" style="flex: 1;" />
        <NInput type="text" :value="obj.value" style="flex: 1;" />
      </div>
      <div
        style="display: flex; flex-direction: row;"
      >
        <NInput type="text" ref="key" :value="pending.key" style="flex: 1;" />
        <NInput type="text" ref="value" :value="pending.value" style="flex: 1;" />
      </div> -->
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
      <NTabPane name="tab-resp-code" disabled><template #tab>
        <NTag
          :type='response.code < 300 ? "success" : response.code < 500 ? "warning" : "error"'
          size="small"
          round
        >{{response.code ?? "N/A"}}</Ntag>
      </template></NTabPane>
      <NTabPane name="tab-resp-body" tab="Body" style="overflow-y: auto;">
        <NCode
          :hljs='hljs'
          :code='response.body'
          :language='responseBodyLanguage(response.headers.find(h => h.key === "Content-Type")?.value ?? "")'
          word-wrap
        />
      </NTabPane>
      <NTabPane name="tab-resp-headers" tab="Headers" style="flex: 1;">
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
</style>
