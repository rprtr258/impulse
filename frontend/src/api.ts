const _baseURL = "http://localhost:8090/api";

export interface Parameter {
  key: string,
  value: string,
}

export const Method = {
  GET:     "GET",
  POST:    "POST",
  PUT:     "PUT",
  DELETE:  "DELETE",
  PATCH:   "PATCH",
  OPTIONS: "OPTIONS",
};
export type Method = keyof typeof Method;

export interface RequestHTTP {
  url: string,
  method: Method,
  body?: string,
  headers: Parameter[],
}

export const Database = {
  postgres:   "PG",
  mysql:      "MY",
  sqlite:     "LITE",
  clickhouse: "CH",
};
export type Database = keyof typeof Database;

export interface RequestSQL {
  dsn: string,
  database: keyof typeof Database,
  query: string
}

export interface ResponseHTTP {
  code: number,
  body: string,
  headers: Parameter[],
}

export type RequestData =
  | {kind: "http"} & RequestHTTP
  | {kind: "sql"} & RequestSQL;

export interface ResponseSQL {
  columns: string[],
  types: ("number" | "string" | "time" | "bool")[],
  rows: unknown[][],
}

export type ResponseData =
  | {kind: "http"} & ResponseHTTP
  | {kind: "sql"} & ResponseSQL;

export type HistoryEntry = {
  request_id: string,
  sent_at: Date,
  received_at: Date,
} & ({
  kind: "http",
  request: RequestHTTP,
  response: ResponseHTTP,
} | {
  kind: "sql",
  request: RequestSQL,
  response: ResponseSQL,
})

export type Request = {
  id: string,
} & (RequestHTTP & {kind: "http"} |
  RequestSQL & {kind: "sql"})

async function apiCall<T>(route: string, params: object): Promise<T/* | {message: string, error: string}*/> { // TODO: handle errors
  const res = await fetch(_baseURL, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      ROUTE: route,
      ...params,
    }),
  });
  const x = await res.json();
  return x;
}

export interface Tree {
  ids: string[] | null, // TODO: remove null
  dirs: Record<string, Tree> | null, // TODO: remove null
}

interface CollectionGetResponse {
  tree: Tree,
  requests: Record<string, Request>,
  history: HistoryEntry[],
}

export type Result<T> = {
  kind: "ok",
  value: T,
} | {
  kind: "err",
  value: string,
}

export const api = {
  async collectionRequests(): Promise<CollectionGetResponse> {
    const x: CollectionGetResponse = await apiCall("/list", {});
    for (const req of x.history) {
      const d = new Date();
      d.setTime(Date.parse(req.sent_at as unknown as string));
      req.sent_at = d;
      // req.sent_at = new Date(req.sent_at as unknown as string);
    }
    x.history.sort((a, b) => b.sent_at.getTime() - a.sent_at.getTime());
    return x;
  },

  async requestCreate(
    name: string,
    kind: "http" | "sql",
  ): Promise<void> {
    return apiCall("/create", {
      id: name,
      kind: kind,
    });
  },

  async requestUpdate(
    reqId: string,
    kind: "http" | "sql",
    req: RequestHTTP | RequestSQL,
    name: string | null = null,
  ): Promise<void> {
    return await apiCall("/update", {
      id: reqId,
      kind: kind,
      name: name ?? reqId,
      request: req,
    });
  },

  async requestPerform(
    reqId: string,
  ): Promise<ResponseData> {
    return await apiCall("/perform", {
      id: reqId,
    });
  },

  async requestDelete(
    reqId: string,
  ): Promise<void> {
    return await apiCall("/delete", {
      id: reqId,
    });
  },

  async jq(
    data: string,
    query: string,
  ): Promise<Result<string[]>> {
    const result = await apiCall<string[]>("/jq", {
      json: data,
      query: query,
    });
    if (result.error) {
      return {kind: "err", value: result.error};
    }
    return {kind: "ok", value: result};
  },
};
