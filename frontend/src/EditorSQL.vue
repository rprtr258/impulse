<script setup lang="ts">
import {onMounted, useTemplateRef, watch} from "vue";
import {ChangeSpec, EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {PostgreSQL, sql} from "@codemirror/lang-sql";
import {defaultEditorExtensions, defaultExtensions} from "./editor";

const {value} = defineProps<{
  value: string,
}>();
const emit = defineEmits<{
  update: [value: string],
}>();

const editorRef = useTemplateRef("editorRef");

let editor: EditorView | null = null;
onMounted(() => {
  const state = EditorState.create({
    doc: value ?? "",
    extensions: [
      ...defaultExtensions,
      ...defaultEditorExtensions((doc: string) => emit("update", doc)),
      sql({
        dialect: PostgreSQL,
      }),
    ],
  });

  editor = new EditorView({
    parent: editorRef.value as Element,
    state: state,
  });
});
watch(() => value, (newValue) => {
  if (!editor || newValue === editor.state.doc.toString()) return;

  editor.dispatch({
    changes: {from: 0, to: editor.state.doc.length, insert: newValue}
  });
}, {immediate: true});
</script>

<template>
  <div ref="editorRef"></div>
</template>
