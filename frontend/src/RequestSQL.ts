import m, {Vnode} from "mithril";
// import {computed, toRefs, h, ref} from "vue";
// import {
//   NButton, NInput, NInputGroup, NSelect,
//   NLayout,
//   NScrollbar, NSplit,
//   NDataTable, NEmpty, NIcon, NTooltip,
// } from "naive-ui";
import {NScrollbar, NSpace} from "./components/layout";
import {Json, NEmpty, NIcon, NTooltip} from "./components/dataview";
import {NInputGroup} from "./components/input";
// import {
//   CheckSquareOutlined, ClockCircleOutlined,
//   FieldNumberOutlined, ItalicOutlined, QuestionCircleOutlined,
// } from "@vicons/antd"
import {database} from "../wailsjs/go/models";
// import {Database} from "./api";
import EditorSQL from "./EditorSQL";
import {use_request} from "./store";

type Request = {kind: database.Kind.SQL} & database.SQLRequest;
type TableBaseColumn = any;

type Props = {
  id: string,
}

const NSplit = {
  view(vnode: Vnode<any, any>) {
    return m("div", vnode.attrs, vnode.children);
  },
}

const NDataTable = {
  view(vnode: Vnode<any, any>) {
    return m(Json, vnode.attrs);
  },
}

export default function(id: string) {
  return {
    view(vnode: Vnode<Props, any>) {
      // const {id} = vnode.attrs;

      // {request, response, is_loading, send} =
      const r = use_request<Request, database.SQLResponse>(id);

      const update_request = (patch: Partial<Request>): void => {
        r.update_request(patch).then(m.redraw);
      };

      const columns = (() => {
        const resp = r.response;
        if (resp === null) {
          return [];
        }

        return (resp.columns ?? []).map(c => {
          return {
            key: c,
            title: (_: TableBaseColumn) => {
              const type = resp.types[resp.columns.indexOf(c)];
              return m(NTooltip, {
                trigger: "hover",
                placement: "bottom-start",
                ontrigger: () => m("div", [
                  m(NIcon, {
                    size: 15,
                    color: "grey",
                    component:
                      type === "number" ? "FieldNumberOutlined" :
                      type === "string" ? "ItalicOutlined" :
                      type === "bool" ? "CheckSquareOutlined" :
                      type === "time" ? "ClockCircleOutlined" :
                      "QuestionCircleOutlined",
                  }),
                  c,
                ]),
                default: () => type,
              });
            },
            render: (rowData: any, rowIndex: number) => {
              const v = rowData[c];
              switch (true) {
              case v === null:
                return "(NULL)"; // TODO: faded
              case typeof v == "boolean":
                return v ? "true" : "false";
              case typeof v == "number" || typeof v == "string":
                return v;
              default:
                console.log(rowData, rowData[c], rowIndex);
                return rowData[c];
              }
            },
          }
        });
      })();
      // TODO: fix duplicate column names
      const data = (() => {
        const resp = r.response;
        if (resp === null) {
          return [];
        }

        return (resp.rows ?? [])
          .map(row =>
            Object.fromEntries(row
              .map((v, i) => [resp.columns[i], v])));
      })();

      if (r.request === null)
        return m(NEmpty, {
          description: "Loading request...",
          class: "h100",
          style: {"justify-content": "center"},
        });

      return m(NSpace, {
        class: "h100",
        id: "gavno",
        header: m(NInputGroup, [
          // <NSelect
          //   :options="Object.keys(Database).map(db => ({label: Database[db as keyof typeof Database], value: db}))"
          //   :value="request.database"
          //   v-on:update:value="(database: Database) => update_request({database: database})"
          //   style="width: 10%;"
          // />
          // <NInput
          //   placeholder="DSN"
          //   :value="request.dsn"
          //   v-on:input="newValue => update_request({dsn: newValue})"
          // />
          // <NButton
          //   type="primary"
          //   v-on:click="send()"
          //   :disabled="is_loading"
          // >Run</NButton>
        ]),
        content: m(NSplit, {class: "h100", direction: "vertical"}, [ // TODO: class="h100" for content
          m(EditorSQL, {
            value: r.request.query,
            on: {update: (query: string) => update_request({query})},
            class: "h100",
          }),
          r.response === null ?
          m(NEmpty, {
            description: "Run query or choose one from history.",
            class: "h100",
            style: {"justify-content": "center"},
          }) :
          m(NScrollbar,
            m(NDataTable, {
              columns: columns,
              data: data,
              "single-line": false,
              size: "small",
              resizable: true,
              "scroll-x": r.response.columns.length * 200,
            })
          ),
        ]),
      });
    },
  };
};

// <style lang="css" scoped>
// .n-tab-pane {
//   height: 100% !important;
// }
// </style>
// <style lang="css">
// /* TODO: как же я ненавижу ебаный цсс блять господи за что */
// #gavno > .n-layout-scroll-container {
//   overflow: hidden;
//   height: 100%;
//   display: grid;
//   grid-template-rows: 34px 1fr;
// }
// </style>
