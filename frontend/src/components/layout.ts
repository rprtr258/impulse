import m, {ChildArrayOrPrimitive, Vnode} from "mithril";

type NTabsProps = {
  type: "card" | "line",
  value: string,
  size: "small",
  class?: string,
  style?: any,
  on: {
    update: (id: string) => void,
  },
  tabs: {
    id: string,
    name: ChildArrayOrPrimitive | Vnode<any, any>,
    style?: any,
    class?: string,
    disabled?: boolean,
    elem?: ChildArrayOrPrimitive | Vnode<any, any>,
  }[],
};
export function NTabs() {
  return {
    view(vnode: Vnode<NTabsProps, any>) {
      const props = vnode.attrs;
      const tabs = props.tabs;
      const value = tabs.findIndex(v => v.id == props.value);
      if (value === -1) {
        throw new Error(`Tab ${props.value} not found in ${tabs.map(v => v.id)}`);
      }

      let tab = tabs[value];
      let elem = tab.elem ?? null;

      return m("div", {
        class: props.class,
        style: props.style,
      },
        m("div", {class: "h100", style: tabStyles.container}, [
          m("div", {style: tabStyles.header}, tabs.map((v, j) =>
            m("button", {
              key: j,
              disabled: v.disabled,
              style:
                j == value ?
                tabStyles.tab.active :
                v.disabled ?
                tabStyles.tab.disabled :
                tabStyles.tab.inactive,
              onclick: () => {
                if (v.disabled || value == j) {
                  return;
                }
                props.on.update(tabs[j].id);
              },
            }, v.name)
          )),
          m("div", {
            style: {
              height: "100%",
              ...tabStyles.content,
              ...tab.style,
            },
            class: tab.class,
          }, elem)])
      );
    },
  };
}
const tabStyles = (() => {
  const base = {
      padding: "4px 5px",
      cursor: "pointer",
      border: "1px solid transparent",
      borderBottom: "none",
      borderRadius: "3px 3px 0 0",
      position: "relative",
  };

  return {
    container: {
      fontFamily: "Arial, sans-serif"
    },
    header: {
      display: "flex",
      borderBottom: "1px solid #303030",
    },
    tab: {
      disabled: {
        ...base,
        background: "gray",
        fontWeight: "italic",
      },
      inactive: {
        ...base,
        background: "#7068ab",
        borderColor: "#454566",
      },
      active: {
        ...base,
        background: "#ddd3f5",
        borderColor: "#656596",
        borderBottom: "3px solid white",
        fontWeight: "bold",
      },
    },
    content: {
      padding: "2px",
    },
  };
})();

export const NSpace = {
  view(vnode: Vnode<any, any>) {
    return m("div", vnode.attrs, vnode.children);
  },
};

function Overlay() {
  let dom : HTMLDivElement;
  let children : m.ChildArrayOrPrimitive | undefined;

  const OverlayContainer = {
    view() {return children},
  }

  return {
    oncreate(vnode: Vnode<any, any>) {
      children = vnode.children;
      // Append a container to the end of body
      dom = document.createElement("div");
      dom.className = "overlay";
      // dom.style = "place-self: center; position: fixed; z-index: 100;";
      document.body.appendChild(dom);
      m.mount(dom, OverlayContainer);
    },
    onremove() {
      m.mount(dom, null); // triggers modal children removal hooks
      document.body.removeChild(dom);
    },
    onbeforeupdate(vnode: Vnode<any, any>) {
      children = vnode.children;
    },
    view() {},
  };
};
type NModalProps = {
  show: boolean,
  preset: "dialog",
  title: string,
  text: {
    positive: string,
    negative: string,
  },
  on: {
    positive_click: () => void,
    negative_click: () => void,
    close: () => void, // TODO: call on close/outer click
  },
};
export const NModal = {
  view(vnode: Vnode<NModalProps, any>) {
    const props = vnode.attrs;
    // return m("dialog", vnode.children);
    return m(Modal, {
      show: props.show,
      title: props.title,
      content: vnode.children,
      buttons: [
        {id: 'positive', text: props.text.positive},
        {id: 'negative', text: props.text.negative},
      ],
      on: {close: (id: "positive" | "negative") => {
        switch (id) {
          case "positive":
            props.on.positive_click();
            break;
          case "negative":
            props.on.negative_click();
            break;
          default:
            // TODO: never called, see not on on.close prop
            props.on.close();
            break;
        }
      }},
    });
  },
};
type ModalProps = {
  show: boolean,
  title: m.Children,
  content: m.Children,
  buttons: {id: string, text: string}[],
  on: {
    close(id: string): void,
  },
};
export function Modal() {
  let clickedId: string;
  return {
    view({attrs: {show, title, content, buttons, on: {close: onClose}}}: Vnode<ModalProps, any>) {
      if (!show || clickedId != null) {
        // We need to allow the Overlay component execute its
        // exit animation. Because it is a child of this component,
        // it will not fire when this component is removed.
        // Instead, we need to remove it first before this component
        // goes away.
        // When a button is clicked, we omit the Overlay component
        // from this Modal component's next view render, which will
        // trigger Overlay's onbeforeremove hook.
        return null
      }
      return m(Overlay,
        {
          onremove() {
            // Wait for the overlay's removal animation to complete.
            // Then we fire our parent's callback, which will
            // presumably remove this Modal component.
            Promise.resolve().then(() => {
              onClose(clickedId);
              m.redraw();
            })
          },
        },
        m('.modal',
          m('h3', title),
          m('.modal-content', content),
          m('.modal-buttons',
            buttons.map(b =>
              m('button', {
                type: 'button',
                disabled: clickedId != null,
                onclick() {
                  clickedId = b.id
                }
              }, b.text)
            )
          ),
        ),
      );
    },
  };
}

export const NScrollbar = {
  view(vnode: Vnode<{trigger: "none"}, any>) {
    return m("div", vnode.children);
  }
}
