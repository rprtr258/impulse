<script setup lang="ts">
import {computed, h, onMounted, ref, VNodeChild, watch} from "vue";
import {useBrowserLocation, useLocalStorage} from "@vueuse/core";
import {
  NTag, NTabs, NTabPane,
  NList, NListItem,
  NResult, NSelect, NIcon,
  NSpace, NScrollbar,
  NInput, NModal, NDropdown, NButton,
  NTree, TreeOption,
  useNotification,
} from "naive-ui";
import {DeleteOutlined, DoubleLeftOutlined, DoubleRightOutlined, EditOutlined, DownOutlined} from "@vicons/antd";
import {ContentCopyFilled} from "@vicons/material";
import {CopySharp} from "@vicons/ionicons5";
import {store, setNotify} from "./store";
import {
  Method as Methods, RequestData, Kinds, Database, Tree,
  ResponseHTTP, RequestHTTP as RequestHTTPT,
  ResponseSQL,  RequestSQL  as RequestSQLT,
  ResponseGRPC, RequestGRPC as RequestGRPCT,
  ResponseJQ,   RequestJQ   as RequestJQT,
} from "./api";
import RequestHTTP from "./RequestHTTP.vue";
import RequestSQL from "./RequestSQL.vue";
import RequestGRPC from "./RequestGRPC.vue";
import RequestJQ from "./RequestJQ.vue";

onMounted(() => {
  const notification = useNotification();
  setNotify((...args) => notification.error({title: "Error", content: args.map(arg => arg.toString()).join("\n")}));
});

function dirname(id: string): string {
  return id.split("/").slice(0, -1).join("/");
}
function basename(id: string): string {
  return id.split("/").pop();
}
const treeData = computed(() => {
  const mapper = (tree: Tree) =>
    Object.entries(tree.dirs ?? {}).map(([k, v]) => ({
      key: k,
      label: basename(k),
      children: mapper(v),
    })).concat((tree.ids ?? []).map(id => {
      return {
        key: id,
        label: basename(id),
      };
    }));
  return mapper(store.requestsTree);
});
function renameRequest(id: string, newID: string) {
  const selectedID = store.requestID;
  store.rename(id, newID);
  if (selectedID !== null && id === selectedID) {
    selectRequest(newID);
  }
}
function drag({node, dragNode, dropPosition}: {
  node: TreeOption,
  dragNode: TreeOption,
  dropPosition: "before" | "inside" | "after",
}): void	{
  const dir = (d: string): string => d === "" ? "" : d + "/";
  const oldID = dragNode.key as string;
  const into = node.key as string;
  switch (dropPosition) {
    case "before":
    case "after":
      renameRequest(oldID, dir(dirname(into)) + basename(oldID));
      break;
    case "inside":
      renameRequest(oldID, dir(into) + basename(oldID));
      break;
  }
}

const location = useBrowserLocation();
function selectRequest(id: string) {
  store.selectRequest(id);
  location.value.hash = id;
}
onMounted(() => {
  store.fetch().then(() => {
    if (location.value.hash !== "#") {
      const id = decodeURI(location.value.hash.slice(1)); // remove '#'
      if (store.requests[id] === undefined) {
        location.value.hash = "";
        return;
      }

      selectRequest(id);
    }
  });
});

