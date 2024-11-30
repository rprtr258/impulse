const _baseURL = "http://localhost:8090/api";

export interface Parameter {
  key: string;
  value: string;
}

export const Method = {
  GET: "GET",
  POST: "POST",
  PUT: "PUT",
  DELETE: "DELETE",
  PATCH: "PATCH",
  OPTIONS: "OPTIONS",
};
export type Method = keyof typeof Method;

export interface RequestHTTP {
  url: string;
  method: Method;
  body?: string;
  headers: Parameter[];
}

export const Database = {
  postgres: "PG",
  mysql: "MY",
  sqlite: "LITE",
  clickhouse: "CH",
};
export type Database = keyof typeof Database;

export interface RequestSQL {
  dsn: string;
  database: keyof typeof Database;
  query: string;
}

export interface ResponseHTTP {
  code: number;
  body: string;
  headers: Parameter[];
}

export type RequestData =
  | ({ kind: "http" } & RequestHTTP)
  | ({ kind: "sql" } & RequestSQL);

export interface ResponseSQL {
  columns: string[];
  types: ("number" | "string" | "time" | "bool")[];
  rows: unknown[][];
}

export type ResponseData =
  | ({ kind: "http" } & ResponseHTTP)
  | ({ kind: "sql" } & ResponseSQL);

type HistoryEntryCommon = {
  request_id: string;
  sent_at: Date;
  received_at: Date;
};
export type HistoryEntry = HistoryEntryCommon &
  (
    | {
        kind: "http";
        request: RequestHTTP;
        response: ResponseHTTP;
      }
    | {
        kind: "sql";
        request: RequestSQL;
        response: ResponseSQL;
      }
  );

export type Request = {
  id: string;
} & ((RequestHTTP & { kind: "http" }) | (RequestSQL & { kind: "sql" }));

interface Collection {
  id: string;
  name: string;
  request_ids: string[];
}

async function apiCall(route: string, params: object) {
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

interface CollectionGetResponse {
  requests: Request[];
  history: HistoryEntry[];
}

export const api = {
  async collectionCreate(name: string): Promise<{ id: string }> {
    return apiCall("/create", { name: name });
  },

  async collectionList(): Promise<Collection[]> {
    return apiCall("/list", {});
  },

  async collectionRequests(id: string): Promise<CollectionGetResponse> {
    const x: CollectionGetResponse = await apiCall("/read", { id: id });
    for (const req of x.history) {
      const d = new Date();
      d.setTime(Date.parse(req.sent_at as unknown as string));
      req.sent_at = d;
      // req.sent_at = new Date(req.sent_at as unknown as string);
    }
    x.history.sort((a, b) => b.sent_at.getTime() - a.sent_at.getTime());
    return x;
  },

  async requestCreate(name: string, kind: "http" | "sql"): Promise<unknown> {
    return apiCall("/requests/create", { name: name, kind: kind });
  },

  async requestUpdate(
    colId: string,
    reqId: string,
    kind: "http" | "sql",
    req: RequestHTTP | RequestSQL
  ): Promise<void> {
    // TODO: remove name
    return await apiCall("/requests/update", {
      id: colId,
      n: reqId,
      kind: kind,
      name: reqId,
      request: req,
    });
  },

  async requestPerform(colId: string, reqId: string): Promise<ResponseData> {
    return await apiCall("/requests/perform", { id: colId, n: reqId });
  },
};
