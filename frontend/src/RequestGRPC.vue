<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {NTag, NTabs, NTabPane, NInput, NButton, NTable, NInputGroup, NSelect, NDynamicInput, NEmpty, useNotification} from "naive-ui";
import {api, RequestGRPC, ResponseGRPC, GRPCCodes, Parameter} from "./api";
import EditorJSON from "./EditorJSON.vue";
import ViewJSON from "./ViewJSON.vue";
import ParamsList from "./ParamsList.vue";

const {request, response = null} = defineProps<{
  request: RequestGRPC,
  response: ResponseGRPC | null,
}>();
const emit = defineEmits<{
  send: [],
  update: [request: RequestGRPC],
}>();
function updateRequest(patch: Partial<RequestGRPC>) {
  emit("update", {...request, ...patch});
}

const methods = ref<{
  service: string,
  methods: string[],
}[]>([]);
const loadingMethods = ref(false);

const notification = useNotification();

watch(() => request.target, async () => {
  loadingMethods.value = true;
  const res = await api.grpcMethods(request.target);
  if (res.kind === "ok") {
    methods.value = res.value;
  } else {
    notification.error({title: "Failed to load methods", content: res.value});
  }
  loadingMethods.value = false;
}, {immediate: true});

function updateMetadata(value: Parameter[]) {
  updateRequest({
    metadata: value.filter(({key, value}) => key !== "" || value !== ""),
  });
}

const selectOptions = computed(() => methods.value.map(svc => ({
  type: "group",
  label: svc.service,
  key: svc.service,
  children: svc.methods.map(method => ({
    label: method,
    value: svc.service + "." + method,
  })),
})));
</script>

<template>
<div class="h100" style="display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 34px 1fr; grid-column-gap: .5em;">
  <NInputGroup style="grid-column: span 2;">
    <NSelect
      :value="request.method"
      v-on:update:value="method => updateRequest({method: method})"
      :options="selectOptions"
      :loading="loadingMethods"
      remote
      style="width: 10%; min-width: 18em;"
    />
    <NInput
      placeholder="Addr"
      :value="request.target"
      v-on:update:value="target => updateRequest({target: target})"
    />
    <NButton type="primary" v-on:click='emit("send")'>Send</NButton>
  </NInputGroup>
  <NTabs
    type="card"
    size="small"
    class="h100"
  >
    <NTabPane
      name="tab-req-request"
      tab='Request'
      class="h100"
      display-directive="show"
    >
      <EditorJSON
        class="h100"
        :value="request.payload"
        v-on:update="value => updateRequest({payload: value})"
      />
    </NTabPane>
    <NTabPane
      name="tab-req-headers"
      tab='Metadata'
      style="display: flex; flex-direction: column; flex: 1;"
      display-directive="show"
    >
      <ParamsList
        :value="request.metadata"
        v-on:update="(value: Parameter[]) => updateMetadata(value)"
      />
      <!-- <div
        style="display: flex; flex-direction: row;"
        v-for="(obj, i) in request.headers"
        :key="i"
      >
        <NInput type="text" :value="obj.key" style="flex: 1;" />
        <NInput type="text" :value="obj.value" style="flex: 1;" />
      </div>
      <div
        style="display: flex; flex-direction: row;"
      >
        <NInput type="text" ref="key" :value="pending.key" style="flex: 1;" />
        <NInput type="text" ref="value" :value="pending.value" style="flex: 1;" />
      </div> -->
    </NTabPane>
  </NTabs>
  <template v-if="response === null">
    <NEmpty
      description="Send request or choose one from history."
      class="h100"
      style="justify-content: center;"
    />
  </template>
  <template v-else>
    <NTabs
      type="card"
      size="small"
      default-value="tab-resp-body"
      style="overflow-y: auto;"
    >
      <NTabPane
        name="tab-resp-code"
        disabled
        display-directive="show"
      ><template #tab>
        <NTag
          :type='response.code === 0 ? "success" : "error"'
          size="small"
          round
        >{{response.code /*TODO: as string*/ ?? "N/A"}} {{GRPCCodes[response.code]}}</Ntag>
      </template></NTabPane>
      <NTabPane
        name="tab-resp-body"
        tab="Body"
        style="overflow-y: auto;"
        display-directive="show"
      >
        <ViewJSON :value="response?.response"></ViewJSON>
      </NTabPane>
      <NTabPane
        name="tab-resp-headers"
        tab="Metadata"
        style="flex: 1;"
        display-directive="show"
      >
        <NTable striped size="small" single-column :single-line="false">
          <colgroup>
            <col style="width: 50%" />
            <col style="width: 50%" />
          </colgroup>
          <thead>
            <tr>
              <th>NAME</th>
              <th>VALUE</th>
            </tr>
          </thead>
          <tr v-for="header in response.metadata" :key="header.key">
            <td>{{header.key}}</td>
            <td>{{header.value}}</td>
          </tr>
        </NTable>
      </NTabPane>
    </NTabs>
  </template>
</div>
</template>

<style lang="css" scoped>
.n-tab-pane {
  height: 100% !important;
}
</style>
