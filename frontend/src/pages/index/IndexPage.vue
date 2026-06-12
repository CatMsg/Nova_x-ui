<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  BarsOutlined,
  ControlOutlined,
  CloudServerOutlined,
  CloudDownloadOutlined,
  CloudUploadOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  AreaChartOutlined,
  GlobalOutlined,
  SwapOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  ThunderboltOutlined,
  DesktopOutlined,
  DatabaseOutlined,
  ForkOutlined,
} from '@ant-design/icons-vue';

const { t } = useI18n();

import { HttpUtil, SizeFormatter, TimeFormatter } from '@/utils';
import { theme as themeState, antdThemeConfig } from '@/composables/useTheme.js';
import { useStatus } from '@/composables/useStatus.js';
import { useMediaQuery } from '@/composables/useMediaQuery.js';
import AppSidebar from '@/components/AppSidebar.vue';
import CustomStatistic from '@/components/CustomStatistic.vue';
import TextModal from '@/components/TextModal.vue';
import StatusCard from './StatusCard.vue';
import XrayStatusCard from './XrayStatusCard.vue';
import PanelUpdateModal from './PanelUpdateModal.vue';
import LogModal from './LogModal.vue';
import BackupModal from './BackupModal.vue';
import SystemHistoryModal from './SystemHistoryModal.vue';
import XrayMetricsModal from './XrayMetricsModal.vue';
import XrayLogModal from './XrayLogModal.vue';
import VersionModal from './VersionModal.vue';

const { status, fetched, refresh } = useStatus();
const { isMobile } = useMediaQuery();

// `/panel/setting/defaultSettings` returns ipLimitEnable; the xray
// card hides its log button when access logs are off.
const ipLimitEnable = ref(false);
HttpUtil.post('/panel/setting/defaultSettings').then((msg) => {
  if (msg?.success && msg.obj) ipLimitEnable.value = !!msg.obj.ipLimitEnable;
});

// Panel-update info — fetched once on mount, drives both the badge
// in QuickActions and the contents of PanelUpdateModal.
const panelUpdateInfo = ref({ currentVersion: '', latestVersion: '', updateAvailable: false });
onMounted(() => {
  HttpUtil.get('/panel/api/server/getPanelUpdateInfo').then((msg) => {
    if (msg?.success && msg.obj) panelUpdateInfo.value = msg.obj;
  });
});

const basePath = window.X_UI_BASE_PATH || '';
const requestUri = window.location.pathname;

// In production, dist.go injects window.X_UI_CUR_VER at serve time.
// In dev, Vite serves the HTML directly so the global is missing — fall
// back to currentVersion from the panel-update API once it answers.
const displayVersion = computed(
  () => panelUpdateInfo.value?.currentVersion || window.X_UI_CUR_VER || '?',
);

// Hide/reveal the public IPv4/IPv6 — same pattern as previous.
const showIp = ref(false);

// Modal open state.
const logsOpen = ref(false);
const backupOpen = ref(false);
const panelUpdateOpen = ref(false);
const sysHistoryOpen = ref(false);
const xrayMetricsOpen = ref(false);
const xrayLogsOpen = ref(false);
const versionOpen = ref(false);
const configTextOpen = ref(false);
const configText = ref('');

// Page-level loading overlay; modals can request it via @busy.
const loading = ref(false);
const loadingTip = ref(t('loading'));
function setBusy({ busy, tip }) {
  loading.value = busy;
  if (tip) loadingTip.value = tip;
}

// Xray controls
async function stopXray() {
  await HttpUtil.post('/panel/api/server/stopXrayService');
  await refresh();
}
async function restartXray() {
  await HttpUtil.post('/panel/api/server/restartXrayService');
  await refresh();
}

function openSystemHistory() { sysHistoryOpen.value = true; }
function openXrayLogs() { xrayLogsOpen.value = true; }
function openVersionSwitch() { versionOpen.value = true; }

function openPanelVersion() {
  if (panelUpdateInfo.value.updateAvailable) {
    panelUpdateOpen.value = true;
  } else {
    window.open('https://github.com/CatMsg/Nova_x-ui/releases', '_blank', 'noopener,noreferrer');
  }
}

function openTelegram() {
  window.open('https://t.me/Nova_x_ui', '_blank', 'noopener,noreferrer');
}

