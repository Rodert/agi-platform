import { createApp, computed, reactive, ref } from "vue";
import {
  ApiClient,
  type AdminUser,
  type AuthToken,
  type DatabaseTable as DBTable,
  type DatabaseTableData,
  type GenerateImageResult,
  type ImageModel,
  type ImageModelRoute,
  type ImageTask,
  type Provider,
  type ProviderKey,
  type User,
  type VideoModel,
  type VideoModelRoute,
  type VideoTask,
  type VideoTaskResult,
  type WalletLog,
  formatCredits,
  parseJSONList,
  resolveApiBaseURL
} from "@agi-platform/shared";
import {
  Database,
  Eye,
  EyeOff,
  KeyRound,
  Loader2,
  LogOut,
  ReceiptText,
  RefreshCw,
  Route,
  Server,
  ShieldCheck,
  UserCircle,
  Users
} from "lucide-vue-next";
import "./styles.css";

const tokenKey = "agi_admin_token";
const apiBaseURL = resolveApiBaseURL();
const defaultImageSupportedSizes = "1024x1024,2048x2048,1536x1024,1024x1536,3840x2160,2160x3840";
const defaultVideoAspectRatios = "9:16,16:9,1:1";
const defaultVideoSeconds = "5,10,15";

const state = reactive({
  token: localStorage.getItem(tokenKey),
  admin: null as AdminUser | null,
  users: [] as User[],
  providers: [] as Provider[],
  providerKeys: [] as ProviderKey[],
  imageModels: [] as ImageModel[],
  videoModels: [] as VideoModel[],
  tasks: [] as ImageTask[],
  videoTasks: [] as VideoTask[],
  databaseTables: [] as DBTable[],
  selectedDatabaseTable: "",
  databaseData: null as DatabaseTableData | null,
  selectedDatabaseRow: null as Record<string, unknown> | null,
  selectedDatabaseRowIndex: -1,
  databaseDDLOpen: false,
  databasePage: 1,
  databasePageSize: 20,
  taskPage: 1,
  taskPageSize: 20,
  selectedTask: null as GenerateImageResult | null,
  selectedVideoTask: null as VideoTaskResult | null,
  walletLogs: [] as WalletLog[],
  walletPage: 1,
  walletPageSize: 20,
  walletHasNext: false,
  profileMessage: "",
  upstreamModels: [] as UpstreamModel[],
  upstreamModelConfigs: [] as UpstreamModelConfig[],
  selectedProviderId: 1,
  activeView: "users" as "users" | "credits" | "providers" | "tasks" | "database" | "taskDetail" | "videoDetail" | "profile",
  userModalOpen: false,
  userModalMode: "create" as "create" | "edit",
  selectedUser: null as User | null,
  upstreamModalOpen: false,
  upstreamModalMode: "create" as "create" | "edit",
  copiedJSON: "",
  loading: false,
  queryingUpstreamModels: false,
  error: ""
});

const menuViews = ["users", "credits", "providers", "tasks", "database", "profile"] as const;
type MenuView = (typeof menuViews)[number];

type UpstreamModel = {
  id: string;
  object?: string;
};

type UpstreamModelConfig = {
  model_type: "image" | "video";
  provider_key_id?: number;
  provider_model_name: string;
  model_code: string;
  display_name: string;
  description: string;
  price_credits: number;
  supported_sizes: string;
  max_images_per_request: number;
  supported_aspect_ratios: string;
  supported_seconds: string;
  extra_config: string;
  priority: number;
  enabled: boolean;
  expanded: boolean;
};

const loginForm = reactive({ username: "", password: "" });
const showPassword = ref(false);
const userForm = reactive({
  id: 0,
  email: "",
  phone: "",
  password: "",
  nickname: "",
  avatar_url: "",
  status: "active"
});
const creditForm = reactive({ user_id: 0, mode: "add" as "add" | "deduct", amount: 100, remark: "manual adjustment" });
const taskFilterForm = reactive({
  id: "",
  keyword: "",
  status: ""
});
const upstreamForm = reactive({
  provider_code: "chongplus",
  provider_name: "ChongPlus",
  provider_type: "openai-compatible",
  base_url: "",
  api_key_name: "default",
  api_key: "",
  model_code: "gpt-image-2",
  display_name: "GPT Image 2",
  description: "OpenAI 兼容图片生成模型",
  provider_model_name: "gpt-image-2",
  price_credits: 8,
  supported_sizes: defaultImageSupportedSizes,
  max_images_per_request: 4,
  extra_config: "",
  timeout_seconds: 120,
  priority: 100,
  enabled: true
});
const providerForm = reactive({
  id: 0,
  code: "chongplus",
  name: "ChongPlus",
  type: "openai-compatible",
  base_url: "",
  enabled: true,
  timeout_seconds: 60,
  retry_count: 1,
  priority: 100,
  remark: ""
});
const providerKeyForm = reactive({ name: "default", api_key: "", status: "active", weight: 100 });
const profileForm = reactive({
  current_password: "",
  new_password: "",
  confirm_password: ""
});
const api = new ApiClient({
  getToken: () => state.token,
  onUnauthorized: () => logout()
});

const combinedTasks = computed(() =>
  [
    ...state.tasks.map((task) => ({
      kind: "image" as const,
      id: task.id,
      task_no: task.task_no,
      user_id: task.user_id,
      status: task.status,
      progress: task.progress,
      credits_used: task.credits_used,
      created_at: task.created_at,
      prompt: task.error_message || task.prompt,
      detail_url: taskDetailURL(task.task_no),
      action_label: "查看图片"
    })),
    ...state.videoTasks.map((task) => ({
      kind: "video" as const,
      id: task.id,
      task_no: task.task_no,
      user_id: task.user_id,
      status: task.status,
      progress: task.progress,
      credits_used: task.credits_used,
      created_at: task.created_at,
      prompt: task.error_message || task.prompt,
      detail_url: videoTaskDetailURL(task.task_no),
      action_label: "查看视频"
    }))
  ].sort((left, right) => new Date(right.created_at).getTime() - new Date(left.created_at).getTime())
);
const pagedTasks = computed(() => {
  const start = (state.taskPage - 1) * state.taskPageSize;
  return combinedTasks.value.slice(start, start + state.taskPageSize);
});
const taskHasNext = computed(() => combinedTasks.value.length > state.taskPage * state.taskPageSize);
const taskRangeText = computed(() => {
  if (!pagedTasks.value.length) return `第 ${state.taskPage} 页`;
  const start = (state.taskPage - 1) * state.taskPageSize + 1;
  return `${start}-${start + pagedTasks.value.length - 1}`;
});
const walletRangeText = computed(() => {
  if (!state.walletLogs.length) return `第 ${state.walletPage} 页`;
  const start = (state.walletPage - 1) * state.walletPageSize + 1;
  return `${start}-${start + state.walletLogs.length - 1}`;
});
const databaseRangeText = computed(() => {
  if (!state.databaseData?.rows.length) return `第 ${state.databasePage} 页`;
  const start = (state.databasePage - 1) * state.databasePageSize + 1;
  return `${start}-${start + state.databaseData.rows.length - 1}`;
});

async function bootstrap() {
  if (state.token) {
    await loadAll();
    await syncRouteFromURL();
  }
}

async function login() {
  await withLoading(async () => {
    const result = await api.post<{ admin: AdminUser; token: AuthToken }>("/admin/auth/login", loginForm);
    state.token = result.token.access_token;
    state.admin = result.admin;
    localStorage.setItem(tokenKey, state.token);
    await loadAll();
    await syncRouteFromURL();
  });
}

async function loadAll() {
  await Promise.all([loadMe(), loadUsers(), loadProviders(), loadModels(), loadVideoModels(), loadTasks(), loadVideoTasks(), loadWalletLogs(), loadDatabaseTables()]);
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
}

async function loadVideoModels() {
  state.videoModels = await api.get<VideoModel[]>("/admin/video-models?limit=50");
}

async function loadTasks() {
  state.tasks = await api.get<ImageTask[]>(`/admin/image-tasks?${taskQueryString()}`);
}

async function loadTaskDetail(taskNo: string) {
  state.selectedTask = await api.get<GenerateImageResult>(`/admin/image-tasks/${taskNo}`);
}

async function loadVideoTasks() {
  state.videoTasks = await api.get<VideoTask[]>(`/admin/video-tasks?${taskQueryString()}`);
}

async function loadVideoTaskDetail(taskNo: string) {
  state.selectedVideoTask = await api.get<VideoTaskResult>(`/admin/video-tasks/${taskNo}`);
}

async function refreshTasks() {
  state.taskPage = 1;
  await loadTaskPage();
}

async function loadTaskPage() {
  await Promise.all([loadTasks(), loadVideoTasks()]);
}

async function changeTaskPage(delta: number) {
  const nextPage = state.taskPage + delta;
  if (nextPage < 1 || (delta > 0 && !taskHasNext.value)) return;
  state.taskPage = nextPage;
  await loadTaskPage();
}

function resetTaskFilters() {
  Object.assign(taskFilterForm, {
    id: "",
    keyword: "",
    status: ""
  });
  void refreshTasks();
}

