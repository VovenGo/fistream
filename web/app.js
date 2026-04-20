const state = {
  access: null,
  activeTab: "create",
  jitsiApi: null,
};

const MESSAGES = {
  serverError: "Ошибка сервера",
  jitsiLoadError: "Не удалось загрузить Jitsi API",
  signInError: "Ошибка входа",
  copied: "Ссылка скопирована",
  roomClosed: "Комната закрыта",
  closeRoomError: "Не удалось закрыть комнату",
};

const TRANSPARENT_LOGO_DATA_URI =
  "data:image/gif;base64,R0lGODlhAQABAAAAACwAAAAAAQABAAA=";

const JITSI_TOOLBAR_BUTTONS = [
  "microphone",
  "camera",
  "desktop",
  "chat",
  "raisehand",
  "tileview",
  "fullscreen",
  "settings",
  "select-background",
];

const els = {
  landing: document.querySelector("#landing"),
  roomView: document.querySelector("#room-view"),
  createForm: document.querySelector("#create-form"),
  joinForm: document.querySelector("#join-form"),
  tabs: document.querySelectorAll(".tab"),
  roomCode: document.querySelector("#room-code"),
  copyRoomLink: document.querySelector("#copy-room-link"),
  closeRoom: document.querySelector("#close-room"),
  leaveRoom: document.querySelector("#leave-room"),
  jitsiContainer: document.querySelector("#jitsi-container"),
  toast: document.querySelector("#toast"),
};

const normalizeDomain = (value) => {
  const raw = String(value || "").trim();
  if (!raw) {
    return "";
  }
  if (raw.startsWith("http://") || raw.startsWith("https://")) {
    return new URL(raw).host;
  }
  return raw;
};

const showToast = (message) => {
  if (!message) {
    return;
  }
  els.toast.textContent = message;
  els.toast.classList.remove("hidden");
  setTimeout(() => {
    els.toast.classList.add("hidden");
  }, 3200);
};

const switchTab = (tab) => {
  state.activeTab = tab;
  els.tabs.forEach((button) => {
    button.classList.toggle("tab--active", button.dataset.tab === tab);
  });
  els.createForm.classList.toggle("hidden", tab !== "create");
  els.joinForm.classList.toggle("hidden", tab !== "join");
};

const extractError = async (response) => {
  try {
    const payload = await response.json();
    return payload.message || payload.error || MESSAGES.serverError;
  } catch {
    return MESSAGES.serverError;
  }
};

const apiRequest = async (path, body, options = {}) => {
  const response = await fetch(path, {
    method: options.method || "POST",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    throw new Error(await extractError(response));
  }

  return response.json();
};

const loadJitsiScript = (domain) =>
  new Promise((resolve, reject) => {
    const normalizedDomain = normalizeDomain(domain);
    if (!normalizedDomain) {
      reject(new Error("Jitsi domain is not configured"));
      return;
    }

    const src = `${window.location.protocol}//${normalizedDomain}/external_api.js`;
    const existing = document.querySelector(`script[src="${src}"]`);
    if (existing) {
      if (window.JitsiMeetExternalAPI) {
        resolve();
      } else {
        existing.addEventListener("load", () => resolve(), { once: true });
        existing.addEventListener("error", () => reject(new Error(MESSAGES.jitsiLoadError)), {
          once: true,
        });
      }
      return;
    }

    const script = document.createElement("script");
    script.src = src;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(MESSAGES.jitsiLoadError));
    document.head.appendChild(script);
  });

const buildJitsiOptions = (jitsi) => ({
  roomName: jitsi.room_name,
  parentNode: els.jitsiContainer,
  width: "100%",
  height: "100%",
  jwt: jitsi.jwt,
  userInfo: {
    displayName: jitsi.display_name,
  },
  configOverwrite: {
    prejoinPageEnabled: false,
    prejoinConfig: {
      enabled: false,
    },
    disableDeepLinking: true,
    disablePolls: true,
    disableInviteFunctions: true,
    notifications: [],
    disableProfile: true,
    disableShortcuts: true,
    disabledNotifications: ["notify.moderator"],
    participantsPane: {
      enabled: false,
    },
    toolbarButtons: JITSI_TOOLBAR_BUTTONS,
    defaultLogoUrl: TRANSPARENT_LOGO_DATA_URI,
    defaultWelcomePageLogoUrl: TRANSPARENT_LOGO_DATA_URI,
  },
  interfaceConfigOverwrite: {
    APP_NAME: "fistream",
    HIDE_INVITE_MORE_HEADER: true,
    SHOW_JITSI_WATERMARK: false,
    SHOW_WATERMARK_FOR_GUESTS: false,
    SHOW_BRAND_WATERMARK: false,
    JITSI_WATERMARK_LINK: "",
    BRAND_WATERMARK_LINK: "",
    SETTINGS_SECTIONS: ["devices", "language", "calendar", "more"],
  },
});

