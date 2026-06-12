<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { UserOutlined, LockOutlined, KeyOutlined, SettingOutlined, DesktopOutlined } from '@ant-design/icons-vue';

import { HttpUtil, LanguageManager } from '@/utils';
import {
  antdThemeConfig,
  currentTheme,
  theme as themeState,
  cycleThemeMode,
  pauseAnimationsUntilLeave,
} from '@/composables/useTheme.js';

const { t } = useI18n();

const fetched = ref(false);
const submitting = ref(false);
const twoFactorEnable = ref(false);

const user = reactive({
  username: '',
  password: '',
  twoFactorCode: '',
});

const basePath = window.X_UI_BASE_PATH || '';

const headlineWords = computed(() => [t('pages.login.hello'), t('pages.login.title')]);
const HEADLINE_INTERVAL_MS = 2000;
const headlineIndex = ref(0);
let headlineTimer = null;

onMounted(() => {
  headlineTimer = window.setInterval(() => {
    headlineIndex.value = (headlineIndex.value + 1) % headlineWords.value.length;
  }, HEADLINE_INTERVAL_MS);
});

onBeforeUnmount(() => {
  if (headlineTimer != null) window.clearInterval(headlineTimer);
});

onMounted(async () => {
  const msg = await HttpUtil.post('/getTwoFactorEnable');
  if (msg.success) twoFactorEnable.value = !!msg.obj;
  fetched.value = true;
});

async function login() {
  submitting.value = true;
  try {
    const msg = await HttpUtil.post('/login', user);
    if (msg.success) window.location.href = basePath + 'panel/';
  } finally {
    submitting.value = false;
  }
}

const lang = ref(LanguageManager.getLanguage());
function onLangChange(next) {
  LanguageManager.setLanguage(next);
}

function cycleTheme() {
  pauseAnimationsUntilLeave('login-theme-cycle');
  cycleThemeMode();
}
</script>

<template>
  <a-config-provider :theme="antdThemeConfig">
    <a-layout class="login-app" :class="{ 'is-dark': themeState.isDark, 'is-ultra': themeState.isUltra }">
      <a-layout-content class="login-content">
        <!-- Floating chrome at top-right: theme cycle (Light/Dark/Ultra)
             plus a language picker hidden behind the gear popover. -->
        <div class="login-toolbar">
          <button type="button" class="theme-cycle" :aria-label="t('menu.theme')" :title="t('menu.theme')"
            @click="cycleTheme">
            <DesktopOutlined v-if="themeState.mode === 'system'" />
            <svg v-else-if="themeState.mode === 'light'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="12" cy="12" r="4" />
              <path
                d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
          </button>

          <a-popover :overlay-class-name="currentTheme" :title="t('pages.settings.language')" placement="bottomRight"
            trigger="click">
            <template #content>
              <a-space direction="vertical" :size="10" class="settings-popover">
                <a-select v-model:value="lang" class="lang-select" @change="onLangChange">
                  <a-select-option v-for="l in LanguageManager.supportedLanguages" :key="l.value" :value="l.value">
                    <span :aria-label="l.name">{{ l.icon }}</span>
                    &nbsp;&nbsp;<span>{{ l.name }}</span>
                  </a-select-option>
                </a-select>
              </a-space>
            </template>
            <a-button shape="circle" class="toolbar-btn" :aria-label="t('menu.settings')">
              <template #icon>
                <SettingOutlined />
              </template>
            </a-button>
          </a-popover>
        </div>

        <div class="login-wrapper">
          <div v-if="!fetched" class="login-loading">
            <a-spin size="large" />
          </div>

          <div v-else class="login-card">
            <div class="brand">
              <span class="brand-name">Nova_x-ui</span>
              <span class="brand-note">面板登录</span>
              <span class="brand-accent" aria-hidden="true"></span>
            </div>
            <h2 class="welcome">
              <Transition name="headline" mode="out-in">
                <b :key="headlineIndex">{{ headlineWords[headlineIndex] }}</b>
              </Transition>
            </h2>

            <a-form layout="vertical" class="login-form" @submit.prevent="login">
              <a-form-item :label="t('username')">
                <a-input v-model:value="user.username" autocomplete="username" name="username" size="large"
                  :placeholder="t('username')" autofocus required>
                  <template #prefix>
                    <UserOutlined />
                  </template>
                </a-input>
              </a-form-item>

              <a-form-item :label="t('password')">
                <a-input-password v-model:value="user.password" autocomplete="current-password" name="password"
                  size="large" :placeholder="t('password')" required>
                  <template #prefix>
                    <LockOutlined />
                  </template>
                </a-input-password>
              </a-form-item>

              <a-form-item v-if="twoFactorEnable" :label="t('twoFactorCode')">
                <a-input v-model:value="user.twoFactorCode" autocomplete="one-time-code" name="twoFactorCode"
                  size="large" :placeholder="t('twoFactorCode')" required>
                  <template #prefix>
                    <KeyOutlined />
                  </template>
                </a-input>
              </a-form-item>

              <a-form-item class="submit-row">
                <a-button type="primary" shape="round" html-type="submit" :loading="submitting" size="large" block>
                  {{ submitting ? '' : t('login') }}
                </a-button>
              </a-form-item>
            </a-form>

          </div>
        </div>
      </a-layout-content>
    </a-layout>
  </a-config-provider>
