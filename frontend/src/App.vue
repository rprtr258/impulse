<script setup lang="ts">
import {computed, onMounted, reactive, ref, watch} from "vue";
import {NConfigProvider, darkTheme, NTag, NTabs, NTabPane, NList, NListItem, NCollapse, NCollapseItem, NResult, NSelect, NInput, NModal} from "naive-ui";
import {api, Method as Methods, type HistoryEntry, RequestData, Database, RequestSQL as RequestSQLT, RequestHTTP as RequestHTTPT, ResponseHTTP, ResponseSQL} from "./api";
import Request from "./Request.vue";
import RequestHTTP from "./RequestHTTP.vue";
import RequestSQL from "./RequestSQL.vue";

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

let requests = ref({} as Record<string, RequestData>); // TODO: tree
let history = ref([] as HistoryEntry[]);

const request = reactive({box: null} as {
  box: {
    id: string,
  } & ({
    kind: "http",
    request: RequestHTTPT,
  } | {
    kind: "sql",
    request: RequestSQLT,
  }) | null,
});
const request_history = computed(() => {
  if (request.box === null) {
    return [];
  }

  return history.value.filter(h => h.request_id === request.box.id);
});

function updateRequest() { // TODO: replace with event
  api
    .requestUpdate(collectionID.value, request.box.id, request.box.kind, request.box.request)
    .then(() => { fetch(collectionID.value); })
    .catch((err) => alert(`Could not save current request: ${err}`));
}
watch(request, updateRequest, { deep: true }); // TODO: replace with events & handler

function fetch(collectionID: string): void {
  api
    .collectionRequests(collectionID)
    .then((json) => {
      requests.value = Object.fromEntries(json.requests.map(r => [r.id, r as RequestHTTPT & {kind: 'http'} | RequestSQLT & {kind: 'sql'}]));
      history.value = json.history;
    })
    .catch((err) => { throw err; }); // TODO: handle error
}

// const collapsed = reactive({});// ref<{ [i: number]: boolean }>({});
// TODO: remove collections altogether, use tree instead
const collectionID = ref("Sanya"); // TODO: default collection must be <unknown>

function selectRequest(id: string, req: RequestHTTPT | RequestSQLT) {
  const kind = requests.value[id].kind;
  switch (kind) {
    case "http":
      request.box = {
        id: id,
        kind: kind,
        request: req as RequestHTTPT,
      };
      break;
    case "sql":
      request.box = {
        id: id,
        kind: kind,
        request: req as RequestSQLT,
      };
      break;
  }
}
function deleteRequest(id: string) {
  api
    .requestDelete(collectionID.value, id)
    .then(() => {
      delete requests.value[id];
      fetch(collectionID.value);
    })
    .catch((err) => alert(`Could not delete request: ${err}`));
}

function selectHistoryRequest(req: HistoryEntry) {
  selectRequest(req.request_id, req.request);
}

const newRequestKind = ref<"http" | "sql" | null>(null);
watch(newRequestKind, function() {
  if (newRequestKind.value === null) {
    return;
  }

  api
    .requestCreate(collectionID.value, "test-create", newRequestKind.value)
    .then(() => {
      fetch(collectionID.value);
    }); // TODO: add name
});

// TODO: fix editing request headers

onMounted(() => {
  fetch(collectionID.value);
});
const openCollections = reactive(['Sanya']); // TODO: save to/load from local storage

const renameID = ref(null as string | null);
const renameValue = ref(null as string | null);
function renameCancel() {
  renameID.value = null;
  renameValue.value = null;
}
function rename() {
  const req = requests.value[renameID.value];
  api.requestUpdate(
    collectionID.value,
    renameID.value,
    req.kind,
    req,
    renameValue.value,
  ).catch((err) => alert(`Could not rename request: ${err}`));
  renameCancel();
  fetch(collectionID.value);
}
</script>

<template>
<NConfigProvider :theme='darkTheme' class="h100">
  <div class="h100" style="display: grid; grid-template-columns: 1fr 6fr;">
    <!-- TODO: highlight selected item -->
    <aside style="width: 272px; overflow: auto; color: rgba(255, 255, 255, 0.82); background-color: rgb(24, 24, 28);">
      <NTabs
        type="card"
        size="small"
      >
        <NTabPane
          name="tab-nav-collection"
          tab="Collection"
          style="flex: 1;"
        >
          <NModal
            :show="renameID !== null"
            v-on:update-show="(show) => {if (!show) { renameCancel(); }}"
            preset="dialog"
            title="Rename request"
            positive-text="Rename"
            negative-text="Cancel"
            @positive-click="rename"
            @negative-click="renameCancel"
          >
            <NInput v-model:value="renameValue" />
          </NModal>
          <NSelect
            v-model:value="newRequestKind"
            placeholder="New"
            clearable
            :options='[{label:"HTTP", value:"http"}, {label:"SQL", value:"sql"}]'
          />
          <NCollapse :default-expanded-names="openCollections" >
            <NCollapseItem name="Sanya" title="Sanya">
              <NList hoverable :border="false">
                <NListItem
                  v-for="(req, id) in requests"
                  :key="id"
                  :animated="false"
                >
                  <Request
                    :id="id"
                    :method='req.kind=="http" ? Methods[req.method] : Database[req.database]'
                    v-on:click="() => selectRequest(id, req)"
                    v-on:rename="() => {renameID = id; renameValue = id;}"
                    v-on:delete="() => deleteRequest(id)"
                  />
                </NListItem>
              </NList>
            </NCollapseItem>
          </NCollapse>
        </NTabPane>
        <NTabPane
          name="tab-nav-history"
          tab="History"
          style="flex: 1;"
        >
          <NList hoverable :border="false">
            <NListItem
              v-for='r, i in history'
              :key='i'
              v-on:click='selectHistoryRequest(r)'
              class='history-card card'
            >
              <div class='headline'>
                <NTag
                  type='info'
                  class='method'
                  size="small"
                  style="width: 4em; justify-content: center;"
                >{{(() => {
                  const tmp = requests[r.request_id];
                  return tmp.kind == "http" ? Methods[tmp.method] : Database[tmp.database];
                })()}}</Ntag>
                <span class='url' style="padding-left: .5em;">{{r.request_id}}</span>
              </div>
              <div class='footer'>
                <span style='color: grey;' class='date'>{{fromNow(r.sent_at)}}</span>
              </div>
            </NListItem>
          </NList>
        </NTabPane>
      </NTabs>
    </aside>
    <div style="color: rgba(255, 255, 255, 0.82); background-color: rgb(16, 16, 20); overflow: hidden;">
      <template v-if="request.box === null">
        <NResult
          status="info"
          title="Pick request"
          description="Pick request to see it, edit and send and do other fun things."
          class="h100"
          style="align-content: center;"
        />
      </template><template v-else-if='request.box.kind === "http"'>
        <RequestHTTP
          :id="request.box.id"
          :collectionID="collectionID"
          :request="request.box.request"
          :response='request_history[0]?.response as ResponseHTTP ?? null'
        />
      </template><template v-else-if='request.box.kind === "sql"'>
        <RequestSQL
          :collectionID="collectionID"
          :id="request.box.id"
          :request="request.box.request"
          :response='request_history[0]?.response as ResponseSQL ?? null'
        />
      </template>
    </div>
  </div>
</NConfigProvider>
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
