import m, {Vnode} from "mithril";
import {NEmpty, NList} from "./dataview";
import {NInput} from "./input";
import {Modal} from "./layout";

type CommandDialogProps = {
  visible: boolean,
  theme: "algolia",
  on: {
    close: () => void
  }
  header: Vnode<any, any>,
  body: Vnode<any, any>,
  footer: Vnode<any, any>,
};
const CommandDialog = {
  view(vnode: Vnode<CommandDialogProps, any>) {
    const props = vnode.attrs;
    return m(Modal, {
      show: props.visible,
      title: "COMMAND PALETTE",
      content: "",
      buttons: [{id: "", text: ""}],
      on: {
        close: props.on.close,
      },
    }, [
      m("div", {class: "command-palette"}, [
        m("div", {class: "command-palette-header"}, props.header),
        m("div", {class: "command-palette-body"}, props.body),
        m("div", {class: "command-palette-footer"}, props.footer),
      ])
    ]);
  },
};

const CommandGroup = {
  view(vnode: Vnode<any, any>) {
    const props = vnode.attrs;
    return m("div", (vnode.children as any[]).flatMap(v => [v, m("hr")]));
  }
};

type ItemProps = {
  value: string,
  shortcut?: string[],
  on: {select?: () => void},
};
const CommandItem = {
  view(vnode: Vnode<ItemProps, any>) {
    const props = vnode.attrs;
    return m("div", [
      m("div", {class: "command-palette-item"}, [
        // ...vnode.children,
        m("span", {class: "command-palette-item-value"}, props.value),
        m("span", {class: "command-palette-item-shortcut"}, props.shortcut),
      ]),
      m("button", {onclick: props.on.select}, "Select"),
    ]);
  },
};

export default {
  Input: NInput,
  Dialog: CommandDialog,
  List: NList,
  Empty: NEmpty,
  Group: CommandGroup,
  Item: CommandItem,
};