</template>

<style scoped>
.login-app {
  --bg-page: #eef3fb;
  --bg-card: rgba(255, 255, 255, 0.76);
  --bg-card-strong: rgba(255, 255, 255, 0.9);
  --color-text: rgba(15, 23, 42, 0.92);
  --color-text-subtle: rgba(71, 85, 105, 0.72);
  --color-accent: #1677ff;
  --color-border: rgba(148, 163, 184, 0.2);
  --shadow-card: 0 24px 56px rgba(15, 23, 42, 0.12);
  --blob-1: rgba(107, 155, 255, 0.34);
  --blob-2: rgba(155, 177, 255, 0.22);
  --blob-3: rgba(94, 234, 212, 0.18);

  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background: var(--bg-page);
}

.login-app.is-dark {
  --bg-page: #11161f;
  --bg-card: rgba(24, 30, 41, 0.88);
  --bg-card-strong: rgba(28, 35, 48, 0.94);
  --color-text: rgba(255, 255, 255, 0.94);
  --color-text-subtle: rgba(255, 255, 255, 0.64);
  --color-accent: #4096ff;
  --color-border: rgba(255, 255, 255, 0.09);
  --shadow-card: 0 24px 58px rgba(0, 0, 0, 0.34);
  --blob-1: rgba(64, 150, 255, 0.28);
  --blob-2: rgba(168, 85, 247, 0.18);
  --blob-3: rgba(34, 211, 238, 0.12);
}

.login-app.is-dark.is-ultra {
  --bg-page: #050608;
  --bg-card: rgba(12, 13, 16, 0.94);
  --bg-card-strong: rgba(16, 18, 22, 0.98);
  --color-border: rgba(255, 255, 255, 0.06);
  --blob-1: rgba(64, 150, 255, 0.18);
  --blob-2: rgba(168, 85, 247, 0.12);
  --blob-3: rgba(34, 211, 238, 0.08);
}

/* Three blurred blobs slowly drift across the page; ::before and
 * ::after carry two of them, the third lives on .login-content::before
 * so we can animate it independently. */
.login-app::before,
.login-app::after {
  content: '';
  position: absolute;
  width: 62vw;
  height: 62vw;
  max-width: 860px;
  max-height: 860px;
  border-radius: 50%;
  filter: blur(100px);
  pointer-events: none;
  z-index: 0;
  will-change: transform;
}

.login-app::before {
  top: -24vw;
  left: -18vw;
  background: radial-gradient(circle, var(--blob-1) 0%, transparent 65%);
  animation: blob-drift-a 24s ease-in-out infinite alternate;
}

.login-app::after {
  bottom: -24vw;
  right: -18vw;
  background: radial-gradient(circle, var(--blob-2) 0%, transparent 65%);
  animation: blob-drift-b 30s ease-in-out infinite alternate;
}

.login-content::before {
  content: '';
  position: absolute;
  top: 30%;
  left: 50%;
  width: 46vw;
  height: 46vw;
  max-width: 680px;
  max-height: 680px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--blob-3) 0%, transparent 65%);
  filter: blur(100px);
  pointer-events: none;
  z-index: 0;
  will-change: transform;
  animation: blob-drift-c 36s ease-in-out infinite alternate;
}

@keyframes blob-drift-a {
  0% {
    transform: translate(0, 0) scale(1);
  }

  50% {
    transform: translate(18vw, 10vh) scale(1.15);
  }

  100% {
    transform: translate(34vw, 22vh) scale(1.25);
  }
}