async function syncRouteFromURL() {
  const params = new URLSearchParams(window.location.search);
  const view = params.get("view");
  const taskNo = params.get("task_no");

  if (view === "task" && taskNo) {
    state.activeView = "taskDetail";
    state.selectedVideoTask = null;
    await loadTaskDetail(taskNo);
    return;
  }

  if (view === "video_task" && taskNo) {
    state.activeView = "videoDetail";
    state.selectedTask = null;
    await loadVideoTaskDetail(taskNo);
    return;
  }

  if (isMenuView(view)) {
    state.activeView = view;
    state.selectedTask = null;
    state.selectedVideoTask = null;
    if (view === "database") {
      await refreshDatabase();
    }
    return;
  }

  state.activeView = "users";
  state.selectedTask = null;
  state.selectedVideoTask = null;
}

async function loadWalletLogs() {
  const offset = (state.walletPage - 1) * state.walletPageSize;
  const logs = await api.get<WalletLog[]>(`/admin/wallet/logs?limit=${state.walletPageSize + 1}&offset=${offset}`);
  state.walletHasNext = logs.length > state.walletPageSize;
  state.walletLogs = logs.slice(0, state.walletPageSize);
}

async function refreshWalletLogs() {
  state.walletPage = 1;
  await loadWalletLogs();
}

async function changeWalletPage(delta: number) {
  const nextPage = state.walletPage + delta;
  if (nextPage < 1 || (delta > 0 && !state.walletHasNext)) return;
  state.walletPage = nextPage;
  await loadWalletLogs();
}

async function changePassword() {
  await withLoading(async () => {
    if (profileForm.new_password !== profileForm.confirm_password) {
      throw new Error("两次输入的新密码不一致");
    }
    await api.post("/admin/me/password", {
      current_password: profileForm.current_password,
      new_password: profileForm.new_password
    });
    Object.assign(profileForm, {
      current_password: "",
      new_password: "",
      confirm_password: ""
    });
    state.profileMessage = "密码已更新，下次登录请使用新密码";
  });
}

async function loadDatabaseTables() {
  state.databaseTables = await api.get<DBTable[]>("/admin/database/tables");
  if (!state.selectedDatabaseTable && state.databaseTables.length) {
    state.selectedDatabaseTable = state.databaseTables[0].name;
  }
}

async function openDatabaseTable(table: string) {
  state.selectedDatabaseTable = table;
  state.databasePage = 1;
  state.databaseDDLOpen = false;
  closeDatabaseRowDetail();
  await loadDatabaseTable();
}

async function loadDatabaseTable() {
  if (!state.selectedDatabaseTable) {
    state.databaseData = null;
    return;
  }
  const offset = (state.databasePage - 1) * state.databasePageSize;
  state.databaseData = await api.get<DatabaseTableData>(
    `/admin/database/tables/${encodeURIComponent(state.selectedDatabaseTable)}?limit=${state.databasePageSize}&offset=${offset}`
  );
  closeDatabaseRowDetail();
}

async function refreshDatabase() {
  await loadDatabaseTables();
  if (!state.selectedDatabaseTable && state.databaseTables.length) {
    state.selectedDatabaseTable = state.databaseTables[0].name;
  }
  await loadDatabaseTable();
}

async function changeDatabasePage(delta: number) {
  const nextPage = state.databasePage + delta;
  if (nextPage < 1 || (delta > 0 && !state.databaseData?.has_next)) return;
  state.databasePage = nextPage;
  await loadDatabaseTable();
}

function openDatabaseRowDetail(row: Record<string, unknown>, rowIndex: number) {
  state.selectedDatabaseRow = row;
  state.selectedDatabaseRowIndex = (state.databasePage - 1) * state.databasePageSize + rowIndex + 1;
}

function closeDatabaseRowDetail() {
  state.selectedDatabaseRow = null;
  state.selectedDatabaseRowIndex = -1;
}

function toggleDatabaseDDL() {
  state.databaseDDLOpen = !state.databaseDDLOpen;
}

async function saveUser() {
  await withLoading(async () => {
    const payload = {
      email: userForm.email.trim(),
      phone: userForm.phone.trim(),
      password: userForm.password,
      nickname: userForm.nickname.trim(),
      avatar_url: userForm.avatar_url.trim(),
      status: userForm.status
    };
    const user =
      state.userModalMode === "edit" && userForm.id
        ? await api.put<User>(`/admin/users/${userForm.id}`, payload)
        : await api.post<User>("/admin/users", payload);
    state.selectedUser = user;
    Object.assign(userForm, {
      id: user.id,
      email: user.email || "",
      phone: user.phone || "",
      password: "",
      nickname: user.nickname,
      avatar_url: user.avatar_url || "",
      status: user.status
    });
    state.userModalMode = "edit";
    creditForm.user_id = user.id;
    await loadUsers();
  });
}

async function adjustCredits() {
  await withLoading(async () => {
    const amount = Math.abs(Number(creditForm.amount));
    if (!amount) {
      throw new Error("请输入积分数量");
    }
    const user = await api.post<User>(`/admin/users/${creditForm.user_id}/credits`, {
      amount: creditForm.mode === "deduct" ? -amount : amount,
      remark: creditForm.remark
    });
    state.selectedUser = user;
    Object.assign(userForm, {
      id: user.id,
      email: user.email || "",
      phone: user.phone || "",
      password: "",
      nickname: user.nickname,
      avatar_url: user.avatar_url || "",
      status: user.status
    });
    Object.assign(creditForm, {
      user_id: user.id,
      mode: "add",
      amount: 100,
      remark: "manual adjustment"
    });
    await Promise.all([loadUsers(), loadWalletLogs()]);
  });
}

function openCreateUserModal() {
  state.userModalMode = "create";
  state.selectedUser = null;
  Object.assign(userForm, {
    id: 0,
    email: "",
    phone: "",
    password: "",
    nickname: "",
    avatar_url: "",
    status: "active"
  });
  Object.assign(creditForm, {
    user_id: 0,
    mode: "add",
    amount: 100,
    remark: "manual adjustment"
  });
  state.userModalOpen = true;
}

function openEditUserModal(user: User) {
  state.userModalMode = "edit";
  state.selectedUser = user;
  Object.assign(userForm, {
    id: user.id,
    email: user.email || "",
    phone: user.phone || "",
    password: "",
    nickname: user.nickname,
    avatar_url: user.avatar_url || "",
    status: user.status
  });
  Object.assign(creditForm, {
    user_id: user.id,
    mode: "add",
    amount: 100,
    remark: "manual adjustment"
  });
  state.userModalOpen = true;
}

