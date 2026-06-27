import { createApp, reactive, ref } from "vue";
import {
  ApiClient,
  type AdminUser,
  type AuthToken,
  type ImageModel,
  type ImageModelRoute,
  type ImageTask,
  type Provider,
  type ProviderKey,
  type User,
  formatCredits
} from "@agi-platform/shared";
import {
  Boxes,
  Database,
  Eye,
  EyeOff,
  KeyRound,
  Loader2,
  LogOut,
  RefreshCw,
  Route,
  Server,
  Users
} from "lucide-vue-next";
import "./styles.css";

const tokenKey = "agi_admin_token";

const state = reactive({
  token: localStorage.getItem(tokenKey),
  admin: null as AdminUser | null,
  users: [] as User[],
  providers: [] as Provider[],
  providerKeys: [] as ProviderKey[],
  imageModels: [] as ImageModel[],
  routes: [] as ImageModelRoute[],
  tasks: [] as ImageTask[],
  selectedProviderId: 1,
  selectedModelId: 1,
  activeView: "users" as "users" | "providers" | "models" | "tasks",
  loading: false,
  error: ""
});

const loginForm = reactive({ username: "admin", password: "admin123" });
const showPassword = ref(false);
const creditForm = reactive({ user_id: 1, amount: 100, remark: "manual bonus" });
const providerForm = reactive({
  id: 0,
  code: "openai",
  name: "OpenAI",
  type: "mock",
  base_url: "",
  enabled: true,
  timeout_seconds: 60,
  retry_count: 1,
  priority: 100,
  remark: ""
});
const providerKeyForm = reactive({ name: "default", api_key: "", status: "active", weight: 100 });
const modelForm = reactive({
  id: 0,
  code: "general-high-quality",
  display_name: "通用高质量",
  description: "适合通用图片、商品图和海报封面",
  cover_url: "",
  price_credits: 8,
  supported_sizes: "1024x1024,1024x1536,1536x1024",
  support_text_to_image: true,
  support_image_to_image: false,
  support_edit: false,
  max_images_per_request: 4,
  auto_refund_on_failure: true,
  enabled: true,
  recommended: true,
  sort_order: 10
});
const routeForm = reactive({
  id: 0,
  provider_id: 1,
  provider_model_name: "mock-image",
  enabled: true,
  priority: 100,
  weight: 100,
  extra_config: "{}"
});

const api = new ApiClient({
  getToken: () => state.token,
  onUnauthorized: () => logout()
});

async function bootstrap() {
  if (state.token) {
    await loadAll();
  }
}

async function login() {
  await withLoading(async () => {
    const result = await api.post<{ admin: AdminUser; token: AuthToken }>("/admin/auth/login", loginForm);
    state.token = result.token.access_token;
    state.admin = result.admin;
    localStorage.setItem(tokenKey, state.token);
    await loadAll();
  });
}

async function loadAll() {
  await Promise.all([loadMe(), loadUsers(), loadProviders(), loadModels(), loadTasks()]);
}

async function loadMe() {
  state.admin = await api.get<AdminUser>("/admin/me");
}

async function loadUsers() {
  state.users = await api.get<User[]>("/admin/users?limit=50");
}

async function loadProviders() {
  state.providers = await api.get<Provider[]>("/admin/providers?limit=50");
  if (state.providers.length && !state.providers.find((provider) => provider.id === state.selectedProviderId)) {
    state.selectedProviderId = state.providers[0].id;
  }
  await loadProviderKeys();
}

async function loadProviderKeys() {
  if (!state.selectedProviderId) return;
  state.providerKeys = await api.get<ProviderKey[]>(`/admin/providers/${state.selectedProviderId}/keys`);
}

async function loadModels() {
  state.imageModels = await api.get<ImageModel[]>("/admin/image-models?limit=50");
  if (state.imageModels.length && !state.imageModels.find((model) => model.id === state.selectedModelId)) {
    state.selectedModelId = state.imageModels[0].id;
  }
  await loadRoutes();
}

async function loadRoutes() {
  if (!state.selectedModelId) return;
  state.routes = await api.get<ImageModelRoute[]>(`/admin/image-models/${state.selectedModelId}/routes`);
}

async function loadTasks() {
  state.tasks = await api.get<ImageTask[]>("/admin/image-tasks?limit=50");
}

