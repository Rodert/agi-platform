import { createApp, computed, reactive, ref } from "vue";
import {
  ApiClient,
  type ApiKey,
  type AuthToken,
  type CreateApiKeyResult,
  type GenerateImageResult,
  type ImageModel,
  type ImageTask,
  type User,
  type VideoModel,
  type VideoTask,
  type VideoTaskResult,
  type WalletLog,
  formatCredits,
  parseJSONList,
  resolveApiBaseURL
} from "@agi-platform/shared";
import {
  ArrowDownToLine,
  Brush,
  Code2,
  Eye,
  EyeOff,
  Image,
  Film,
  KeyRound,
  ListChecks,
  Loader2,
  LogOut,
  ReceiptText,
  RefreshCw,
  ShieldCheck,
  Sparkles,
  X,
  UserCircle
} from "lucide-vue-next";
import "./styles.css";

const tokenKey = "agi_user_token";
const storedToken = localStorage.getItem(tokenKey);
const apiBaseURL = resolveApiBaseURL();
const commonSizes = ["1K", "2K", "4K"];
const openAIRatiosBySize: Record<string, string[]> = {
  "1K": ["1:1"],
  "2K": ["1:1", "3:2", "2:3"],
  "4K": ["16:9", "9:16"]
};
const openAISizeMap: Record<string, Record<string, string>> = {
  "1K": {
    "1:1": "1024x1024"
  },
  "2K": {
    "1:1": "2048x2048",
    "3:2": "1536x1024",
    "2:3": "1024x1536"
  },
  "4K": {
    "16:9": "3840x2160",
    "9:16": "2160x3840"
  }
};

const state = reactive({
  token: storedToken,
  user: null as User | null,
  models: [] as ImageModel[],
  videoModels: [] as VideoModel[],
  apiKeys: [] as ApiKey[],
  tasks: [] as ImageTask[],
  videoTasks: [] as VideoTask[],
  walletLogs: [] as WalletLog[],
  taskPage: 1,
  taskPageSize: 20,
  walletPage: 1,
  walletPageSize: 20,
  walletHasNext: false,
  result: null as GenerateImageResult | null,
  selectedTask: null as GenerateImageResult | null,
  videoResult: null as VideoTaskResult | null,
  selectedVideoTask: null as VideoTaskResult | null,
  previewAsset: null as { type: "image" | "video" | "audio"; url: string; title: string } | null,
  createdApiKey: "",
  profileMessage: "",
  loading: false,
  error: "",
  authMode: "login" as "login" | "register",
  activeView: "create" as "create" | "video" | "tasks" | "taskDetail" | "videoDetail" | "wallet" | "keys" | "docs" | "profile"
});

const menuViews = ["create", "video", "tasks", "wallet", "keys", "docs", "profile"] as const;
type MenuView = (typeof menuViews)[number];

type UploadReferenceResult = {
  url: string;
  filename: string;
  mime_type: string;
  size: number;
};

const videoReferenceLimits = {
  image: 4,
  video: 3,
  audio: 1
};
const videoSecondOptions = [5, 10, 15];

const authForm = reactive({
  email: "",
  password: "",
  nickname: ""
});

const generateForm = reactive({
  model: "general-high-quality",
  prompt: "一张科技感十足的 AI 芯片海报，蓝黑色背景，电影光效",
  resolution: "2K",
  aspect_ratio: "1:1",
  n: 1,
  reference_images: ""
});

const videoForm = reactive({
  model: "video-ds-2.0-fast",
  prompt: "A cinematic 9:16 video of a cat running through warm sunlight",
  seconds: 15,
  aspect_ratio: "9:16",
  images: "",
  videos: "",
  audios: ""
});

const keyForm = reactive({
  name: "default"
});

const profileForm = reactive({
  current_password: "",
  new_password: "",
  confirm_password: ""
});

const showPassword = ref(false);

const api = new ApiClient({
  getToken: () => state.token,
  onUnauthorized: () => logout()
});

async function bootstrap() {
  await Promise.all([loadModels(), loadVideoModels()]);
  if (state.token) {
    await Promise.allSettled([loadMe(), loadApiKeys(), loadTasks(), loadVideoTasks(), loadWalletLogs()]);
    await syncRouteFromURL();
  }
}

async function loadModels() {
  state.models = await api.get<ImageModel[]>("/api/models");
  if (state.models.length && !state.models.find((model) => model.code === generateForm.model)) {
    generateForm.model = state.models[0].code;
  }
}

async function loadVideoModels() {
  state.videoModels = await api.get<VideoModel[]>("/api/video-models");
  if (state.videoModels.length && !state.videoModels.find((model) => model.code === videoForm.model)) {
    videoForm.model = state.videoModels[0].code;
  }
}

async function loadMe() {
  state.user = await api.get<User>("/api/me");
}

async function loadApiKeys() {
  state.apiKeys = await api.get<ApiKey[]>("/api/api-keys");
}

async function loadTasks() {
  const limit = state.taskPage * state.taskPageSize + 1;
  state.tasks = await api.get<ImageTask[]>(`/api/images/tasks?limit=${limit}`);
}

async function loadTaskDetail(taskNo: string) {
  state.selectedTask = await api.get<GenerateImageResult>(`/api/images/tasks/${taskNo}`);
}

async function loadVideoTasks() {
  const limit = state.taskPage * state.taskPageSize + 1;
  state.videoTasks = await api.get<VideoTask[]>(`/api/videos/tasks?limit=${limit}`);
}

