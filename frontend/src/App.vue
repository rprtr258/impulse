<script setup lang="ts">
import {computed, h, onMounted, ref, VNodeChild, watch} from "vue";
import {useBrowserLocation, useLocalStorage} from "@vueuse/core";
import {
  NIcon, NTag, NTabs, NTabPane,
  NList, NListItem,
  NResult, NEmpty, NSpace, NScrollbar,
  NSelect, NInput, NModal, NDropdown, NButton,
  NTree, TreeOption,
  useNotification,
} from "naive-ui";
import {
  DeleteOutlined, DoubleLeftOutlined, DoubleRightOutlined, EditOutlined, DownOutlined,
} from "@vicons/antd";
import {ContentCopyFilled} from "@vicons/material";
import {CopySharp} from "@vicons/ionicons5";
import {use_history, useStore} from "./store";
import {Method, Kinds, Database, api} from "./api";
import {database, app} from "wailsjs/go/models";
import RequestHTTP from "./RequestHTTP.vue";
import RequestSQL from "./RequestSQL.vue";
import RequestGRPC from "./RequestGRPC.vue";
import RequestJQ from "./RequestJQ.vue";
import RequestRedis from "./RequestRedis.vue";

const notification = useNotification();
const store = useStore();

function dirname(id: string): string {
  return id.split("/").slice(0, -1).join("/");
}
function basename(id: string): string {
  return id.split("/").pop() ?? "";
}
const treeData = computed(() => {
  const mapper = (tree: app.Tree): TreeOption[] =>
    Object.entries(tree.Dirs ?? {}).map(([k, v]) => ({
      key: k,
      label: basename(k),
      children: mapper(v),
    } as TreeOption)).concat(tree.IDs.map(id => {
      return {
        key: id,
        label: basename(id),
      } as TreeOption;
    }));
  return mapper(store.requestsTree.value);
});
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
      store.rename(oldID, dir(dirname(into)) + basename(oldID));
      break;
    case "inside":
      store.rename(oldID, dir(into) + basename(oldID));
      break;
  }
}

const location = useBrowserLocation();
function selectRequest(id: string) {
  store.selectRequest(id);
  location.value.hash = id;
}
function handleClose(id: string) {
  const v = store.tabs.value;
  if (v === null) {
    return;
  }
  if (v.map.list.length === 1) {
    store.tabs.value = null;
    return;
  }

  // adjust index
  const idx = v.map.index(id);
  if (idx === null) {
    return;
  }
  if (idx <= v.index) {
    v.index = Math.max(v.index - 1, 0);
  }
  v!.map.remove(id);
  store.tabs.value = {map: v.map, index: v.index};
  selectRequest(v.map.list[v.index]);
}
onMounted(() => {
  store.fetch().then(() => {
    const oldTabs = localStorage.getItem("tabs");
    if (oldTabs !== null) {
      for (const id of JSON.parse(oldTabs)) {
        store.selectRequest(id);
      }
    }

    if (location.value.hash !== "#") {
      const id = decodeURI((location.value.hash ?? "").slice(1)); // remove '#'
      if (store.requests[id] === undefined) {
        location.value.hash = "";
        return;
      }

      selectRequest(id);
    }
  });
});
watch(() => store.tabs.value, () => {
  if (store.tabs.value) {
    localStorage.setItem("tabs", JSON.stringify(store.tabs.value?.map.list ?? []));
  } else {
    localStorage.removeItem("tabs");
  }
}, {deep: true});

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

const renameID = ref<string | null>(null);
const renameValue = ref<string | null>(null);
function renameCancel() {
  renameID.value = null;
  renameValue.value = null;
}
function rename() {
  const fromID = renameID.value;
  if (fromID === null) {
    notification.error({title: "Invalid request", content: `No request to rename: ${renameID.value} -> ${renameValue.value}`});
    return;
  }

  const toID = renameValue.value;
  if (toID === null) {
    notification.error({title: "Invalid request", content: "No new name"});
    return;
  }

  if (toID === fromID) {
    return;
  }

  store.rename(fromID, toID);
  renameCancel();
}

