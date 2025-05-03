<script setup lang="ts">
import {computed, h, onMounted, ref, VNodeChild, watch} from "vue";
import {useBrowserLocation, useLocalStorage, useMagicKeys} from "@vueuse/core";
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
import {Command, useCommandState} from 'vue-command-palette'
import {useStore} from "./store";
import {Method, Kinds, Database, api, HistoryEntry} from "./api";
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
    store.clearTabs();
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
const newRequestName = ref<string | null>(null);
function create() {
  const kind = newRequestKind.value!;
  const name = newRequestName.value!;
  store.createRequest(name, kind);
  createCancel();
}
function createCancel() {
  newRequestKind.value = null;
  newRequestName.value = null;
}
watch(() => newRequestKind.value, () => {
  newRequestName.value = new Date().toUTCString();
});

// TODO: fix editing request headers

const renameID = ref<string | null>(null);
const renameValue = ref<string | null>(null);
function renameInit(id: string) {
  renameID.value = id;
  renameValue.value = id;
}
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
          onClick: () => renameInit(id),
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

// TODO: reverse history order
// const history = computed(() => use_request(store.requestID()!).value?.history ?? []);
const history: HistoryEntry[] = [];
const requestKind = computed(() => {
  const requestID = store.requestID();
  if (requestID === null) {
    return null;
  }
  return store.requests[requestID].Kind;
});





const commandBarNewRequestKindVisible = ref(false);
const commandBarVisible = ref(false);
const commandBarOpenVisible = ref(false);
const items = [
  {
    label: "Requests",
    items: [
      {
        label: "Create new",
        shortcut: ["Ctrl", "N"],
        perform: () => {
          commandBarVisible.value = false;

          commandBarNewRequestKindVisible.value = true;
        },
      },
      {
        label: "Rename current", // TODO: hide if no request selected
        // shortcut: ["Ctrl", "R"],
        perform: () => {
          commandBarVisible.value = false;

          const currentID = store.requestID();
          if (currentID === null) {return;}
          renameInit(currentID);
        },
      },
      {
        label: "Open",
        shortcut: ["Alt", "T"],
        perform: () => {
          commandBarVisible.value = false;

          commandBarOpenVisible.value = true;
        },
      },
      {
        label: "Run",
        shortcut: ["Ctrl", "Enter"],
      },
      {
        label: "Duplicate",
      },
      {
        label: "Delete",
      },
    ],
  },
  {
    label: "Tabs",
    items: [
      {
        label: "Next tab",
        shortcut: ["Ctrl", "PgDown"],
      },
      {
        label: "Previous tab",
        shortcut: ["Ctrl", "PgUp"],
      },
      {
        label: "Close tab",
        shortcut: ["Ctrl", "W"],
        perform: () => {
          commandBarVisible.value = false;

          const currentID = store.requestID();
          if (currentID === null) {return;}
          handleClose(currentID);
        },
      },
      {
        label: "Move tab right",
        shortcut: ["Ctrl", "Shift", "PgDown"],
      },
      {
        label: "Move tab left",
        shortcut: ["Ctrl", "Shift", "PgUp"],
      },
    ],
  },
  {
    label: "Other",
    items: [
      {
        label: "Create new directory",
      },
    ],
  },
];
const openItems = computed(() => {
  return Object.entries(store.requests).map(([id, preview]) => ({
    id: id,
    kind: preview.Kind,
  }));
});
const keys = useMagicKeys();
const anyModalIsOpen = computed(() => newRequestName.value !== null || renameID.value !== null);
watch(keys['Alt+K'], (v) => {
  if (!v || anyModalIsOpen.value) {
    return;
  }
  commandBarVisible.value = !commandBarVisible.value;
});
watch(keys['Alt+T'], (v) => {
  if (!v || anyModalIsOpen.value) {
    return;
  }
  commandBarOpenVisible.value = !commandBarOpenVisible.value;
});
watch(keys['Escape'], (v) => {
  if (!v) {
    return;
  }
  renameCancel();
  createCancel();
  commandBarVisible.value = false;
  commandBarNewRequestKindVisible.value = false;
  commandBarOpenVisible.value = false;
});


</script>

