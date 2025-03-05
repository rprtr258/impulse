import {reactive, ref, watch} from "vue";
import {useNotification} from "naive-ui";
import type { RequestData,
  ResponseData} from "./api";
import {
  api, type HistoryEntry
} from "./api";
import {app} from '../wailsjs/go/models';

interface OrderedSet {
  list: string[],
  set: Set<string>,
  length: () => number,
  has: (key: string) => boolean,
  index: (key: string) => number | null,
  add: (key: string) => void,
  remove: (key: string) => void,
  removeAt: (index: number) => void,
  rename: (keyOld: string, keyNew: string) => void,
}

function orderedMap(...elems: readonly string[]): OrderedSet {
  const set = new Set(elems);
  return {
    list: [...elems],
    set,
    length(): number {
      return this.list.length;
    },
    has(key: string): boolean {
      return set.has(key);
    },
    index(key: string): number | null {
      const index = this.list.indexOf(key);
      return index === -1 ? null : index;
    },
    add(key: string): void {
      if (this.has(key)) {
        return;
      }
      this.list.push(key);
      set.add(key);
    },
    remove(key: string): void {
      const index = this.list.indexOf(key);
      if (index === -1) {
        return;
      }
      this.removeAt(index);
    },
    removeAt(index: number): void {
      this.list.splice(index, 1);
      set.delete(this.list[index]);
    },
    rename(keyOld, keyNew): void {
      const index = this.list.indexOf(keyOld);
      if (index === -1) {
        return;
      }
      this.list[index] = keyNew;
      set.delete(keyOld);
      set.add(keyNew);
    },
  };
}

const requestsTree = ref<app.Tree>(new app.Tree({IDs: [], Dirs: {}}));
const requests = reactive<Record<string, RequestData>>({});
const history = reactive<HistoryEntry[]>([]);

export function useStore() {
  const usenotification = useNotification();
  const notify = (...args: readonly unknown[]): void => {
    usenotification.error({title: "Error", content: args.map(arg => String(arg)).join("\n")});
  };
  const tabs = reactive<{value: {
    map: OrderedSet,
    index: number,
  } | null}>({value: null});
  watch(() => requests, () => {
    if (!tabs.value) {
      return;
    }

    const indexesToRemove = tabs.value.map.list
      .map((id: string, i: number) => [id, i] as [string, number])
      .filter(([id]: readonly [string, number]) => !requests.hasOwnProperty(id))  // Changed condition here
      .map(([, i]: readonly [string, number]) => i);

    if (indexesToRemove.length === 0) {
      return;
    } else if (indexesToRemove.length === tabs.value.map.length()) {
      tabs.value = null;
      return;
    }

    // Remove tabs from highest index to lowest to avoid shifting issues
    for (const i of indexesToRemove.sort((a, b) => b - a)) {
      tabs.value.map.removeAt(i);
      if (tabs.value.index === i && tabs.value.index > 0) {
        tabs.value.index--;
      }
    }
  }, {immediate: true, deep: true});

  return {
    requestsTree,
    requests,
    history,
    tabs,
    request(): RequestData | null {
      const tabsValue = tabs.value;
      if (tabsValue === null) {
        return null;
      }
      const {map: requestIDs, index} = tabsValue;
      return requests[requestIDs.list[index]] ?? null;
    },
    getResponse(id: string): Omit<ResponseData, "kind"> | null {
      return history.find((h: Readonly<HistoryEntry>) => h.RequestId === id)?.response ?? null;
    },
    selectRequest(id: string): void {
      const tabsValue = tabs.value;
      if (tabsValue === null) {
        // no tabs open, create one
        tabs.value = {
          map: orderedMap(id),
          index: 0,
        };
        return;
      }

      const requestIDs = tabsValue.map;
      const indexNew = requestIDs.index(id);
      if (indexNew === null) {
        // tab with such id not found, add new
        requestIDs.add(id);
        tabs.value = {
          map: requestIDs,
          index: requestIDs.length() - 1,
        };
        return;
      }

      tabsValue.index = indexNew;
    },
    async fetch(): Promise<void> {
      const json = await api.collectionRequests();
      if (json.kind === "err") {
        notify(`Could not fetch requests: ${json.value}`);
        return;
      }

      const res = json.value;
      requestsTree.value = res.Tree;

      // Don't overwrite currently edited request
      const currentRequest = this.request();
      const currentRequestId = currentRequest ? tabs.value?.map.list[tabs.value.index] : null;

      // Update requests, preserving the current one if it exists
      Object.entries(res.Requests).forEach(([id, request]) => {
        if (id !== currentRequestId) {
          requests[id] = request;
        }
      });

      // Update history
      Object.assign(history, res.History);
    },
    async createRequest(id: string, kind: RequestData["kind"]): Promise<void> {
      const res = await api.requestCreate(id, kind);
      if (res.kind === "err") {
        notify(`Could not create request: ${res.value}`);
        return;
      }

      await this.fetch();
    },
    async duplicate(id: string): Promise<void> {
      const res = await api.requestDuplicate(id);
      if (res.kind === "err") {
        notify(`Could not duplicate: ${res.value}`);
        return;
      }

      await this.fetch();
    },
    async deleteRequest(id: string): Promise<void> {
      const res = await api.requestDelete(id);
      if (res.kind === "err") {
        notify(`Could not delete request: ${res.value}`);
        return;
      }
      if (requests.hasOwnProperty(id)) {
        delete requests[id];
      }
      await this.fetch();
    },
    async send(id: string): Promise<void> {
      const res = await api.requestPerform(id);
      if (res.kind === "err") {
        notify(`Could not perform request: ${res.value}`);
        return;
      }

      await this.fetch();
      this.selectRequest(id);
    },
    async update(id: string, req: Omit<RequestData, "kind">): Promise<void> {
      this.selectRequest(id);
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
    async rename(id: string, newID: string): Promise<void> {
      const request = requests[id];
      const res = await api.requestUpdate(id, request.kind, request, newID);
      if (res.kind === "err") {
        notify(`Could not rename request: ${res.value}`);
        return;
      }

      tabs.value?.map.rename(id, newID);

      await this.fetch();
    },
  };
}
