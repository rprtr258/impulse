<script setup lang="ts">
import {h, ref} from "vue";
import {NButton, NDataTable, NEmpty, NIcon, NInput, NInputGroup, NLayout, NLayoutContent, NLayoutHeader, NSelect, NSplit, NTooltip} from "naive-ui";
import {CheckSquareOutlined, ClockCircleOutlined, FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined} from "@vicons/antd"
import {api, Database, RequestSQL, ResponseSQL} from "./api";

const {id, collectionID} = defineProps<{
  id: string,
  collectionID: string,
}>();

let request = defineModel<RequestSQL>("request");
let response = defineModel<ResponseSQL>("response");

function send() {
  api
    .requestPerform(collectionID, id)
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
        <NInput
          type="textarea"
          class="h100"
          v-model:value='request.query'
          @update:value="value => request.query = value"
        />
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
