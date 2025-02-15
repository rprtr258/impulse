import {err, ok, Result} from "./result";
import * as App from "../wailsjs/go/app/App";
import {database, app} from '../wailsjs/go/models';

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
  headers: database.KV[],
}

export interface RequestGRPC {
  target: string,
  method: string,
  payload: string,
  metadata: database.KV[],
}

export interface ResponseGRPC {
  response: string,
  code: GRPCCode,
  metadata: database.KV[],
}

export interface RequestJQ {
  query: string,
  json: string,
}

export interface ResponseJQ {
  response: string[],
}

export type RequestData =
  | {kind: "http"} & database.HTTPRequest
  | {kind: "sql"} & RequestSQL
  | {kind: "grpc"} & RequestGRPC
  | {kind: "jq"} & RequestJQ;

export type Request = {
  id: string,
} & RequestData;

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
  request: database.HTTPRequest,
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
  async collectionRequests(): Promise<Result<app.ListResponse>> {
    const y = await wrap(() => App.List());
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
  ): Promise<Result<app.ResponseNewRequest>> {
    return await wrap(() => App.Create(name, kind));
  },

  async requestDuplicate(
    name: string,
  ): Promise<Result<void>> {
    return await wrap(() => App.Duplicate(name));
  },

  async requestUpdate(
    reqId: string,
    kind: RequestData["kind"],
    req: Omit<RequestData, "kind">,
    name: string | null = null,
  ): Promise<Result<void>> {
    return await wrap(() => App.Update(reqId, kind, name ?? reqId, req));
  },

  async requestPerform(
    reqId: string,
  ): Promise<Result<ResponseData>> {
    return await wrap(() => App.Perform(reqId)) as Result<ResponseData>;
  },

  async requestDelete(
    reqId: string,
  ): Promise<Result<void>> {
    return await wrap(() => App.Delete(reqId));
  },

  async jq(
    data: string,
    query: string,
  ): Promise<Result<string[]>> {
    return await wrap(() => App.JQ(data, query));
  },

  async grpcMethods(target: string): Promise<Result<app.grpcServiceMethods[]>> {
    return await wrap(() => App.GRPCMethods(target));
  },
};