function fromNow(date: Date): string {
  const now = new Date();

  const milliseconds = now.getTime() - date.getTime();
  const seconds = Math.floor(milliseconds / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  const weeks = Math.floor(days / 7);
  const months = Math.floor(weeks / 4);
  const years = Math.floor(months / 12);

  if (years > 0) {
    return years === 1 ? "a year ago" : years + " years ago";
  } else if (months > 0) {
    return months === 1 ? "a month ago" : months + " months ago";
  } else if (weeks > 0) {
    return weeks === 1 ? "a week ago" : weeks + " weeks ago";
  } else if (days > 0) {
    return days === 1 ? "yesterday" : days + " days ago";
  } else if (hours > 0) {
    return hours === 1 ? "an hour ago" : hours + " hours ago";
  } else if (minutes > 0) {
    return minutes === 1 ? "a minute ago" : minutes + " minutes ago";
  } else {
    return "just now";
  }
}

const treeNodeProps = ({ option }: { option: TreeOption }) => {
  return {
    onClick() {
      if (!option.children && !option.disabled) {
        const id = option.key as string;
        selectRequest(id);
      }
    }
  }
}

const expandedKeys = useLocalStorage<string[]>("expanded-keys", []);

const newRequestKind = ref<typeof Kinds[number] | null>(null);
watch(newRequestKind, async () => {
  if (newRequestKind.value === null) {
    return;
  }

  const kind = newRequestKind.value;
  newRequestKind.value = null;
  const id = new Date().toUTCString();
  store.createRequest(id, kind);
  // TODO: add name
});

// TODO: fix editing request headers

const renameID = ref(null as string | null);
const renameValue = ref(null as string | null);
function renameCancel() {
  renameID.value = null;
  renameValue.value = null;
}
function rename() {
  renameRequest(renameID.value, renameValue.value);
  renameCancel();
}

function badge(req: RequestData): [string, string] {
  switch (req.kind) {
  case "http": return [Methods[req.method], "lime"];
  case "sql": return [Database[req.database], "bluewhite"];
  case "grpc": return ["GRPC", "cyan"];
  case "jq": return ["JQ", "violet"];
  }
}
function renderPrefix(info: {option: TreeOption, checked: boolean, selected: boolean}): VNodeChild {
  const option = info.option;
  const req = store.requests[option.key];
  if (req === undefined) {
    return null;
  }
  const [method, color] = badge(req);
  return h(NTag, {
    type: Methods[method] ? "success" : "info",
    class: 'method',
    style: `width: 4em; justify-content: center; color: ${color};`,
    size: "small",
  }, () => method)
}
function renderSuffix(info: {option: TreeOption}): VNodeChild {
  const option = info.option;
  const id = option.key as string;
  return h(NDropdown, {
    trigger: "hover",
    options: [
      {
        label: "Rename",
        key: "rename",
        icon: () => h(NIcon, {component: EditOutlined}),
        props: {
          onClick: () => {
            renameID.value = id;
            renameValue.value = id;
          }
        }
      },
      {
        label: "Duplicate",
        key: "duplicate",
        icon: () => h(NIcon, {component: CopySharp}),
        props: {
          onClick: () => {
            store.duplicate(id);
          }
        }
      },
      {
        label: "Copy as curl",
        key: "copy-as-curl",
        icon: () => h(NIcon, {component: ContentCopyFilled}),
        show: store.requests[id]?.kind === "http",
        props: {
          onClick: () => {
            const req = store.requests[id];
            if (!req) {
              return;
            }
            const httpToCurl = ({url, method, body, headers}: RequestHTTPT) => {
              return `curl -X ${method} ${url}` +
                (headers.length > 0 ? " " + headers.map(({key, value}) => `-H "${key}: ${value}"`).join(" ") : "") +
                ((body) ? ` -d '${body}'` : "");
            };
            console.log(req);
            navigator.clipboard.writeText(httpToCurl(req as RequestHTTPT));
          }
        }
      },
      {
        label: "Delete",
        key: "delete",
        icon: () => h(NIcon, {color: "red", component: DeleteOutlined}),
        props: {
          onClick: () => {
            store.deleteRequest(id);
          },
        },
      },
    ],
    "v-on:select": (key: string | number) => {
      // message.info(String(key))
    },
  }, () => [h(NIcon, {component: DownOutlined})]);
}

const sidebarHidden = ref(false);
</script>

<template>
<div
  class="h100"
  :style='{
    display: "grid",
    "grid-template-columns": `${sidebarHidden ? "3em" : "300px"} 6fr`,
  }'