function badge(req: app.requestPreview): [string, string] {
  switch (req.Kind) {
  case "http": return [Method[req.SubKind as keyof typeof Method], "lime"];
  case "sql": return [Database[req.SubKind as keyof typeof Database], "bluewhite"];
  case "grpc": return ["GRPC", "cyan"];
  case "jq": return ["JQ", "violet"];
  case "redis": return ["REDIS", "red"];
  }
  throw "unreachable";
}
function renderPrefix(info: {option: TreeOption, checked: boolean, selected: boolean}): VNodeChild {
  const option = info.option;
  if (option.key === undefined) {
    return null;
  }

  const req = store.requests[option.key];
  if (req === undefined) {
    return null;
  }
  const [method, color] = badge(req);
  return h(NTag, {
    type: req.Kind === "http" ? "success" : "info", // TODO: replace with color
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
        show: store.requests[id]?.Kind === "http",
        props: {
          onClick: () => {
            api.get(id).then((r) => {
              if (r.kind === "err") {
                useNotification().error({title: "Error", content: `Failed to load request: ${r.value}`});
                return;
              }

              const req = r.value as unknown as database.HTTPRequest; // TODO: remove unknown cast
              const httpToCurl = ({url, method, body, headers}: database.HTTPRequest) => {
                const headersStr = headers?.length > 0 ? " " + headers.map(({key, value}) => `-H "${key}: ${value}"`).join(" ") : "";
                const bodyStr = (body) ? ` -d '${body}'` : "";
                return `curl -X ${method} ${url}${headersStr}${bodyStr}`;
              };
              navigator.clipboard.writeText(httpToCurl(req));
            })
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

const history = use_history(() => store.requestID()!);
const requestKind = computed(() => {
  const requestID = store.requestID();
  if (requestID === null) {
    return null;
  }
  return store.requests[requestID].Kind;
});
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
            :selected-keys='[(store.tabs.value ? store.tabs.value.map.list[store.tabs.value.index] : "")]'
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
        <NEmpty
          v-if="store.requestID() === null"
          description="No request picked yet"
          class="h100"
          style="justify-content: center;"
        />
        <NEmpty
          v-else-if="history.value.length === 0"
          description="No history yet"
          class="h100"
          style="justify-content: center;"
        />
        <NList v-else hoverable :border="false">
          <NListItem
            v-for="r, i in history.value"
            :key="i"
            class="history-card card"
          >
            <!-- v-on:click="selectRequest(r.RequestId)" -->
            <span style='color: grey;' class='date'>{{fromNow(r.sent_at)}}</span>
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
    <NResult
      v-if="store.tabs.value.map.length() === 0"
      status="info"
      title="Pick request"
      description="Pick request to see it, edit and send and do other fun things."
      class="h100"
      style="align-content: center;"
    />
    <NTabs
      v-else
      closable
      v-on:close="handleClose"
      class="h100"
      type="card"
      size="small"
      :value="store.tabs.value.map.list[store.tabs.value.index]"
      v-on:update-value="(id) => selectRequest(id)"
    ><NTabPane
      v-for="id in store.tabs.value.map.list"
      :key="id"
      :name="id"
      :tab="id"
      class="h100"
      display-directive="if"
    >
      <RequestHTTP
        v-if='requestKind === "http"'
        :id="id"
        v-on:send="() => store.send(id)"
        v-on:update="(request) => store.update(id, request)"
      />
      <RequestSQL
        v-else-if='requestKind === "sql"'
        :id="id"
        v-on:update="(request) => store.update(id, request)"
      />
      <RequestGRPC
        v-else-if='requestKind === "grpc"'
        :id="id"
        v-on:send="() => store.send(id)"
        v-on:update="(request) => store.update(id, request)"
      />
      <RequestJQ
        v-else-if='requestKind === "jq"'
        :id="id"
        v-on:send="() => store.send(id)"
        v-on:update="(request) => store.update(id, request)"
      />
      <RequestRedis
        v-else-if='requestKind === "redis"'
        :id="id"
        v-on:send="() => store.send(id)"
        v-on:update="(request) => store.update(id, request)"
      />
    </NTabPane></NTabs>
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
.n-tabs.n-tabs--top .n-tab-pane {
  --n-pane-padding-top: 0;
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
</style>
