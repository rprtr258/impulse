<script setup lang="ts">
import {computed, h, onMounted, ref} from "vue";
import {NButton, NDataTable, NEmpty, NIcon, NInput, NInputGroup, NLayout, NLayoutContent, NLayoutHeader, NSelect, NSplit, NTooltip} from "naive-ui";
import {CheckSquareOutlined, ClockCircleOutlined, FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined} from "@vicons/antd"
import * as monaco from "monaco-editor";
import {Database, RequestSQL, ResponseSQL} from "./api";
import {watch} from "vue";

const {response} = defineProps<{
  response: ResponseSQL | null,
}>();

const emit = defineEmits<{
  (e: "send"): void,
}>();

let request = defineModel<RequestSQL>("request");

const codeRef = ref(null);
let editor = null as monaco.editor.IStandaloneCodeEditor | null;
onMounted(() => {
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
watch(request, () => {
  editor?.setValue(request.value?.query ?? "");
});

const columns = computed(() => (response?.columns ?? []).map(c => {
  return {
    key: c,
    title: _ => {
      const type = response.types[response.columns.indexOf(c)];
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
const data = computed(() => (response?.rows ?? []).map(row =>
  Object.fromEntries(row.map((v, i) => [response.columns[i], v]))
));
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
      <NInput v-model:value="request.dsn"/>
      <NButton type="primary" v-on:click='emit("send")'>Run</NButton>
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
.monaco-editor, .n-tab-pane {
  height: 100% !important;
}
</style>
