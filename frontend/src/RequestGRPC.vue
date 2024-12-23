<script setup lang="ts">
import {onMounted, ref, useTemplateRef, watch} from "vue";
import {NTag, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NDynamicInput, NEmpty} from "naive-ui";
import * as monaco from "monaco-editor";
import {api, RequestGRPC, ResponseGRPC, Result, GRPCCodes} from "./api";

const {response} = defineProps<{
  response: ResponseGRPC | null,
}>();

const emit = defineEmits<{
  (e: "send"): void,
}>();

let request = defineModel<RequestGRPC>("request");
const query = ref("");
const jqerror = ref<string | null>(null);

const methods = ref<{
  service: string,
  methods: string[],
}[]>([]);
const loadingMethods = ref(false);

watch(() => request.value.target, () => {
  loadingMethods.value = true;
  api.grpcMethods(request.value.target)
    .then(v => {
      methods.value = v;
      loadingMethods.value = false;
    })
    .catch(err => {
      loadingMethods.value = false;
    });
}, {immediate: true});

async function transform(body: string, query: string): Promise<Result<string>> {
  if (query === "") {
    return {kind: "ok", value: body};
  }

  const res = await api.jq(body, query);
  if (res.kind === "err") {
    return {kind: "err", value: res.value};
  }
  return {kind: "ok", value: res.value.map(v => JSON.stringify(JSON.parse(v), null, 2)).join("\n")};
};

const requestRef = useTemplateRef("requestRef");
let editorReq = null as monaco.editor.IStandaloneCodeEditor | null;
onMounted(() => {
  editorReq = monaco.editor.create(requestRef.value, {
    value: request.value.payload,
    language: "json",
    theme: "material-ocean",
    readOnly: false,
    folding: true,
    minimap: {enabled: false},
    wordWrap: "on",
    lineNumbers: "off",
  });
  editorReq.onDidChangeModelContent(() => {
    request.value.payload = editorReq.getValue();
  });
});

const responseRef = useTemplateRef("responseRef");
let responseEditor = null as monaco.editor.IStandaloneCodeEditor | null;
watch([
  () => response,
  responseRef,
], () => {
  // NOTE: cant init in onMounted since response is optional
  if (responseRef.value !== null && responseEditor === null) {
    responseEditor = monaco.editor.create(responseRef.value, {
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
}, {deep: true, immediate: true});
watch([
  () => response,
  query,
], () => {
  const body = response?.response ?? "";
  transform(body, query.value)
    .then(v => {
      switch (v.kind) {
      case "ok":
        if (responseEditor !== null) {
          responseEditor.setValue(v.value);
        }
        jqerror.value = null;
        break;
      case "err":
        jqerror.value = v.value;
        break;
      }
    });
}, {immediate: true});
watch(() => response, () => {
  query.value = "";
  jqerror.value = null;
});

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
      v-model:value="request.method"
      :options='methods.map(svc => ({
        type: "group",
        label: svc.service,
        key: svc.service,
        children: svc.methods.map(method => ({
          label: method,
          value: svc.service + "." + method,
        })),
      }))'
      :loading="loadingMethods"
      remote
      style="width: 10%; min-width: 18em;"
    />
    <NInput v-model:value="request.target"/>
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
      <div
        id="code"
        ref="requestRef"
        style="height: 100%;"
      ></div>
    </NTabPane>
    <NTabPane
      name="tab-req-headers"
      tab='Metadata'
      style="display: flex; flex-direction: column; flex: 1;"
      display-directive="show"
    >
      <NDynamicInput
        :value='request.metadata'
        @update:value='value => request.metadata=value.filter(({key, value}) => key!=="" || value!=="").concat([{key: "", value: ""}])'
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
      <NTabPane
        name="tab-resp-code"
        disabled
        display-directive="show"
      ><template #tab>
        <NTag
          :type='response.code === 0 ? "success" : "error"'
          size="small"
          round
        >{{response.code /*TODO: as string*/ ?? "N/A"}} {{GRPCCodes[response.code]}}</Ntag>
      </template></NTabPane>
      <NTabPane
        name="tab-resp-body"
        tab="Body"
        style="overflow-y: auto;"
        display-directive="show"
      >
        <div style="display: grid; grid-template-rows: auto 3em; height: 100%;">
          <div
            id="code"
            ref="responseRef"
            style="height: 100%;"
          ></div>
          <div v-if="jqerror !== null" style="position: fixed; color: red; bottom: 3em;">{{jqerror}}</div>
          <NInput
            :status='jqerror !== null ? "error" : "success"'
            v-model:value="query"
            style="height: 100%;"
          />
        </div>
      </NTabPane>
      <NTabPane
        name="tab-resp-headers"
        tab="Metadata"
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
          <tr v-for="header in response.metadata" :key="header.key">
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