@keyframes blob-drift-b {
  0% {
    transform: translate(0, 0) scale(1);
  }

  50% {
    transform: translate(-16vw, -10vh) scale(1.12);
  }

  100% {
    transform: translate(-30vw, -22vh) scale(1.2);
  }
}

@keyframes blob-drift-c {
  0% {
    transform: translate(-50%, -50%) scale(1);
  }

  50% {
    transform: translate(-20%, -20%) scale(1.1);
  }

  100% {
    transform: translate(-80%, -10%) scale(1.05);
  }
}

@media (prefers-reduced-motion: reduce) {

  .login-app::before,
  .login-app::after,
  .login-content::before {
    animation: none;
  }
}

.login-app :deep(.ant-layout-content) {
  background: transparent;
}

.login-content {
  position: relative;
}

.login-content>* {
  position: relative;
  z-index: 1;
}

.login-toolbar {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 10;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.toolbar-btn {
  width: 40px;
  height: 40px;
}

.theme-cycle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--bg-card);
  color: var(--color-text);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  transition: background-color 0.2s, transform 0.15s, color 0.2s, box-shadow 0.2s;
}

.theme-cycle:hover,
.theme-cycle:focus-visible {
  background-color: var(--bg-card-strong);
  color: #4096ff;
  transform: translateY(-1px) scale(1.04);
  box-shadow: 0 14px 28px rgba(22, 119, 255, 0.14);
  outline: none;
}

.theme-cycle svg {
  width: 18px;
  height: 18px;
}

.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
}

.login-loading {
  text-align: center;
}

.login-card {
  width: 100%;
  max-width: 430px;
  background: var(--bg-card);
  border: 1px solid var(--color-border);
  border-radius: 28px;
  padding: 42px 34px 30px;
  box-shadow: var(--shadow-card);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  position: relative;
  overflow: hidden;
}

.login-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.22), transparent 28%);
  pointer-events: none;
}

.login-card > * {
  position: relative;
  z-index: 1;
}

@media (max-width: 480px) {
  .login-card {
    max-width: none;
    padding: 30px 20px 24px;
  }
}

.brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
}

.brand-name {
  font-size: 30px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--color-text);
}

.brand-note {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--color-text-subtle);
}

.brand-accent {
  display: block;
  width: 52px;
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(22, 119, 255, 0.1), var(--color-accent), rgba(22, 119, 255, 0.1));
}

.welcome {
  text-align: center;
  color: var(--color-text);
  font-size: 30px;
  font-weight: 700;
  line-height: 1.2;
  min-height: 40px;
  margin: 14px 0 28px;
  letter-spacing: 0.3px;
}

.welcome b {
  display: inline-block;
  font-weight: inherit;
}

.headline-enter-active,
.headline-leave-active {
  transition: opacity 280ms ease, transform 280ms ease;
}

.headline-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.headline-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.login-form :deep(.ant-form-item-label > label) {
  color: var(--color-text);
  font-weight: 500;
}

.login-form :deep(.ant-input-affix-wrapper),
.login-form :deep(.ant-input),
.login-form :deep(.ant-select-selector) {
  background: rgba(255, 255, 255, 0.86) !important;
  border-color: rgba(148, 163, 184, 0.22) !important;
  border-radius: 18px !important;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.login-form :deep(.ant-input-affix-wrapper),
.login-form :deep(.ant-select-selector) {
  min-height: 54px;
  padding-inline: 14px;
}

.login-form :deep(.ant-input),
.login-form :deep(.ant-input-affix-wrapper input) {
  color: var(--color-text);
}

.login-form :deep(.ant-input-prefix) {
  color: rgba(100, 116, 139, 0.78);
}

.submit-row :deep(.ant-btn-primary) {
  min-height: 52px;
  border: none;
  background: linear-gradient(135deg, #4c88ff 0%, #2f6df6 55%, #2559d6 100%);
  box-shadow: 0 18px 28px rgba(37, 89, 214, 0.22);
}

.submit-row :deep(.ant-btn-primary:hover),
.submit-row :deep(.ant-btn-primary:focus-visible) {
  background: linear-gradient(135deg, #5f97ff 0%, #3877ff 55%, #2b63ea 100%);
}

.submit-row :deep(.ant-btn-round) {
  border-radius: 999px;
}

.submit-row {
  margin-bottom: 0;
}

.settings-popover {
  min-width: 220px;
}

.lang-select {
  width: 100%;
}
</style>
