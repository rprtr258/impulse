import {err, ok, Result} from "./result";

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

export const GRPCCodes = {
  0: "OK",
  1: "CANCELLED",
  2: "UNKNOWN",
  3: "INVALID_ARGUMENT",
  4: "DEADLINE_EXCEEDED",
  5: "NOT_FOUND",
  6: "ALREADY_EXISTS",
  7: "PERMISSION_DENIED",
  8: "RESOURCE_EXHAUSTED",
  9: "FAILED_PRECONDITION",
  10: "ABORTED",
  11: "OUT_OF_RANGE",
  12: "UNIMPLEMENTED",
  13: "INTERNAL",
  14: "UNAVAILABLE",
  15: "DATA_LOSS",
  16: "UNAUTHENTICATED",
} as const;
export type GRPCCode = keyof typeof GRPCCodes;

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

export interface RequestGRPC {
  target: string,
  method: string,
  payload: string,
  metadata: Parameter[],
}

export interface ResponseGRPC {
  response: string,
  code: GRPCCode,
  metadata: Parameter[],
}

export interface RequestJQ {
  query: string,
  json: string,
}

export interface ResponseJQ {
  response: string[],
}

export type RequestData =
  | {kind: "http"} & RequestHTTP
  | {kind: "sql"} & RequestSQL
  | {kind: "grpc"} & RequestGRPC
  | {kind: "jq"} & RequestJQ;

export interface ResponseSQL {
  columns: string[],
  types: ("number" | "string" | "time" | "bool")[],
  rows: unknown[][],
};

export const Kinds = ["http", "sql", "grpc", "jq"] as const;
export type ResponseData =
  | {kind: "http"} & ResponseHTTP
  | {kind: "sql"} & ResponseSQL
  | {kind: "grpc"} & ResponseGRPC
  | {kind: "jq"} & ResponseJQ;

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
} | {
  kind: "grpc",
  request: RequestGRPC,
  response: ResponseGRPC,
} | {
  kind: "jq",
  request: RequestJQ,
  response: ResponseJQ,
})

export type Request = {
  id: string,
} & RequestData;

async function apiCall<T>(route: string, params: object): Promise<Result<T>> { // TODO: handle errors
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
  if (res.status !== 200) {
    const x = await res.json();
    return err(x.error);
  }
  return ok(await res.json());
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

function parseTime(s: string): Date {
  const d = new Date();
  d.setTime(Date.parse(s));
  return d;
};

export const api = {
  async collectionRequests(): Promise<Result<CollectionGetResponse>> {
    const y = await apiCall<CollectionGetResponse>("/list", {});
    return y.map(x => {
      for (const req of x.history) {
        req.sent_at = parseTime(req.sent_at as unknown as string);
      }
      x.history.sort((a, b) => b.sent_at.getTime() - a.sent_at.getTime());
      return x;
    });
  },

  async requestCreate(
    name: string,
    kind: RequestData["kind"],
  ): Promise<Result<void>> {
    return apiCall("/create", {
      id: name,
      kind: kind,
    });
  },

  async requestDuplicate(
    name: string,
  ): Promise<Result<void>> {
    return apiCall("/duplicate", {
      id: name,
    });
  },

  async requestUpdate(
    reqId: string,
    kind: RequestData["kind"],
    req: Omit<RequestData, "kind">,
    name: string | null = null,
  ): Promise<Result<void>> {
    return await apiCall("/update", {
      id: reqId,
      kind: kind,
      name: name ?? reqId,
      request: req,
    });
  },

  async requestPerform(
    reqId: string,
  ): Promise<Result<ResponseData>> {
    return await apiCall("/perform", {
      id: reqId,
    });
  },

  async requestDelete(
    reqId: string,
  ): Promise<Result<void>> {
    return await apiCall("/delete", {
      id: reqId,
    });
  },

  async jq(
    data: string,
    query: string,
  ): Promise<Result<string[]>> {
    return await apiCall<string[]>("/jq", {
      json: data,
      query: query,
    });
  },

  async grpcMethods(target: string): Promise<Result<{
    service: string,
    methods: string[],
  }[]>> {
    return await apiCall<{
      service: string,
      methods: string[],
    }[]>("/grpc/methods", {
      target: target,
    });
  },
};
