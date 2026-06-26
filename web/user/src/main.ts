import { createApp, computed, reactive, ref } from "vue";
import {
  ApiClient,
  type ApiKey,
  type AuthToken,
  type CreateApiKeyResult,
  type GenerateImageResult,
  type ImageModel,
  type User,
  formatCredits,
  parseJSONList
} from "@agi-platform/shared";
import {
  ArrowDownToLine,
  Brush,
  Code2,
  Image,
  KeyRound,
  Loader2,
  LogOut,
  RefreshCw,
  Sparkles,
  UserCircle
} from "lucide-vue-next";
import "./styles.css";

const tokenKey = "agi_user_token";
const storedToken = localStorage.getItem(tokenKey);

const state = reactive({
  token: storedToken,
  user: null as User | null,
  models: [] as ImageModel[],
  apiKeys: [] as ApiKey[],
  result: null as GenerateImageResult | null,
  createdApiKey: "",
  loading: false,
  error: "",
  authMode: "login" as "login" | "register",
  activeView: "create" as "create" | "keys" | "docs"
});

const authForm = reactive({
  email: "user@example.com",
  password: "secret123",
  nickname: "User"
});

const generateForm = reactive({
  model: "general-high-quality",
  prompt: "一张科技感十足的 AI 芯片海报，蓝黑色背景，电影光效",
  negative_prompt: "",
  size: "1024x1024",
  n: 1
});

const keyForm = reactive({
  name: "default"
});

const api = new ApiClient({
  getToken: () => state.token,
  onUnauthorized: () => logout()
});

async function bootstrap() {
  await loadModels();
  if (state.token) {
    await Promise.allSettled([loadMe(), loadApiKeys()]);
  }
}

async function loadModels() {
  state.models = await api.get<ImageModel[]>("/api/models");
  if (state.models.length && !state.models.find((model) => model.code === generateForm.model)) {
    generateForm.model = state.models[0].code;
  }
}

async function loadMe() {
  state.user = await api.get<User>("/api/me");
}

async function loadApiKeys() {
  state.apiKeys = await api.get<ApiKey[]>("/api/api-keys");
}

async function submitAuth() {
  await withLoading(async () => {
    const path = state.authMode === "login" ? "/api/auth/login" : "/api/auth/register";
    const payload =
      state.authMode === "login"
        ? { email: authForm.email, password: authForm.password }
        : { email: authForm.email, password: authForm.password, nickname: authForm.nickname };
    const result = await api.post<{ user: User; token: AuthToken }>(path, payload);
    state.token = result.token.access_token;
    state.user = result.user;
    localStorage.setItem(tokenKey, state.token);
    await loadApiKeys();
  });
}

async function generateImage() {
  await withLoading(async () => {
    state.result = await api.post<GenerateImageResult>("/api/images/generate", {
      model: generateForm.model,
      prompt: generateForm.prompt,
      negative_prompt: generateForm.negative_prompt,
      size: generateForm.size,
      n: Number(generateForm.n)
    });
    await loadMe();
  });
}

async function createApiKey() {
  await withLoading(async () => {
    const result = await api.post<CreateApiKeyResult>("/api/api-keys", { name: keyForm.name });
    state.createdApiKey = result.plain;
    await loadApiKeys();
  });
}

async function revokeApiKey(id: number) {
  await withLoading(async () => {
    await api.delete(`/api/api-keys/${id}`);
    await loadApiKeys();
  });
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
  state.user = null;
  state.apiKeys = [];
  localStorage.removeItem(tokenKey);
}

const selectedModel = computed(() => state.models.find((model) => model.code === generateForm.model));

