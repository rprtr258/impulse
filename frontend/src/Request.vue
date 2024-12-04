<script setup lang="ts">
import {NButton, NIcon, NSpace, NTag} from "naive-ui";
import {DeleteOutlined, EditOutlined} from "@vicons/antd"
import {Method as Methods} from "./api";

const {collapsible = false, collapsed = false, method, id} = defineProps<{
  id: string,
  collapsible?: boolean,
  collapsed?: boolean,
  method: string,
}>();
const emit = defineEmits<{
  (e: 'click'): void,
  (e: 'rename'): void,
  (e: 'delete'): void,
}>();
</script>

<template class='item nested'>
  <div class='collection-item' v-on:click='emit("click")'>
    <template v-if="collapsible">
        <template v-if="collapsed">
          <svg class="icon" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 16 16"><path fill="currentColor" fillRule="evenodd" d="M4.646 1.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1 0 .708l-6 6a.5.5 0 0 1-.708-.708L10.293 8L4.646 2.354a.5.5 0 0 1 0-.708"></path></svg>
        </template><template v-else>
          <svg class="icon" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 16 16"><path fill="currentColor" fillRule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708"></path></svg>
        </template>
    </template>
    <span style="display: grid; grid-template-columns: 20% auto 30%; width: 100%;">
      <NTag
        :type='Methods[method] ? "success" : "info"'
        class='method'
        style="width: 4em; justify-content: center;"
        size="small"
      >{{method}}</Ntag>
      <span style="padding-left: .5em;">{{id}}</span>
      <span style="display: grid; grid-template-columns: 1fr 1fr; grid-column-gap: .5em;">
        <NButton size="small" @click='emit("rename")'><NIcon :component="EditOutlined"></NIcon></NButton>
        <NButton size="small" @click='emit("delete")'><NIcon color="red" :component="DeleteOutlined"></NIcon></NButton>
      </span>
    </span>
  </div>
</template>

<style lang="css" scoped>
.collection-item {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 5px;
}
</style>