>
  <aside
    style="
      overflow: auto;
      color: rgba(255, 255, 255, 0.82);
      background-color: rgb(24, 24, 28);
      display: grid;
      grid-template-rows: auto 3em;
      height: 100%;
    "
  >
    <NTabs
      type="card"
      size="small"
      v-if="!sidebarHidden"
    >
      <NTabPane
        name="tab-nav-collection"
        tab="Collection"
        display-directive="show"
        style="
          overflow: auto;
          display: grid;
          padding-top: 0px;
          grid-template-rows: 2.5em auto;
          height: 100%;
        "
      >
        <NModal
          :show="renameID !== null"
          v-on:update-show="(show) => {if (!show) { renameCancel(); }}"
          preset="dialog"
          title="Rename request"
          positive-text="Rename"
          negative-text="Cancel"
          v-on:positive-click="rename"
          v-on:negative-click="renameCancel"
        >
          <NInput v-model:value="renameValue" />
        </NModal>
        <NSelect
          v-model:value="newRequestKind"
          placeholder="New"
          clearable
          :options="Kinds.map(kind => ({label: kind.toUpperCase(), value: kind}))"
        />
        <NScrollbar trigger="none">
          <NTree
            block-line
            expand-on-click
            :selected-keys='[(store.requestID ?? "")]'
            :show-line="true"
            :data="treeData"
            :draggable="true"
            :default-expanded-keys="expandedKeys as unknown as string[]"
            v-on:update:expanded-keys="(keys: string[]) => expandedKeys = keys"
            v-on:drop="drag"
            :node-props="treeNodeProps"
            :render-prefix="renderPrefix"
            :render-suffix="renderSuffix"
          />
        </NScrollbar>
      </NTabPane>
      <NTabPane
        name="tab-nav-history"
        tab="History"
        style="flex: 1;"
        display-directive="show"
      >
        <NList hoverable :border="false">
          <NListItem
            v-for="r, i in store.history"
            :key="i"
            v-on:click="selectRequest(r.request_id)"
            class="history-card card"
          >
            <div class="headline">
              <NTag
                class="method"
                size="small"
                :style="`width: 4em; justify-content: center; color: ${badge(store.requests[r.request_id])[1]}`"
              >{{badge(store.requests[r.request_id])[0]}}</Ntag>
              <span class='url' style="padding-left: .5em;">{{r.request_id}}</span>
            </div>
            <div class='footer'>
              <span style='color: grey;' class='date'>{{fromNow(r.sent_at)}}</span>
            </div>
          </NListItem>
        </NList>
      </NTabPane>
    </NTabs>
    <div v-else></div>
    <NButton
      class="h100"
      v-on:click="sidebarHidden = !sidebarHidden"
      style="display: grid; grid-template-columns: 1fr 1fr; grid-column-gap: .5em;"
      v-if="!sidebarHidden"
    >
      <NSpace>
        <NIcon :component="DoubleLeftOutlined" />
        hide
      </NSpace>
    </NButton>
    <NButton
      class="h100"
      v-on:click="sidebarHidden = !sidebarHidden"
      style="display: grid; grid-template-columns: 1fr 1fr; grid-column-gap: .5em;"
      v-else
    >
      <NSpace>
        <NIcon :component="DoubleRightOutlined" />
      </NSpace>
    </NButton>
  </aside>
  <div style="color: rgba(255, 255, 255, 0.82); background-color: rgb(16, 16, 20); overflow: hidden;">
    <template v-if="store.request() === null">
      <NResult
        status="info"
        title="Pick request"
        description="Pick request to see it, edit and send and do other fun things."
        class="h100"
        style="align-content: center;"
      />
    </template><template v-else-if='store.request().kind === "http"'>
      <RequestHTTP
        :request="store.request() as RequestHTTPT"
        :response="store.response as ResponseHTTP | null"
        v-on:send="() => store.send(store.requestID)"
        v-on:update="(request) => store.update(store.requestID, request)"
      />
    </template><template v-else-if='store.request().kind === "sql"'>
      <RequestSQL
        :request="store.request() as RequestSQLT"
        :response="store.response as ResponseSQL | null"
        v-on:send="() => store.send(store.requestID)"
        v-on:update="(request) => store.update(store.requestID, request)"
      />
    </template><template v-else-if='store.request().kind === "grpc"'>
      <RequestGRPC
        :request="store.request() as RequestGRPCT"
        :response="store.response as ResponseGRPC | null"
        v-on:send="() => store.send(store.requestID)"
        v-on:update="(request) => store.update(store.requestID, request)"
      />
    </template><template v-else-if='store.request().kind === "jq"'>
      <RequestJQ
        :request="store.request() as RequestJQT"
        :response="store.response as ResponseJQ | null"
        v-on:send="() => store.send(store.requestID)"
        v-on:update="(request) => store.update(store.requestID, request)"
      />
    </template>
  </div>
</div>
</template>

<style scoped>
</style>

<style>
div, div:hover, ul, ul:hover, li, li:hover {
  transition: none !important;
}

