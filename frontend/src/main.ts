import {createApp} from "vue";
import * as monaco from "monaco-editor";
import App from "./App.vue";
import {theme} from "./editor";

monaco.editor.defineTheme("material-ocean", theme);
createApp(App)
  .mount("#app");
