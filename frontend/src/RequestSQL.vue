<script setup lang="ts">
import {computed, toRefs, h, ref} from "vue";
import {
  NButton, NInput, NInputGroup, NSelect,
  NLayout, NLayoutContent, NLayoutHeader,
  NScrollbar, NSplit,
  NDataTable, NEmpty, NIcon, NTooltip,
} from "naive-ui";
import {TableBaseColumn} from "naive-ui/es/data-table/src/interface";
import {
  CheckSquareOutlined, ClockCircleOutlined,
  FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined,
} from "@vicons/antd"
import {database} from "wailsjs/go/models";
import {Database} from "./api";
import EditorSQL from "./EditorSQL.vue";
import {use_request} from "./store";

type Request = {kind: database.Kind.SQL} & database.SQLRequest;

const {id} = defineProps<{
  id: string,
}>();

const {request, response, is_loading, update_request, send} = toRefs(use_request<Request, database.SQLResponse>(ref(id)));

const columns = computed(() => {
  const resp = response.value;
  if (resp === null) {
    return [];
  }

  return (resp.columns ?? []).map(c => {
    return {
      key: c,
      title: (_: TableBaseColumn) => {
        const type = resp.types[resp.columns.indexOf(c)];
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
      render: (rowData: any, rowIndex: number) => {
        const v = rowData[c];
        switch (true) {
        case v === null:
          return "(NULL)"; // TODO: faded
        case typeof v == "boolean":
          return v ? "true" : "false";
        case typeof v == "number" || typeof v == "string":
          return v;
        default:
          console.log(rowData, rowData[c], rowIndex);
          return rowData[c];
        }
      },
    }
  });
});
// TODO: fix duplicate column names
const data = computed(() => {
  const resp = response.value;
  if (resp === null) {
    return [];
  }

  return (resp.rows ?? [])
    .map(row =>
      Object.fromEntries(row
        .map((v, i) => [resp.columns[i], v])));
});
</script>

<template>
<NEmpty
  v-if="request === null"
  description="Loading request..."
  class="h100"
  style="justify-content: center;"
/>
<NLayout
  v-else
  class="h100"
  id="gavno"
>
  <NLayoutHeader>
    <NInputGroup>
      <NSelect
        :options="Object.keys(Database).map(db => ({label: Database[db as keyof typeof Database], value: db}))"
        :value="request.database"
        v-on:update:value="(database: Database) => update_request({database: database})"
        style="width: 10%;"
      />
      <NInput
        placeholder="DSN"
        :value="request.dsn"
        v-on:input="newValue => update_request({dsn: newValue})"
      />
      <NButton
        type="primary"
        v-on:click="send()"
        :disabled="is_loading"
      >Run</NButton>
    </NInputGroup>
  </NLayoutHeader>
  <NLayoutContent class="h100">
    <NSplit class="h100" direction="vertical">
      <template #1>
        <EditorSQL
          :value="request.query"
          v-on:update="newValue => update_request({query: newValue})"
          class="h100"
        />
      </template>
      <template #2>
        <template v-if="response === null">
          <NEmpty
            description="Run query or choose one from history."
            class="h100"
            style="justify-content: center;"
          />
        </template><template v-else>
          <NScrollbar>
            <NDataTable
              :columns="columns"
              :data="data"
              :single-line="false"
              size="small"
              resizable
              :scroll-x="response.columns.length * 200"
            />
          </NScrollbar>
        </template>
      </template>
    </NSplit>
  </NLayoutContent>
</NLayout>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
<style lang="css">
/* TODO: как же я ненавижу ебаный цсс блять господи за что */
#gavno > .n-layout-scroll-container {
  overflow: hidden;
  height: 100%;
  display: grid;
  grid-template-rows: 34px 1fr;
}
</style>