html, body, #app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  /* background: rgb(10, 10, 20); */
  /* color: white; */
  height: 100%;
  max-height: 100%;
  margin: 0;
  --margin-bottom: 0;
}
.h100 {
  height: 100%;
}
.tab-body, .input-wrapper, .input {
  height: 100%;
}
.tabs .tab > div.tab-body {
  padding: 0px;
}
.dropdown {
  left: 0px !important;
}
.dropdown-item, .dropdown-body {
  max-width: 8rem !important;
}
.collapsible > .collapsible-item > .collapsible-body > .content {
  padding: 0px !important;
}
.list-group > .list-group-item:hover {
  background: var(--color-dark-tint-50);
}
table {
  table-layout: fixed;
  /* border-spacing: 0;
  width: 100%;
  column-width: 50%;
  text-align: left; */
}
.n-list-item {
  padding: .1em .5rem !important; /* TODO: remove? */
}
div.n-dynamic-input-item__action {
  display: none !important; /* TODO: remove? */
}
textarea {
  height: 100%;
}
/*
.page {
  display: grid;
  padding: 1em 10px 0px 10px;
  grid-template-columns: 2fr 10fr;
  grid-gap: 10px;
}

.page .sidebar {
  scrollbar-width: none;
  padding: 0 0.5em;
}

.history-card {
  margin-bottom: 10px;
  cursor: pointer;
}

.request-bar {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  position: relative;
}

.request-bar .bar {
  width: 100%;
}

.nested {
  margin-left: 10px;
}

#request-response {
  margin-top: 20px;
  padding-bottom: 30px;
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: minmax(100px, 1fr) minmax(100px, 1fr);
  grid-column-gap: 1em;
}

td {
  padding: 3px;
}
tr:nth-child(odd) {
  background: gray;
  color: white;
}

.input {
  background: lightgray;
  border: 1px solid #2c3e50;
  padding: 15px;
  width: calc(100% - 144px);
}

.input:focus {
  outline: 1.5px solid #2c3e50;
}

.icon {
  width: 1.2em;
  vertical-align: middle;
  display: inline-block;
}

.name {
  padding-left: 10px;
}

.method {
  background: #2c3e50;
  color: #ffffff;
  padding: 1px;
}

.badge {
  background: #2c3e50;
  color: #ffffff;
  padding: 3px 6px;
  width: 3em;
  text-align: center;
}

.base-field {
  align-content: center;
  text-align: center;
  height: 100%;
}

.text {
  vertical-align: middle;
  text-align: center;
}

.pretty-select {
  background: #2c3e50;
  cursor: pointer;
}

.option {
  padding: 2px;
  background: lightgray;
  color: black;
}

.option:hover {
  background: #2c3e50;
  color: white;
}

.options-container {
  position: absolute;
  z-index: 1;
  border: 1px solid #2c3e50;
  background: white;
  width: inherit;
  left: 0px;
}

.url {
  display: inline-flex;
  margin-left: 10px;
  flex-direction: row;
  align-items: center;
  justify-content: center;
}

.card {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-content: flex-start;
  border: 1px solid #2c3e50;
  padding: 10px;
}

.card div {
  display: flex;
  flex-direction: row;
  align-content: flex-start;
}

.footer {
  margin-top: 5px;
  font-size: 0.8rem;
}

.card:hover {
  border: 1px solid #ffffff;
  background: #2c3e50;
  color: #ffffff;
}

.card:hover .method {
  background: #2c3e50;
  color: #2c3e50;
}

.editor {
  text-align: left;
  min-width: 100%;
  max-width: 100%;
  height: 100%;
  border: 1px solid #2c3e50;
  background: inherit;
  color: inherit;
}

.h-100 {
  height: 100%;
}

.active {
  font-weight: bold;
  border-bottom: 2px solid #2c3e50;
}

.button {
  width: 10rem;
  font-size: 16pt;
  height: 100%;
  border: 1px solid #2c3e50;
  background: #3f8ddb;
  color: white;
  padding: 15px;
}
.button:hover {
  background: #2c3e50;
  cursor: pointer;
}

.json-viewer {
  border: 1px solid #2c3e50;
  padding: 0 1em;
  height: calc(100% - 2 * 10px);
  text-align: left;
  word-break: break-all;
  font-family: monospace;
}

.parameter-row {
  display: flex;
  align-content: stretch;
  justify-content: stretch;
  margin: -1px;
}
.parameter-row input {
  flex-grow: 1;
  background: #ffffff;
  border: 1px solid #2c3e50;
  padding: 10px;
  width: calc(100% - (2 * 15px) - 1px);
}
.parameter-row input:focus {
  outline: 2px solid #2c3e50;
  border-left: 1px solid #2c3e50;
}
.parameter-row input:nth-child(2) {
  border-left: none;
} */
</style>