function closeUserModal() {
  state.userModalOpen = false;
  state.selectedUser = null;
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

async function saveUpstreamIntegration() {
  await withLoading(async () => {
    const configs = state.upstreamModelConfigs.length ? state.upstreamModelConfigs : [modelConfigFromForm()];
    if (
      !upstreamForm.provider_code ||
      !upstreamForm.provider_name ||
      !upstreamForm.provider_type ||
      !upstreamForm.base_url ||
      !configs.length
    ) {
      throw new Error("请填写上游 API 和模型信息");
    }

    const providerPayload = {
      code: upstreamForm.provider_code.trim(),
      name: upstreamForm.provider_name.trim(),
      type: upstreamForm.provider_type,
      base_url: upstreamForm.base_url.trim(),
      enabled: upstreamForm.enabled,
      timeout_seconds: Number(upstreamForm.timeout_seconds) || 120,
      retry_count: 1,
      priority: Number(upstreamForm.priority) || 100,
      remark: ""
    };
    const existingProvider = state.providers.find((item) => item.code === providerPayload.code);
    const provider = existingProvider
      ? await updateAndReturnProvider(existingProvider, providerPayload)
      : await api.post<Provider>("/admin/providers", providerPayload);

    if (upstreamForm.api_key.trim()) {
      await api.post(`/admin/providers/${provider.id}/keys`, {
        name: upstreamForm.api_key_name || "default",
        api_key: upstreamForm.api_key.trim(),
        status: "active",
        weight: 100
      });
      upstreamForm.api_key = "";
    }

    state.selectedProviderId = provider.id;
    await loadProviderKeys();

    for (const config of configs) {
      if (!config.model_code.trim() || !config.display_name.trim() || !config.provider_model_name.trim()) {
        throw new Error("请检查模型名称、显示名称和上游模型名");
      }
      if (config.model_type === "video") {
        await saveVideoModelConfig(provider.id, config);
      } else {
        await saveImageModelConfig(provider.id, config);
      }
    }

    await Promise.all([loadProviders(), loadModels(), loadVideoModels()]);
    closeUpstreamModal();
  });
}

async function toggleProviderEnabled(provider: Provider) {
  await withLoading(async () => {
    await api.put(`/admin/providers/${provider.id}`, {
      code: provider.code,
      name: provider.name,
      type: provider.type,
      base_url: provider.base_url,
      enabled: !provider.enabled,
      timeout_seconds: provider.timeout_seconds,
      retry_count: provider.retry_count,
      priority: provider.priority,
      remark: provider.remark || ""
    });
    await loadProviders();
  });
}

async function openCreateUpstreamModal() {
  state.upstreamModalMode = "create";
  state.upstreamModels = [];
  state.upstreamModelConfigs = [];
  resetUpstreamForm();
  applyDefaultImageModel();
  state.upstreamModalOpen = true;
}

async function openEditUpstreamModal(provider: Provider) {
  state.upstreamModalMode = "edit";
  state.upstreamModels = [];
  state.upstreamModelConfigs = [];
  resetUpstreamForm();
  Object.assign(upstreamForm, {
    provider_code: provider.code,
    provider_name: provider.name,
    provider_type: provider.type,
    base_url: provider.base_url,
    api_key_name: "default",
    api_key: "",
    timeout_seconds: provider.timeout_seconds || 120,
    priority: provider.priority || 100,
    enabled: provider.enabled
  });

  state.upstreamModelConfigs = await configsForProvider(provider);
  if (state.upstreamModelConfigs.length) {
    const first = state.upstreamModelConfigs[0];
    Object.assign(upstreamForm, {
      provider_model_name: first.provider_model_name,
      model_code: first.model_code,
      display_name: first.display_name,
      description: first.description,
      price_credits: first.price_credits,
      supported_sizes: first.supported_sizes,
      max_images_per_request: first.max_images_per_request,
      enabled: provider.enabled && first.enabled
    });
  }
  state.upstreamModalOpen = true;
}

function closeUpstreamModal() {
  state.upstreamModalOpen = false;
  state.upstreamModelConfigs = [];
}

function resetUpstreamForm() {
  Object.assign(upstreamForm, {
    provider_code: "chongplus",
    provider_name: "ChongPlus",
    provider_type: "openai-compatible",
    base_url: "",
    api_key_name: "default",
    api_key: "",
    model_code: "gpt-image-2",
    display_name: "GPT Image 2",
    description: "OpenAI 兼容图片生成模型",
    provider_model_name: "gpt-image-2",
    price_credits: 8,
    supported_sizes: defaultImageSupportedSizes,
    max_images_per_request: 4,
    timeout_seconds: 120,
    priority: 100,
    enabled: true
  });
}

function applyDefaultImageModel() {
  Object.assign(upstreamForm, {
    model_code: "gpt-image-2",
    display_name: "GPT Image 2",
    description: "OpenAI 兼容图片生成模型",
    provider_model_name: "gpt-image-2",
    price_credits: 8,
    supported_sizes: defaultImageSupportedSizes,
    max_images_per_request: 4
  });
  state.upstreamModelConfigs = [modelConfigFromForm(true)];
}

function applyDefaultVideoModel() {
  Object.assign(upstreamForm, {
    model_code: "video-ds-2.0-fast",
    display_name: "AI 视频 Fast",
    description: "OpenAI 兼容视频生成模型",
    provider_model_name: "video-ds-2.0-fast",
    price_credits: 80,
    supported_sizes: defaultImageSupportedSizes,
    max_images_per_request: 1
  });
  state.upstreamModelConfigs = [modelConfigFromForm(true, "video")];
}

function applyDefaultGrokVideoModels() {
  Object.assign(upstreamForm, {
    provider_code: "grok-video",
    provider_name: "Grok Video",
    provider_type: "grok-video",
    base_url: upstreamForm.base_url || "https://api.119337.xyz"
  });
  state.upstreamModelConfigs = [
    modelConfigFromName("grok-image-video", true),
    modelConfigFromName("grok-video-1.5", false)
  ];
}

async function queryUpstreamModels() {
  state.queryingUpstreamModels = true;
  state.error = "";
  try {
    if (!upstreamForm.base_url.trim() || !upstreamForm.api_key.trim()) {
      throw new Error("请先填写 Base URL 和 API Key");
    }
    const models = await api.post<UpstreamModel[]>("/admin/upstream/models/query", {
      base_url: upstreamForm.base_url.trim(),
      api_key: upstreamForm.api_key.trim()
    });
    state.upstreamModels = models;
    state.upstreamModelConfigs = models.map((item, index) => modelConfigFromName(item.id, index === 0));
  } catch (error) {
    state.error = error instanceof Error ? error.message : "查询上游模型失败";
  } finally {
    state.queryingUpstreamModels = false;
  }
}

function applyUpstreamModel(modelID: string) {
  const model = modelID.trim();
  if (!model) return;
  upstreamForm.provider_model_name = model;
  upstreamForm.model_code = model;
  upstreamForm.display_name = displayNameFromModel(model);
}

function toggleModelConfig(config: UpstreamModelConfig) {
  config.expanded = !config.expanded;
}

function modelConfigFromForm(expanded = false, modelType: "image" | "video" = "image"): UpstreamModelConfig {
  return {
    model_type: modelType,
    provider_key_id: state.providerKeys.find((key) => key.provider_id === state.selectedProviderId)?.id,
    provider_model_name: upstreamForm.provider_model_name,
    model_code: upstreamForm.model_code,
    display_name: upstreamForm.display_name,
    description: upstreamForm.description,
    price_credits: Number(upstreamForm.price_credits) || (modelType === "video" ? 80 : 8),
    supported_sizes: upstreamForm.supported_sizes,
    max_images_per_request: modelType === "video" ? 1 : Number(upstreamForm.max_images_per_request) || 4,
    supported_aspect_ratios: defaultVideoAspectRatios,
    supported_seconds: defaultVideoSeconds,
    extra_config: "",
    priority: Number(upstreamForm.priority) || 100,
    enabled: upstreamForm.enabled,
    expanded
  };
}

function modelConfigFromName(model: string, expanded = false): UpstreamModelConfig {
  const modelType = guessModelType(model);
  return {
    model_type: modelType,
    provider_key_id: state.providerKeys.find((key) => key.provider_id === state.selectedProviderId)?.id,
    provider_model_name: model,
    model_code: model,
    display_name: displayNameFromModel(model),
    description: modelType === "video" ? "OpenAI 兼容视频生成模型" : "OpenAI 兼容图片生成模型",
    price_credits: modelType === "video" ? 80 : 8,
    supported_sizes: defaultImageSupportedSizes,
    max_images_per_request: modelType === "video" ? 1 : 4,
    supported_aspect_ratios: defaultVideoAspectRatios,
    supported_seconds: defaultVideoSeconds,
    extra_config: defaultExtraConfigForModel(model, modelType),
    priority: Number(upstreamForm.priority) || 100,
    enabled: true,
    expanded
  };
}

async function saveImageModelConfig(providerID: number, config: UpstreamModelConfig) {
  const sizes = splitCSV(config.supported_sizes);
  if (!sizes.length) {
    throw new Error("请填写图片模型支持尺寸");
  }
  const modelPayload = {
    code: config.model_code.trim(),
    display_name: config.display_name.trim(),
    description: config.description.trim(),
    cover_url: "",
    price_credits: Number(config.price_credits) || 0,
    supported_sizes: sizes,
    support_text_to_image: true,
    support_image_to_image: true,
    support_edit: true,
    max_images_per_request: Number(config.max_images_per_request) || 1,
    auto_refund_on_failure: true,
    enabled: upstreamForm.enabled && config.enabled,
    recommended: true,
    sort_order: 10
  };
  const existingModel = state.imageModels.find((item) => item.code === modelPayload.code);
  const imageModel = existingModel
    ? await updateAndReturnModel(existingModel, modelPayload)
    : await api.post<ImageModel>("/admin/image-models", modelPayload);
  const routePayload = modelRoutePayload(providerID, config);
  const routes = await api.get<ImageModelRoute[]>(`/admin/image-models/${imageModel.id}/routes`);
  const existingRoute = routes.find((item) => item.provider_id === providerID);
  if (existingRoute) {
    await api.put(`/admin/image-model-routes/${existingRoute.id}`, routePayload);
  } else {
    await api.post(`/admin/image-models/${imageModel.id}/routes`, routePayload);
  }
}

async function saveVideoModelConfig(providerID: number, config: UpstreamModelConfig) {
  const ratios = splitCSV(config.supported_aspect_ratios);
  const seconds = splitNumberCSV(config.supported_seconds);
  if (!ratios.length || !seconds.length) {
    throw new Error("请填写视频模型支持比例和时长");
  }
  const modelPayload = {
    code: config.model_code.trim(),
    display_name: config.display_name.trim(),
    description: config.description.trim(),
    price_credits: Number(config.price_credits) || 0,
    supported_aspect_ratios: ratios,
    supported_seconds: seconds,
    enabled: upstreamForm.enabled && config.enabled,
    recommended: true,
    sort_order: 20
  };
  const existingModel = state.videoModels.find((item) => item.code === modelPayload.code);
  const videoModel = existingModel
    ? await updateAndReturnVideoModel(existingModel, modelPayload)
    : await api.post<VideoModel>("/admin/video-models", modelPayload);
  const routePayload = modelRoutePayload(providerID, config);
  const routes = await api.get<VideoModelRoute[]>(`/admin/video-models/${videoModel.id}/routes`);
  const existingRoute = routes.find((item) => item.provider_id === providerID);
  if (existingRoute) {
    await api.put(`/admin/video-model-routes/${existingRoute.id}`, routePayload);
  } else {
    await api.post(`/admin/video-models/${videoModel.id}/routes`, routePayload);
  }
}

function modelRoutePayload(providerID: number, config: UpstreamModelConfig) {
  return {
    provider_id: providerID,
    provider_key_id: config.provider_key_id || providerKeyIDForProvider(providerID),
    provider_model_name: config.provider_model_name.trim(),
    enabled: upstreamForm.enabled && config.enabled,
    priority: Number(config.priority) || Number(upstreamForm.priority) || 100,
    weight: 100,
    extra_config: parseExtraConfig(config.extra_config)
  };
}

function providerKeyIDForProvider(providerID: number) {
  return state.providerKeys.find((key) => key.provider_id === providerID && key.status === "active")?.id;
}

async function configsForProvider(provider: Provider): Promise<UpstreamModelConfig[]> {
  const configs: UpstreamModelConfig[] = [];
  for (const model of state.imageModels) {
    const routes = await api.get<ImageModelRoute[]>(`/admin/image-models/${model.id}/routes`);
    const route = routes.find((item) => item.provider_id === provider.id);
    if (route) {
      configs.push({
        model_type: "image",
        provider_key_id: route.provider_key_id,
        provider_model_name: route.provider_model_name,
        model_code: model.code || route.provider_model_name,
        display_name: model.display_name || displayNameFromModel(route.provider_model_name),
        description: model.description || "",
        price_credits: model.price_credits ?? 8,
        supported_sizes: formatSupportedSizes(model.supported_sizes),
        max_images_per_request: model.max_images_per_request ?? 4,
        supported_aspect_ratios: defaultVideoAspectRatios,
        supported_seconds: defaultVideoSeconds,
        extra_config: formatExtraConfig(route.extra_config),
        priority: route.priority || provider.priority || 100,
        enabled: provider.enabled && model.enabled && route.enabled,
        expanded: configs.length === 0
      });
    }
  }
  for (const model of state.videoModels) {
    const routes = await api.get<VideoModelRoute[]>(`/admin/video-models/${model.id}/routes`);
    const route = routes.find((item) => item.provider_id === provider.id);
    if (route) {
      configs.push({
        model_type: "video",
        provider_key_id: route.provider_key_id,
        provider_model_name: route.provider_model_name,
        model_code: model.code || route.provider_model_name,
        display_name: model.display_name || displayNameFromModel(route.provider_model_name),
        description: model.description || "",
        price_credits: model.price_credits ?? 80,
        supported_sizes: defaultImageSupportedSizes,
        max_images_per_request: 1,
        supported_aspect_ratios: formatList(model.supported_aspect_ratios, defaultVideoAspectRatios),
        supported_seconds: formatList(model.supported_seconds, defaultVideoSeconds),
        extra_config: formatExtraConfig(route.extra_config),
        priority: route.priority || provider.priority || 100,
        enabled: provider.enabled && model.enabled && route.enabled,
        expanded: configs.length === 0
      });
    }
  }
  return configs;
}

function formatSupportedSizes(value: ImageModel["supported_sizes"] | undefined): string {
  return Array.isArray(value) && value.length ? value.map(String).join(",") : defaultImageSupportedSizes;
}

function formatList(value: unknown, fallback: string): string {
  return Array.isArray(value) && value.length ? value.map(String).join(",") : fallback;
}

function defaultExtraConfigForModel(model: string, modelType: "image" | "video") {
  if (modelType !== "video") return "";
  if (model === "grok-video-1.5") {
    return requestJSON({
      resolution: "720p",
      max_reference_images: 1,
      require_exactly_one_image: true,
      image_field: "image_urls"
    });
  }
  if (model === "grok-image-video") {
    return requestJSON({
      resolution: "720p",
      max_reference_images: 7,
      multi_image_max_seconds: 10,
      image_field: "image_urls"
    });
  }
  return "";
}

function formatExtraConfig(value: unknown) {
  if (!value || (typeof value === "object" && !Object.keys(value as Record<string, unknown>).length)) {
    return "";
  }
  return requestJSON(value);
}

function parseExtraConfig(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return {};
  try {
    return JSON.parse(trimmed);
  } catch {
    throw new Error("扩展配置必须是合法 JSON");
  }
}

function displayNameFromModel(model: string): string {
  return model
    .split("-")
    .filter(Boolean)
    .map((part) => (part.length <= 3 ? part.toUpperCase() : part.charAt(0).toUpperCase() + part.slice(1)))
    .join(" ");
}

async function updateAndReturnProvider(existing: Provider, payload: Partial<Provider>): Promise<Provider> {
  await api.put(`/admin/providers/${existing.id}`, payload);
  return { ...existing, ...payload };
}

async function updateAndReturnModel(existing: ImageModel, payload: Partial<ImageModel> & { supported_sizes: string[] }): Promise<ImageModel> {
  await api.put(`/admin/image-models/${existing.id}`, payload);
  return { ...existing, ...payload };
}

async function updateAndReturnVideoModel(
  existing: VideoModel,
  payload: Partial<VideoModel> & { supported_aspect_ratios: string[]; supported_seconds: number[] }
): Promise<VideoModel> {
  await api.put(`/admin/video-models/${existing.id}`, payload);
  return { ...existing, ...payload };
}

async function deleteImageModel(model: ImageModel) {
  if (!window.confirm(`确认删除图片模型「${model.display_name || model.code}」？已有任务引用的模型不能删除，可以改为停用。`)) {
    return;
  }
  await withLoading(async () => {
    await api.delete(`/admin/image-models/${model.id}`);
    await loadModels();
  });
}

async function deleteVideoModel(model: VideoModel) {
  if (!window.confirm(`确认删除视频模型「${model.display_name || model.code}」？已有任务引用的模型不能删除，可以改为停用。`)) {
    return;
  }
  await withLoading(async () => {
    await api.delete(`/admin/video-models/${model.id}`);
    await loadVideoModels();
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

function editProvider(item: Provider) {
  Object.assign(providerForm, item);
}

function resetProviderForm() {
  Object.assign(providerForm, { id: 0, code: "", name: "", type: "openai-compatible", base_url: "", enabled: true, timeout_seconds: 60, retry_count: 1, priority: 100, remark: "" });
}

function splitCSV(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function splitNumberCSV(value: string): number[] {
  return splitCSV(value)
    .map((item) => Number(item))
    .filter((item) => Number.isFinite(item) && item > 0);
}

function guessModelType(model: string): "image" | "video" {
  const normalized = model.toLowerCase();
  return normalized.includes("video") || normalized.includes("veo") || normalized.includes("kling") ? "video" : "image";
}

function taskQueryString() {
  const params = new URLSearchParams({ limit: String(state.taskPage * state.taskPageSize + 1) });
  const id = taskFilterForm.id.trim();
  const keyword = taskFilterForm.keyword.trim();
  if (id) {
    params.set(/^\d+$/.test(id) ? "id" : "task_no", id);
  }
  if (keyword) {
    params.set("keyword", keyword);
  }
  if (taskFilterForm.status) {
    params.set("status", taskFilterForm.status);
  }
  return params.toString();
}

function walletTypeText(type: string) {
  const labels: Record<string, string> = {
    consume: "生成消费",
    refund: "失败退款",
    admin_add: "管理员增加",
    admin_deduct: "管理员扣减",
    register_gift: "注册赠送"
  };
  return labels[type] ?? type;
}

function signedCredits(value: number) {
  return `${value > 0 ? "+" : ""}${value} 积分`;
}

function formatDateTime(value: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString("zh-CN", { hour12: false });
}

function formatDatabaseValue(value: unknown) {
  if (value === null || value === undefined) return "-";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  return JSON.stringify(value);
}

function databaseRowJSON(row: Record<string, unknown> | null) {
  return requestJSON(row || {});
}

function tableComment(tableName: string) {
  const table = state.databaseTables.find((item) => item.name === tableName);
  return table?.comment || "";
}

function taskStatusText(status: string) {
  const labels: Record<string, string> = {
    pending: "排队中",
    running: "生成中",
    succeeded: "已完成",
    failed: "失败",
    timeout: "生成超时 24 小时"
  };
  return labels[status] ?? status;
}

function imageModelName(modelID: number) {
  const imageModel = state.imageModels.find((item) => item.id === modelID);
  return imageModel?.display_name || `#${modelID}`;
}

function videoModelName(modelID: number) {
  const videoModel = state.videoModels.find((item) => item.id === modelID);
  return videoModel?.display_name || `#${modelID}`;
}

function imageTaskRequestParams(task: ImageTask) {
  return [
    { label: "模型", value: imageModelName(task.model_id) },
    { label: "模型 ID", value: task.model_id },
    { label: "尺寸", value: task.size || "-" },
    { label: "张数", value: task.num_images },
    { label: "负向词", value: task.negative_prompt || "-" },
    { label: "来源", value: task.source || "-" }
  ];
}

function imageTaskRequestPayload(task: ImageTask) {
  return {
    type: "image",
    model: imageModelName(task.model_id),
    model_id: task.model_id,
    prompt: task.prompt,
    negative_prompt: task.negative_prompt || "",
    size: task.size,
    n: task.num_images,
    source: task.source,
    user_id: task.user_id,
    task_no: task.task_no
  };
}

function videoTaskRequestParams(task: VideoTask) {
  return [
    { label: "模型", value: videoModelName(task.model_id) },
    { label: "模型 ID", value: task.model_id },
    { label: "时长", value: `${task.seconds || 0}s` },
    { label: "比例", value: task.aspect_ratio || "-" },
    { label: "参考图", value: mediaCount(task.images) },
    { label: "参考视频", value: mediaCount(task.videos) },
    { label: "参考音频", value: mediaCount(task.audios) },
    { label: "来源", value: task.source || "-" }
  ];
}

function videoTaskRequestPayload(task: VideoTask) {
  return {
    type: "video",
    model: videoModelName(task.model_id),
    model_id: task.model_id,
    prompt: task.prompt,
    seconds: task.seconds,
    aspect_ratio: task.aspect_ratio,
    images: parseJSONList(task.images),
    videos: parseJSONList(task.videos),
    audios: parseJSONList(task.audios),
    source: task.source,
    user_id: task.user_id,
    task_no: task.task_no
  };
}

function requestJSON(payload: unknown) {
  return JSON.stringify(payload, null, 2);
}

function providerResponsePayload(task: ImageTask | VideoTask) {
  return task.provider_response || {};
}

async function copyJSON(payload: unknown) {
  const text = requestJSON(payload);
  await copyText(text);
}

async function copyText(text: string) {
  let copied = false;
  try {
    await navigator.clipboard.writeText(text);
    copied = true;
  } catch {
    copied = fallbackCopyText(text);
  }
  state.copiedJSON = copied ? "copied" : "failed";
  window.setTimeout(() => {
    state.copiedJSON = "";
  }, 1600);
}

function fallbackCopyText(text: string) {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "true");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  textarea.style.top = "0";
  document.body.appendChild(textarea);
  textarea.select();
  const copied = document.execCommand("copy");
  document.body.removeChild(textarea);
  return copied;
}

function mediaCount(value: unknown) {
  return parseJSONList(value).length;
}

function imageUrl(value: string) {
  if (!value || value.startsWith("http://") || value.startsWith("https://") || value.startsWith("data:")) {
    return value;
  }
  if (value.startsWith("/")) {
    return `${apiBaseURL}${value}`;
  }
  return value;
}

function taskDetailURL(taskNo: string) {
  return `${window.location.origin}${window.location.pathname}?view=task&task_no=${encodeURIComponent(taskNo)}`;
}

function videoTaskDetailURL(taskNo: string) {
  return `${window.location.origin}${window.location.pathname}?view=video_task&task_no=${encodeURIComponent(taskNo)}`;
}

function updateBrowserURL(url: string, replace = false) {
  if (window.location.href === url) {
    return;
  }
  const method = replace ? "replaceState" : "pushState";
  window.history[method]({}, "", url);
}

function setActiveView(view: MenuView) {
  state.activeView = view;
  state.selectedTask = null;
  state.selectedVideoTask = null;
  updateBrowserURL(menuViewURL(view));
  if (view === "database") {
    void refreshDatabase();
  }
}

function isMenuView(view: string | null): view is MenuView {
  return menuViews.includes(view as MenuView);
}

function menuViewURL(view: MenuView) {
  return `${window.location.origin}${window.location.pathname}?view=${encodeURIComponent(view)}`;
}

function activeViewTitle() {
  if (state.activeView === "users") return "用户管理";
  if (state.activeView === "credits") return "积分管理";
  if (state.activeView === "providers") return "上游 API 管理";
  if (state.activeView === "database") return "数据表浏览";
  if (state.activeView === "profile") return "个人中心";
  return "生成任务";
}

async function withLoading(task: () => Promise<void>) {
  state.loading = true;
  state.error = "";
  state.profileMessage = "";
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
  components: { Database, Eye, EyeOff, KeyRound, Loader2, LogOut, ReceiptText, RefreshCw, Route, Server, ShieldCheck, UserCircle, Users },
  setup() {
    bootstrap();
    return {
      state,
      loginForm,
      showPassword,
      userForm,
      creditForm,
      profileForm,
      taskFilterForm,
      upstreamForm,
      providerForm,
      providerKeyForm,
      formatCredits,
      combinedTasks,
      pagedTasks,
      taskHasNext,
      taskRangeText,
      walletRangeText,
      databaseRangeText,
      activeViewTitle,
      walletTypeText,
      signedCredits,
      formatDateTime,
      formatDatabaseValue,
      databaseRowJSON,
      tableComment,
      formatList,
      taskStatusText,
      imageModelName,
      videoModelName,
      imageTaskRequestParams,
      imageTaskRequestPayload,
      videoTaskRequestParams,
      videoTaskRequestPayload,
      providerResponsePayload,
      requestJSON,
      copyJSON,
      copyText,
      imageUrl,
      taskDetailURL,
      videoTaskDetailURL,
      setActiveView,
      login,
      loadAll,
      loadTaskDetail,
      loadVideoTaskDetail,
      refreshTasks,
      changeTaskPage,
      resetTaskFilters,
      loadUsers,
      loadWalletLogs,
      refreshWalletLogs,
      changeWalletPage,
      changePassword,
      loadDatabaseTables,
      openDatabaseTable,
      loadDatabaseTable,
      refreshDatabase,
      changeDatabasePage,
      openDatabaseRowDetail,
      closeDatabaseRowDetail,
      toggleDatabaseDDL,
      loadProviderKeys,
      saveUser,
      adjustCredits,
      openCreateUserModal,
      openEditUserModal,
      closeUserModal,
      saveProvider,
      saveUpstreamIntegration,
      toggleProviderEnabled,
      openCreateUpstreamModal,
      openEditUpstreamModal,
      closeUpstreamModal,
      applyDefaultImageModel,
      applyDefaultVideoModel,
      applyDefaultGrokVideoModels,
      queryUpstreamModels,
      applyUpstreamModel,
      toggleModelConfig,
      deleteImageModel,
      deleteVideoModel,
      createProviderKey,
      deleteProviderKey,
      editProvider,
      resetProviderForm,
      logout
    };
  },
  template: `
    <main class="admin-shell">
      <aside class="sidebar">
        <div class="brand"><Database :size="22" /><strong>AGI Admin</strong></div>
        <button class="nav-item" :class="{ active: state.activeView === 'users' }" @click="setActiveView('users')"><Users :size="18" />用户</button>
        <button class="nav-item" :class="{ active: state.activeView === 'credits' }" @click="setActiveView('credits')"><ReceiptText :size="18" />积分</button>
        <button class="nav-item" :class="{ active: state.activeView === 'providers' }" @click="setActiveView('providers')"><Server :size="18" />上游 API</button>
        <button class="nav-item" :class="{ active: state.activeView === 'tasks' }" @click="setActiveView('tasks')"><Route :size="18" />任务</button>
        <button class="nav-item" :class="{ active: state.activeView === 'database' }" @click="setActiveView('database')"><Database :size="18" />数据表</button>
        <button class="nav-item" :class="{ active: state.activeView === 'profile' }" @click="setActiveView('profile')"><UserCircle :size="18" />个人中心</button>
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
              <h1>{{ activeViewTitle() }}</h1>
              <p>{{ state.admin.username }} · {{ state.admin.role }}</p>
            </div>
            <button class="ghost-button" @click="loadAll"><RefreshCw :size="16" />刷新</button>
          </header>
          <p v-if="state.error" class="error">{{ state.error }}</p>

          <section v-if="state.activeView === 'users'" class="panel">
            <div class="panel-heading">
              <div>
                <h2>用户列表</h2>
                <p>查看用户基础信息、状态和当前积分余额。</p>
              </div>
              <div class="button-row">
                <button class="primary-button" type="button" @click="openCreateUserModal"><Users :size="16" />新增用户</button>
                <button class="ghost-button" @click="loadUsers"><RefreshCw :size="16" />刷新用户</button>
              </div>
            </div>
            <table>
              <thead><tr><th>ID</th><th>邮箱</th><th>昵称</th><th>积分</th><th>状态</th><th></th></tr></thead>
              <tbody>
                <tr v-for="user in state.users" :key="user.id">
                  <td>{{ user.id }}</td>
                  <td>{{ user.email || '-' }}</td>
                  <td>{{ user.nickname }}</td>
                  <td>{{ formatCredits(user.credits) }}</td>
                  <td>{{ user.status }}</td>
                  <td><button class="small-button" type="button" @click="openEditUserModal(user)">编辑</button></td>
                </tr>
              </tbody>
            </table>
            <div v-if="state.userModalOpen" class="modal-backdrop">
              <div class="modal-panel user-modal">
                <div class="modal-header">
                  <div>
                    <h2>{{ state.userModalMode === 'create' ? '新增用户' : '编辑用户' }}</h2>
                    <p v-if="state.selectedUser">用户 {{ state.selectedUser.id }} · 当前 {{ formatCredits(state.selectedUser.credits) }}</p>
                    <p v-else>创建用户后，可以继续在这里调整积分和管理扩展能力。</p>
                  </div>
                  <button class="ghost-button" type="button" @click="closeUserModal">关闭</button>
                </div>
                <div class="form-section">
                  <h3>基础信息</h3>
                  <div class="form-grid">
                    <label>邮箱<input v-model="userForm.email" type="email" placeholder="user@example.com" /></label>
                    <label>手机号<input v-model="userForm.phone" placeholder="可选" /></label>
                    <label>昵称<input v-model="userForm.nickname" placeholder="留空时使用邮箱前缀" /></label>
                    <label>状态
                      <select v-model="userForm.status">
                        <option value="active">active</option>
                        <option value="disabled">disabled</option>
                      </select>
                    </label>
                    <label>密码
                      <input v-model="userForm.password" type="password" :placeholder="state.userModalMode === 'edit' ? '留空表示不修改' : '至少 6 位'" />
                    </label>
                    <label>头像 URL<input v-model="userForm.avatar_url" placeholder="可选" /></label>
                  </div>
                  <div class="modal-actions inline-actions">
                    <button class="primary-button" type="button" @click="saveUser" :disabled="state.loading">
                      <Loader2 v-if="state.loading" class="spin" :size="16" />保存基础信息
                    </button>
                  </div>
                </div>

                <div v-if="state.userModalMode === 'edit'" class="form-section">
                  <h3>积分调整</h3>
                  <div class="segmented">
                    <button :class="{ active: creditForm.mode === 'add' }" @click="creditForm.mode = 'add'" type="button">增加</button>
                    <button :class="{ active: creditForm.mode === 'deduct' }" @click="creditForm.mode = 'deduct'" type="button">扣减</button>
                  </div>
                  <div class="form-grid">
                    <label>积分数量<input v-model.number="creditForm.amount" type="number" min="1" /></label>
                    <label>备注<input v-model="creditForm.remark" /></label>
                  </div>
                  <div class="modal-actions inline-actions">
                    <button class="primary-button" type="button" @click="adjustCredits" :disabled="state.loading">
                      <Loader2 v-if="state.loading" class="spin" :size="16" />{{ creditForm.mode === 'deduct' ? '确认扣减' : '确认增加' }}
                    </button>
                  </div>
                </div>
                <div class="modal-actions">
                  <button class="ghost-button" type="button" @click="closeUserModal">关闭</button>
                </div>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'credits'" class="panel">
              <div class="panel-heading">
                <div>
                  <h2>积分流水</h2>
                  <p>记录所有用户积分增加、扣减、生成消费和退款。</p>
                </div>
                <button class="ghost-button" @click="refreshWalletLogs"><ReceiptText :size="16" />刷新流水</button>
              </div>
              <table>
                <thead><tr><th>ID</th><th>用户</th><th>时间</th><th>类型</th><th>变动</th><th>变动前</th><th>变动后</th><th>关联</th><th>操作方</th><th>备注</th></tr></thead>
                <tbody>
                  <tr v-for="log in state.walletLogs" :key="log.id">
                    <td>{{ log.id }}</td>
                    <td>{{ log.user_id }}</td>
                    <td class="nowrap">{{ formatDateTime(log.created_at) }}</td>
                    <td>{{ walletTypeText(log.type) }}</td>
                    <td class="amount-cell" :class="{ positive: log.amount > 0, negative: log.amount < 0 }">{{ signedCredits(log.amount) }}</td>
                    <td>{{ formatCredits(log.balance_before) }}</td>
                    <td>{{ formatCredits(log.balance_after) }}</td>
                    <td>{{ log.related_type || '-' }}<span v-if="log.related_id"> #{{ log.related_id }}</span></td>
                    <td>{{ log.operator_type }}<span v-if="log.operator_id"> #{{ log.operator_id }}</span></td>
                    <td class="prompt-cell">{{ log.remark || '-' }}</td>
                  </tr>
                </tbody>
              </table>
              <div class="pagination-bar">
                <span>当前 {{ walletRangeText }}</span>
                <div>
                  <button class="ghost-button small-button" :disabled="state.walletPage <= 1" @click="changeWalletPage(-1)">上一页</button>
                  <span class="page-number">第 {{ state.walletPage }} 页</span>
                  <button class="ghost-button small-button" :disabled="!state.walletHasNext" @click="changeWalletPage(1)">下一页</button>
                </div>
              </div>
          </section>

          <section v-if="state.activeView === 'profile'" class="panel">
            <div class="panel-heading">
              <div>
                <h2>个人中心</h2>
                <p>查看管理员账号信息并修改登录密码。</p>
              </div>
              <button class="ghost-button" @click="loadMe"><RefreshCw :size="16" />刷新</button>
            </div>
            <div class="profile-layout">
              <div class="profile-card">
                <UserCircle :size="38" />
                <div>
                  <strong>{{ state.admin.nickname || state.admin.username }}</strong>
                  <span>{{ state.admin.username }}</span>
                </div>
                <div class="profile-meta">
                  <span>角色</span>
                  <strong>{{ state.admin.role }}</strong>
                </div>
                <div class="profile-meta">
                  <span>状态</span>
                  <strong>{{ state.admin.status }}</strong>
                </div>
              </div>

              <div class="profile-form">
                <div class="section-title">
                  <ShieldCheck :size="18" />
                  <h3>修改密码</h3>
                </div>
                <p v-if="state.profileMessage" class="secret-line">{{ state.profileMessage }}</p>
                <div class="form-grid">
                  <label>当前密码<input v-model="profileForm.current_password" type="password" autocomplete="current-password" /></label>
                  <label>新密码<input v-model="profileForm.new_password" type="password" autocomplete="new-password" /></label>
                  <label>确认新密码<input v-model="profileForm.confirm_password" type="password" autocomplete="new-password" /></label>
                </div>
                <button class="primary-button" @click="changePassword" :disabled="state.loading">
                  <Loader2 v-if="state.loading" class="spin" :size="16" />保存新密码
                </button>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'providers'" class="upstream-page">
            <div class="panel full">
              <div class="panel-heading">
                <div>
                  <h2>已接入上游</h2>
                  <p>统一管理管理员配置的上游 API、密钥、模型和路由关系。</p>
                </div>
                <button class="primary-button" @click="openCreateUpstreamModal">添加上游 API</button>
              </div>
              <table>
                <thead><tr><th>ID</th><th>Code</th><th>名称</th><th>类型</th><th>Base URL</th><th>优先级</th><th>启用</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="item in state.providers" :key="item.id">
                    <td>{{ item.id }}</td>
                    <td>{{ item.code }}</td>
                    <td>{{ item.name }}</td>
                    <td>{{ item.type }}</td>
                    <td>{{ item.base_url || '-' }}</td>
                    <td>{{ item.priority }}</td>
                    <td><span class="status-pill" :class="{ active: item.enabled }">{{ item.enabled ? '启用中' : '已停用' }}</span></td>
                    <td class="action-cell">
                      <button class="small-button" @click="toggleProviderEnabled(item)">{{ item.enabled ? '停用' : '启用' }}</button>
                      <button class="small-button" @click="openEditUpstreamModal(item)">编辑</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="panel full">
              <h2>API Key 状态</h2>
              <div class="toolbar key-toolbar">
                <select v-model.number="state.selectedProviderId" @change="loadProviderKeys"><option v-for="item in state.providers" :key="item.id" :value="item.id">{{ item.name }}</option></select>
              </div>
              <table><thead><tr><th>ID</th><th>名称</th><th>状态</th><th>权重</th><th>错误</th><th></th></tr></thead><tbody><tr v-for="key in state.providerKeys" :key="key.id"><td>{{ key.id }}</td><td>{{ key.name }}</td><td>{{ key.status }}</td><td>{{ key.weight }}</td><td>{{ key.last_error || '-' }}</td><td><button class="danger-button" @click="deleteProviderKey(key.id)">删除</button></td></tr></tbody></table>
            </div>

            <div class="panel full">
              <h2>图片模型概览</h2>
              <table>
                <thead><tr><th>模型 ID</th><th>模型 Code</th><th>显示名称</th><th>价格</th><th>启用</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="item in state.imageModels" :key="item.id">
                    <td>{{ item.id }}</td>
                    <td>{{ item.code }}</td>
                    <td>{{ item.display_name }}</td>
                    <td>{{ item.price_credits }}</td>
                    <td>{{ item.enabled ? '是' : '否' }}</td>
                    <td><button class="danger-button" type="button" @click="deleteImageModel(item)">删除</button></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="panel full">
              <h2>视频模型概览</h2>
              <table>
                <thead><tr><th>模型 ID</th><th>模型 Code</th><th>显示名称</th><th>每次积分</th><th>支持比例</th><th>支持时长</th><th>启用</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="item in state.videoModels" :key="item.id">
                    <td>{{ item.id }}</td>
                    <td>{{ item.code }}</td>
                    <td>{{ item.display_name }}</td>
                    <td>{{ item.price_credits }}</td>
                    <td>{{ formatList(item.supported_aspect_ratios, '-') }}</td>
                    <td>{{ formatList(item.supported_seconds, '-') }}</td>
                    <td>{{ item.enabled ? '是' : '否' }}</td>
                    <td><button class="danger-button" type="button" @click="deleteVideoModel(item)">删除</button></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="state.upstreamModalOpen" class="modal-backdrop">
              <div class="modal-panel">
                <div class="modal-header">
                  <div>
                    <h2>{{ state.upstreamModalMode === 'create' ? '添加上游 API' : '编辑上游 API' }}</h2>
                    <p>填写 URL 和 API Key 后可同步上游模型，也可以手动填写或套用默认图片/视频模型配置。</p>
                  </div>
                  <button class="ghost-button" type="button" @click="closeUpstreamModal">关闭</button>
                </div>

                <div class="form-section">
                  <h3>上游 API</h3>
                  <div class="form-grid">
                    <label>Provider Code<input v-model="upstreamForm.provider_code" placeholder="chongplus" /></label>
                    <label>显示名称<input v-model="upstreamForm.provider_name" placeholder="ChongPlus" /></label>
                    <label>接口类型
                      <select v-model="upstreamForm.provider_type">
                        <option value="openai-compatible">OpenAI 兼容接口（图片/视频）</option>
                        <option value="openai">OpenAI</option>
                        <option value="grok-video">Grok 视频接口</option>
                        <option value="mock">Mock</option>
                      </select>
                    </label>
                    <label>Base URL<input v-model="upstreamForm.base_url" placeholder="https://api.example.com" /></label>
                    <label>Key 名称<input v-model="upstreamForm.api_key_name" placeholder="default" /></label>
                    <label>API Key<input v-model="upstreamForm.api_key" type="password" autocomplete="new-password" :placeholder="state.upstreamModalMode === 'edit' ? '留空则不新增 Key' : 'sk-...'" /></label>
                    <div class="span-two query-row">
                      <button class="ghost-button" type="button" @click="queryUpstreamModels" :disabled="state.queryingUpstreamModels">
                        <Loader2 v-if="state.queryingUpstreamModels" class="spin" :size="16" />同步上游模型
                      </button>
                      <button class="ghost-button" type="button" @click="applyDefaultImageModel">填入默认生图模型</button>
                      <button class="ghost-button" type="button" @click="applyDefaultVideoModel">填入默认视频模型</button>
                      <button class="ghost-button" type="button" @click="applyDefaultGrokVideoModels">填入 Grok 视频模型</button>
                      <span v-if="state.upstreamModelConfigs.length">已导入 {{ state.upstreamModelConfigs.length }} 个模型</span>
                    </div>
                  </div>
                </div>

                <div class="form-section">
                  <h3>模型配置</h3>
                  <div class="model-config-list">
                    <article v-for="config in state.upstreamModelConfigs" :key="config.provider_model_name" class="model-config-item">
                      <div class="model-config-summary">
                        <div>
                          <strong>{{ config.provider_model_name }}</strong>
                          <span>{{ config.display_name }} · {{ config.model_type === 'video' ? '视频' : '图片' }} · {{ config.price_credits }} 积分/{{ config.model_type === 'video' ? '次' : '张' }}</span>
                        </div>
                        <div class="action-cell">
                          <span class="status-pill" :class="{ active: config.enabled }">{{ config.enabled ? '启用' : '停用' }}</span>
                          <button class="small-button" type="button" @click="toggleModelConfig(config)">{{ config.expanded ? '收起' : '详情/编辑' }}</button>
                        </div>
                      </div>
                      <div v-if="config.expanded" class="form-grid model-config-detail">
                        <label>模型类型
                          <select v-model="config.model_type">
                            <option value="image">图片生成</option>
                            <option value="video">视频生成</option>
                          </select>
                        </label>
                        <label>平台模型 Code<input v-model="config.model_code" /></label>
                        <label>显示名称<input v-model="config.display_name" /></label>
                        <label>上游模型名<input v-model="config.provider_model_name" /></label>
                        <label>调用 API Key
                          <select v-model.number="config.provider_key_id">
                            <option :value="undefined">自动选择 active key</option>
                            <option v-for="key in state.providerKeys" :key="key.id" :value="key.id">{{ key.name }} · {{ key.status }}</option>
                          </select>
                        </label>
                        <label>{{ config.model_type === 'video' ? '每次积分' : '每张积分' }}<input v-model.number="config.price_credits" type="number" min="0" /></label>
                        <label v-if="config.model_type === 'image'">单次最大张数<input v-model.number="config.max_images_per_request" type="number" min="1" /></label>
                        <label>优先级<input v-model.number="config.priority" type="number" /></label>
                        <label class="check-row"><input v-model="config.enabled" type="checkbox" />启用这个模型</label>
                        <label v-if="config.model_type === 'image'" class="span-two">支持尺寸
                          <input v-model="config.supported_sizes" placeholder="1024x1024,2048x2048,1536x1024,1024x1536,3840x2160,2160x3840" />
                        </label>
                        <label v-if="config.model_type === 'video'">支持比例
                          <input v-model="config.supported_aspect_ratios" placeholder="9:16,16:9,1:1" />
                        </label>
                        <label v-if="config.model_type === 'video'">支持时长秒数
                          <input v-model="config.supported_seconds" placeholder="15" />
                        </label>
                        <label v-if="config.model_type === 'video'" class="span-two">扩展配置 JSON
                          <textarea v-model="config.extra_config" rows="5" placeholder='{"resolution":"720p"}' />
                        </label>
                        <label class="span-two">模型描述<textarea v-model="config.description" rows="3" /></label>
                      </div>
                    </article>
                  </div>
                </div>

                <div class="modal-actions">
                  <button class="ghost-button" type="button" @click="closeUpstreamModal">取消</button>
                  <button class="primary-button" type="button" @click="saveUpstreamIntegration" :disabled="state.loading">
                    <Loader2 v-if="state.loading" class="spin" :size="16" />保存
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'database'" class="database-page">
            <div class="panel table-list-panel">
              <div class="panel-heading compact-heading">
                <div>
                  <h2>数据表</h2>
                  <p>{{ state.databaseTables.length }} 张表</p>
                </div>
                <button class="ghost-button small-button" type="button" @click="refreshDatabase"><RefreshCw :size="15" />刷新</button>
              </div>
              <div class="db-table-list">
                <button
                  v-for="table in state.databaseTables"
                  :key="table.name"
                  class="db-table-item"
                  :class="{ active: state.selectedDatabaseTable === table.name }"
                  type="button"
                  @click="openDatabaseTable(table.name)"
                >
                  <strong>{{ table.comment || table.name }}</strong>
                  <span>{{ table.name }} · {{ table.table_type }} · 约 {{ table.rows }} 行</span>
                </button>
              </div>
            </div>

            <div class="panel db-data-panel">
              <div class="panel-heading">
                <div>
                  <h2>{{ state.selectedDatabaseTable || '请选择数据表' }}</h2>
                  <p v-if="state.databaseData">{{ state.databaseData.comment || tableComment(state.selectedDatabaseTable) || '暂无中文说明' }} · {{ state.databaseData.columns.length }} 列 · 当前 {{ databaseRangeText }}</p>
                  <p v-else>点击左侧表名查看数据。</p>
                </div>
                <button class="ghost-button" type="button" :disabled="!state.selectedDatabaseTable" @click="loadDatabaseTable"><RefreshCw :size="16" />刷新数据</button>
              </div>

              <div v-if="!state.databaseData" class="empty-state">
                <span>请选择一张表</span>
              </div>
              <template v-else>
                <div class="ddl-panel">
                  <div class="ddl-panel-heading">
                    <div>
                      <strong>DDL</strong>
                      <span>{{ state.databaseData.table }} 建表语句</span>
                    </div>
                    <div class="action-cell">
                      <button class="small-button" type="button" @click="copyText(state.databaseData.ddl)">
                        {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 DDL' }}
                      </button>
                      <button class="small-button" type="button" @click="toggleDatabaseDDL">{{ state.databaseDDLOpen ? '收起' : '展开' }}</button>
                    </div>
                  </div>
                  <pre v-if="state.databaseDDLOpen">{{ state.databaseData.ddl }}</pre>
                </div>
                <div class="table-wrap db-table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th class="db-action-column">操作</th>
                        <th v-for="column in state.databaseData.columns" :key="column.name">
                          {{ column.comment || column.name }}
                          <span class="column-name">{{ column.name }}</span>
                          <span class="column-meta">{{ column.type }}{{ column.primary_key ? ' · PK' : '' }}{{ column.nullable ? '' : ' · NN' }}</span>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(row, rowIndex) in state.databaseData.rows" :key="rowIndex">
                        <td class="db-action-column"><button class="small-button" type="button" @click="openDatabaseRowDetail(row, rowIndex)">详情</button></td>
                        <td v-for="column in state.databaseData.columns" :key="column.name" class="db-value-cell" :title="formatDatabaseValue(row[column.name])">
                          {{ formatDatabaseValue(row[column.name]) }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                  <div v-if="!state.databaseData.rows.length" class="empty-state compact-empty">
                    <span>当前页没有数据</span>
                  </div>
                  <div class="pagination-bar">
                    <span>当前 {{ databaseRangeText }}</span>
                    <div>
                      <button class="ghost-button small-button" :disabled="state.databasePage <= 1" @click="changeDatabasePage(-1)">上一页</button>
                      <span class="page-number">第 {{ state.databasePage }} 页</span>
                      <button class="ghost-button small-button" :disabled="!state.databaseData.has_next" @click="changeDatabasePage(1)">下一页</button>
                    </div>
                  </div>
                </div>
              </template>
            </div>

            <div v-if="state.selectedDatabaseRow && state.databaseData" class="modal-backdrop">
              <div class="modal-panel database-row-modal">
                <div class="modal-header">
                  <div>
                    <h2>记录详情</h2>
                    <p>{{ state.selectedDatabaseTable }} · 第 {{ state.selectedDatabaseRowIndex }} 条</p>
                  </div>
                  <button class="ghost-button" type="button" @click="closeDatabaseRowDetail">关闭</button>
                </div>

                <div class="request-param-heading">
                  <h3>完整字段</h3>
                  <button class="small-button" type="button" @click="copyJSON(state.selectedDatabaseRow)">
                    {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 JSON' }}
                  </button>
                </div>
                <div class="db-row-detail-list">
                  <div v-for="column in state.databaseData.columns" :key="column.name" class="db-row-detail-item">
                    <div>
                      <strong>{{ column.comment || column.name }}</strong>
                      <span>{{ column.name }} · {{ column.type }}{{ column.primary_key ? ' · PK' : '' }}{{ column.nullable ? '' : ' · NN' }}</span>
                    </div>
                    <pre>{{ formatDatabaseValue(state.selectedDatabaseRow[column.name]) }}</pre>
                  </div>
                </div>
                <pre class="request-json">{{ databaseRowJSON(state.selectedDatabaseRow) }}</pre>
                <div class="modal-actions">
                  <button class="ghost-button" type="button" @click="closeDatabaseRowDetail">关闭</button>
                </div>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'tasks'" class="panel">
            <div class="panel-heading">
              <div>
                <h2>生成任务</h2>
                <p>管理员可查看用户提交的任务状态、请求参数和生成图片。</p>
              </div>
              <button class="ghost-button" @click="refreshTasks"><RefreshCw :size="16" />刷新任务</button>
            </div>
            <div class="task-filter-bar">
              <label>任务 ID / 任务号
                <input v-model="taskFilterForm.id" placeholder="输入数字 ID 或 task_ / video_" @keyup.enter="refreshTasks" />
              </label>
              <label>任务状态
                <select v-model="taskFilterForm.status" @change="refreshTasks">
                  <option value="">全部状态</option>
                  <option value="pending">排队中</option>
                  <option value="running">生成中</option>
                  <option value="succeeded">已完成</option>
                  <option value="failed">失败</option>
                  <option value="timeout">生成超时 24 小时</option>
                </select>
              </label>
              <label>关键词
                <input v-model="taskFilterForm.keyword" placeholder="任务号或提示词" @keyup.enter="refreshTasks" />
              </label>
              <button class="primary-button" @click="refreshTasks">查询</button>
              <button class="ghost-button" @click="resetTaskFilters">重置</button>
            </div>
            <div class="table-wrap">
              <table>
                <thead><tr><th>类型</th><th>ID</th><th>任务号</th><th>用户</th><th>状态</th><th>进度</th><th>积分</th><th>提示词</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="task in pagedTasks" :key="task.kind + '-' + task.id">
                    <td>{{ task.kind === 'image' ? '图片' : '视频' }}</td>
                    <td>{{ task.id }}</td>
                    <td class="nowrap">{{ task.task_no }}</td>
                    <td>{{ task.user_id }}</td>
                    <td>{{ taskStatusText(task.status) }}</td>
                    <td>{{ task.progress }}%</td>
                    <td>{{ task.credits_used }}</td>
                    <td class="prompt-cell">{{ task.prompt }}</td>
                    <td><a class="small-button" :href="task.detail_url" target="_blank" rel="noreferrer">{{ task.action_label }}</a></td>
                  </tr>
                </tbody>
              </table>
              <div class="pagination-bar">
                <span>当前 {{ taskRangeText }}</span>
                <div>
                  <button class="ghost-button small-button" :disabled="state.taskPage <= 1" @click="changeTaskPage(-1)">上一页</button>
                  <span class="page-number">第 {{ state.taskPage }} 页</span>
                  <button class="ghost-button small-button" :disabled="!taskHasNext" @click="changeTaskPage(1)">下一页</button>
                </div>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'taskDetail'" class="panel">
            <div class="panel-heading">
              <div>
                <h2>任务详情</h2>
                <p v-if="state.selectedTask">用户 {{ state.selectedTask.task.user_id }} · {{ state.selectedTask.task.task_no }}</p>
              </div>
              <button v-if="state.selectedTask" class="ghost-button" @click="loadTaskDetail(state.selectedTask.task.task_no)"><RefreshCw :size="16" />刷新任务</button>
            </div>
            <div v-if="!state.selectedTask" class="empty-state">
              <Loader2 class="spin" :size="28" />
              <span>正在加载任务</span>
            </div>
            <div v-else class="task-detail standalone-detail">
              <div class="task-line">
                <strong>{{ state.selectedTask.task.task_no }}</strong>
                <span>用户 {{ state.selectedTask.task.user_id }} · {{ taskStatusText(state.selectedTask.task.status) }} · {{ state.selectedTask.task.progress }}%</span>
              </div>
              <div class="request-param-panel">
                <div class="request-param-heading">
                  <h3>请求参数</h3>
                  <button class="small-button" type="button" @click="copyJSON(imageTaskRequestPayload(state.selectedTask.task))">
                    {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 JSON' }}
                  </button>
                </div>
                <div class="request-param-grid">
                  <div v-for="param in imageTaskRequestParams(state.selectedTask.task)" :key="param.label">
                    <span>{{ param.label }}</span>
                    <strong>{{ param.value }}</strong>
                  </div>
                </div>
                <pre class="request-json">{{ requestJSON(imageTaskRequestPayload(state.selectedTask.task)) }}</pre>
              </div>
              <div class="request-param-panel">
                <div class="request-param-heading">
                  <h3>上游返回</h3>
                  <button class="small-button" type="button" @click="copyJSON(providerResponsePayload(state.selectedTask.task))">
                    {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 JSON' }}
                  </button>
                </div>
                <pre class="request-json">{{ requestJSON(providerResponsePayload(state.selectedTask.task)) }}</pre>
              </div>
              <p class="prompt-detail">{{ state.selectedTask.task.prompt }}</p>
              <p v-if="state.selectedTask.task.error_message" class="error">{{ state.selectedTask.task.error_message }}</p>
              <div v-if="!state.selectedTask.images.length" class="empty-state">
                <Loader2 v-if="state.selectedTask.task.status === 'pending' || state.selectedTask.task.status === 'running'" class="spin" :size="28" />
                <span>{{ state.selectedTask.task.status === 'pending' || state.selectedTask.task.status === 'running' ? '后台生成中，稍后刷新查看' : '这个任务没有图片结果' }}</span>
              </div>
              <div v-else class="image-grid">
                <article v-for="image in state.selectedTask.images" :key="image.id" class="image-card">
                  <img :src="imageUrl(image.url)" :alt="state.selectedTask.task.prompt" />
                  <a :href="imageUrl(image.url)" target="_blank" rel="noreferrer">打开图片</a>
                </article>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'videoDetail'" class="panel">
            <div class="panel-heading">
              <div>
                <h2>视频任务详情</h2>
                <p v-if="state.selectedVideoTask">用户 {{ state.selectedVideoTask.task.user_id }} · {{ state.selectedVideoTask.task.task_no }}</p>
              </div>
              <button v-if="state.selectedVideoTask" class="ghost-button" @click="loadVideoTaskDetail(state.selectedVideoTask.task.task_no)"><RefreshCw :size="16" />刷新任务</button>
            </div>
            <div v-if="!state.selectedVideoTask" class="empty-state">
              <Loader2 class="spin" :size="28" />
              <span>正在加载视频任务</span>
            </div>
            <div v-else class="task-detail standalone-detail">
              <div class="task-line">
                <strong>{{ state.selectedVideoTask.task.task_no }}</strong>
                <span>用户 {{ state.selectedVideoTask.task.user_id }} · {{ taskStatusText(state.selectedVideoTask.task.status) }} · {{ state.selectedVideoTask.task.progress }}%</span>
              </div>
              <div class="request-param-panel">
                <div class="request-param-heading">
                  <h3>请求参数</h3>
                  <button class="small-button" type="button" @click="copyJSON(videoTaskRequestPayload(state.selectedVideoTask.task))">
                    {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 JSON' }}
                  </button>
                </div>
                <div class="request-param-grid">
                  <div v-for="param in videoTaskRequestParams(state.selectedVideoTask.task)" :key="param.label">
                    <span>{{ param.label }}</span>
                    <strong>{{ param.value }}</strong>
                  </div>
                </div>
                <pre class="request-json">{{ requestJSON(videoTaskRequestPayload(state.selectedVideoTask.task)) }}</pre>
              </div>
              <div class="request-param-panel">
                <div class="request-param-heading">
                  <h3>上游返回</h3>
                  <button class="small-button" type="button" @click="copyJSON(providerResponsePayload(state.selectedVideoTask.task))">
                    {{ state.copiedJSON === 'copied' ? '已复制' : state.copiedJSON === 'failed' ? '复制失败' : '复制 JSON' }}
                  </button>
                </div>
                <pre class="request-json">{{ requestJSON(providerResponsePayload(state.selectedVideoTask.task)) }}</pre>
              </div>
              <p class="prompt-detail">{{ state.selectedVideoTask.task.prompt }}</p>
              <p v-if="state.selectedVideoTask.task.error_message" class="error">{{ state.selectedVideoTask.task.error_message }}</p>
              <div v-if="!state.selectedVideoTask.videos.length" class="empty-state">
                <Loader2 v-if="state.selectedVideoTask.task.status === 'pending' || state.selectedVideoTask.task.status === 'running'" class="spin" :size="28" />
                <span>{{ state.selectedVideoTask.task.status === 'pending' || state.selectedVideoTask.task.status === 'running' ? '后台生成中，稍后刷新查看' : '这个任务没有视频结果' }}</span>
              </div>
              <div v-else class="video-result-list">
                <article v-for="video in state.selectedVideoTask.videos" :key="video.id" class="video-card">
                  <video :src="imageUrl(video.url)" controls playsinline />
                  <a :href="imageUrl(video.url)" target="_blank" rel="noreferrer">下载/打开 MP4</a>
                </article>
              </div>
            </div>
          </section>
        </template>
      </section>
    </main>
  `
};

createApp(App).mount("#app");

window.addEventListener("popstate", () => {
  if (state.token) {
    void syncRouteFromURL();
  }
});
