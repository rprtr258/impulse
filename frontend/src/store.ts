import {reactive, ref} from "vue";
import {useNotification} from "naive-ui";
import {
  api, type HistoryEntry, RequestData,
  ResponseSQL, ResponseHTTP, ResponseGRPC, ResponseJQ,
} from "./api";
import {app} from '../wailsjs/go/models';
// import { useTimeoutPoll } from "@vueuse/core";

const requestsTree = ref<app.Tree>(new app.Tree({IDs: [], Dirs: {}}));
const requests = reactive<Record<string, RequestData>>({});
const history = reactive<HistoryEntry[]>([]);
const requestID = ref<string | null>(null);
const response = reactive<{box: ResponseHTTP | ResponseSQL | ResponseGRPC | ResponseJQ | null}>({box: null});

export function useStore() {
  const notify = (...args) => useNotification().error({title: "Error", content: args.map(arg => arg.toString()).join("\n")});

  return {
    requestsTree,
    requests,
    history,
    requestID,
    response,
    request(): RequestData | null {
      const id = requestID.value;
      if (id === null) {
        return null;
      }
      return requests[id] ?? null;
    },
    selectRequest(id: string) {
      requestID.value = id;
      response.box = history.find((h: HistoryEntry) => h.RequestId === id)?.response ?? null;
    },
    async fetch(): Promise<void> {
      const json = await api.collectionRequests();
      if (json.kind === "err") {
        notify(`Could not fetch requests: ${json.value}`);
        return;
      }

      const res = json.value;
      requestsTree.value = res.Tree;
      Object.assign(requests, res.Requests);
      Object.assign(history, res.History);
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

      if (requests[id]) {
        delete requests[id];
      }
      await this.fetch();
    },
    async send(id: string) {
      const res = await api.requestPerform(id);
      if (res.kind === "err") {
        notify(`Could not perform request: ${res.value}`);
        return;
      }

      response.box = res.value;
      await this.fetch();
      this.selectRequest(id);
    },
    async update(id: string, req: Omit<RequestData, "kind">) {
      requestID.value = id;
      const request = this.request();
      if (request === null) {
        notify(`Could update request: ${id}`);
        return;
      }

      requests[id] = {kind: request.kind, ...req} as RequestData;
      const res = await api.requestUpdate(id, request.kind, req);
      if (res.kind === "err") {
        notify(`Could not save current request: ${res.value}`);
        return;
      }

      await this.fetch();
    },
    async rename(id: string, newID: string) {
      const request = requests[id];
      const res = await api.requestUpdate(id, request.kind, request, newID);
      if (res.kind === "err") {
        notify(`Could not rename request: ${res.value}`);
        return;
      }

      await this.fetch();
    },
  };
}

// TODO: use SSE
// const {} = useTimeoutPoll(() => {
//   if (store) {
//     store.fetch();
//   }
// }, 1000);
