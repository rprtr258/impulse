<script setup lang="ts">
import {ref, watch} from "vue";
import {NTag, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NDynamicInput, NEmpty, darkTheme} from "naive-ui";
import * as monaco from "monaco-editor";
import {Method as Methods, RequestHTTP, ResponseHTTP} from "./api";

const {response} = defineProps<{
  response: ResponseHTTP | null,
}>();

const emit = defineEmits<{
  (e: "send"): void,
}>();

let request = defineModel<RequestHTTP>("request");

const codeRef = ref(null);
let editor = null as monaco.editor.IStandaloneCodeEditor | null;
watch([() => response, codeRef], () => {
  if (codeRef.value !== null && editor === null) {
    editor = monaco.editor.create(codeRef.value, {
      value: "",
      language: "json",
      theme: "material-ocean",
      readOnly: true,
      folding: true,
      minimap: {enabled: false},
      wordWrap: "on",
      lineNumbers: "off",
    });
  }

  editor?.setValue(response?.body ?? "");
}, {deep: true, immediate: true});

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
        <div
          id="code"
          ref="codeRef"
          style="height: 100%;"
        ></div>
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
.monaco-editor, .n-tab-pane {
  height: 100% !important;
}
</style>