const mountJitsi = async (jitsi) => {
  if (state.jitsiApi) {
    state.jitsiApi.dispose();
    state.jitsiApi = null;
  }

  await loadJitsiScript(jitsi.domain);

  const domain = normalizeDomain(jitsi.domain);
  state.jitsiApi = new window.JitsiMeetExternalAPI(domain, buildJitsiOptions(jitsi));
};

const enterRoom = async (access) => {
  state.access = access;
  els.roomCode.textContent = access.room_code;
  els.closeRoom.classList.toggle("hidden", access.role !== "host");
  els.landing.classList.add("hidden");
  els.roomView.classList.remove("hidden");
  await mountJitsi(access.jitsi);
};

const leaveRoomView = () => {
  if (state.jitsiApi) {
    state.jitsiApi.dispose();
    state.jitsiApi = null;
  }

  state.access = null;
  els.roomView.classList.add("hidden");
  els.landing.classList.remove("hidden");
  window.history.pushState({}, "", "/");
};

const handleCreateRoom = async (event) => {
  event.preventDefault();

  const form = new FormData(event.currentTarget);
  const payload = {
    display_name: String(form.get("display_name") || "").trim(),
    service_password: String(form.get("service_password") || "").trim(),
  };

  try {
    const access = await apiRequest("/api/v1/rooms/create", payload);
    window.history.pushState({}, "", `/room/${access.room_code}`);
    await enterRoom(access);
  } catch (error) {
    showToast(error.message || MESSAGES.signInError);
  }
};

const handleJoinRoom = async (event) => {
  event.preventDefault();

  const form = new FormData(event.currentTarget);
  const payload = {
    display_name: String(form.get("display_name") || "").trim(),
    room_code: String(form.get("room_code") || "").trim().toUpperCase(),
    service_password: String(form.get("service_password") || "").trim(),
  };

  try {
    const access = await apiRequest("/api/v1/rooms/join", payload);
    window.history.pushState({}, "", `/room/${access.room_code}`);
    await enterRoom(access);
  } catch (error) {
    showToast(error.message || MESSAGES.signInError);
  }
};

const handleCopyRoomLink = async () => {
  if (!state.access?.room_code) {
    return;
  }
  await navigator.clipboard.writeText(`${window.location.origin}/room/${state.access.room_code}`);
  showToast(MESSAGES.copied);
};

const handleCloseRoom = async () => {
  if (!state.access) {
    return;
  }

  try {
    await apiRequest(`/api/v1/rooms/${state.access.room_code}/close`, null, {
      headers: {
        Authorization: `Bearer ${state.access.api_token}`,
      },
    });
    showToast(MESSAGES.roomClosed);
    leaveRoomView();
  } catch (error) {
    showToast(error.message || MESSAGES.closeRoomError);
  }
};

const attachEventHandlers = () => {
  els.tabs.forEach((button) => {
    button.addEventListener("click", () => switchTab(button.dataset.tab));
  });

  els.createForm.addEventListener("submit", handleCreateRoom);
  els.joinForm.addEventListener("submit", handleJoinRoom);
  els.copyRoomLink.addEventListener("click", handleCopyRoomLink);
  els.closeRoom.addEventListener("click", handleCloseRoom);
  els.leaveRoom.addEventListener("click", leaveRoomView);
};

const boot = () => {
  const roomMatch = window.location.pathname.match(/^\/room\/([A-Za-z0-9]+)$/);
  if (roomMatch) {
    switchTab("join");
    els.joinForm.elements.room_code.value = roomMatch[1].toUpperCase();
  }
};

attachEventHandlers();
boot();