<template>
<Command.Dialog :visible="commandBarVisible" theme="algolia">
  <template #header>
    <Command.Input placeholder="Type a command or search..." />
  </template>
  <template #body>
    <Command.List>
      <Command.Empty>No results found.</Command.Empty>
      <Command.Group
        v-for="group in items"
        :heading="group.label"
      >
        <Command.Item
          v-for="item in group.items"
          :data-value="item.label"
          :shortcut="item.shortcut"
          v-on:select="item.perform"
        >
          <div>{{item.label}}</div>
          <div
            command-linear-shortcuts
            class="hidden"
            v-show="item.shortcut !== undefined"
            style="color: var(--gray10); align-items: center;"
          >
            <template v-for="(k, i) in item.shortcut">
              <div v-if="i !== 0" style="padding: 0px 2px;">+</div>
              <kbd>{{k}}</kbd>
            </template>
          </div>
        </Command.Item>
      </Command.Group>
    </Command.List>
  </template>
  <template #footer>
    <ul class="command-palette-commands">
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Enter key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M12 3.53088v3c0 1-1 2-2 2H4M7 11.53088l-3-3 3-3"
            />
          </g></svg></kbd><span class="command-palette-Label">to select</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow down" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 3.5v8M10.5 8.5l-3 3-3-3" />
          </g></svg></kbd><kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow up" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 11.5v-8M10.5 6.5l-3-3-3 3" />
          </g></svg></kbd><span class="command-palette-Label">to navigate</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Escape key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M13.6167 8.936c-.1065.3583-.6883.962-1.4875.962-.7993 0-1.653-.9165-1.653-2.1258v-.5678c0-1.2548.7896-2.1016 1.653-2.1016.8634 0 1.3601.4778 1.4875 1.0724M9 6c-.1352-.4735-.7506-.9219-1.46-.8972-.7092.0246-1.344.57-1.344 1.2166s.4198.8812 1.3445.9805C8.465 7.3992 8.968 7.9337 9 8.5c.032.5663-.454 1.398-1.4595 1.398C6.6593 9.898 6 9 5.963 8.4851m-1.4748.5368c-.2635.5941-.8099.876-1.5443.876s-1.7073-.6248-1.7073-2.204v-.4603c0-1.0416.721-2.131 1.7073-2.131.9864 0 1.6425 1.031 1.5443 2.2492h-2.956"
            />
          </g></svg></kbd><span class="command-palette-Label">to close</span>
      </li>
    </ul>
  </template>
</Command.Dialog>
<Command.Dialog :visible="commandBarNewRequestKindVisible" theme="algolia">
  <template #header>
    <Command.Input placeholder="Enter kind of new reqeust" />
  </template>
  <template #body>
    <Command.List>
      <Command.Empty>No kinds found.</Command.Empty>
      <Command.Item
        v-for="kind in Kinds"
        :data-value="kind"
        v-on:select="() => {newRequestKind = kind; commandBarNewRequestKindVisible = false;}"
      >
        {{kind}}
      </Command.Item>
    </Command.List>
  </template>
  <template #footer>
    <ul class="command-palette-commands">
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Enter key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M12 3.53088v3c0 1-1 2-2 2H4M7 11.53088l-3-3 3-3"
            />
          </g></svg></kbd><span class="command-palette-Label">to select</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow down" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 3.5v8M10.5 8.5l-3 3-3-3" />
          </g></svg></kbd><kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow up" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 11.5v-8M10.5 6.5l-3-3-3 3" />
          </g></svg></kbd><span class="command-palette-Label">to navigate</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Escape key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M13.6167 8.936c-.1065.3583-.6883.962-1.4875.962-.7993 0-1.653-.9165-1.653-2.1258v-.5678c0-1.2548.7896-2.1016 1.653-2.1016.8634 0 1.3601.4778 1.4875 1.0724M9 6c-.1352-.4735-.7506-.9219-1.46-.8972-.7092.0246-1.344.57-1.344 1.2166s.4198.8812 1.3445.9805C8.465 7.3992 8.968 7.9337 9 8.5c.032.5663-.454 1.398-1.4595 1.398C6.6593 9.898 6 9 5.963 8.4851m-1.4748.5368c-.2635.5941-.8099.876-1.5443.876s-1.7073-.6248-1.7073-2.204v-.4603c0-1.0416.721-2.131 1.7073-2.131.9864 0 1.6425 1.031 1.5443 2.2492h-2.956"
            />
          </g></svg></kbd><span class="command-palette-Label">to close</span>
      </li>
    </ul>
  </template>
