import type {Result} from "./result";
import {err, ok} from "./result";
import * as App from "../wailsjs/go/app/App";
import type {app} from '../wailsjs/go/models';
import {database} from '../wailsjs/go/models';

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

export const Database: Record<database.Database, string> = {
  [database.Database.POSTGRES]:   "PG",
  [database.Database.MYSQL]:      "MY",
  [database.Database.SQLITE]:     "LITE",
  [database.Database.CLICKHOUSE]: "CH",
} as const;
export type Database = keyof typeof Database;

export type RequestData =
  | {kind: database.Kind.HTTP} & database.HTTPRequest
  | {kind: database.Kind.SQL} & database.SQLRequest
  | {kind: database.Kind.GRPC} & database.GRPCRequest
  | {kind: database.Kind.JQ} & database.JQRequest
  | {kind: database.Kind.REDIS} & database.RedisRequest
;

export type Request = {
  id: string,
} & RequestData;

export const Kinds = Object.values(database.Kind);
export type ResponseData =
  | {kind: database.Kind.HTTP} & database.HTTPResponse
  | {kind: database.Kind.SQL} & database.SQLResponse
  | {kind: database.Kind.GRPC} & database.GRPCResponse
  | {kind: database.Kind.JQ} & database.JQResponse
  | {kind: database.Kind.REDIS} & database.RedisResponse
;

export type HistoryEntry = {
  RequestId: string,
  sent_at: Date,
  received_at: Date,
} & ({
  kind: database.Kind.HTTP,
  request: database.HTTPRequest,
  response: database.HTTPResponse,
} | {
  kind: database.Kind.SQL,
  request: database.SQLRequest,
  response: database.SQLResponse,
} | {
  kind: database.Kind.GRPC,
  request: database.GRPCRequest,
  response: database.GRPCResponse,
} | {
  kind: database.Kind.JQ,
  request: database.JQRequest,
  response: database.JQResponse,
} | {
  kind: database.Kind.REDIS,
  request: database.RedisRequest,
  response: database.RedisResponse,
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
    return err(String(e));
  }
}

export const api = {
  async collectionRequests(): Promise<Result<app.ListResponse>> {
    return await wrap(async () => App.List());
  },

  async get(id: string): Promise<Result<database.Request>> {
    return await wrap(async () => App.Get(id));
  },

  async history(id: string): Promise<Result<HistoryEntry[]>> {
    const y = await wrap(async () => App.History(id) as HistoryEntry[]);
    return y.map((x: HistoryEntry[]): HistoryEntry[] => {
      for (const req of x) {
        req.sent_at = parseTime(req.sent_at as unknown as string);
      }
      x.sort((a, b) => b.sent_at.getTime() - a.sent_at.getTime());
      return x;
    });
  },

  async requestCreate(
    name: string,
    kind: database.Kind,
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
    kind: database.Kind,
    req: Omit<RequestData, "kind">,
    name: string | null = null,
  ): Promise<Result<void>> {
    return await wrap(() => App.Update(reqId, kind, name ?? reqId, req));
  },

  async requestPerform(
    reqId: string,
  ): Promise<Result<void>> {
    return await wrap(() => App.Perform(reqId)) as Result<void>;
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
