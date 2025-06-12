import m, {Vnode} from "mithril";

type NInputProps = {
  value: string,
  on: {update: (value: string) => void},
};
export const NInput = {
  view(vnode: Vnode<any, any>) {
    const props = vnode.attrs as NInputProps;
    return m("input", {
      value: props.value,
      oninput: (e: any) => props.on.update(e.target.value),
    });
  }
}

export const NInputGroup = {
  view(vnode: Vnode<any, any>) {
    const props = vnode.attrs;
    return m("div", props, vnode.children);
  }
}

type NDropdownProps = {
  trigger: "hover",
  options: any[],
  on: {select: (key: string | number) => void},
};
export const NDropdown = {
  view(vnode: Vnode<NDropdownProps, any>) {
    return m("span", vnode.children);
  }
};

// TODO: make generic over value type
type SelectOption = {
  label: string,
  value: string,
};
type NSelectProps = {
  value: string,
  options: (SelectOption | {
    type: "group",
    label: string,
    key: string,
    children: SelectOption[],
  })[],
  placeholder?: string,
  clearable?: true,
  style?: any,
  loading?: boolean,
  remote?: true,
  on: {update: (value: string) => void},
};
export function NSelect() {
  return {
    open: false,
    current: null as number | null,
    view(vnode: Vnode<NSelectProps, any>) {
      const props = vnode.attrs;
      // TODO: use groups
      const options = props.options.filter(v => !("type" in v)).map(v => v as SelectOption);

      const idx = options.findIndex(v => v.value == props.value);
      if (idx != -1) {
        this.current = idx;
      }

      return m("select", {
        style: props.style,
        onchange: (e: InputEvent) => {
          const i = parseInt((e.target! as HTMLSelectElement).value);
          const value = options[i].value;
          this.current = i;
          props.on.update(value);
        },
      }, [m("option", {
        value: "",
        disabled: true,
        selected: this.current === null,
        hidden: true,
      }, props.placeholder)].concat(props.options.map(({label}, i) =>
        m("option", {
          value: i,
          selected: i == this.current,
        }, label)
      )));
    }
  };
}

export const NButton = {
  view(vnode: Vnode<{
    type?: "primary",
    disabled?: boolean,
    class?: string,
    style?: any,
    on: {click: () => void},
  }, any>) {
    const props = vnode.attrs;
    return m("button", {
      class: props.class,
      style: props.style,
      onclick: props.on.click,
    }, vnode.children);
  },
};