</Command.Dialog>
<Command.Dialog :visible="commandBarOpenVisible" theme="algolia">
  <template #header>
    <Command.Input placeholder="Type or search request id to open" />
  </template>
  <template #body>
    <Command.List>
      <Command.Empty>No requests found.</Command.Empty>
      <Command.Item
        v-for="item in openItems"
        :data-value="item.id"
        v-on:select="() => {commandBarOpenVisible = false; selectRequest(item.id);}"
      >
        <div>[{{item.kind}}] {{item.id}}</div>
      </Command.Item>
    </Command.List>
  </template>
  <template #footer>
    <ul class="command-palette-commands">
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Enter key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M12 3.53088v3c0 1-1 2-2 2H4M7 11.53088l-3-3 3-3"
            />
          </g></svg></kbd><span class="command-palette-Label">to select</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow down" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 3.5v8M10.5 8.5l-3 3-3-3" />
          </g></svg></kbd><kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Arrow up" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path d="M7.5 11.5v-8M10.5 6.5l-3-3-3 3" />
          </g></svg></kbd><span class="command-palette-Label">to navigate</span>
      </li>
      <li>
        <kbd class="command-palette-commands-key"><svg width="15" height="15" aria-label="Escape key" role="img">
          <g
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.2"
          >
            <path
              d="M13.6167 8.936c-.1065.3583-.6883.962-1.4875.962-.7993 0-1.653-.9165-1.653-2.1258v-.5678c0-1.2548.7896-2.1016 1.653-2.1016.8634 0 1.3601.4778 1.4875 1.0724M9 6c-.1352-.4735-.7506-.9219-1.46-.8972-.7092.0246-1.344.57-1.344 1.2166s.4198.8812 1.3445.9805C8.465 7.3992 8.968 7.9337 9 8.5c.032.5663-.454 1.398-1.4595 1.398C6.6593 9.898 6 9 5.963 8.4851m-1.4748.5368c-.2635.5941-.8099.876-1.5443.876s-1.7073-.6248-1.7073-2.204v-.4603c0-1.0416.721-2.131 1.7073-2.131.9864 0 1.6425 1.031 1.5443 2.2492h-2.956"
            />
          </g></svg></kbd><span class="command-palette-Label">to close</span>
      </li>
    </ul>
  </template>
</Command.Dialog>
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
          :show="newRequestKind !== null"
          v-on:update-show="(show: boolean) => {if (!show) { createCancel(); }}"
          preset="dialog"
          title="Create request"
          positive-text="Create"
          negative-text="Cancel"
          v-on:positive-click="create"
          v-on:negative-click="createCancel"
        >
          <NInput v-model:value="newRequestName" />
        </NModal>
        <NModal
          :show="renameID !== null"
          v-on:update-show="(show: boolean) => {if (show) { renameCancel(); }}"
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
          :options="Kinds.map((kind: database.Kind) => ({label: kind.toUpperCase(), value: kind}))"
        />
        <NScrollbar trigger="none">
          <NTree
            block-line
            expand-on-click
            :selected-keys='[(store.tabs.value ? store.tabs.value.map.list[store.tabs.value.index] : "")]'
            :show-line="true"
            :data="treeData"
            :draggable="true"
            :default-expanded-keys="expandedKeys"
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
          v-else-if="history.length === 0"
          description="No history yet"
          class="h100"
          style="justify-content: center;"
        />
        <NList v-else hoverable :border="false">
          <NListItem
            v-for="r, i in history"
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
      v-on:update:value="(id: string) => selectRequest(id)"
    ><NTabPane
      v-for="id in store.tabs.value.map.list"
      :key="id"
      :name="id"
      :tab="id"
      class="h100"
      display-directive="if"
    >
      <RequestHTTP       v-if='requestKind === "http"'  :id="id" />
      <RequestSQL   v-else-if='requestKind === "sql"'   :id="id" />
      <RequestGRPC  v-else-if='requestKind === "grpc"'  :id="id" />
      <RequestJQ    v-else-if='requestKind === "jq"'    :id="id" />
      <RequestRedis v-else-if='requestKind === "redis"' :id="id" />
    </NTabPane></NTabs>
  </div>
</div>
</template>

<style scoped>
</style>

