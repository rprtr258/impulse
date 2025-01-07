<script setup lang="ts">
import {computed, ref, useTemplateRef, watch} from "vue";
import {NTag, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NDynamicInput, NEmpty} from "naive-ui";
import * as monaco from "monaco-editor";
import {api, Method as Methods, RequestJQ, ResponseJQ, Result} from "./api";
import {formatResponse} from "./utils";

const {response} = defineProps<{
  response: ResponseJQ | null,
}>();

const emit = defineEmits<{
  (e: "send"): void,
}>();

let request = defineModel<RequestJQ>("request");
const query = ref("");
const jqerror = ref<string | null>(null);

async function transform(body: string, query: string): Promise<Result<string>> {
  if (query === "") {
    return {kind: "ok", value: formatResponse(body)};
  }

  const res = await api.jq(body, query);
  if (res.kind === "err") {
    return {kind: "err", value: res.value};
  }
  return {kind: "ok", value: res.value.map(v => formatResponse(v)).join("\n")};
};

const responseRef = useTemplateRef("responseRef");
let responseEditor = null as monaco.editor.IStandaloneCodeEditor | null;
watch([
  () => response,
  responseRef,
], () => {
  const dermo = () => monaco.editor.create(responseRef.value, {
    language: "json",
    theme: "material-ocean",
    readOnly: true,
    folding: true,
    minimap: {enabled: false},
    wordWrap: "on",
    lineNumbers: "off",
  });
  // NOTE: cant init in onMounted since response is optional
  if (responseRef.value !== null && responseEditor === null) {
    responseEditor = dermo();
  } else if (responseEditor !== null) {
    responseEditor.dispose();
    // NOTE: эту хуйню нужно пересоздавать после получения нового ответа
    responseEditor = responseRef.value === null ? null : dermo();
  }
}, {deep: true, immediate: true});
const responseText = computed(() => response.response.join("\n"));
watch([
  () => response,
  query,
], () => {
  const body = responseText.value ?? "";
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
</script>

<template>
<div class="h100" style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;">
  <NInputGroup style="grid-column: span 2;">
    <NInput v-model:value="request.query"/>
    <!-- TODO: autosend -->
    <NButton type="primary" v-on:click='emit("send")'>Send</NButton>
  </NInputGroup>
  <NInput
    type="textarea"
    class="h100"
    :value='request.json.join("\n")'
    @update:value='value => request.json = value.split("\n")'
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
    <div
      ref="responseRef"
      style="height: 100%;"
    ></div>
  </template>
</div>
</template>

<style lang="css" scoped>
.monaco-editor, .n-tab-pane {
  height: 100% !important;
}
</style>
