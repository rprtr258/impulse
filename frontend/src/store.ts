import {onUnmounted, Reactive, reactive, Ref, ref, UnwrapRef, watch} from "vue";
import {useNotification} from "naive-ui";
import {api, type RequestData, type HistoryEntry} from "./api";
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
    async rename(id: string, newID: string): Promise<void> {
      const request = requests[id];
      const res = await api.request_update(id, request.Kind, request, newID);
      if (res.kind === "err") {
        notify(`Could not rename request: ${res.value}`);
        return;
      }

      tabs.value?.map.rename(id, newID);

      await this.fetch();
    },
  };
}

type UseRequest<Request extends object, Response extends object> = {
  request: Request | null,
  history: HistoryEntry[],
  response: Response | null,
  is_loading: boolean,
  update_request: (patch: Partial<Request>) => Promise<void>,
  send: () => Promise<void>,
};
export function use_request<
  Request extends object,
  Response extends object,
>(request_id: Ref<string>): Reactive<UseRequest<Request, Response>> {
  const notify = useNotify();

  const state = reactive<UseRequest<Request, Response>>({
    request: null,
    history: [],
    response: null,
    is_loading: true,
    send: async () => {
      if (state.request === null || state.is_loading) return;

      state.is_loading = true;
      const res = await api.requestPerform(request_id.value);
      state.is_loading = false;
      if (res.kind === "err") {
        notify(`Could not perform request ${request_id.value}: ${res.value}`);
        return;
      }

      state.history.push(res.value);
      state.response = res.value.response as UnwrapRef<Response>;
    },
    update_request: async (patch: Partial<Request>) => {
      if (state.request === null || state.is_loading) return;

      state.is_loading = true;
      const old_request = state.request;
      const new_request = {...state.request, ...patch} as RequestData;
      state.request = new_request as UnwrapRef<Request>; // NOTE: optimistic update
      const res = await api.request_update(request_id.value, new_request.kind, new_request);
      state.is_loading = false;
      if (res.kind === "err") {
        state.request = old_request; // NOTE: undo change
        notify(`Could not save current request: ${res.value}`);
        return;
      }
    },
  });
  const fetchData = async () => {
    state.is_loading = true;
    const res = await api.get(request_id.value);
    state.is_loading = false;
    if (res.kind === "err") {
      notify("load request", request_id.value, res.value);
      return;
    }

    state.request = res.value.Request as UnwrapRef<Request>;
    state.history = res.value.History as unknown as HistoryEntry[];
    state.response = state.history[state.history.length - 1]?.response as UnwrapRef<Response> ?? null;
  };

  const stopWatch = watch(() => request_id.value, fetchData, {immediate: true});
  onUnmounted(() => {
    stopWatch();
  });

  return state;
}