<style lang="scss">
.algolia {
  --gray1: hsl(0, 0%, 8.5%);
  --gray2: hsl(0, 0%, 11%);
  --gray3: hsl(0, 0%, 13.6%);
  --gray4: hsl(0, 0%, 15.8%);
  --gray5: hsl(0, 0%, 17.9%);
  --gray6: hsl(0, 0%, 20.5%);
  --gray7: hsl(0, 0%, 24.3%);
  --gray8: hsl(0, 0%, 31.2%);
  --gray9: hsl(0, 0%, 43.9%);
  --gray10: hsl(0, 0%, 49.4%);
  --gray11: hsl(0, 0%, 62.8%);
  --gray12: hsl(0, 0%, 93%);

  --lowContrast: #ffffff;
  --highContrast: #000000;
  --vcp-c-brand: #44bd87;
  --vcp-c-accent: #35495e;

  --bg: var(--gray1);

  [command-root] {}

  [command-input] {
    box-sizing: border-box;
    font-family: var(--font-sans);
    width: 100%;
    font-size: 18px;
    padding: 12px;
    outline: none;
    background: var(--bg);
    color: var(--gray12);
    border-bottom: 1px solid var(--gray6);
    border-radius: 4px;
    caret-color: var(--vcp-c-brand);
    border: 1px solid var(--vcp-c-brand);

    &::placeholder {
      color: var(--gray9);
    }
  }

  [command-list] {
    height: var(--command-list-height);
    max-height: 360px;
    overflow: auto;
    overscroll-behavior: contain;
  }

  [command-item] {
    position: relative;
    content-visibility: auto;
    cursor: pointer;
    height: 2.5em;
    font-size: 14px;
    display: flex;
    align-items: center;
    padding: 0px 16px;
    color: var(--gray12);
    user-select: none;
    will-change: background, color;
    border-radius: 4px;

    [command-linear-shortcuts] {
      color: var(--gray8);
      margin-left: auto;
      display: flex;

      &:hover {
        color: #fff;
      }

      kbd {
        align-items: center;
        background: var(--gray5);
        border-radius: 2px;
        display: flex;
        height: 18px;
        justify-content: center;
        color: var(--gray11);
        border: 0;
        padding: 0px 5px;

        font-family: var(--font-sans);
        font-size: 13px;
      }
    }

    &[aria-selected='true'],
    &:hover {
      background: var(--vcp-c-brand);
      color: #fff;

      svg {
        color: #fff;
      }
    }

    &[aria-disabled='true'] {
      color: var(--gray8);
      cursor: not-allowed;
    }

    &:active {
      background: var(--gray4);
    }

    svg {
      width: 16px;
      height: 16px;
      color: var(--gray10);
    }
  }

  [command-empty=''] {
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 64px;
    white-space: pre-wrap;
    color: var(--gray11);
  }

  [command-dialog-mask] {
    background-color: rgba(75, 75, 75, 0.35);
    left: 0;
    position: fixed;
    top: 0;
    width: 100vw;
    height: 100vh;
  }

  [command-dialog-header] {}

  [command-dialog-body] {}

  [command-dialog-wrapper] {
    margin: 20vh auto auto;
    max-width: 560px;
    background: var(--gray2);
    border-radius: 6px;
  }

  [command-dialog-footer] {
    align-items: center;
    border-radius: 0 0 8px 8px;
    box-shadow: none;
    display: flex;
    flex-direction: row;
    flex-shrink: 0;
    height: 44px;
    justify-content: space-between;
    padding: 0 12px;
    position: relative;
    user-select: none;
    width: 100%;
    z-index: 300;
    border-top: 1px solid var(--gray6);
    background: var(--gray4);
    box-sizing: border-box;
  }

  [command-group-heading] {
    color: var(--vcp-c-brand);
    font-size: 0.85em;
    font-weight: 600;
    line-height: 1.2rem;
    top: 0;
    z-index: 10;
    width: 100%;
    box-sizing: border-box;
    padding: 5px 12px;
  }

  .command-palette-commands {
    color: var(--gray11);
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;

    li {
      display: flex;
      align-items: center;
    }
    li:not(:last-of-type) {
      margin-right: 0.8em;
    }
  }

  .command-palette-logo {
    a {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    svg {
      height: 24px;
      width: 24px;
    }
  }

  .command-palette-commands-key {
    align-items: center;
    background: var(--gray3);
    border-radius: 2px;
    display: flex;
    height: 18px;
    justify-content: center;
    margin-right: 0.4em;
    padding: 0 0 1px;
    color: var(--gray11);
    border: 0;
    width: 20px;
  }
}

.dark .algolia [command-dialog-footer] {
  box-shadow: none;
}








































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