async function adjustCredits() {
  await withLoading(async () => {
    await api.post<User>(`/admin/users/${creditForm.user_id}/credits`, {
      amount: Number(creditForm.amount),
      remark: creditForm.remark
    });
    await loadUsers();
  });
}

async function saveProvider() {
  await withLoading(async () => {
    const payload = { ...providerForm };
    if (providerForm.id) {
      await api.put(`/admin/providers/${providerForm.id}`, payload);
    } else {
      await api.post("/admin/providers", payload);
    }
    resetProviderForm();
    await loadProviders();
  });
}

async function createProviderKey() {
  await withLoading(async () => {
    await api.post(`/admin/providers/${state.selectedProviderId}/keys`, providerKeyForm);
    providerKeyForm.api_key = "";
    await loadProviderKeys();
  });
}

async function deleteProviderKey(id: number) {
  await withLoading(async () => {
    await api.delete(`/admin/provider-keys/${id}`);
    await loadProviderKeys();
  });
}

async function saveModel() {
  await withLoading(async () => {
    const payload = {
      ...modelForm,
      supported_sizes: modelForm.supported_sizes.split(",").map((item) => item.trim()).filter(Boolean)
    };
    if (modelForm.id) {
      await api.put(`/admin/image-models/${modelForm.id}`, payload);
    } else {
      await api.post("/admin/image-models", payload);
    }
    resetModelForm();
    await loadModels();
  });
}

async function saveRoute() {
  await withLoading(async () => {
    const payload = {
      ...routeForm,
      extra_config: JSON.parse(routeForm.extra_config || "{}")
    };
    if (routeForm.id) {
      await api.put(`/admin/image-model-routes/${routeForm.id}`, payload);
    } else {
      await api.post(`/admin/image-models/${state.selectedModelId}/routes`, payload);
    }
    resetRouteForm();
    await loadRoutes();
  });
}

function editProvider(item: Provider) {
  Object.assign(providerForm, item);
}

function resetProviderForm() {
  Object.assign(providerForm, { id: 0, code: "", name: "", type: "mock", base_url: "", enabled: true, timeout_seconds: 60, retry_count: 1, priority: 100, remark: "" });
}

function editModel(item: ImageModel) {
  Object.assign(modelForm, {
    ...item,
    supported_sizes: Array.isArray(item.supported_sizes) ? item.supported_sizes.join(",") : "1024x1024"
  });
}

function resetModelForm() {
  Object.assign(modelForm, { id: 0, code: "", display_name: "", description: "", cover_url: "", price_credits: 1, supported_sizes: "1024x1024", support_text_to_image: true, support_image_to_image: false, support_edit: false, max_images_per_request: 1, auto_refund_on_failure: true, enabled: true, recommended: false, sort_order: 100 });
}

function editRoute(item: ImageModelRoute) {
  Object.assign(routeForm, {
    ...item,
    extra_config: JSON.stringify(item.extra_config ?? {}, null, 2)
  });
}

function resetRouteForm() {
  Object.assign(routeForm, { id: 0, provider_id: state.providers[0]?.id ?? 1, provider_model_name: "", enabled: true, priority: 100, weight: 100, extra_config: "{}" });
}

async function withLoading(task: () => Promise<void>) {
  state.loading = true;
  state.error = "";
  try {
    await task();
  } catch (error) {
    state.error = error instanceof Error ? error.message : "请求失败";
  } finally {
    state.loading = false;
  }
}

function logout() {
  state.token = null;
  state.admin = null;
  localStorage.removeItem(tokenKey);
}

