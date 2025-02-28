<script setup lang="ts">
import {computed, h, ref, watch} from "vue";
import {NButton, NDataTable, NEmpty, NIcon, NInput, NInputGroup, NLayout, NLayoutContent, NLayoutHeader, NScrollbar, NSelect, NSplit, NTooltip} from "naive-ui";
import {TableBaseColumn} from "naive-ui/es/data-table/src/interface";
import {CheckSquareOutlined, ClockCircleOutlined, FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined} from "@vicons/antd"
import {database} from "wailsjs/go/models";
import {Database} from "./api";
import EditorSQL from "./EditorSQL.vue";
import {useStore} from "./store";

const store = useStore();

const {id, request, response} = defineProps<{
  id: string,
  request: database.SQLRequest,
  response: database.SQLResponse | null,
}>();
const emit = defineEmits<{
  update: [request: database.SQLRequest],
}>();
function updateRequest(patch: Partial<database.SQLRequest>) {
  emit("update", {...request, ...patch});
}

const dsn = ref(request.dsn);
function onInputChange(newValue: string) {
  dsn.value = newValue;
  updateRequest({dsn: newValue});
}

const query = ref(request.query);
function onQueryChange(newValue: string) {
  query.value = newValue;
  updateRequest({query: newValue});
}

const buttonDisabled = ref(false);
function onButtonClick() {
  store.send(id).then(() => {
    buttonDisabled.value = false;
  });
  buttonDisabled.value = true;
}

watch(() => store.tabs, () => {
  dsn.value = request.dsn;
  query.value = request.query;
  buttonDisabled.value = false;
});

const columns = computed(() => {
  if (response === null) {
    return [];
  }

  return (response.columns ?? []).map(c => {
    return {
      key: c,
      title: (_: TableBaseColumn) => {
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
  if (response === null) {
    return [];
  }

  return (response.rows ?? [])
    .map(row =>
      Object.fromEntries(row
        .map((v, i) => [response.columns[i], v])));
});
</script>

<template>
<NLayout
  class="h100"
  id="gavno"
>
  <NLayoutHeader>
    <NInputGroup>
      <NSelect
        :options="Object.keys(Database).map(db => ({label: Database[db as keyof typeof Database], value: db}))"
        :value="request.database"
        v-on:update:value="(database: Database) => updateRequest({database: database})"
        style="width: 10%;"
      />
      <NInput
        placeholder="DSN"
        :value="dsn"
        v-on:input="onInputChange"
      />
      <NButton
        type="primary"
        v-on:click='onButtonClick'
        :disabled="buttonDisabled"
      >Run</NButton>
    </NInputGroup>
  </NLayoutHeader>
  <NLayoutContent class="h100">
    <NSplit class="h100" direction="vertical">
      <template #1>
        <EditorSQL
          :value="query"
          v-on:update="onQueryChange"
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
