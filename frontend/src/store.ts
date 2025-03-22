import {Reactive, reactive, Ref, ref, UnwrapRef, watch} from "vue";
import {useNotification} from "naive-ui";
import type {RequestData, HistoryEntry} from "./api";
import {api} from "./api";
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
  elems = elems ?? [];
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
      set.delete(this.list[index]);
      this.list.splice(index, 1);
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
const requests = reactive<Record<string, app.requestPreview>>({});

function useNotify() {
  const usenotification = useNotification();
  return (...args: readonly unknown[]): void => {
    usenotification.error({title: "Error", content: args.map(arg => String(arg)).join("\n")});
  };
}

export function useStore() {
  const notify = useNotify();
  const tabs = reactive<{value: {
    map: OrderedSet,
    index: number,
  }}>({value: {
    map: orderedMap(),
    index: -1,
  }});
  watch(() => requests, () => {
    if (!tabs.value) {
      return;
    }

    const indexesToRemove = tabs.value.map.list
      .map((id: string, i: number) => [id, i] as [string, number])
      .filter(([id]: readonly [string, number]) => !requests.hasOwnProperty(id))
      .map(([, i]: readonly [string, number]) => i);
    if (indexesToRemove.length === 0) {
      return;
    }

    for (const i of indexesToRemove) {
      tabs.value.map.removeAt(i);
      if (tabs.value.index === i && tabs.value.index > 0) {
        tabs.value.index--;
      }
    }
  }, {immediate: true, deep: true});

  return {
    requestsTree,
    requests,
    tabs,
    clearTabs() {
      tabs.value = {
        map: orderedMap(),
        index: -1,
      };
    },
    requestID(): string | null {
      const tabsValue = tabs.value;
      if (tabsValue === null) {
        return null;
      }
      const {map: requestIDs, index} = tabsValue;
      return requestIDs.list[index] ?? null;
    },
    request(): Promise<RequestData> | null {
      const requestID = this.requestID();
      if (requestID === null) {
        return null;
      }
      return api.get(requestID).then(res => res.kind === "ok" ? res.value.Data : null);
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

      this.fetch();
    },
    async fetch(): Promise<void> {
      const json = await api.collectionRequests();
      if (json.kind === "err") {
        notify(`Could not fetch requests: ${json.value}`);
        return;
      }

      const res = json.value;
      requestsTree.value = res.Tree;

      const currentRequestId = this.requestID();

      for (const id in res.Requests) {
        if (id !== currentRequestId) {
          requests[id] = res.Requests[id];
        }
      }
      for (const id in requests) {
        if (!res.Requests.hasOwnProperty(id)) {
          delete requests[id];
        }
      }
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
    async update(id: string, req: RequestData): Promise<void> {
      this.selectRequest(id);

      const res = await api.requestUpdate(id, req.kind, req);
      if (res.kind === "err") {
        notify(`Could not save current request: ${res.value}`);
        return;
      }

      await this.fetch();
    },
    async rename(id: string, newID: string): Promise<void> {
      const request = requests[id];
      const res = await api.requestUpdate(id, request.Kind, request, newID);
      if (res.kind === "err") {
        notify(`Could not rename request: ${res.value}`);
        return;
      }

      tabs.value?.map.rename(id, newID);

      await this.fetch();
    },
  };
}

export function use_request<R extends object>(request_id: string): Reactive<{value: {
  request: R,
} | null}> {
  const notify = useNotify();

  const request = reactive<{value: {request: R} | null}>({value: null});
  api.get(request_id).then(res => {
    if (res.kind === "err") {
      notify("load request", request_id, res.value);
      return;
    }
    request.value = {
      request: res.value as UnwrapRef<R>,
    };
  });
  return request;
}

export function use_history(request_id: () => string): Reactive<{value: HistoryEntry[]}> {
  const notify = useNotify();

  const response = reactive<{value: HistoryEntry[]}>({value: []});
  watch(request_id, (id) => {
    if (id === null) {
      return;
    }

    api.history(id).then(history => {
      if (history.kind === "err") {
        notify("load history", request_id, history.value);
        return null;
      }

      if (history.value.length === 0) {
        return null;
      }

      response.value = history.value ?? [];
    });
  }, {immediate: true});
  return response;
}

export function use_response<R>(request_id: () => string): Ref<Omit<R, "kind"> | null> {
  const history = use_history(request_id);

  const response = ref<Omit<R, "kind"> | null>(null) as Ref<Omit<R, "kind"> | null>;
  watch(() => history.value, () => {
    if (history.value.length !== 0) {
      response.value = history.value[0].response as R;
    }
  }, {immediate: true});
  return response;
}