<script setup lang="ts">
import {onMounted, useTemplateRef, watch} from "vue";
import {ChangeSpec, EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {json} from "@codemirror/lang-json";
import {defaultEditorExtensions, defaultExtensions} from "./editor";

const {value = null} = defineProps<{
  value: string | null,
}>();
const emit = defineEmits<{
  update: [value: string],
}>();

const editorRef = useTemplateRef("editorRef");

let editor: EditorView;
onMounted(() => {
  const state = EditorState.create({
    doc: value ?? "",
    extensions: [
      ...defaultExtensions,
      ...defaultEditorExtensions((doc: string) => emit("update", doc)),
      json(),
    ],
  });

  editor = new EditorView({
    parent: editorRef.value as Element,
    state: state as EditorState,
  });
});
watch(() => value, () => {
  if (value === editor.state.doc.toString()) return;

  editor.dispatch({
    changes: {from: 0, to: editor.state.doc.length, insert: value} as ChangeSpec,
  });
});
</script>

<template>
  <div ref="editorRef"></div>
</template>
