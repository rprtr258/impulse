import m from "mithril";
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

// TODO: <NNotificationProvider :max="1" placement="bottom-right">
export function useNotification() {
  const notify = (...args: readonly unknown[]): void => {
    alert(JSON.stringify({title: "Error", content: args.map(arg => String(arg)).join("\n")}));
  };
  return {
    error: notify,
  };
}
export const notification = useNotification();

const tabs : {value: {
  map: OrderedSet,
  index: number,
}} = {value: {
  map: orderedMap(),
  index: -1,
}};

function updateLocalstorageTabs() {
  if (store.tabs.value) {
    localStorage.setItem("tabs", JSON.stringify(store.tabs.value?.map.list ?? []));
  } else {
    localStorage.removeItem("tabs");
  }
}

export function handleCloseTab(id: string) {
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
  store.selectRequest(v.map.list[v.index]);
  updateLocalstorageTabs();
}

export function useStore() {
  const notify = notification.error;

  const load = () => {
    // if (!tabs.value) {
    //   return;
    // }

    // const indexesToRemove = tabs.value.map.list
    //   .map((id: string, i: number) => [id, i] as [string, number])
    //   .filter(([id]: readonly [string, number]) => !requests.hasOwnProperty(id))
    //   .map(([, i]: readonly [string, number]) => i);
    // if (indexesToRemove.length === 0) {
    //   return;
    // }

    // for (const i of indexesToRemove) {
    //   tabs.value.map.removeAt(i);
    //   if (tabs.value.index === i && tabs.value.index > 0) {
    //     tabs.value.index--;
    //   }
    // }
  };

  return {
    requestsTree : new app.Tree({IDs: [], Dirs: {}}) as app.Tree,
    requests : {} as Record<string, app.requestPreview>,
    requests2: {} as Record<string, UseRequest<any, any>>,
    load,
    tabs,
    clearTabs() {
      tabs.value = {
        map: orderedMap(),
        index: -1,
      };
      updateLocalstorageTabs();
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
        updateLocalstorageTabs();
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
        updateLocalstorageTabs();
        return;
      }

      tabsValue.index = indexNew;
      this.fetch();
    },
    async fetch(): Promise<void> {
      const json = await api.collectionRequests();
      if (json.kind === "err") {
        notification.error(`Could not fetch requests: ${json.value}`);
        return;
      }

      const res = json.value;
      this.requestsTree = res.Tree;

      const currentRequestId = this.requestID();

      for (const id in res.Requests) {
        if (id !== currentRequestId) {
          this.requests[id] = res.Requests[id];
        }
      }
      for (const id in this.requests) {
        if (!res.Requests.hasOwnProperty(id)) {
          delete this.requests[id];
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
      if (this.requests.hasOwnProperty(id)) {
        delete this.requests[id];
      }
      await this.fetch();
    },
    async rename(id: string, newID: string): Promise<void> {
      const res = await api.rename(id, newID);
      if (res.kind === "err") {
        notify(`Could not rename request: ${res.value}`);
        return;
      }

      this.requests[newID] = Object.assign({}, this.requests[id]);
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
>(request_id: string): UseRequest<Request, Response> {
  const notify = notification.error;

  store.requests2[request_id] = store.requests2[request_id] ?? {
    request: null,
    history: [],
    response: null,
    is_loading: true,
    send: async () => {
      if (store.requests2[request_id].request === null || store.requests2[request_id].is_loading) return;

      store.requests2[request_id].is_loading = true;
      const res = await api.requestPerform(request_id);
      store.requests2[request_id].is_loading = false;
      if (res.kind === "err") {
        notify(`Could not perform request ${request_id}: ${res.value}`);
        return;
      }

      store.requests2[request_id].history.push(res.value);
      store.requests2[request_id].response = res.value.response as Response;
    },
    update_request: async (patch: Partial<Request>) => {
      if (store.requests2[request_id].request === null || store.requests2[request_id].is_loading) return;

      store.requests2[request_id].is_loading = true;
      const old_request = store.requests2[request_id].request;
      const new_request = {...store.requests2[request_id].request, ...patch} as RequestData;
      store.requests2[request_id].request = new_request as Request; // NOTE: optimistic update
      const res = await api.request_update(request_id, new_request.kind, new_request);
      store.requests2[request_id].is_loading = false;
      if (res.kind === "err") {
        store.requests2[request_id].request = old_request; // NOTE: undo change
        notify(`Could not save current request: ${res.value}`);
        return;
      }
    },
  };
  const fetchData = async () => {
    store.requests2[request_id].is_loading = true;
    const res = await api.get(request_id);
    store.requests2[request_id].is_loading = false;
    if (res.kind === "err") {
      notify("load request", request_id, res.value);
      return;
    }

    store.requests2[request_id].request = res.value.Request as Request;
    store.requests2[request_id].history = res.value.History as unknown as HistoryEntry[];
    store.requests2[request_id].response = store.requests2[request_id].history[store.requests2[request_id].history.length - 1]?.response as Response ?? null;
  };

  if (store.requests2[request_id].is_loading) {
    fetchData().then(m.redraw);
  }
  // const stopWatch = watch(() => request_id, fetchData, {immediate: true});
  // onUnmounted(() => {
  //   stopWatch();
  // });

  return store.requests2[request_id];
}

export const store = useStore();
(async () => {
  await store.fetch()
  const oldTabs = localStorage.getItem("tabs");
  if (oldTabs !== null) {
    for (const id of JSON.parse(oldTabs)) {
      store.selectRequest(id);
    }
  }

  if (location.hash !== "#") {
    const id = decodeURI((location.hash ?? "").slice(1)); // remove '#'
    if (store.requests[id] === undefined) {
      location.hash = "";
      return;
    }

    store.selectRequest(id);
  }
})().then(m.redraw);