const App = {
  components: {
    ArrowDownToLine,
    Brush,
    Code2,
    Image,
    KeyRound,
    Loader2,
    LogOut,
    RefreshCw,
    Sparkles,
    UserCircle
  },
  setup() {
    bootstrap();
    return {
      state,
      authForm,
      generateForm,
      keyForm,
      selectedModel,
      formatCredits,
      parseJSONList,
      submitAuth,
      generateImage,
      createApiKey,
      revokeApiKey,
      logout
    };
  },
  template: `
    <main class="app-shell">
      <aside class="sidebar">
        <div class="brand">
          <div class="brand-mark"><Sparkles :size="20" /></div>
          <div>
            <strong>AGI Platform</strong>
            <span>AI Image Console</span>
          </div>
        </div>

        <button class="nav-item" :class="{ active: state.activeView === 'create' }" @click="state.activeView = 'create'">
          <Brush :size="18" /> AI 画图
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'keys' }" @click="state.activeView = 'keys'">
          <KeyRound :size="18" /> API Key
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'docs' }" @click="state.activeView = 'docs'">
          <Code2 :size="18" /> API 接入
        </button>

        <div class="account" v-if="state.user">
          <UserCircle :size="28" />
          <div>
            <strong>{{ state.user.nickname }}</strong>
            <span>{{ formatCredits(state.user.credits) }}</span>
          </div>
          <button class="icon-button" @click="logout"><LogOut :size="16" /></button>
        </div>
      </aside>

      <section class="content">
        <div v-if="!state.user" class="auth-panel">
          <div>
            <h1>AGI Platform</h1>
            <p>登录后开始使用多模型 AI 生图和开发者 API。</p>
          </div>
          <div class="form-grid">
            <div class="segmented">
              <button :class="{ active: state.authMode === 'login' }" @click="state.authMode = 'login'">登录</button>
              <button :class="{ active: state.authMode === 'register' }" @click="state.authMode = 'register'">注册</button>
            </div>
            <label>邮箱<input v-model="authForm.email" /></label>
            <label>密码<input v-model="authForm.password" type="password" /></label>
            <label v-if="state.authMode === 'register'">昵称<input v-model="authForm.nickname" /></label>
            <button class="primary-button" @click="submitAuth" :disabled="state.loading">
              <Loader2 v-if="state.loading" class="spin" :size="17" /> {{ state.authMode === 'login' ? '登录' : '创建账号' }}
            </button>
          </div>
        </div>

        <template v-else>
          <header class="topbar">
            <div>
              <h1>{{ state.activeView === 'create' ? 'AI 画图' : state.activeView === 'keys' ? 'API Key' : 'API 接入' }}</h1>
              <p>{{ state.user.email }} · {{ formatCredits(state.user.credits) }}</p>
            </div>
            <button class="ghost-button" @click="loadMe"><RefreshCw :size="16" />刷新余额</button>
          </header>

          <p v-if="state.error" class="error">{{ state.error }}</p>

          <div v-if="state.activeView === 'create'" class="workspace">
            <section class="panel controls">
              <label>模型
                <select v-model="generateForm.model">
                  <option v-for="model in state.models" :key="model.id" :value="model.code">
                    {{ model.display_name }} · {{ model.price_credits }} 积分/张
                  </option>
                </select>
              </label>
              <div class="model-hint" v-if="selectedModel">
                <strong>{{ selectedModel.display_name }}</strong>
                <span>{{ selectedModel.description }}</span>
              </div>
              <label>提示词
                <textarea v-model="generateForm.prompt" rows="7" />
              </label>
              <label>反向提示词
                <input v-model="generateForm.negative_prompt" />
              </label>
              <div class="inline-grid">
                <label>尺寸
                  <select v-model="generateForm.size">
                    <option v-for="size in parseJSONList(selectedModel?.supported_sizes)" :key="size" :value="size">{{ size }}</option>
                    <option v-if="!parseJSONList(selectedModel?.supported_sizes).length" value="1024x1024">1024x1024</option>
                  </select>
                </label>
                <label>数量
                  <input v-model.number="generateForm.n" type="number" min="1" :max="selectedModel?.max_images_per_request || 4" />
                </label>
              </div>
              <button class="primary-button" @click="generateImage" :disabled="state.loading">
                <Loader2 v-if="state.loading" class="spin" :size="17" /> 生成图片
              </button>
            </section>

            <section class="panel result-panel">
              <div class="empty-result" v-if="!state.result">
                <Image :size="40" />
                <span>生成结果会显示在这里</span>
              </div>
              <div v-else>
                <div class="task-line">
                  <strong>{{ state.result.task.task_no }}</strong>
                  <span>{{ state.result.task.status }} · 消耗 {{ state.result.task.credits_used }} 积分</span>
                </div>
                <div class="image-grid">
                  <article v-for="image in state.result.images" :key="image.id" class="image-card">
                    <img :src="image.url" :alt="state.result.task.prompt" />
                    <a :href="image.url" target="_blank" rel="noreferrer"><ArrowDownToLine :size="16" /> 打开图片</a>
                  </article>
                </div>
              </div>
            </section>
          </div>

          <section v-if="state.activeView === 'keys'" class="panel list-panel">
            <div class="toolbar">
              <label>名称<input v-model="keyForm.name" /></label>
              <button class="primary-button" @click="createApiKey" :disabled="state.loading"><KeyRound :size="16" />创建</button>
            </div>
            <p v-if="state.createdApiKey" class="secret-line">{{ state.createdApiKey }}</p>
            <table>
              <thead><tr><th>名称</th><th>Prefix</th><th>状态</th><th>最近使用</th><th></th></tr></thead>
              <tbody>
                <tr v-for="key in state.apiKeys" :key="key.id">
                  <td>{{ key.name || '-' }}</td>
                  <td>{{ key.key_prefix }}</td>
                  <td>{{ key.status }}</td>
                  <td>{{ key.last_used_at || '-' }}</td>
                  <td><button class="danger-button" @click="revokeApiKey(key.id)">删除</button></td>
                </tr>
              </tbody>
            </table>
          </section>

          <section v-if="state.activeView === 'docs'" class="panel docs-panel">
            <h2>OpenAI 风格接口</h2>
            <pre>POST /v1/images/generations
Authorization: Bearer agi_xxx

{
  "model": "general-high-quality",
  "prompt": "一张科技感海报",
  "size": "1024x1024",
  "n": 1
}</pre>
          </section>
        </template>
      </section>
    </main>
  `
};

createApp(App).mount("#app");