async function loadVideoTaskDetail(taskNo: string) {
  state.selectedVideoTask = await api.get<VideoTaskResult>(`/api/videos/tasks/${taskNo}`);
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
    return;
  }

  state.activeView = "create";
  state.selectedTask = null;
  state.selectedVideoTask = null;
}

async function loadWalletLogs() {
  const offset = (state.walletPage - 1) * state.walletPageSize;
  const logs = await api.get<WalletLog[]>(`/api/wallet/logs?limit=${state.walletPageSize + 1}&offset=${offset}`);
  state.walletHasNext = logs.length > state.walletPageSize;
  state.walletLogs = logs.slice(0, state.walletPageSize);
}

async function refreshCurrentView() {
  if (state.activeView === "profile") {
    await loadMe();
    return;
  }
  if (state.activeView === "tasks") {
    await Promise.all([loadTasks(), loadVideoTasks()]);
    return;
  }
  if (state.activeView === "wallet") {
    await loadWalletLogs();
    return;
  }
  if (state.activeView === "keys") {
    await loadApiKeys();
    return;
  }
  await loadMe();
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
    await Promise.all([loadApiKeys(), loadTasks(), loadVideoTasks(), loadWalletLogs()]);
    await syncRouteFromURL();
  });
}

async function generateImage() {
  await withLoading(async () => {
    state.result = await api.post<GenerateImageResult>("/api/images/generate", {
      model: generateForm.model,
      prompt: generateForm.prompt,
      size: requestSize(generateForm.resolution, generateForm.aspect_ratio),
      n: Number(generateForm.n),
      reference_images: buildReferenceImages()
    });
    state.selectedTask = state.result;
    openTaskDetail(state.result.task.task_no, false);
    await Promise.all([loadMe(), loadTasks(), loadWalletLogs()]);
  });
}