const App = {
  components: { Boxes, Database, Eye, EyeOff, KeyRound, Loader2, LogOut, RefreshCw, Route, Server, Users },
  setup() {
    bootstrap();
    return {
      state,
      loginForm,
      showPassword,
      creditForm,
      providerForm,
      providerKeyForm,
      modelForm,
      routeForm,
      formatCredits,
      login,
      loadAll,
      loadProviderKeys,
      loadRoutes,
      adjustCredits,
      saveProvider,
      createProviderKey,
      deleteProviderKey,
      saveModel,
      saveRoute,
      editProvider,
      resetProviderForm,
      editModel,
      resetModelForm,
      editRoute,
      resetRouteForm,
      logout
    };
  },
  template: `
    <main class="admin-shell">
      <aside class="sidebar">
        <div class="brand"><Database :size="22" /><strong>AGI Admin</strong></div>
        <button class="nav-item" :class="{ active: state.activeView === 'users' }" @click="state.activeView = 'users'"><Users :size="18" />用户</button>
        <button class="nav-item" :class="{ active: state.activeView === 'providers' }" @click="state.activeView = 'providers'"><Server :size="18" />Provider</button>
        <button class="nav-item" :class="{ active: state.activeView === 'models' }" @click="state.activeView = 'models'"><Boxes :size="18" />模型</button>
        <button class="nav-item" :class="{ active: state.activeView === 'tasks' }" @click="state.activeView = 'tasks'"><Route :size="18" />任务</button>
        <button v-if="state.admin" class="nav-item logout" @click="logout"><LogOut :size="18" />退出</button>
      </aside>

      <section class="content">
        <section v-if="!state.admin" class="login-panel">
          <h1>管理后台</h1>
          <label>账号<input v-model="loginForm.username" /></label>
          <label>密码
            <span class="password-field">
              <input v-model="loginForm.password" :type="showPassword ? 'text' : 'password'" />
              <button
                class="password-toggle"
                type="button"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                @click="showPassword = !showPassword"
              >
                <EyeOff v-if="showPassword" :size="17" />
                <Eye v-else :size="17" />
              </button>
            </span>
          </label>
          <button class="primary-button" @click="login" :disabled="state.loading">
            <Loader2 v-if="state.loading" class="spin" :size="16" />登录
          </button>
          <p v-if="state.error" class="error">{{ state.error }}</p>
        </section>

        <template v-else>
          <header class="topbar">
            <div>
              <h1>{{ state.activeView === 'users' ? '用户管理' : state.activeView === 'providers' ? 'Provider 管理' : state.activeView === 'models' ? '模型管理' : '生成任务' }}</h1>
              <p>{{ state.admin.username }} · {{ state.admin.role }}</p>
            </div>
            <button class="ghost-button" @click="loadAll"><RefreshCw :size="16" />刷新</button>
          </header>
          <p v-if="state.error" class="error">{{ state.error }}</p>

          <section v-if="state.activeView === 'users'" class="grid-two">
            <div class="panel">
              <h2>手动加减积分</h2>
              <label>用户 ID<input v-model.number="creditForm.user_id" type="number" /></label>
              <label>积分变动<input v-model.number="creditForm.amount" type="number" /></label>
              <label>备注<input v-model="creditForm.remark" /></label>
              <button class="primary-button" @click="adjustCredits">提交</button>
            </div>
            <div class="panel wide">
              <h2>用户列表</h2>
              <table>
                <thead><tr><th>ID</th><th>邮箱</th><th>昵称</th><th>积分</th><th>状态</th></tr></thead>
                <tbody><tr v-for="user in state.users" :key="user.id"><td>{{ user.id }}</td><td>{{ user.email || '-' }}</td><td>{{ user.nickname }}</td><td>{{ formatCredits(user.credits) }}</td><td>{{ user.status }}</td></tr></tbody>
              </table>
            </div>
          </section>

          <section v-if="state.activeView === 'providers'" class="grid-two">
            <div class="panel">
              <h2>{{ providerForm.id ? '编辑 Provider' : '创建 Provider' }}</h2>
              <label>Code<input v-model="providerForm.code" /></label>
              <label>名称<input v-model="providerForm.name" /></label>
              <label>类型<input v-model="providerForm.type" /></label>
              <label>Base URL<input v-model="providerForm.base_url" /></label>
              <div class="inline"><label><input v-model="providerForm.enabled" type="checkbox" />启用</label><label>优先级<input v-model.number="providerForm.priority" type="number" /></label></div>
              <button class="primary-button" @click="saveProvider">保存</button>
              <button class="ghost-button" @click="resetProviderForm">重置</button>
            </div>
            <div class="panel wide">
              <h2>Provider 列表</h2>
              <table>
                <thead><tr><th>ID</th><th>Code</th><th>名称</th><th>类型</th><th>启用</th><th></th></tr></thead>
                <tbody><tr v-for="item in state.providers" :key="item.id"><td>{{ item.id }}</td><td>{{ item.code }}</td><td>{{ item.name }}</td><td>{{ item.type }}</td><td>{{ item.enabled ? '是' : '否' }}</td><td><button class="small-button" @click="editProvider(item)">编辑</button></td></tr></tbody>
              </table>
            </div>
            <div class="panel wide full">
              <h2>Provider Key</h2>
              <div class="toolbar">
                <select v-model.number="state.selectedProviderId" @change="loadProviderKeys"><option v-for="item in state.providers" :key="item.id" :value="item.id">{{ item.name }}</option></select>
                <input v-model="providerKeyForm.name" placeholder="名称" />
                <input v-model="providerKeyForm.api_key" placeholder="上游 API Key" />
                <button class="primary-button" @click="createProviderKey"><KeyRound :size="16" />添加</button>
              </div>
              <table><thead><tr><th>ID</th><th>名称</th><th>状态</th><th>权重</th><th>错误</th><th></th></tr></thead><tbody><tr v-for="key in state.providerKeys" :key="key.id"><td>{{ key.id }}</td><td>{{ key.name }}</td><td>{{ key.status }}</td><td>{{ key.weight }}</td><td>{{ key.last_error || '-' }}</td><td><button class="danger-button" @click="deleteProviderKey(key.id)">删除</button></td></tr></tbody></table>
            </div>
          </section>

          <section v-if="state.activeView === 'models'" class="grid-two">
            <div class="panel">
              <h2>{{ modelForm.id ? '编辑模型' : '创建模型' }}</h2>
              <label>Code<input v-model="modelForm.code" /></label>
              <label>显示名称<input v-model="modelForm.display_name" /></label>
              <label>描述<textarea v-model="modelForm.description" rows="3" /></label>
              <label>价格积分<input v-model.number="modelForm.price_credits" type="number" /></label>
              <label>支持尺寸<input v-model="modelForm.supported_sizes" /></label>
              <div class="inline"><label><input v-model="modelForm.enabled" type="checkbox" />启用</label><label><input v-model="modelForm.recommended" type="checkbox" />推荐</label></div>
              <button class="primary-button" @click="saveModel">保存</button>
              <button class="ghost-button" @click="resetModelForm">重置</button>
            </div>
            <div class="panel wide">
              <h2>模型列表</h2>
              <table><thead><tr><th>ID</th><th>Code</th><th>名称</th><th>价格</th><th>启用</th><th></th></tr></thead><tbody><tr v-for="item in state.imageModels" :key="item.id"><td>{{ item.id }}</td><td>{{ item.code }}</td><td>{{ item.display_name }}</td><td>{{ item.price_credits }}</td><td>{{ item.enabled ? '是' : '否' }}</td><td><button class="small-button" @click="editModel(item)">编辑</button></td></tr></tbody></table>
            </div>
            <div class="panel wide full">
              <h2>模型路由</h2>
              <div class="toolbar">
                <select v-model.number="state.selectedModelId" @change="loadRoutes"><option v-for="item in state.imageModels" :key="item.id" :value="item.id">{{ item.display_name }}</option></select>
                <select v-model.number="routeForm.provider_id"><option v-for="item in state.providers" :key="item.id" :value="item.id">{{ item.name }}</option></select>
                <input v-model="routeForm.provider_model_name" placeholder="上游模型名" />
                <button class="primary-button" @click="saveRoute">保存路由</button>
              </div>
              <table><thead><tr><th>ID</th><th>Provider</th><th>上游模型</th><th>启用</th><th>优先级</th><th></th></tr></thead><tbody><tr v-for="route in state.routes" :key="route.id"><td>{{ route.id }}</td><td>{{ route.provider_id }}</td><td>{{ route.provider_model_name }}</td><td>{{ route.enabled ? '是' : '否' }}</td><td>{{ route.priority }}</td><td><button class="small-button" @click="editRoute(route)">编辑</button></td></tr></tbody></table>
            </div>
          </section>

          <section v-if="state.activeView === 'tasks'" class="panel">
            <h2>生成任务</h2>
            <table><thead><tr><th>ID</th><th>任务号</th><th>用户</th><th>来源</th><th>状态</th><th>积分</th><th>提示词</th></tr></thead><tbody><tr v-for="task in state.tasks" :key="task.id"><td>{{ task.id }}</td><td>{{ task.task_no }}</td><td>{{ task.user_id }}</td><td>{{ task.source }}</td><td>{{ task.status }}</td><td>{{ task.credits_used }}</td><td class="prompt-cell">{{ task.prompt }}</td></tr></tbody></table>
          </section>
        </template>
      </section>
    </main>
  `
};

createApp(App).mount("#app");
