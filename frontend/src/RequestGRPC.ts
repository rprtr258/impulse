import m, {Vnode} from "mithril";
import {NInput, NButton, NInputGroup, NSelect} from "./components/input";
import {NTabs} from "./components/layout";
import {NTag, NTable, NEmpty} from "./components/dataview";
import {api, GRPCCodes} from "./api";
import {database} from '../wailsjs/go/models';
import EditorJSON from "./components/EditorJSON";
import ViewJSON from "./components/ViewJSON";
import ParamsList from "./components/ParamsList";
import {use_request, useNotification} from "./store";

type Request = {kind: database.Kind.GRPC} & database.GRPCRequest;
type Props = {
  id: string,
};

export default function() {
  const methods : {
    service: string,
    methods: string[],
  }[] = [];
  const loadingMethods = false;
  let requestTab = "tab-req-request";
  let responseTab = "tab-resp-body";
  return {
    view(vnode: Vnode<Props, any>) {
      const {id} = vnode.attrs;

      // {request, response, is_loading, update_request, send}
      const r = use_request<Request, database.GRPCResponse>(id);
      const notification = useNotification();

      // watch(() => request.value?.target, async () => {
      //   loadingMethods.value = true;
      //   const res = await api.grpcMethods(id);
      //   loadingMethods.value = false;
      //   if (res.kind === "err") {
      //     notification.error({title: "Error fetching GRPC methods", content: res.value});
      //     return;
      //   }
      //   methods.value = res.value;
      // }, {immediate: true});

      const selectOptions = methods.map(svc => ({
        type: "group" as const,
        label: svc.service,
        key: svc.service,
        children: svc.methods.map(method => ({
          label: method,
          value: svc.service + "." + method,
        })),
      }));

      function responseBadge(response: {code: number}): Vnode<any, any> {
        const code = response.code;
        return m(NTag, {
          type: (code === 0 ? "success" : "error") as "success" | "info" | "warning",
          size: "small",
          round: true,
        }, `${code ?? "N/A"} ${GRPCCodes[code as keyof typeof GRPCCodes]}`);
      }

      if (r.request === null)
        return m(NEmpty, {
          description: "Loading request...",
          class: "h100",
          style: {"justify-content": "center"},
        });

      const request = r.request;
      const update_request = (patch: Partial<Request>): void => {
        r.update_request(patch).then(m.redraw);
      };

      return m("div", {
        class: "h100",
        style: {display: "grid", "grid-template-columns": "1fr 1fr", "grid-template-rows": "34px 1fr", "grid-column-gap": ".5em"},
      }, [
        m(NInputGroup, {style: {"grid-column": "span 2"}}, [
          m(NSelect, {
            value: request.method,
            options: selectOptions,
            loading: loadingMethods,
            remote: true,
            style: {width: "10%", "min-width": "18em"},
            on: {update: (method: string) => update_request({method})},
          }),
          m(NInput, {
            placeholder: "Addr",
            value: request.target,
            on: {update: (target: string) => update_request({target})},
          }),
          m(NButton, {
            type: "primary",
            on: {click: r.send},
            disabled: r.is_loading,
          }, "Send"),
        ]),
        m(NTabs, {
          value: requestTab,
          type: "line",
          size: "small",
          class: "h100",
          on: {update: (id: string) => requestTab = id},
          tabs: [
            {
              id: "tab-req-request",
              name: "Request",
              class: "h100",
              elem: m(EditorJSON, {
                class: "h100",
                value: request.payload,
                on: {update: (payload: string) => update_request({payload})},
              }),
            },
            {
              id: "tab-req-headers",
              name: "Metadata",
              style: {display: "flex", "flex-direction": "column", flex: 1},
              elem: [
                m(ParamsList, {
                  value: request.metadata,
                  on: {update: (value: database.KV[]) => update_request({metadata: value.filter(({key, value}) => key !== "" || value !== "")})},
                }),
                // ...request.headers.map((obj, i) => m("div", {
                //   style: {display: "flex", "flex-direction": "row"},
                //   key: i,
                // }, [
                //   m(NInput, {type: "text", value: obj.key,   style: {flex: 1}}),
                //   m(NInput, {type: "text", value: obj.value, style: {flex: 1}}),
                // ])),
                // m("div", {
                //   style: {display: "flex", "flex-direction": "row"},
                // }, [
                //   m(NInput, {type: "text", ref: "key",   value: pending.key,   style: {flex: 1}),
                //   m(NInput, {type: "text", ref: "value", value: pending.value, style: {flex: 1}),
                // ]),
              ],
            },
          ],
        }),
        r.response === null ?
        m(NEmpty, {
          description: "Send request or choose one from history.",
          class: "h100",
          style: {"justify-content": "center"},
        }) :
        ((response: database.GRPCResponse) => m(NTabs, {
          type: "card",
          size: "small",
          style: {"overflow-y": "auto"},
          value: responseTab,
          on: {update: (id: string) => responseTab = id},
          tabs: [
            {
              id: "tab-resp-code",
              name: responseBadge(response),
              disabled: true,
            },
            {
              id: "tab-resp-body",
              name: "Body",
              style: {"overflow-y": "auto"},
              elem: m(ViewJSON, {value: response?.response}),
            },
            {
              id: "tab-resp-headers",
              name: "Metadata",
              style: {flex: 1},
              elem: m(NTable, {striped: true, size: "small", "single-column": true, "single-line": false}, [
                m("colgroup", [
                  m("col", {style: {width: "50%"}}),
                  m("col", {style: {width: "50%"}}),
                ]),
                m("thead", [
                  m("tr", [
                    m("th", "NAME"),
                    m("th", "VALUE"),
                  ]),
                ]),
                ...response.metadata.map(header => m("tr", {key: header.key}, [
                  m("td", header.key),
                  m("td", header.value),
                ])),
              ]),
            },
          ],
        }))(r.response),
      ]);
    },
  };
}

// <style lang="css" scoped>
// .n-tab-pane {
//   height: 100% !important;
// }
// </style>