// Previous "Config" action — fetch the rendered xray config and show
// it as JSON in the shared TextModal (same UX as main).
async function openConfig() {
  loading.value = true;
  try {
    const msg = await HttpUtil.get('/panel/api/server/getConfigJson');
    if (!msg?.success) return;
    configText.value = JSON.stringify(msg.obj, null, 2);
    configTextOpen.value = true;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <a-config-provider :theme="antdThemeConfig">
    <a-layout class="index-page" :class="{ 'is-dark': themeState.isDark, 'is-ultra': themeState.isUltra }">
      <AppSidebar :base-path="basePath" :request-uri="requestUri" />

      <a-layout class="content-shell">
        <a-layout-content class="content-area">
          <a-spin :spinning="loading || !fetched" :delay="200" :tip="loading ? loadingTip : t('loading')" size="large">
            <div v-if="!fetched" class="loading-spacer" />

            <a-row v-else :gutter="[isMobile ? 8 : 16, 12]">
              <a-col :span="24">
                <StatusCard :status="status" :is-mobile="isMobile" />
              </a-col>

              <a-col :xs="24" :lg="12">
                <XrayStatusCard :status="status" :is-mobile="isMobile" :ip-limit-enable="ipLimitEnable"
                  @stop-xray="stopXray" @restart-xray="restartXray" @open-xray-logs="openXrayLogs"
                  @open-logs="logsOpen = true" @open-version-switch="openVersionSwitch" />
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('menu.link')" hoverable class="glass-card action-card">
                  <template #actions>
                    <a-space class="action" @click="logsOpen = true">
                      <BarsOutlined />
                      <span v-if="!isMobile">{{ t('pages.index.logs') }}</span>
                    </a-space>
                    <a-space class="action" @click="openConfig">
                      <ControlOutlined />
                      <span v-if="!isMobile">{{ t('pages.index.config') }}</span>
                    </a-space>
                    <a-space class="action" @click="backupOpen = true">
                      <CloudServerOutlined />
                      <span v-if="!isMobile">{{ t('pages.index.backupTitle') }}</span>
                    </a-space>
                  </template>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card title="Nova_x-ui" hoverable class="glass-card brand-card">
                  <template #actions>
                    <a-space class="action" @click="openTelegram">
                      <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" class="tg-icon"
                        aria-hidden="true">
                        <path
                          d="M21.93 4.34a1.5 1.5 0 0 0-2.05-1.6L2.97 9.6c-.92.36-.91 1.66.02 1.99l4.32 1.53 1.7 5.23a1 1 0 0 0 1.68.36l2.43-2.43 4.36 3.21a1.5 1.5 0 0 0 2.36-.91l3.09-13.86a1.5 1.5 0 0 0 0-.38ZM9.97 14.66l-.55 3.36-1.36-4.2 9.8-7.05-7.89 7.89Z" />
                      </svg>
                      <span v-if="!isMobile">@Nova_x_ui</span>
                    </a-space>
                    <a-space class="action" :class="{ 'action-update': panelUpdateInfo.updateAvailable }"
                      @click="openPanelVersion">
                      <CloudDownloadOutlined />
                      <span v-if="!isMobile">
                        {{ panelUpdateInfo.updateAvailable
                          ? `${t('update')} ${panelUpdateInfo.latestVersion}`
                          : `v${displayVersion}` }}
                      </span>
                    </a-space>
                  </template>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.charts')" hoverable class="glass-card action-card">
                  <template #actions>
                    <a-space class="action" @click="openSystemHistory">
                      <AreaChartOutlined />
                      <span v-if="!isMobile">{{ t('pages.index.systemHistoryTitle') }}</span>
                    </a-space>
                    <a-space class="action" @click="xrayMetricsOpen = true">
                      <AreaChartOutlined />
                      <span v-if="!isMobile">{{ t('pages.index.xrayMetricsTitle') }}</span>
                    </a-space>
                  </template>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.operationHours')" hoverable class="glass-card metric-card">
                  <a-row :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="12">
                      <CustomStatistic title="Xray" :value="TimeFormatter.formatSecond(status.appStats.uptime)">
                        <template #prefix>
                          <ThunderboltOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="12">
                      <CustomStatistic title="OS" :value="TimeFormatter.formatSecond(status.uptime)">
                        <template #prefix>
                          <DesktopOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('usage')" hoverable class="glass-card metric-card">
                  <a-row :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.memory')"
                        :value="SizeFormatter.sizeFormat(status.appStats.mem)">
                        <template #prefix>
                          <DatabaseOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.threads')" :value="status.appStats.threads">
                        <template #prefix>
                          <ForkOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.overallSpeed')" hoverable class="glass-card metric-card">
                  <a-row :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.upload')"
                        :value="SizeFormatter.sizeFormat(status.netIO.up)">
                        <template #prefix>
                          <ArrowUpOutlined />
                        </template>
                        <template #suffix>/s</template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.download')"
                        :value="SizeFormatter.sizeFormat(status.netIO.down)">
                        <template #prefix>
                          <ArrowDownOutlined />
                        </template>
                        <template #suffix>/s</template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.totalData')" hoverable class="glass-card metric-card">
                  <a-row :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.sent')"
                        :value="SizeFormatter.sizeFormat(status.netTraffic.sent)">
                        <template #prefix>
                          <CloudUploadOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="12">
                      <CustomStatistic :title="t('pages.index.received')"
                        :value="SizeFormatter.sizeFormat(status.netTraffic.recv)">
                        <template #prefix>
                          <CloudDownloadOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.ipAddresses')" hoverable class="glass-card metric-card">
                  <template #extra>
                    <a-tooltip :title="t('pages.index.toggleIpVisibility')" :placement="isMobile ? 'topRight' : 'top'">
                      <component :is="showIp ? EyeOutlined : EyeInvisibleOutlined" class="ip-toggle-icon"
                        @click="showIp = !showIp" />
                    </a-tooltip>
                  </template>
                  <a-row :class="showIp ? 'ip-visible' : 'ip-hidden'" :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="isMobile ? 24 : 12">
                      <CustomStatistic title="IPv4" :value="status.publicIP.ipv4">
                        <template #prefix>
                          <GlobalOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="isMobile ? 24 : 12">
                      <CustomStatistic title="IPv6" :value="status.publicIP.ipv6">
                        <template #prefix>
                          <GlobalOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>

              <a-col :xs="24" :lg="12">
                <a-card :title="t('pages.index.connectionCount')" hoverable class="glass-card metric-card">
                  <a-row :gutter="isMobile ? [8, 8] : 0">
                    <a-col :span="12">
                      <CustomStatistic title="TCP" :value="status.tcpCount">
                        <template #prefix>
                          <SwapOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                    <a-col :span="12">
                      <CustomStatistic title="UDP" :value="status.udpCount">
                        <template #prefix>
                          <SwapOutlined />
                        </template>
                      </CustomStatistic>
                    </a-col>
                  </a-row>
                </a-card>
              </a-col>
            </a-row>
          </a-spin>
        </a-layout-content>
      </a-layout>

      <PanelUpdateModal v-model:open="panelUpdateOpen" :info="panelUpdateInfo" @busy="setBusy" />
      <LogModal v-model:open="logsOpen" />
      <BackupModal v-model:open="backupOpen" :base-path="basePath" @busy="setBusy" />
      <SystemHistoryModal v-model:open="sysHistoryOpen" :status="status" />
      <XrayMetricsModal v-model:open="xrayMetricsOpen" />
      <XrayLogModal v-model:open="xrayLogsOpen" />
      <VersionModal v-model:open="versionOpen" :status="status" @busy="setBusy" />
      <TextModal v-model:open="configTextOpen" :title="t('pages.index.config')" :content="configText"
        file-name="config.json" />
    </a-layout>
  </a-config-provider>
</template>

<style scoped>
.index-page {
  position: relative;
  isolation: isolate;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 15%, rgba(117, 151, 255, 0.34), transparent 0 22%),
    radial-gradient(circle at 82% 14%, rgba(92, 224, 214, 0.2), transparent 0 18%),
    radial-gradient(circle at 86% 84%, rgba(255, 145, 190, 0.18), transparent 0 22%),
    radial-gradient(circle at 14% 86%, rgba(255, 214, 102, 0.12), transparent 0 20%),
    linear-gradient(145deg, #f7fbff 0%, #eef4ff 44%, #f7f8fe 100%);
}

.index-page::before,
.index-page::after {
  content: '';
  position: fixed;
  pointer-events: none;
  z-index: 0;
  border-radius: 999px;
  filter: blur(72px);
  opacity: 0.9;
}

.index-page::before {
  width: 42vw;
  height: 42vw;
  top: -10vw;
  right: -8vw;
  background: radial-gradient(circle, rgba(115, 135, 255, 0.36) 0%, rgba(115, 135, 255, 0.12) 36%, transparent 72%);
}

.index-page::after {
  width: 52vw;
  height: 52vw;
  left: -12vw;
  bottom: -18vw;
  background: radial-gradient(circle, rgba(70, 225, 220, 0.24) 0%, rgba(70, 225, 220, 0.08) 34%, transparent 68%);
}

.index-page.is-dark {
  background:
    radial-gradient(circle at 18% 15%, rgba(89, 140, 255, 0.25), transparent 0 22%),
    radial-gradient(circle at 82% 14%, rgba(58, 208, 196, 0.16), transparent 0 18%),
    radial-gradient(circle at 86% 84%, rgba(255, 120, 178, 0.14), transparent 0 22%),
    radial-gradient(circle at 14% 86%, rgba(255, 210, 102, 0.1), transparent 0 20%),
    linear-gradient(145deg, #0d1118 0%, #121824 44%, #0b0d12 100%);
}

.index-page.is-dark::before {
  background: radial-gradient(circle, rgba(87, 148, 255, 0.22) 0%, rgba(87, 148, 255, 0.08) 36%, transparent 72%);
}

.index-page.is-dark::after {
  background: radial-gradient(circle, rgba(78, 228, 219, 0.16) 0%, rgba(78, 228, 219, 0.06) 34%, transparent 68%);
}

.index-page.is-dark.is-ultra {
  background:
    radial-gradient(circle at 18% 15%, rgba(89, 140, 255, 0.16), transparent 0 22%),
    radial-gradient(circle at 82% 14%, rgba(58, 208, 196, 0.1), transparent 0 18%),
    radial-gradient(circle at 86% 84%, rgba(255, 120, 178, 0.08), transparent 0 22%),
    radial-gradient(circle at 14% 86%, rgba(255, 210, 102, 0.06), transparent 0 20%),
    linear-gradient(145deg, #050608 0%, #090c12 44%, #030304 100%);
}

.index-page.is-dark.is-ultra::before {
  background: radial-gradient(circle, rgba(79, 146, 255, 0.16) 0%, rgba(79, 146, 255, 0.06) 36%, transparent 72%);
}

.index-page.is-dark.is-ultra::after {
  background: radial-gradient(circle, rgba(68, 214, 202, 0.1) 0%, rgba(68, 214, 202, 0.04) 34%, transparent 68%);
}

.index-page.is-dark :deep(.glass-card) {
  background:
    linear-gradient(180deg, rgba(23, 30, 48, 0.56), rgba(12, 16, 27, 0.34));
  border-color: rgba(115, 145, 255, 0.18);
  border-bottom-color: rgba(255, 255, 255, 0.08);
  box-shadow:
    0 22px 56px rgba(0, 0, 0, 0.34),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.index-page.is-dark :deep(.glass-card:hover) {
  border-color: rgba(140, 170, 255, 0.22);
  box-shadow:
    0 28px 64px rgba(0, 0, 0, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.index-page.is-dark :deep(.glass-card .ant-card-head) {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.04)),
    rgba(255, 255, 255, 0.02);
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

.index-page.is-dark :deep(.glass-card .ant-card-head-title),
.index-page.is-dark :deep(.xray-card .ant-card-extra) {
  color: rgba(255, 255, 255, 0.92);
}

.index-page.is-dark :deep(.glass-card .ant-card-actions) {
  background: rgba(255, 255, 255, 0.06);
  border-top-color: rgba(255, 255, 255, 0.08);
}

.index-page.is-dark :deep(.action),
.index-page.is-dark :deep(.glass-card .ant-card-actions > li > span) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(140, 170, 255, 0.14);
  color: rgba(255, 255, 255, 0.9);
}

.index-page.is-dark.is-ultra :deep(.glass-card) {
  background: rgba(9, 11, 16, 0.6);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow:
    0 24px 62px rgba(0, 0, 0, 0.48),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.index-page.is-dark.is-ultra :deep(.glass-card .ant-card-head) {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03));
}

.index-page.is-dark.is-ultra :deep(.glass-card .ant-card-head-title),
.index-page.is-dark.is-ultra :deep(.xray-card .ant-card-extra) {
  color: rgba(255, 255, 255, 0.96);
}

.index-page :deep(.ant-layout),
.index-page :deep(.ant-layout-content) {
  background: transparent;
}

.content-shell {
  background: transparent;
  position: relative;
  z-index: 1;
}

.content-area {
  padding: 18px 18px 22px 16px;
}

@media (max-width: 768px) {
  .content-area {
    padding: 12px 12px 18px;
    padding-top: 66px;
  }
}

.index-page :deep(.glass-card) {
  position: relative;
  background: rgba(255, 255, 255, 0.42);
  border: 1px solid rgba(255, 255, 255, 0.46);
  border-bottom-color: rgba(255, 255, 255, 0.3);
  border-radius: 28px;
  box-shadow:
    0 18px 48px rgba(46, 65, 105, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.68);
  backdrop-filter: blur(28px) saturate(180%);
  -webkit-backdrop-filter: blur(28px) saturate(180%);
  overflow: hidden;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease,
    background 0.2s ease;
}

.index-page :deep(.glass-card:hover) {
  transform: translateY(-4px);
  box-shadow:
    0 26px 60px rgba(46, 65, 105, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.74);
  border-color: rgba(255, 255, 255, 0.58);
}

.index-page :deep(.glass-card .ant-card-head) {
  min-height: 58px;
  padding-inline: 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.26);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.34), rgba(255, 255, 255, 0.12));
}

.index-page :deep(.glass-card .ant-card-head-title) {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.01em;
  color: rgba(15, 23, 42, 0.84);
}

.index-page :deep(.glass-card .ant-card-extra) {
  color: rgba(22, 119, 255, 0.96);
}

.index-page :deep(.glass-card .ant-card-body) {
  padding: 22px;
}

.index-page :deep(.glass-card .ant-card-actions) {
  margin: 0 14px 14px;
  background: rgba(255, 255, 255, 0.34);
  border-top: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 18px;
  overflow: hidden;
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}

.index-page :deep(.glass-card .ant-card-actions > li) {
  margin: 0;
  padding: 8px 0;
}

.index-page :deep(.glass-card .ant-card-actions > li > span) {
  min-height: 46px;
}

.index-page :deep(.hero-card .ant-card-body) {
  padding: 24px 20px 20px;
}

.index-page :deep(.hero-card .ant-progress-text) {
  font-size: 15px !important;
  font-weight: 700;
}

.index-page :deep(.brand-card .ant-card-body),
.index-page :deep(.action-card .ant-card-body) {
  padding-top: 18px;
}

.index-page :deep(.metric-card .ant-card-body) {
  padding-top: 18px;
}

.index-page :deep(.xray-card .ant-card-extra),
.index-page :deep(.xray-card .ant-card-head-title) {
  color: rgba(15, 23, 42, 0.88);
}

.index-page :deep(.xray-card .ant-card-actions > li > span),
.index-page :deep(.action-card .ant-card-actions > li > span) {
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-spacer {
  min-height: calc(100vh - 120px);
}

.action {
  cursor: pointer;
  justify-content: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.46), rgba(255, 255, 255, 0.22));
  border: 1px solid rgba(255, 255, 255, 0.42);
  box-shadow: 0 10px 22px rgba(46, 65, 105, 0.08);
}

.action :deep(.anticon) {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(22, 119, 255, 0.12);
  color: #1677ff;
}

.index-page.is-dark :deep(.action .anticon) {
  background: rgba(115, 145, 255, 0.16);
  color: rgba(170, 198, 255, 0.96);
}

.action-update {
  color: #fa8c16;
  font-weight: 600;
}

.action-update :deep(.anticon) {
  color: #fa8c16;
}

.history-tag {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-inline-end: 0;
}

.tg-icon {
  display: inline-block;
  vertical-align: -2px;
}

.ip-toggle-icon {
  cursor: pointer;
  font-size: 16px;
}

.ip-hidden :deep(.ant-statistic-content-value) {
  filter: blur(6px);
  transition: filter 0.2s ease;
}

.ip-visible :deep(.ant-statistic-content-value) {
  filter: none;
}

@media (prefers-reduced-motion: reduce) {
  .index-page::before,
  .index-page::after,
  .index-page :deep(.glass-card),
  .index-page :deep(.glass-card:hover) {
    transition: none;
  }
}
</style>
