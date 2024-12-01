<script setup lang="ts">
import {onMounted, ref, useTemplateRef, watch} from "vue";
import {NInput, NTooltip} from "naive-ui";
import {EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {json} from "@codemirror/lang-json";
import {transform} from "./utils";
import {defaultExtensions} from "./editor";

const {value = null} = defineProps<{
  value?: string | undefined | null, // TODO: optional govno is ignored in typing
}>();

const query = ref("");
const jqerror = ref<string | null>(null);

const editorRef = useTemplateRef("editorRef");

let editor: EditorView;
onMounted(() => {
  const state = EditorState.create({
    doc: value ?? "",
    extensions: [
      ...defaultExtensions,
      EditorState.readOnly.of(true),
      json(),
    ],
  });

  editor = new EditorView({
    parent: editorRef.value as Element,
    state: state as EditorState,
  });
});

function localTransform() {
  const body = value ?? "";
  transform(body, query.value) // TODO: can't transform if body is stream of jsons
    .then(v => {
      switch (v.kind) {
      case "ok":
        editor.dispatch({
          changes: {from: 0, to: editor.state.doc.length, insert: v.value},
        });
        jqerror.value = null;
        break;
      case "err":
        jqerror.value = v.value;
        break;
      }
    });
}
watch([
  () => value,
  () => query,
], localTransform, {immediate: true});

watch(() => value, () => {
  query.value = "";
  jqerror.value = null;
});
</script>

<template>
  <div style="height: 100%; display: grid; grid-template-rows: 2rem auto;">
    <!-- TODO: put input under editor -->
    <NTooltip :show="jqerror !== null" placement="bottom">
      <template #trigger>
        <NInput
          placeholder="JQ query"
          :status='jqerror !== null ? "error" : "success"'
          :value="query"
          v-on:update:value="value => {query = value; localTransform();}"
        />
      </template>
      {{jqerror}}
    </NTooltip>
    <div ref="editorRef"></div>
  </div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
<style lang="css">
.cm-editor { height: 100% }
.cm-scroller { overflow: auto }
</style>
