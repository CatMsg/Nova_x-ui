<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Modal } from 'ant-design-vue';
import {
  SettingOutlined,
  SafetyOutlined,
  MessageOutlined,
  CloudServerOutlined,
  CodeOutlined,
} from '@ant-design/icons-vue';

import { HttpUtil, PromiseUtil } from '@/utils';
import { theme as themeState, antdThemeConfig } from '@/composables/useTheme.js';
import { useMediaQuery } from '@/composables/useMediaQuery.js';
import AppSidebar from '@/components/AppSidebar.vue';
import { useAllSetting } from './useAllSetting.js';
import GeneralTab from './GeneralTab.vue';
import SecurityTab from './SecurityTab.vue';
import TelegramTab from './TelegramTab.vue';
import SubscriptionGeneralTab from './SubscriptionGeneralTab.vue';
import SubscriptionFormatsTab from './SubscriptionFormatsTab.vue';

const { t } = useI18n();

const { fetched, spinning, saveDisabled, allSetting, saveAll } = useAllSetting();
const { isMobile } = useMediaQuery();

const basePath = window.X_UI_BASE_PATH || '';
const requestUri = window.location.pathname;

// AD-Vue 4's <a-back-top> calls `target()` after mount to find the
// scrolled element. Inline-arrow `() => document.getElementById(...)`
// in the template threw "Cannot read properties of undefined (reading
// 'getElementById')" because of how Vue 3 evaluates the expression
// outside the script-setup scope — wrap in a regular function so
// `document` resolves to the window global at call time.
function scrollTarget() {
  return document.getElementById('content-layout');
}

// `entry*` mirrors the URL the user opened the panel with so the page
// can rebuild it after a restart that may change host/port/scheme.
const entryHost = ref('');
const entryPort = ref('');
const entryIsIP = ref(false);

function isIp(h) {
  if (typeof h !== 'string') return false;
  // IPv4: four dot-separated octets 0-255.
  const v4 = h.split('.');
  if (v4.length === 4 && v4.every((p) => /^\d{1,3}$/.test(p) && Number(p) <= 255)) return true;
  // IPv6: hex groups, optional single :: compression.
  if (!h.includes(':') || h.includes(':::')) return false;
  const parts = h.split('::');
  if (parts.length > 2) return false;
  const split = (s) => (s ? s.split(':').filter(Boolean) : []);
  const head = split(parts[0]);
  const tail = split(parts[1]);
  const valid = (seg) => /^[0-9a-fA-F]{1,4}$/.test(seg);
  if (![...head, ...tail].every(valid)) return false;
  const groups = head.length + tail.length;
  return parts.length === 2 ? groups < 8 : groups === 8;
}

onMounted(() => {
  entryHost.value = window.location.hostname;
  entryPort.value = window.location.port;
  entryIsIP.value = isIp(entryHost.value);
});

