import {reactive} from "vue";
import {
  api, type HistoryEntry, RequestData, Tree,
  ResponseSQL, ResponseHTTP, ResponseGRPC, ResponseJQ,
} from "./api";
// import { useTimeoutPoll } from "@vueuse/core";

let notify = (...args) => alert(...args);

export function setNotify(f: (...args) => void) {
  notify = f;
}

export const store = reactive({
  requestsTree: {} as Tree,
  requests: {} as Record<string, RequestData>, // TODO: tree
  history: [] as HistoryEntry[],
  requestID: null as string | null,
  response: null as ResponseHTTP | ResponseSQL | ResponseGRPC | ResponseJQ | null,
  request(): RequestData | null {
    return this.requests[this.requestID] ?? null;
  },
  selectRequest(id: string) {
    this.requestID = id;
    this.response = this.history.find((h: HistoryEntry) => h.request_id === id)?.response ?? null;
  },
  async fetch(): Promise<void> {
    const json = await api.collectionRequests();
    if (json.kind === "err") {
      notify(`Could not fetch requests: ${json.value}`);
      return;
    }

    const res = json.value;
    this.requestsTree = res.tree;
    this.requests = res.requests;
    this.history = res.history;
    // TODO: unselect request if it does not exist anymore
  },
  async createRequest(id: string, kind: RequestData["kind"]) {
    const res = await api.requestCreate(id, kind);
    if (res.kind === "err") {
      notify(`Could not create request: ${res.value}`);
      return;
    }

    await this.fetch();
  },
  async duplicate(id: string) {
    const res = await api.requestDuplicate(id);
    if (res.kind === "err") {
      notify(`Could not duplicate: ${res.value}`);
      return;
    }

    await this.fetch();
  },
  async deleteRequest(id: string) {
    const res = await api.requestDelete(id);
    if (res.kind === "err") {
      notify(`Could not delete request: ${res.value}`);
      return;
    }

    if (this.requests[id]) {
      delete this.requests[id];
    }
    await this.fetch();
  },
  async send(id: string) {
    const res = await api.requestPerform(id);
    if (res.kind === "err") {
      notify(`Could not perform request: ${res.value}`);
      return;
    }

    this.response = res.value;
    await this.fetch();
    this.selectRequest(id);
  },
  async update(id: string, req: Omit<RequestData, "kind">) {
    this.requestID = id;
    this.requests[id] = req;
    const request = this.request();

    const res = await api.requestUpdate(id, request.kind, {kind: request.kind, ...request});
    if (res.kind === "err") {
      notify(`Could not save current request: ${res.value}`);
      return;
    }

    await store.fetch();
  },
  async rename(id: string, newID: string) {
    const request = this.requests[id];
    const res = await api.requestUpdate(id, request.kind, request, newID);
    if (res.kind === "err") {
      notify(`Could not rename request: ${res.value}`);
      return;
    }

    await store.fetch();
  },
});

// TODO: use SSE
// const {} = useTimeoutPoll(() => {
//   if (store) {
//     store.fetch();
//   }
// }, 1000);
