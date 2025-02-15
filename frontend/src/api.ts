import {err, ok, Result} from "./result";
import * as app from "../wailsjs/go/service/App";
import {database, service} from '../wailsjs/go/models';

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

export type RequestHTTP = database.HTTPRequest;

export const Database = {
  postgres:   "PG",
  mysql:      "MY",
  sqlite:     "LITE",
  clickhouse: "CH",
} as const;
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
  RequestId: string,
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
  try {
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
  } catch (e) {
    return err(`${e}`);
  }
}

function parseTime(s: string): Date {
  const d = new Date();
  d.setTime(Date.parse(s));
  return d;
};

async function wrap<T>(f: () => Promise<T>): Promise<Result<T>> {
  try {
    return ok(await f());
  } catch (e) {
    return err(`${e}`);
  }
}

export const api = {
  async collectionRequests(): Promise<Result<service.ListResponse>> {
    const y = await wrap(() => app.List());
    return y.map(x => {
      for (const req of x.History) {
        req.sent_at = parseTime(req.sent_at as unknown as string);
      }
      x.History.sort((a, b) => b.sent_at.getTime() - a.sent_at.getTime());
      return x;
    });
  },

  async requestCreate(
    name: string,
    kind: RequestData["kind"],
  ): Promise<Result<service.ResponseNewRequest>> {
    return await wrap(() => app.Create(name, kind));
  },

  async requestDuplicate(
    name: string,
  ): Promise<Result<void>> {
    return await wrap(() => app.Duplicate(name));
  },

  async requestUpdate(
    reqId: string,
    kind: RequestData["kind"],
    req: Omit<RequestData, "kind">,
    name: string | null = null,
  ): Promise<Result<void>> {
    return await wrap(() => app.Update(reqId, kind, name ?? reqId, req));
  },

  async requestPerform(
    reqId: string,
  ): Promise<Result<ResponseData>> {
    return await wrap(() => app.Perform(reqId)) as Result<ResponseData>;
  },

  async requestDelete(
    reqId: string,
  ): Promise<Result<void>> {
    return await wrap(() => app.Delete(reqId));
  },

  async jq(
    data: string,
    query: string,
  ): Promise<Result<string[]>> {
    return await wrap(() => app.JQ(data, query));
  },

  async grpcMethods(target: string): Promise<Result<service.grpcServiceMethods[]>> {
    return await wrap(() => app.GRPCMethods(target));
  },
};
