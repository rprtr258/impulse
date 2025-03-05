<script setup lang="ts">
import {onMounted, useTemplateRef, watch} from "vue";
import {ChangeSpec, EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {PostgreSQL, sql} from "@codemirror/lang-sql";
import {defaultEditorExtensions, defaultExtensions} from "./editor";

const {value = null} = defineProps<{
  value: string,
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
    changes: {
      from: 0,
      to: editor.state.doc.length,
      insert: newValue ?? ""
    }
  });
}, { immediate: true });
</script>

<template>
  <div ref="editorRef" class="editor-container"></div>
</template>

<style scoped>
.editor-container {
  height: 100%;
  width: 100%;
}
</style>

<style>
.cm-editor {
  height: 100%;
}
.cm-scroller {
  overflow: auto;
}
</style>
