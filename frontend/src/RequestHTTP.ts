import m from "mithril";
import {database} from '../wailsjs/go/models';
import {Method as Methods} from "./api";
import {VNodeChild} from "./components";
import {NInputGroup, NInput, NSelect, NButton} from "./components/input";
import {NTabs} from "./components/layout";
import {NTag, NTable, NEmpty} from "./components/dataview";
import EditorJSON from "./components/EditorJSON";
import ViewJSON from "./components/ViewJSON";
import ParamsList from "./components/ParamsList";
import {use_request} from "./store";
import Mithril from "mithril";

type Request = {kind: database.Kind.HTTP} & database.HTTPRequest;

// function responseBodyLanguage(contentType: string): string {
//   for (const [key, value] of Object.entries({
//     "application/json;": "json",
//     "text/html;": "html",
//   })) {
//     if (contentType.startsWith(key)) {
//       return value;
//     }
//   }
//   return "text";
// };

function responseBadge(response?: any): VNodeChild {
  const code = response!.code;
  return m(NTag, {
    type: (
      code < 300 ? "success" :
      code < 500 ? "warning" :
                   "error"
    ) as "success" | "info" | "warning",
    size: "small",
    round: true,
  }, code ?? "N/A");
}

export default function(id: string): Mithril.ComponentTypes<any, any> {
  let requestTab = "tab-req-request";
  let responseTab = "tab-resp-body";
  return {
    view() {
      // const {id} = vnode.attrs;
      const r = use_request<Request, database.HTTPResponse>(id);
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
        style: {
          display: "grid",
          "grid-template-columns": "1fr 1fr",
          "grid-template-rows": "auto 1fr",
          "grid-column-gap": ".5em",
        },
      }, [
        m(NInputGroup, {style: {
          "grid-column": "span 2",
          display: "grid",
          "grid-template-columns": "1fr 10fr 1fr",
        }}, [
          m(NSelect, {
            value: r.request.method,
            options: Object.keys(Methods).map(method => ({label: method, value: method})),
            on: {update: (method: string) => update_request({method})},
          }),
          m(NInput, {
            placeholder: "URL",
            value: r.request.url,
            on: {update: (url: string) => update_request({url})},
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
                // class: "h100",
                value: request.body ?? null,
                on: {update: (body: string) => update_request({body})},
              }),
            },
            {
              name: "Headers",
              id: "tab-req-headers",
              style: {display: "flex", "flex-direction": "column", flex: 1},
              elem: m(ParamsList, {
                value: request.headers,
                on: {update: (value: database.KV[]): void => update_request({headers: value.filter(({key, value}) => key!=="" || value!=="")})},
              }),
            },
          ],
        }),
        r.response === null ?
        m(NEmpty, {
          description: "Send request or choose one from history.",
          class: "h100",
          style: {"justify-content": "center"},
        }) : ((response: database.HTTPResponse) => {
          return m(NTabs, {
            value: responseTab,
            type: "card",
            size: "small",
            style: {"overflow-y": "auto"},
            on: {update: (id) => responseTab = id},
            tabs: [
              {
                id: "tab-resp-code",
                name: responseBadge(r.response),
                disabled: true,
              },
              {
                id: "tab-resp-body",
                name: "Body",
                style: {"overflow-y": "auto"},
                elem: m(ViewJSON, {value: response.body}),
              },
              {
                id: "tab-resp-headers",
                name: "Headers",
                style: {flex: 1},
                elem: m(NTable, {
                  striped: true,
                  size: "small",
                  "single-column": true,
                  "single-line": false,
                }, [
                  m("colgroup", {key: "__colgroup"}, [
                    m("col", {style: {width: "50%"}}),
                    m("col", {style: {width: "50%"}}),
                  ]),
                  m("thead", {key: "__head"}, [
                    m("tr", [
                      m("th", "NAME"),
                      m("th", "VALUE"),
                    ]),
                  ]),
                  ...response.headers.map(header =>
                    m("tr", {key: header.key}, [
                      m("td", header.key),
                      m("td", header.value),
                    ])),
                ])
              },
            ],
          });
        })(r.response),
      ]);
    },
  };
}