// Rebuild the URL after a restart — host/port/scheme may have changed
// (cert toggled on, port edited, base path edited).
function rebuildUrlAfterRestart() {
  const { webDomain, webPort, webBasePath, webCertFile, webKeyFile } = allSetting;
  const newProtocol = (webCertFile || webKeyFile) ? 'https:' : 'http:';

  let base = webBasePath ? webBasePath.replace(/^\//, '') : '';
  if (base && !base.endsWith('/')) base += '/';

  if (!entryIsIP.value) {
    const url = new URL(window.location.href);
    url.pathname = `/${base}panel/settings`;
    url.protocol = newProtocol;
    return url.toString();
  }

  let finalHost = entryHost.value;
  let finalPort = entryPort.value || '';
  if (webDomain && isIp(webDomain)) finalHost = webDomain;
  if (webPort && Number(webPort) !== Number(entryPort.value)) finalPort = String(webPort);

  const url = new URL(`${newProtocol}//${finalHost}`);
  if (finalPort) url.port = finalPort;
  url.pathname = `/${base}panel/settings`;
  return url.toString();
}

function restartPanel() {
  Modal.confirm({
    title: t('pages.settings.restartPanel'),
    content: t('pages.settings.restartPanelDesc'),
    okText: t('pages.settings.restartPanel'),
    okButtonProps: { danger: true },
    cancelText: t('cancel'),
    async onOk() {
      spinning.value = true;
      try {
        const msg = await HttpUtil.post('/panel/setting/restartPanel');
        if (!msg?.success) return;
        await PromiseUtil.sleep(5000);
        window.location.replace(rebuildUrlAfterRestart());
      } finally {
        spinning.value = false;
      }
    },
  });
}

// Conf alerts mirror the previous banner — pure derivation off allSetting.
const confAlerts = computed(() => {
  const out = [];
  if (window.location.protocol !== 'https:') {
    out.push('Panel is served over plain HTTP — set up TLS for production.');
  }
  if (allSetting.webPort === 2026) {
    out.push('Default port 2026 is well-known — change it to a random port.');
  }
  const segs = window.location.pathname.split('/').length < 4;
  if (segs && allSetting.webBasePath === '/') {
    out.push('Default base path "/" is well-known — change it to a random path.');
  }
  if (allSetting.subEnable) {
    let subPath = allSetting.subPath;
    if (allSetting.subURI) {
      try { subPath = new URL(allSetting.subURI).pathname; } catch (_e) { }
    }
    if (subPath === '/sub/') {
      out.push('Default subscription path "/sub/" is well-known — change it.');
    }
  }
  if (allSetting.subJsonEnable) {
    let p = allSetting.subJsonPath;
    if (allSetting.subJsonURI) {
      try { p = new URL(allSetting.subJsonURI).pathname; } catch (_e) { }
    }
    if (p === '/json/') {
      out.push('Default JSON subscription path "/json/" is well-known — change it.');
    }
  }
  return out;
});

const alertVisible = ref(true);
</script>

<template>
  <a-config-provider :theme="antdThemeConfig">
    <a-layout class="settings-page" :class="{ 'is-dark': themeState.isDark, 'is-ultra': themeState.isUltra }">
      <AppSidebar :base-path="basePath" :request-uri="requestUri" />

      <a-layout class="content-shell">
        <a-layout-content id="content-layout" class="content-area">
          <a-spin :spinning="spinning || !fetched" :delay="200" tip="Loading…" size="large">
            <div v-if="!fetched" class="loading-spacer" />

            <template v-else>
              <a-alert v-if="confAlerts.length > 0 && alertVisible" type="error" show-icon closable class="conf-alert"
                @close="alertVisible = false">
                <template #message>Security warnings</template>
                <template #description>
                  <b>Your panel may be exposed:</b>
                  <ul>
                    <li v-for="(msg, i) in confAlerts" :key="i">{{ msg }}</li>
                  </ul>
                </template>
              </a-alert>

              <a-row :gutter="[isMobile ? 8 : 16, isMobile ? 0 : 12]">
                <a-col :span="24">
                  <a-card hoverable class="settings-hero">
                    <a-row class="header-row">
                      <a-col :xs="24" :sm="10" class="header-actions">
                        <a-space direction="horizontal">
                          <a-button type="primary" :disabled="saveDisabled" @click="saveAll">
                            {{ t('pages.settings.save') }}
                          </a-button>
                          <a-button type="primary" danger :disabled="!saveDisabled" @click="restartPanel">
                            {{ t('pages.settings.restartPanel') }}
                          </a-button>
                        </a-space>
                      </a-col>
                      <a-col :xs="24" :sm="14" class="header-info">
                        <a-back-top :target="scrollTarget" :visibility-height="200" />
                        <a-alert type="warning" show-icon :message="t('pages.settings.infoDesc')" />
                      </a-col>
                    </a-row>
                  </a-card>
                </a-col>

                <a-col :span="24">
                  <a-tabs default-active-key="1" class="settings-tabs">
                    <a-tab-pane key="1" class="tab-pane">
                      <template #tab>
                        <SettingOutlined />
                        <span>{{ t('pages.settings.panelSettings') }}</span>
                      </template>
                      <GeneralTab :all-setting="allSetting" />
                    </a-tab-pane>
                    <a-tab-pane key="2" class="tab-pane">
                      <template #tab>
                        <SafetyOutlined />
                        <span>{{ t('pages.settings.securitySettings') }}</span>
                      </template>
                      <SecurityTab :all-setting="allSetting" />
                    </a-tab-pane>
                    <a-tab-pane key="3" class="tab-pane">
                      <template #tab>
                        <MessageOutlined />
                        <span>{{ t('pages.settings.TGBotSettings') }}</span>
                      </template>
                      <TelegramTab :all-setting="allSetting" />
                    </a-tab-pane>
                    <a-tab-pane key="4" class="tab-pane">
                      <template #tab>
                        <CloudServerOutlined />
                        <span>{{ t('pages.settings.subSettings') }}</span>
                      </template>
                      <SubscriptionGeneralTab :all-setting="allSetting" />
                    </a-tab-pane>
                    <a-tab-pane v-if="allSetting.subJsonEnable || allSetting.subClashEnable" key="5" class="tab-pane">
                      <template #tab>
                        <CodeOutlined />
                        <span>{{ t('pages.settings.subSettings') }} (Formats)</span>
                      </template>
                      <SubscriptionFormatsTab :all-setting="allSetting" />
                    </a-tab-pane>
                  </a-tabs>
                </a-col>
              </a-row>
            </template>
          </a-spin>
        </a-layout-content>
      </a-layout>
    </a-layout>
  </a-config-provider>
</template>

<style scoped>
.settings-page {
  --bg-page: #eef3fb;
  --bg-card: rgba(255, 255, 255, 0.78);
  --card-border: rgba(148, 163, 184, 0.2);
  --card-shadow: 0 22px 54px rgba(15, 23, 42, 0.1);
  --card-hover-shadow: 0 26px 64px rgba(15, 23, 42, 0.13);

  min-height: 100vh;
  background:
    radial-gradient(circle at top right, rgba(22, 119, 255, 0.12), transparent 28%),
    radial-gradient(circle at bottom left, rgba(249, 115, 22, 0.08), transparent 24%),
    linear-gradient(180deg, #f9fbff 0%, var(--bg-page) 56%, #e9eef7 100%);
}

.settings-page.is-dark {
  --bg-page: #11161f;
  --bg-card: rgba(22, 28, 39, 0.86);
  --card-border: rgba(255, 255, 255, 0.08);
  --card-shadow: 0 24px 58px rgba(0, 0, 0, 0.34);
  --card-hover-shadow: 0 28px 66px rgba(0, 0, 0, 0.42);
  background:
    radial-gradient(circle at top right, rgba(64, 150, 255, 0.16), transparent 28%),
    radial-gradient(circle at bottom left, rgba(249, 115, 22, 0.08), transparent 24%),
    linear-gradient(180deg, #151b25 0%, var(--bg-page) 48%, #0b0f15 100%);
}

.settings-page.is-dark.is-ultra {
  --bg-page: #040405;
  --bg-card: rgba(10, 10, 12, 0.92);
  --card-border: rgba(255, 255, 255, 0.06);
  --card-shadow: 0 26px 64px rgba(0, 0, 0, 0.5);
  --card-hover-shadow: 0 30px 74px rgba(0, 0, 0, 0.58);
  background:
    radial-gradient(circle at top right, rgba(102, 170, 255, 0.08), transparent 28%),
    radial-gradient(circle at bottom left, rgba(249, 115, 22, 0.06), transparent 24%),
    linear-gradient(180deg, #0b0c10 0%, var(--bg-page) 48%, #040404 100%);
}

.settings-page :deep(.ant-layout),
.settings-page :deep(.ant-layout-content) {
  background: transparent;
}

.content-shell {
  background: transparent;
}

.content-area {
  padding: 22px 24px 30px;
}

.settings-page :deep(.ant-card) {
  background: var(--bg-card);
  border: 1px solid var(--card-border);
  border-radius: 28px;
  box-shadow: var(--card-shadow);
  backdrop-filter: blur(22px) saturate(180%);
  -webkit-backdrop-filter: blur(22px) saturate(180%);
  overflow: hidden;
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
}

.settings-page :deep(.ant-card:hover) {
  transform: translateY(-2px);
  box-shadow: var(--card-hover-shadow);
  border-color: rgba(22, 119, 255, 0.16);
}

.settings-page :deep(.ant-card-head) {
  min-height: 58px;
  padding-inline: 20px;
  border-bottom-color: var(--card-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.2), transparent);
}

.settings-page :deep(.ant-card-head-title) {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.settings-page :deep(.ant-card-body) {
  padding: 20px;
}

.settings-page :deep(.settings-hero .ant-card-body) {
  padding: 18px;
}

.settings-page :deep(.settings-hero .ant-space) {
  flex-wrap: wrap;
}

.settings-page :deep(.settings-hero .ant-btn) {
  min-height: 44px;
  border-radius: 999px;
  padding-inline: 18px;
}

.settings-page :deep(.settings-hero .ant-btn-primary) {
  background: linear-gradient(135deg, #4d88ff 0%, #2f6df6 100%);
  border: none;
  box-shadow: 0 16px 28px rgba(37, 89, 214, 0.18);
}

.settings-page :deep(.settings-hero .ant-btn-dangerous) {
  color: #a9772b;
  border-color: rgba(169, 119, 43, 0.36);
  background: rgba(255, 255, 255, 0.7);
}

.settings-page :deep(.settings-hero .ant-btn-dangerous:hover),
.settings-page :deep(.settings-hero .ant-btn-dangerous:focus-visible) {
  color: #8f611d;
  border-color: rgba(169, 119, 43, 0.48);
  background: rgba(255, 250, 240, 0.92);
}

.settings-page :deep(.settings-hero .ant-alert) {
  margin: 0;
}

.settings-page :deep(.settings-tabs) {
  display: block;
}

.settings-page :deep(.ant-tabs-nav) {
  margin-bottom: 18px;
  padding: 10px;
  border: 1px solid var(--card-border);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.62);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.07);
  backdrop-filter: blur(18px) saturate(170%);
  -webkit-backdrop-filter: blur(18px) saturate(170%);
}

.settings-page :deep(.ant-tabs-nav-wrap) {
  padding-inline: 2px;
}

.settings-page :deep(.ant-tabs-tab) {
  border-radius: 12px;
  padding: 11px 16px;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.settings-page :deep(.ant-tabs-tab-active) {
  background: rgba(22, 119, 255, 0.09);
}

.settings-page :deep(.ant-tabs-ink-bar) {
  height: 3px;
  border-radius: 999px;
  background: #1677ff;
}

.settings-page :deep(.ant-tabs-tab-btn) {
  font-weight: 600;
}

.loading-spacer {
  min-height: calc(100vh - 120px);
}

.conf-alert {
  margin-bottom: 12px;
}

.settings-page :deep(.conf-alert.ant-alert) {
  background: rgba(255, 255, 255, 0.8);
  border-radius: 22px;
  border: 1px solid rgba(239, 68, 68, 0.18);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(18px) saturate(170%);
  -webkit-backdrop-filter: blur(18px) saturate(170%);
}

.settings-page :deep(.conf-alert .ant-alert-message) {
  font-weight: 700;
}

.settings-page :deep(.conf-alert .ant-alert-description) {
  color: rgba(15, 23, 42, 0.82);
}

.settings-page :deep(.ant-collapse) {
  background: transparent;
  border: none;
}

.settings-page :deep(.ant-collapse > .ant-collapse-item) {
  margin-bottom: 14px;
  border: 1px solid var(--card-border);
  border-radius: 24px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: var(--card-shadow);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
}

.settings-page :deep(.ant-collapse > .ant-collapse-item:last-child) {
  margin-bottom: 0;
}

.settings-page :deep(.ant-collapse-header) {
  min-height: 58px;
  align-items: center !important;
  padding: 18px 20px !important;
  font-weight: 700;
  color: rgba(15, 23, 42, 0.9);
}

.settings-page :deep(.ant-collapse-content) {
  border-top: 1px solid var(--card-border);
  background: transparent;
}

.settings-page :deep(.ant-collapse-content-box) {
  padding: 0;
}

.settings-page :deep(.ant-list-item) {
  margin: 10px 14px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.66);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
  overflow: hidden;
}

.settings-page :deep(.ant-list-item-meta-title) {
  font-weight: 700;
  color: rgba(15, 23, 42, 0.9);
}

.settings-page :deep(.ant-list-item-meta-description) {
  color: rgba(71, 85, 105, 0.72);
}

.settings-page :deep(.ant-input),
.settings-page :deep(.ant-input-affix-wrapper),
.settings-page :deep(.ant-input-number),
.settings-page :deep(.ant-select-selector) {
  border-radius: 18px !important;
  border-color: rgba(148, 163, 184, 0.22) !important;
  background: rgba(255, 255, 255, 0.88) !important;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.settings-page :deep(.ant-input-affix-wrapper),
.settings-page :deep(.ant-select-selector),
.settings-page :deep(.ant-input-number) {
  min-height: 52px;
}

.settings-page :deep(.ant-input),
.settings-page :deep(.ant-select-selector),
.settings-page :deep(.ant-input-number-input) {
  color: rgba(15, 23, 42, 0.9);
}

.settings-page :deep(.ant-select-selection-item),
.settings-page :deep(.ant-input-number-input) {
  line-height: 50px;
}

.settings-page :deep(.ant-select-arrow),
.settings-page :deep(.ant-input-suffix) {
  color: rgba(100, 116, 139, 0.8);
}

.settings-page :deep(.ant-switch) {
  min-width: 46px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.settings-page :deep(.hy2-backend-tag) {
  border-radius: 999px;
  padding-inline: 12px;
  font-weight: 600;
  border: none;
}

.header-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.header-actions {
  padding: 4px;
}

.header-info {
  display: flex;
  justify-content: flex-end;
}

.tab-pane {
  padding-top: 18px;
}

.settings-page :deep(.ant-card-actions) {
  background: rgba(255, 255, 255, 0.35);
}

@media (max-width: 768px) {
  .content-area {
    padding: 14px 12px 18px;
  }

  .settings-page :deep(.ant-tabs-nav) {
    padding: 8px;
  }

  .settings-page :deep(.ant-list-item) {
    margin-inline: 10px;
  }
}
</style>
