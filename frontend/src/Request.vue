<script setup lang="ts">
import {} from 'vue';
import {NTag} from 'naive-ui';
import {Method as Methods} from './api';

const {collapsible = false, collapsed = false, method, id} = defineProps<{
  id: string,
  collapsible?: boolean,
  collapsed?: boolean,
  method: string,
}>();
const emit = defineEmits<{
  (e: 'click'): void
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
    <NTag
      :type='Methods[method] ? "success" : "info"'
      class='method'
      style="width: 4em; justify-content: center;"
      size="small"
    >{{method}}</Ntag>
    <span style="padding-left: .5em;">{{id}}</span>
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