async function generateVideo() {
  await withLoading(async () => {
    validateVideoReferences();
    validateSelectedVideoModelReferences();
    state.videoResult = await api.post<VideoTaskResult>("/api/videos/generate", {
      model: videoForm.model,
      prompt: videoForm.prompt,
      seconds: Number(videoForm.seconds),
      aspect_ratio: videoForm.aspect_ratio,
      images: splitMediaList(videoForm.images),
      videos: splitMediaList(videoForm.videos),
      audios: splitMediaList(videoForm.audios)
    });
    state.selectedVideoTask = state.videoResult;
    openVideoTaskDetail(state.videoResult.task.task_no, false);
    await Promise.all([loadMe(), loadTasks(), loadVideoTasks(), loadWalletLogs()]);
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

async function changePassword() {
  await withLoading(async () => {
    if (profileForm.new_password !== profileForm.confirm_password) {
      throw new Error("两次输入的新密码不一致");
    }
    await api.post("/api/me/password", {
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
  state.user = null;
  state.apiKeys = [];
  state.tasks = [];
  state.videoTasks = [];
  state.walletLogs = [];
  state.selectedTask = null;
  state.selectedVideoTask = null;
  state.previewAsset = null;
  localStorage.removeItem(tokenKey);
}

const selectedModel = computed(() => state.models.find((model) => model.code === generateForm.model));
const selectedVideoModel = computed(() => state.videoModels.find((model) => model.code === videoForm.model));
const selectedVideoReferencePolicy = computed(() => selectedVideoModel.value?.reference_policy);
const selectedVideoRequiresOneImage = computed(() => selectedVideoReferencePolicy.value?.require_exactly_one_image === true);
const selectedVideoMaxImages = computed(() => selectedVideoReferencePolicy.value?.max_reference_images || 4);
const availableRatios = computed(() => openAIRatiosBySize[generateForm.resolution] ?? openAIRatiosBySize["2K"]);
const selectedRequestSize = computed(() => requestSize(generateForm.resolution, generateForm.aspect_ratio));
const combinedTasks = computed(() =>
  [
    ...state.tasks.map((task) => ({
      kind: "image" as const,
      id: task.id,
      task_no: task.task_no,
      status: task.status,
      meta: `${modelName(task.model_id)} · ${task.size || "-"} · ${task.num_images} 张`,
      credits_used: task.credits_used,
      created_at: task.created_at,
      prompt: task.error_message || task.prompt,
      detail_url: taskDetailURL(task.task_no)
    })),
    ...state.videoTasks.map((task) => ({
      kind: "video" as const,
      id: task.id,
      task_no: task.task_no,
      status: task.status,
      meta: `${videoModelName(task.model_id)} · ${task.aspect_ratio || "-"} · ${task.seconds || 0}s`,
      credits_used: task.credits_used,
      created_at: task.created_at,
      prompt: task.error_message || task.prompt,
      detail_url: videoTaskDetailURL(task.task_no)
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

async function changeTaskPage(delta: number) {
  const nextPage = state.taskPage + delta;
  if (nextPage < 1 || (delta > 0 && !taskHasNext.value)) return;
  state.taskPage = nextPage;
  await Promise.all([loadTasks(), loadVideoTasks()]);
}

async function changeWalletPage(delta: number) {
  const nextPage = state.walletPage + delta;
  if (nextPage < 1 || (delta > 0 && !state.walletHasNext)) return;
  state.walletPage = nextPage;
  await loadWalletLogs();
}

async function uploadImageReferenceFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) {
    return;
  }
  await withLoading(async () => {
    const uploaded = await uploadReferenceFile(file, "image");
    appendMediaValue("image", uploaded.url);
  });
}

async function uploadVideoReferenceFile(event: Event, kind: "image" | "video" | "audio") {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) {
    return;
  }
  await withLoading(async () => {
    const uploaded = await uploadReferenceFile(file, kind);
    appendMediaValue(kind, uploaded.url);
  });
}

async function uploadReferenceFile(file: File, kind: "image" | "video" | "audio") {
  const formData = new FormData();
  formData.set("kind", kind);
  formData.set("file", file);
  return api.upload<UploadReferenceResult>("/api/uploads/references", formData);
}

function appendMediaValue(kind: "image" | "video" | "audio", value: string) {
  const key = kind === "image" ? "images" : kind === "video" ? "videos" : "audios";
  if (kind === "image" && state.activeView === "create") {
    generateForm.reference_images = appendListValue(generateForm.reference_images, value);
    return;
  }
  ensureVideoReferenceCapacity(kind, 1);
  videoForm[key] = appendListValue(videoForm[key], value);
}

function appendListValue(current: string, value: string) {
  const trimmed = current.trim();
  return trimmed ? `${trimmed}\n${value}` : value;
}

function removeImageReference(index: number) {
  generateForm.reference_images = removeListItem(generateForm.reference_images, index);
}

function clearImageReferences() {
  generateForm.reference_images = "";
}

function removeVideoReference(kind: "image" | "video" | "audio", index: number) {
  const key = videoReferenceKey(kind);
  videoForm[key] = removeListItem(videoForm[key], index);
}

function clearVideoReferences(kind: "image" | "video" | "audio") {
  videoForm[videoReferenceKey(kind)] = "";
}

function removeListItem(value: string, index: number) {
  return splitMediaList(value)
    .filter((_, itemIndex) => itemIndex !== index)
    .join("\n");
}

function videoReferenceKey(kind: "image" | "video" | "audio") {
  return kind === "image" ? "images" : kind === "video" ? "videos" : "audios";
}

function validateVideoReferences() {
  ensureVideoReferenceLimit("image", splitMediaList(videoForm.images).length);
  ensureVideoReferenceLimit("video", splitMediaList(videoForm.videos).length);
  ensureVideoReferenceLimit("audio", splitMediaList(videoForm.audios).length);
}

function ensureVideoReferenceCapacity(kind: "image" | "video" | "audio", adding: number) {
  ensureVideoReferenceLimit(kind, splitMediaList(videoForm[videoReferenceKey(kind)]).length + adding);
}

function ensureVideoReferenceLimit(kind: "image" | "video" | "audio", count: number) {
  const limit = videoReferenceLimit(kind);
  if (count > limit) {
    const label = kind === "image" ? "参考图片" : kind === "video" ? "参考视频" : "参考音频";
    throw new Error(`${label}最多 ${limit} 个`);
  }
}

function videoReferenceLimit(kind: "image" | "video" | "audio") {
  const policy = selectedVideoReferencePolicy.value;
  if (kind === "image") return policy?.max_reference_images || videoReferenceLimits.image;
  if (kind === "video") return policy?.max_reference_videos || videoReferenceLimits.video;
  return policy?.max_reference_audios || videoReferenceLimits.audio;
}

function handleResolutionChange() {
  if (!availableRatios.value.includes(generateForm.aspect_ratio)) {
    generateForm.aspect_ratio = availableRatios.value[0] ?? "1:1";
  }
}

function requestSize(resolution: string, ratio: string): string {
  const sizeMap = openAISizeMap[resolution] ?? openAISizeMap["2K"];
  return sizeMap[ratio] ?? Object.values(sizeMap)[0] ?? "2048x2048";
}

function buildReferenceImages(): Array<{ url?: string; data_url?: string; filename?: string }> {
  return generateForm.reference_images
    .split(/\n|,/)
    .map((item) => item.trim())
    .filter(Boolean)
    .map((url) => ({ url }));
}

function splitMediaList(value: string): string[] {
  return value
    .split(/\n|,/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function videoModelOptionLabel(model: VideoModel): string {
  const suffix = model.reference_policy?.require_exactly_one_image ? " · 必须传 1 张参考图" : "";
  return `${model.display_name} · ${model.price_credits} 积分/次${suffix}`;
}

function validateSelectedVideoModelReferences() {
  if (!selectedVideoRequiresOneImage.value) return;
  if (splitMediaList(videoForm.images).length !== 1) {
    throw new Error("当前视频模型必须上传 1 张参考图");
  }
}

function activeViewTitle() {
  if (state.activeView === "create") return "AI 画图";
  if (state.activeView === "video") return "AI 视频";
  if (state.activeView === "tasks") return "任务记录";
  if (state.activeView === "taskDetail") return "任务详情";
  if (state.activeView === "videoDetail") return "视频详情";
  if (state.activeView === "wallet") return "积分流水";
  if (state.activeView === "keys") return "API Key";
  if (state.activeView === "profile") return "个人中心";
  return "API 文档";
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

function taskCanRefresh(status: string) {
  return status === "pending" || status === "running";
}

function modelName(modelID: number) {
  const model = state.models.find((item) => item.id === modelID);
  return model?.display_name || `#${modelID}`;
}

function videoModelName(modelID: number) {
  const model = state.videoModels.find((item) => item.id === modelID);
  return model?.display_name || `#${modelID}`;
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

function imageUrl(value: string) {
  if (!value || value.startsWith("http://") || value.startsWith("https://") || value.startsWith("data:")) {
    return value;
  }
  if (value.startsWith("/")) {
    return `${apiBaseURL}${value}`;
  }
  return value;
}

function openImagePreview(url: string, alt: string) {
  openAssetPreview("image", url, alt);
}

function openAssetPreview(type: "image" | "video" | "audio", url: string, title: string) {
  state.previewAsset = { type, url: imageUrl(url), title };
}

function closeAssetPreview() {
  state.previewAsset = null;
}

function openTaskDetail(taskNo: string, replace = false) {
  state.activeView = "taskDetail";
  updateBrowserURL(taskDetailURL(taskNo), replace);
}

function openVideoTaskDetail(taskNo: string, replace = false) {
  state.activeView = "videoDetail";
  updateBrowserURL(videoTaskDetailURL(taskNo), replace);
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
}

function isMenuView(view: string | null): view is MenuView {
  return menuViews.includes(view as MenuView);
}

function menuViewURL(view: MenuView) {
  return `${window.location.origin}${window.location.pathname}?view=${encodeURIComponent(view)}`;
}

function taskDetailURL(taskNo: string) {
  return `${window.location.origin}${window.location.pathname}?view=task&task_no=${encodeURIComponent(taskNo)}`;
}

function videoTaskDetailURL(taskNo: string) {
  return `${window.location.origin}${window.location.pathname}?view=video_task&task_no=${encodeURIComponent(taskNo)}`;
}

const App = {
  components: {
    ArrowDownToLine,
    Brush,
    Code2,
    Eye,
    EyeOff,
    Film,
    Image,
    KeyRound,
    ListChecks,
    Loader2,
    LogOut,
    ReceiptText,
    RefreshCw,
    ShieldCheck,
    Sparkles,
    X,
    UserCircle
  },
  setup() {
    bootstrap();
    return {
      state,
      authForm,
      generateForm,
      videoForm,
      keyForm,
      profileForm,
      showPassword,
      selectedModel,
      selectedVideoModel,
      selectedVideoRequiresOneImage,
      selectedVideoMaxImages,
      combinedTasks,
      pagedTasks,
      taskHasNext,
      taskRangeText,
      walletRangeText,
      commonSizes,
      videoSecondOptions,
      availableRatios,
      selectedRequestSize,
      apiBaseURL,
      formatCredits,
      parseJSONList,
      loadMe,
      loadTasks,
      loadTaskDetail,
      loadVideoTasks,
      loadVideoTaskDetail,
      loadWalletLogs,
      changeTaskPage,
      changeWalletPage,
      refreshCurrentView,
      activeViewTitle,
      taskStatusText,
      taskCanRefresh,
      modelName,
      videoModelName,
      videoModelOptionLabel,
      walletTypeText,
      signedCredits,
      formatDateTime,
      imageUrl,
      taskDetailURL,
      videoTaskDetailURL,
      setActiveView,
      splitMediaList,
      openImagePreview,
      openAssetPreview,
      closeAssetPreview,
      submitAuth,
      generateImage,
      generateVideo,
      createApiKey,
      revokeApiKey,
      changePassword,
      uploadImageReferenceFile,
      uploadVideoReferenceFile,
      removeImageReference,
      clearImageReferences,
      removeVideoReference,
      clearVideoReferences,
      handleResolutionChange,
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

        <button class="nav-item" :class="{ active: state.activeView === 'create' }" @click="setActiveView('create')">
          <Brush :size="18" /> AI 画图
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'video' }" @click="setActiveView('video')">
          <Film :size="18" /> AI 视频
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'tasks' }" @click="setActiveView('tasks')">
          <ListChecks :size="18" /> 任务记录
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'wallet' }" @click="setActiveView('wallet')">
          <ReceiptText :size="18" /> 积分流水
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'keys' }" @click="setActiveView('keys')">
          <KeyRound :size="18" /> API Key
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'docs' }" @click="setActiveView('docs')">
          <Code2 :size="18" /> API 文档
        </button>
        <button class="nav-item" :class="{ active: state.activeView === 'profile' }" @click="setActiveView('profile')">
          <UserCircle :size="18" /> 个人中心
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
            <label>密码
              <span class="password-field">
                <input v-model="authForm.password" :type="showPassword ? 'text' : 'password'" />
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
            <label v-if="state.authMode === 'register'">昵称<input v-model="authForm.nickname" /></label>
            <button class="primary-button" @click="submitAuth" :disabled="state.loading">
              <Loader2 v-if="state.loading" class="spin" :size="17" /> {{ state.authMode === 'login' ? '登录' : '创建账号' }}
            </button>
            <p v-if="state.error" class="error">{{ state.error }}</p>
          </div>
        </div>

        <template v-else>
          <header class="topbar">
            <div>
              <h1>{{ activeViewTitle() }}</h1>
              <p>{{ state.user.email }} · {{ formatCredits(state.user.credits) }}</p>
            </div>
            <button class="ghost-button" @click="refreshCurrentView"><RefreshCw :size="16" />刷新</button>
          </header>

          <p v-if="state.error" class="error">{{ state.error }}</p>

          <div v-if="state.activeView === 'create'" class="workspace studio-workspace image-studio">
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
              <label>参考图上传
                <span class="upload-row">
                  <input type="file" accept="image/*" @change="uploadImageReferenceFile" />
                </span>
              </label>
              <div v-if="splitMediaList(generateForm.reference_images).length" class="reference-list">
                <div class="reference-list-heading">
                  <span>已上传参考图 {{ splitMediaList(generateForm.reference_images).length }} 张</span>
                  <button class="ghost-button small-button" type="button" @click="clearImageReferences">清空</button>
                </div>
                <div v-for="(item, index) in splitMediaList(generateForm.reference_images)" :key="item" class="reference-item">
                  <span>第 {{ index + 1 }} 个</span>
                  <div class="reference-actions">
                    <button class="ghost-button small-button" type="button" @click="openAssetPreview('image', item, '参考图 ' + (index + 1))">预览</button>
                    <button class="ghost-button small-button" type="button" @click="removeImageReference(index)">删除</button>
                  </div>
                </div>
              </div>
              <div class="inline-grid">
                <label>清晰度
                  <select v-model="generateForm.resolution" @change="handleResolutionChange">
                    <option v-for="size in commonSizes" :key="size" :value="size">{{ size }}</option>
                  </select>
                </label>
                <label>比例
                  <select v-model="generateForm.aspect_ratio">
                    <option v-for="ratio in availableRatios" :key="ratio" :value="ratio">{{ ratio }}</option>
                  </select>
                </label>
                <label>数量
                  <input v-model.number="generateForm.n" type="number" min="1" :max="selectedModel?.max_images_per_request || 4" />
                </label>
              </div>
              <div class="request-size-line">请求尺寸：{{ selectedRequestSize }}</div>
              <button class="primary-button" @click="generateImage" :disabled="state.loading">
                <Loader2 v-if="state.loading" class="spin" :size="17" /> 提交任务
              </button>
            </section>

            <section class="panel result-panel">
              <div class="empty-result" v-if="!state.result">
                <Image :size="40" />
                <span>提交后可在任务记录查看进度和结果</span>
              </div>
              <div v-else>
                <div class="task-line">
                  <strong>{{ state.result.task.task_no }}</strong>
                  <span>{{ taskStatusText(state.result.task.status) }} · 消耗 {{ state.result.task.credits_used }} 积分</span>
                </div>
                <div v-if="!state.result.images.length" class="empty-result compact-empty">
                  <Loader2 v-if="taskCanRefresh(state.result.task.status)" class="spin" :size="28" />
                  <span>{{ taskCanRefresh(state.result.task.status) ? '任务已提交，后台正在生成' : '暂无图片' }}</span>
                </div>
                <div v-else class="image-grid">
                  <article v-for="image in state.result.images" :key="image.id" class="image-card">
                    <button class="image-preview-button" type="button" @click="openImagePreview(image.url, state.result.task.prompt)">
                      <img :src="imageUrl(image.url)" :alt="state.result.task.prompt" />
                    </button>
                    <a :href="imageUrl(image.url)" target="_blank" rel="noreferrer"><ArrowDownToLine :size="16" /> 打开图片</a>
                  </article>
                </div>
              </div>
            </section>
          </div>

          <div v-if="state.activeView === 'video'" class="workspace studio-workspace video-studio">
            <section class="panel controls">
              <label>模型
                <select v-model="videoForm.model">
                  <option v-for="model in state.videoModels" :key="model.id" :value="model.code">
                    {{ videoModelOptionLabel(model) }}
                  </option>
                </select>
              </label>
              <div class="model-hint" v-if="selectedVideoModel">
                <strong>
                  {{ selectedVideoModel.display_name }}
                  <span v-if="selectedVideoRequiresOneImage" class="required-badge">必须传 1 张参考图</span>
                </strong>
                <span>{{ selectedVideoModel.description }}</span>
                <span v-if="selectedVideoRequiresOneImage" class="danger-hint">这个模型需要上传且只能上传 1 张参考图片。</span>
              </div>
              <label>提示词
                <textarea v-model="videoForm.prompt" rows="7" />
              </label>
              <div class="inline-grid">
                <label>时长
                  <select v-model.number="videoForm.seconds">
                    <option v-for="seconds in videoSecondOptions" :key="seconds" :value="seconds">{{ seconds }} 秒</option>
                  </select>
                </label>
                <label>比例
                  <select v-model="videoForm.aspect_ratio">
                    <option value="9:16">9:16</option>
                    <option value="16:9">16:9</option>
                    <option value="1:1">1:1</option>
                  </select>
                </label>
              </div>
              <label>参考图片
                <span class="limit-line">已上传 {{ splitMediaList(videoForm.images).length }}/{{ selectedVideoMaxImages }}</span>
                <span class="upload-row">
                  <input type="file" accept="image/*" :disabled="splitMediaList(videoForm.images).length >= selectedVideoMaxImages" @change="uploadVideoReferenceFile($event, 'image')" />
                </span>
                <div v-if="splitMediaList(videoForm.images).length" class="reference-list">
                  <div class="reference-list-heading">
                    <span>{{ splitMediaList(videoForm.images).length }} 个参考图片已就绪</span>
                    <button class="ghost-button small-button" type="button" @click="clearVideoReferences('image')">清空</button>
                  </div>
                  <div v-for="(item, index) in splitMediaList(videoForm.images)" :key="item" class="reference-item">
                    <span>第 {{ index + 1 }} 个</span>
                    <div class="reference-actions">
                      <button class="ghost-button small-button" type="button" @click="openAssetPreview('image', item, '参考图片 ' + (index + 1))">预览</button>
                      <button class="ghost-button small-button" type="button" @click="removeVideoReference('image', index)">删除</button>
                    </div>
                  </div>
                </div>
              </label>
              <label>参考视频
                <span class="limit-line">已上传 {{ splitMediaList(videoForm.videos).length }}/3</span>
                <span class="upload-row">
                  <input type="file" accept="video/*" :disabled="splitMediaList(videoForm.videos).length >= 3" @change="uploadVideoReferenceFile($event, 'video')" />
                </span>
                <div v-if="splitMediaList(videoForm.videos).length" class="reference-list">
                  <div class="reference-list-heading">
                    <span>{{ splitMediaList(videoForm.videos).length }} 个参考视频已就绪</span>
                    <button class="ghost-button small-button" type="button" @click="clearVideoReferences('video')">清空</button>
                  </div>
                  <div v-for="(item, index) in splitMediaList(videoForm.videos)" :key="item" class="reference-item">
                    <span>第 {{ index + 1 }} 个</span>
                    <div class="reference-actions">
                      <button class="ghost-button small-button" type="button" @click="openAssetPreview('video', item, '参考视频 ' + (index + 1))">预览</button>
                      <button class="ghost-button small-button" type="button" @click="removeVideoReference('video', index)">删除</button>
                    </div>
                  </div>
                </div>
              </label>
              <label>参考音频
                <span class="limit-line">已上传 {{ splitMediaList(videoForm.audios).length }}/1</span>
                <span class="upload-row">
                  <input type="file" accept="audio/*" :disabled="splitMediaList(videoForm.audios).length >= 1" @change="uploadVideoReferenceFile($event, 'audio')" />
                </span>
                <div v-if="splitMediaList(videoForm.audios).length" class="reference-list">
                  <div class="reference-list-heading">
                    <span>{{ splitMediaList(videoForm.audios).length }} 个参考音频已就绪</span>
                    <button class="ghost-button small-button" type="button" @click="clearVideoReferences('audio')">清空</button>
                  </div>
                  <div v-for="(item, index) in splitMediaList(videoForm.audios)" :key="item" class="reference-item">
                    <span>第 {{ index + 1 }} 个</span>
                    <div class="reference-actions">
                      <button class="ghost-button small-button" type="button" @click="openAssetPreview('audio', item, '参考音频 ' + (index + 1))">预览</button>
                      <button class="ghost-button small-button" type="button" @click="removeVideoReference('audio', index)">删除</button>
                    </div>
                  </div>
                </div>
              </label>
              <button class="primary-button" @click="generateVideo" :disabled="state.loading">
                <Loader2 v-if="state.loading" class="spin" :size="17" /> 提交视频任务
              </button>
            </section>

            <section class="panel result-panel">
              <div class="empty-result" v-if="!state.videoResult">
                <Film :size="40" />
                <span>提交后可在任务记录查看进度和结果</span>
              </div>
              <div v-else>
                <div class="task-line">
                  <strong>{{ state.videoResult.task.task_no }}</strong>
                  <span>{{ taskStatusText(state.videoResult.task.status) }} · 消耗 {{ state.videoResult.task.credits_used }} 积分</span>
                </div>
                <div class="empty-result compact-empty">
                  <Loader2 v-if="taskCanRefresh(state.videoResult.task.status)" class="spin" :size="28" />
                  <span>{{ taskCanRefresh(state.videoResult.task.status) ? '视频任务已提交，后台正在生成' : '请到任务记录查看结果' }}</span>
                </div>
              </div>
            </section>
          </div>

          <section v-if="state.activeView === 'tasks'" class="panel list-panel">
            <div class="panel-heading">
              <div>
                <h2>我的生成任务</h2>
                <p>这里会记录你在网页端发起的图片和视频生成任务，包括状态、消耗积分和请求参数。</p>
              </div>
              <button class="ghost-button" @click="refreshCurrentView"><RefreshCw :size="16" />刷新</button>
            </div>
            <p class="retention-warning">视频和图片资源只保留 3 天，请提前下载到本地存储。</p>
            <div v-if="!combinedTasks.length" class="empty-state">
              <ListChecks :size="36" />
              <span>还没有任务记录</span>
            </div>
            <div v-else class="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>类型</th>
                    <th>任务号</th>
                    <th>状态</th>
                    <th>参数</th>
                    <th>积分</th>
                    <th>创建时间</th>
                    <th>提示词</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="task in pagedTasks" :key="task.kind + '-' + task.id">
                    <td>{{ task.kind === 'image' ? '图片' : '视频' }}</td>
                    <td class="mono-cell">{{ task.task_no }}</td>
                    <td><span class="status-pill" :class="task.status">{{ taskStatusText(task.status) }}</span></td>
                    <td>{{ task.meta }}</td>
                    <td>{{ formatCredits(task.credits_used) }}</td>
                    <td class="muted-cell">{{ formatDateTime(task.created_at) }}</td>
                    <td class="prompt-cell" :title="task.prompt">{{ task.prompt }}</td>
                    <td><a class="ghost-button small-button" :href="task.detail_url" target="_blank" rel="noreferrer">查看</a></td>
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

          <section v-if="state.activeView === 'taskDetail'" class="panel list-panel">
            <div class="panel-heading">
              <div>
                <h2>任务详情</h2>
                <p v-if="state.selectedTask">{{ state.selectedTask.task.task_no }} · {{ taskStatusText(state.selectedTask.task.status) }}</p>
              </div>
              <button v-if="state.selectedTask" class="ghost-button" @click="loadTaskDetail(state.selectedTask.task.task_no)"><RefreshCw :size="16" />刷新任务</button>
            </div>
            <div v-if="!state.selectedTask" class="empty-state compact-empty">
              <Loader2 class="spin" :size="28" />
              <span>正在加载任务</span>
            </div>
            <div v-else class="task-detail standalone-detail">
              <div class="task-line">
                <strong>{{ state.selectedTask.task.task_no }}</strong>
                <span>{{ taskStatusText(state.selectedTask.task.status) }} · {{ state.selectedTask.task.progress }}% · 消耗 {{ state.selectedTask.task.credits_used }} 积分</span>
              </div>
              <p class="prompt-detail">{{ state.selectedTask.task.prompt }}</p>
              <p v-if="state.selectedTask.task.error_message" class="error">{{ state.selectedTask.task.error_message }}</p>
              <div v-if="!state.selectedTask.images.length" class="empty-state compact-empty">
                <Loader2 v-if="taskCanRefresh(state.selectedTask.task.status)" class="spin" :size="28" />
                <span>{{ taskCanRefresh(state.selectedTask.task.status) ? '后台生成中，稍后刷新查看' : '这个任务没有图片结果' }}</span>
              </div>
              <div v-else class="image-grid">
                <article v-for="image in state.selectedTask.images" :key="image.id" class="image-card">
                  <button class="image-preview-button" type="button" @click="openImagePreview(image.url, state.selectedTask.task.prompt)">
                    <img :src="imageUrl(image.url)" :alt="state.selectedTask.task.prompt" />
                  </button>
                  <a :href="imageUrl(image.url)" target="_blank" rel="noreferrer"><ArrowDownToLine :size="16" /> 打开图片</a>
                </article>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'videoDetail'" class="panel list-panel">
            <div class="panel-heading">
              <div>
                <h2>视频详情</h2>
                <p v-if="state.selectedVideoTask">{{ state.selectedVideoTask.task.task_no }} · {{ taskStatusText(state.selectedVideoTask.task.status) }}</p>
              </div>
              <button v-if="state.selectedVideoTask" class="ghost-button" @click="loadVideoTaskDetail(state.selectedVideoTask.task.task_no)"><RefreshCw :size="16" />刷新任务</button>
            </div>
            <div v-if="!state.selectedVideoTask" class="empty-state compact-empty">
              <Loader2 class="spin" :size="28" />
              <span>正在加载视频任务</span>
            </div>
            <div v-else class="task-detail standalone-detail">
              <div class="task-line">
                <strong>{{ state.selectedVideoTask.task.task_no }}</strong>
                <span>{{ taskStatusText(state.selectedVideoTask.task.status) }} · {{ state.selectedVideoTask.task.progress }}% · 消耗 {{ state.selectedVideoTask.task.credits_used }} 积分</span>
              </div>
              <p class="prompt-detail">{{ state.selectedVideoTask.task.prompt }}</p>
              <p v-if="state.selectedVideoTask.task.error_message" class="error">{{ state.selectedVideoTask.task.error_message }}</p>
              <div v-if="!state.selectedVideoTask.videos.length" class="empty-state compact-empty">
                <Loader2 v-if="taskCanRefresh(state.selectedVideoTask.task.status)" class="spin" :size="28" />
                <span>{{ taskCanRefresh(state.selectedVideoTask.task.status) ? '后台生成中，稍后刷新查看' : '这个任务没有视频结果' }}</span>
              </div>
              <div v-else class="video-result-list">
                <article v-for="video in state.selectedVideoTask.videos" :key="video.id" class="video-card">
                  <video :src="imageUrl(video.url)" controls playsinline />
                  <a :href="imageUrl(video.url)" target="_blank" rel="noreferrer"><ArrowDownToLine :size="16" /> 下载/打开 MP4</a>
                </article>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'wallet'" class="panel list-panel">
            <div class="panel-heading">
              <div>
                <h2>积分流水</h2>
                <p>这里会记录积分增加、扣减、生成消费和失败退款。</p>
              </div>
              <button class="ghost-button" @click="loadWalletLogs"><RefreshCw :size="16" />刷新</button>
            </div>
            <div v-if="!state.walletLogs.length" class="empty-state">
              <ReceiptText :size="36" />
              <span>还没有积分记录</span>
            </div>
            <div v-else class="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>类型</th>
                    <th>变动</th>
                    <th>变动前</th>
                    <th>变动后</th>
                    <th>关联</th>
                    <th>备注</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="log in state.walletLogs" :key="log.id">
                    <td class="muted-cell">{{ formatDateTime(log.created_at) }}</td>
                    <td>{{ walletTypeText(log.type) }}</td>
                    <td class="amount-cell" :class="{ positive: log.amount > 0, negative: log.amount < 0 }">{{ signedCredits(log.amount) }}</td>
                    <td>{{ formatCredits(log.balance_before) }}</td>
                    <td>{{ formatCredits(log.balance_after) }}</td>
                    <td>{{ log.related_type || '-' }}<span v-if="log.related_id"> #{{ log.related_id }}</span></td>
                    <td class="prompt-cell" :title="log.remark">{{ log.remark || '-' }}</td>
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
            </div>
          </section>

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

          <section v-if="state.activeView === 'docs'" class="docs-page">
            <div class="docs-hero">
              <div>
                <h2>API 文档</h2>
                <p>使用 API Key 调用平台图片和视频生成能力，接口风格兼容 OpenAI 常见调用方式。</p>
              </div>
              <a class="ghost-button" href="#api-quickstart">快速开始</a>
            </div>

            <div class="docs-layout">
              <aside class="docs-toc">
                <a href="#api-quickstart">快速开始</a>
                <a href="#api-auth">鉴权</a>
                <a href="#api-image">图片生成</a>
                <a href="#api-video">视频生成</a>
                <a href="#api-errors">错误处理</a>
              </aside>

              <div class="docs-content">
                <article id="api-quickstart" class="doc-section">
                  <h3>快速开始</h3>
                  <p>先在“API Key”页面创建密钥，然后在请求头中使用 Bearer Token。</p>
                  <pre><code>curl {{ apiBaseURL }}/health</code></pre>
                </article>

                <article id="api-auth" class="doc-section">
                  <h3>鉴权</h3>
                  <p>所有外部开放接口都需要传入 API Key。</p>
                  <div class="param-table">
                    <div><strong>Header</strong><span>Authorization</span></div>
                    <div><strong>格式</strong><span>Bearer agi_xxx</span></div>
                    <div><strong>Content-Type</strong><span>application/json</span></div>
                  </div>
                  <pre><code>Authorization: Bearer agi_xxx
Content-Type: application/json</code></pre>
                </article>

                <article id="api-image" class="doc-section">
                  <h3>图片生成</h3>
                  <p>提交图片生成请求，响应会返回平台任务和生成结果。参考图建议使用公网可访问 URL。</p>
                  <div class="endpoint-line"><span>POST</span><code>/v1/images/generations</code></div>
                  <div class="param-table">
                    <div><strong>model</strong><span>必填，模型编码</span></div>
                    <div><strong>prompt</strong><span>必填，图片提示词</span></div>
                    <div><strong>size</strong><span>如 1024x1024、2048x2048、1536x1024</span></div>
                    <div><strong>n</strong><span>生成张数，受模型配置限制</span></div>
                    <div><strong>reference_images</strong><span>可选，参考图 URL 或对象数组</span></div>
                  </div>
                  <pre><code>curl -X POST {{ apiBaseURL }}/v1/images/generations \\
  -H "Authorization: Bearer agi_xxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-2",
    "prompt": "一张科技感十足的 AI 芯片海报，蓝黑色背景，电影光效",
    "size": "2048x2048",
    "n": 1,
    "reference_images": []
  }'</code></pre>
                </article>

                <article id="api-video" class="doc-section">
                  <h3>视频生成</h3>
                  <p>视频是异步任务。先创建任务，再轮询状态；成功后可通过 content 接口播放或下载 mp4。</p>
                  <div class="endpoint-stack">
                    <div class="endpoint-line"><span>POST</span><code>/v1/videos</code></div>
                    <div class="endpoint-line"><span>GET</span><code>/v1/videos/{task_id}</code></div>
                    <div class="endpoint-line"><span>GET</span><code>/v1/videos/{task_id}/content</code></div>
                  </div>
                  <div class="param-table">
                    <div><strong>model</strong><span>必填，视频模型编码</span></div>
                    <div><strong>prompt</strong><span>必填，视频提示词</span></div>
                    <div><strong>seconds</strong><span>支持 5、10、15 秒</span></div>
                    <div><strong>aspect_ratio</strong><span>9:16、16:9、1:1</span></div>
                    <div><strong>images/videos/audios</strong><span>可选，参考素材 URL 数组</span></div>
                  </div>
                  <pre><code>curl -X POST {{ apiBaseURL }}/v1/videos \\
  -H "Authorization: Bearer agi_xxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "video-ds-2.0-fast",
    "prompt": "A cinematic 9:16 video of a cat running through warm sunlight",
    "seconds": 15,
    "aspect_ratio": "9:16",
    "images": [],
    "videos": [],
    "audios": []
  }'

curl {{ apiBaseURL }}/v1/videos/{task_id} \\
  -H "Authorization: Bearer agi_xxx"

curl -L {{ apiBaseURL }}/v1/videos/{task_id}/content \\
  -H "Authorization: Bearer agi_xxx" \\
  -o result.mp4</code></pre>
                </article>

                <article id="api-errors" class="doc-section">
                  <h3>错误处理</h3>
                  <p>客户端应同时判断 HTTP 状态码和响应体里的 code/message。</p>
                  <div class="param-table">
                    <div><strong>401</strong><span>API Key 缺失或无效</span></div>
                    <div><strong>402</strong><span>积分不足</span></div>
                    <div><strong>404</strong><span>任务或模型不存在</span></div>
                    <div><strong>429</strong><span>请求过快或上游限流</span></div>
                    <div><strong>500/502</strong><span>平台或上游生成失败</span></div>
                  </div>
                  <pre><code>{
  "code": 402,
  "message": "insufficient credits"
}</code></pre>
                </article>
              </div>
            </div>
          </section>

          <section v-if="state.activeView === 'profile'" class="panel list-panel">
            <div class="panel-heading">
              <div>
                <h2>个人中心</h2>
                <p>查看账号信息并修改登录密码。</p>
              </div>
              <button class="ghost-button" @click="loadMe"><RefreshCw :size="16" />刷新</button>
            </div>
            <div class="profile-layout">
              <div class="profile-card">
                <UserCircle :size="38" />
                <div>
                  <strong>{{ state.user.nickname }}</strong>
                  <span>{{ state.user.email || '-' }}</span>
                </div>
                <div class="profile-meta">
                  <span>积分余额</span>
                  <strong>{{ formatCredits(state.user.credits) }}</strong>
                </div>
                <div class="profile-meta">
                  <span>账号状态</span>
                  <strong>{{ state.user.status }}</strong>
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
        </template>
      </section>

      <div v-if="state.previewAsset" class="preview-backdrop" @click.self="closeAssetPreview">
        <div class="preview-panel">
          <button class="preview-close" type="button" aria-label="关闭预览" @click="closeAssetPreview"><X :size="20" /></button>
          <img v-if="state.previewAsset.type === 'image'" :src="state.previewAsset.url" :alt="state.previewAsset.title" />
          <video v-else-if="state.previewAsset.type === 'video'" :src="state.previewAsset.url" controls autoplay playsinline />
          <div v-else class="audio-preview">
            <strong>{{ state.previewAsset.title }}</strong>
            <audio :src="state.previewAsset.url" controls autoplay />
          </div>
        </div>
      </div>
    </main>
  `
};

createApp(App).mount("#app");

window.addEventListener("popstate", () => {
  if (state.token) {
    void syncRouteFromURL();
  }
});
