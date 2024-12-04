<script setup lang="ts">
import {h, onMounted, ref, watch} from "vue";
import {NButton, NDataTable, NEmpty, NIcon, NInput, NInputGroup, NLayout, NLayoutContent, NLayoutHeader, NSelect, NSplit, NTooltip} from "naive-ui";
import {CheckSquareOutlined, ClockCircleOutlined, FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined} from "@vicons/antd"
import * as monaco from "monaco-editor";
import {api, Database, RequestSQL, ResponseSQL} from "./api";

const {id} = defineProps<{
  id: string,
}>();

let request = defineModel<RequestSQL>("request");
let response = defineModel<ResponseSQL>("response");

const codeRef = ref(null);
let editor = null as monaco.editor.IStandaloneCodeEditor | null;
onMounted(() => {
  monaco.editor.defineTheme("material-ocean", {
    base: "vs-dark",
    inherit: true,
    rules: [
      {token: "number",            foreground: "F78C6C"},
      {token: "string.key.json",   foreground: "49C1AE"},
      {token: "string.value.json", foreground: "BEE28A"},
      {token: "keyword.json",      foreground: "C792EA"},
    ],
    colors: {
      'editor.background': '#0F111A',
      'editor.foreground': '#8F93A2',
      'editor.lineHighlightBackground': '#00000050',
      'editor.selectionBackground': '#717CB450',
    },
  });
  editor = monaco.editor.create(codeRef.value, {
    value: request.value?.query,
    language: "sql",
    theme: "material-ocean",
    folding: true,
    minimap: {enabled: false},
    wordWrap: "on",
    lineNumbers: "off",
  });
  editor.onDidChangeModelContent(() => {
    request.value.query = editor.getValue();
  });
});


function send() {
  api
    .requestPerform(id)
    .then(v => {
      if (v.kind === "sql") {
        response.value = {
          columns: v.columns,
          types: v.types,
          rows: v.rows,
        }
      } else {
        // TODO: fail
        throw new Error("Unexpected response kind: " + v.kind);
      }
    })
    .catch((err) => alert(`Could not perform request: ${err}`));
}
const columns = ref(response.value.columns.map(c => {
  return {
    key: c,
    title: _ => {
      const type = response.value.types[response.value.columns.indexOf(c)];
      return h(NTooltip, {trigger: "hover", placement: "bottom-start"}, {
        trigger: () => h("div", {}, [
          h(NIcon, {size: "15", color: "grey"}, () => [
            h(
              type === "number" ? FieldNumberOutlined :
              type === "string" ? ItalicOutlined :
              type === "bool" ? CheckSquareOutlined :
              type === "time" ? ClockCircleOutlined :
              QuestionCircleOutlined,
            )
          ]),
          c,
        ]),
        default: () => type,
      });
    },
  }
}));
const data = ref(response.value.rows.map(row => Object.fromEntries(row.map((v, i) => [response.value.columns[i], v]))));
</script>

<template>
<NLayout class="h100">
  <NLayoutHeader>
    <NInputGroup>
      <NSelect
        :options="Object.keys(Database).map(db => ({label: Database[db], value: db}))"
        v-model:value="request.database"
        style="width: 10%;"
      />
      <NInput :value="request.dsn"/>
      <NButton type="primary" v-on:click="send()">Run</NButton>
    </NInputGroup>
  </NLayoutHeader>
  <NLayoutContent style="height: 90%;">
    <NSplit class="h100">
      <template #1>
        <div
          id="code"
          ref="codeRef"
          class="h100"
        ></div>
      </template>
      <template #2>
        <template v-if="response === null">
          <NEmpty
            description="Run query or choose one from history."
            class="h100"
            style="justify-content: center;"
          />
        </template>
        <template v-else>
          <NDataTable
            :columns="columns"
            :data="data"
            :single-line="false"
            resizable
          />
        </template>
      </template>
    </NSplit>
  </NLayoutContent>
</NLayout>
</template>

<style lang="css" scoped>
</style>
