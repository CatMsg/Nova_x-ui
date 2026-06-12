<script setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  DashboardOutlined,
  UserOutlined,
  SettingOutlined,
  ToolOutlined,
  ClusterOutlined,
  LogoutOutlined,
  CloseOutlined,
  MenuOutlined,
  ApiOutlined,
} from '@ant-design/icons-vue';

import { theme, currentTheme, toggleTheme, toggleUltra, pauseAnimationsUntilLeave } from '@/composables/useTheme.js';

const { t } = useI18n();

const SIDEBAR_COLLAPSED_KEY = 'isSidebarCollapsed';

const props = defineProps({
  basePath: { type: String, default: '' },
  // Current request URI so the matching menu item highlights.
  requestUri: { type: String, default: '' },
});


const iconByName = {
  dashboard: DashboardOutlined,
  user: UserOutlined,
  setting: SettingOutlined,
  tool: ToolOutlined,
  cluster: ClusterOutlined,
  logout: LogoutOutlined,
  apidocs: ApiOutlined,
};

const prefix = props.basePath?.startsWith('/') ? props.basePath : `/${props.basePath || ''}`;

const tabs = computed(() => [
  { key: `${prefix}panel/`, icon: 'dashboard', title: t('menu.dashboard') },
  { key: `${prefix}panel/inbounds`, icon: 'user', title: t('menu.inbounds') },
  { key: `${prefix}panel/nodes`, icon: 'cluster', title: t('menu.nodes') },
  { key: `${prefix}panel/settings`, icon: 'setting', title: t('menu.settings') },
  { key: `${prefix}panel/xray`, icon: 'tool', title: t('menu.xray') },
  { key: `${prefix}panel/api-docs`, icon: 'apidocs', title: t('menu.apiDocs') },
  { key: `${prefix}logout`, icon: 'logout', title: t('logout') },
]);

const navTabs = computed(() => tabs.value.filter((tab) => tab.icon !== 'logout'));
const utilTabs = computed(() => tabs.value.filter((tab) => tab.icon === 'logout'));
const activeTab = ref([props.requestUri]);
const drawerOpen = ref(false);
const collapsed = ref(JSON.parse(localStorage.getItem(SIDEBAR_COLLAPSED_KEY) || 'false'));
const drawerWidth = 'min(82vw, 320px)';

function openLink(key) {
  if (key.startsWith('http')) {
    window.open(key);
  } else {
    window.location.href = key;
  }
}

function onCollapse(isCollapsed, type) {
  // Only persist explicit toggle clicks, not breakpoint-triggered collapses.
  if (type === 'clickTrigger') {
    localStorage.setItem(SIDEBAR_COLLAPSED_KEY, isCollapsed);
    collapsed.value = isCollapsed;
  }
}

function toggleDrawer() {
  drawerOpen.value = !drawerOpen.value;
}

function closeDrawer() {
  drawerOpen.value = false;
}

function cycleTheme() {
  pauseAnimationsUntilLeave('theme-cycle');
  if (!theme.isDark) {
    toggleTheme();
    if (theme.isUltra) toggleUltra();
  } else if (!theme.isUltra) {
    toggleUltra();
  } else {
    toggleUltra();
    toggleTheme();
  }
}
</script>

<template>
  <div class="ant-sidebar">
    <a-layout-sider :theme="currentTheme" collapsible :collapsed="collapsed" breakpoint="md" @collapse="onCollapse">
      <div class="sider-brand" :class="{ 'sider-brand-collapsed': collapsed }">
        <span class="brand-text">{{ collapsed ? 'Nova' : 'Nova_x-ui' }}</span>
        <button v-if="!collapsed" id="theme-cycle" type="button" class="theme-cycle" :aria-label="t('menu.theme')"
          :title="t('menu.theme')" @click="cycleTheme">
          <svg v-if="!theme.isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="4" />
            <path
              d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
          </svg>
          <svg v-else-if="!theme.isUltra" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            <path fill="none" d="M19 3l0.7 1.4 1.4 0.7-1.4 0.7L19 7.2l-0.7-1.4-1.4-0.7 1.4-0.7z" />
          </svg>
        </button>
      </div>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="sider-nav"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in navTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="sider-utility"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in utilTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-drawer placement="left" :closable="false" :open="drawerOpen" :wrap-class-name="currentTheme"
      :wrap-style="{ padding: 0 }" :width="drawerWidth"
      :body-style="{ padding: 0, display: 'flex', flexDirection: 'column', height: '100%' }"
      :header-style="{ display: 'none' }" @close="closeDrawer">
      <div class="drawer-header">
        <span class="drawer-brand">Nova_x-ui</span>
        <div class="drawer-header-actions">
          <button id="theme-cycle-drawer" type="button" class="theme-cycle" :aria-label="t('menu.theme')"
            :title="t('menu.theme')" @click="cycleTheme">
            <svg v-if="!theme.isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="12" cy="12" r="4" />
              <path
                d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
            </svg>
            <svg v-else-if="!theme.isUltra" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"
              stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              <path fill="none" d="M19 3l0.7 1.4 1.4 0.7-1.4 0.7L19 7.2l-0.7-1.4-1.4-0.7 1.4-0.7z" />
            </svg>
          </button>
          <button class="drawer-close" type="button" :aria-label="t('close')" @click="closeDrawer">
            <CloseOutlined />
          </button>
        </div>
      </div>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="drawer-menu drawer-nav"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in navTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
      <a-menu :theme="currentTheme" mode="inline" :selected-keys="activeTab" class="drawer-menu drawer-utility"
        @click="({ key }) => openLink(key)">
        <a-menu-item v-for="tab in utilTabs" :key="tab.key">
          <component :is="iconByName[tab.icon]" />
          <span>{{ tab.title }}</span>
        </a-menu-item>
      </a-menu>
    </a-drawer>

    <button v-show="!drawerOpen" class="drawer-handle" type="button" :aria-label="t('menu.dashboard')"
      @click="toggleDrawer">
      <MenuOutlined />
    </button>
  </div>
</template>

<style scoped>
.ant-sidebar {
  --sidebar-surface: rgba(255, 255, 255, 0.9);
  --sidebar-surface-soft: rgba(248, 250, 252, 0.92);
  --sidebar-border: rgba(15, 23, 42, 0.08);
  --sidebar-shadow: 0 18px 45px rgba(15, 23, 42, 0.12);
  --sidebar-text: rgba(15, 23, 42, 0.88);
  --sidebar-text-muted: rgba(15, 23, 42, 0.56);
  --sidebar-accent: #1677ff;
  --sidebar-accent-soft: rgba(22, 119, 255, 0.1);
}

body.dark .ant-sidebar {
  --sidebar-surface: rgba(20, 24, 33, 0.94);
  --sidebar-surface-soft: rgba(32, 37, 48, 0.96);
  --sidebar-border: rgba(255, 255, 255, 0.08);
  --sidebar-shadow: 0 22px 54px rgba(0, 0, 0, 0.42);
  --sidebar-text: rgba(255, 255, 255, 0.92);
  --sidebar-text-muted: rgba(255, 255, 255, 0.62);
  --sidebar-accent: #4096ff;
  --sidebar-accent-soft: rgba(64, 150, 255, 0.18);
}

html[data-theme='ultra-dark'] .ant-sidebar {
  --sidebar-surface: rgba(8, 8, 10, 0.98);
  --sidebar-surface-soft: rgba(18, 18, 20, 0.98);
  --sidebar-border: rgba(255, 255, 255, 0.06);
  --sidebar-shadow: 0 22px 54px rgba(0, 0, 0, 0.58);
  --sidebar-text: rgba(255, 255, 255, 0.96);
  --sidebar-text-muted: rgba(255, 255, 255, 0.68);
  --sidebar-accent: #66aaff;
  --sidebar-accent-soft: rgba(102, 170, 255, 0.16);
}

.ant-sidebar>.ant-layout-sider {
  position: sticky;
  top: 0;
  height: 100vh;
  align-self: flex-start;
  margin: 16px 0 16px 16px;
  background:
    radial-gradient(circle at top, rgba(22, 119, 255, 0.18), transparent 42%),
    radial-gradient(circle at bottom, rgba(90, 224, 214, 0.1), transparent 36%),
    linear-gradient(180deg, var(--sidebar-surface) 0%, var(--sidebar-surface-soft) 100%);
  border: 1px solid var(--sidebar-border);
  border-radius: 30px;
  box-shadow: var(--sidebar-shadow);
  backdrop-filter: blur(28px) saturate(180%);
  -webkit-backdrop-filter: blur(28px) saturate(180%);
  overflow: hidden;
}

.sider-brand,
.drawer-brand {
  font-weight: 600;
  font-size: 16px;
  letter-spacing: 0.08em;
  color: var(--sidebar-text);
  text-transform: uppercase;
}

.sider-brand {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 20px 18px 16px;
  border-bottom: 1px solid var(--sidebar-border);
  background:
    linear-gradient(135deg, rgba(22, 119, 255, 0.12), rgba(255, 255, 255, 0.02) 55%),
    rgba(255, 255, 255, 0.03);
  user-select: none;
}

/* Collapsed sider only has room for the 'Nova' brand — center it and
 * hide the theme cycle button (which is `v-if`-ed out in template). */
.sider-brand-collapsed {
  justify-content: center;
  font-size: 15px;
  padding: 18px 4px 16px;
  letter-spacing: 0.04em;
}

.brand-text {
  flex: 1 1 auto;
}

.sider-brand-collapsed .brand-text {
  flex: 0 0 auto;
}

.theme-cycle {
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.24);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--sidebar-text);
  padding: 0;
  flex-shrink: 0;
  transition: background-color 0.2s, transform 0.15s, color 0.2s, box-shadow 0.2s;
}

.theme-cycle:hover,
.theme-cycle:focus-visible {
  background: rgba(255, 255, 255, 0.28);
  color: var(--sidebar-accent);
  transform: translateY(-1px) scale(1.04);
  box-shadow: 0 14px 26px rgba(22, 119, 255, 0.18);
  outline: none;
}

.theme-cycle svg {
  width: 17px;
  height: 17px;
}

.drawer-header-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.drawer-handle {
  position: fixed;
  top: 14px;
  left: 14px;
  z-index: 1100;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.34), rgba(255, 255, 255, 0.16));
  color: var(--sidebar-text);
  border: 1px solid rgba(255, 255, 255, 0.34);
  width: 44px;
  height: 44px;
  border-radius: 50%;
  cursor: pointer;
  display: none;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  box-shadow: 0 16px 32px rgba(46, 65, 105, 0.18);
  backdrop-filter: blur(22px) saturate(180%);
  -webkit-backdrop-filter: blur(22px) saturate(180%);
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid var(--sidebar-border);
  background:
    linear-gradient(135deg, rgba(22, 119, 255, 0.12), rgba(255, 255, 255, 0.02) 55%),
    rgba(255, 255, 255, 0.03);
}

.drawer-close {
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.24);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  color: var(--sidebar-text);
  transition: background-color 0.2s, transform 0.15s, box-shadow 0.2s;
}

.drawer-close:hover,
.drawer-close:focus-visible {
  background: rgba(255, 255, 255, 0.28);
  box-shadow: 0 14px 26px rgba(22, 119, 255, 0.16);
  transform: translateY(-1px);
}

.drawer-menu :deep(.ant-menu-item) {
  height: 46px;
  line-height: 46px;
  margin: 6px 12px;
  border-radius: 16px;
  padding-inline: 16px !important;
  transition: transform 0.18s ease, background-color 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}

.drawer-menu :deep(.ant-menu-item .anticon) {
  font-size: 16px;
}

.drawer-utility {
  margin-top: auto;
  border-top: 1px solid var(--sidebar-border);
  padding-top: 8px;
}

.ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-children) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sider-brand {
  flex: 0 0 auto;
}

.sider-nav {
  flex: 1 1 auto;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
  padding: 10px 0 6px;
}

.sider-utility {
  flex: 0 0 auto;
  border-top: 1px solid var(--sidebar-border);
  padding-top: 10px;
}

@media (max-width: 768px) {
  .drawer-handle {
    display: inline-flex;
  }

  .ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-children),
  .ant-sidebar>.ant-layout-sider :deep(.ant-layout-sider-trigger) {
    display: none;
  }

  .ant-sidebar>.ant-layout-sider {
    margin: 0;
    flex: 0 0 0 !important;
    max-width: 0 !important;
    min-width: 0 !important;
    width: 0 !important;
  }
}
</style>

<style>
.sider-nav .ant-menu-item-selected,
.sider-utility .ant-menu-item-selected,
.drawer-menu .ant-menu-item-selected {
  background: linear-gradient(135deg, rgba(22, 119, 255, 0.26), rgba(22, 119, 255, 0.14)) !important;
  color: var(--sidebar-accent) !important;
  box-shadow:
    inset 0 0 0 1px rgba(22, 119, 255, 0.14),
    0 10px 22px rgba(22, 119, 255, 0.12);
}

.sider-nav .ant-menu-item-active:not(.ant-menu-item-selected),
.sider-utility .ant-menu-item-active:not(.ant-menu-item-selected),
.drawer-menu .ant-menu-item-active:not(.ant-menu-item-selected),
.sider-nav .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover,
.sider-utility .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover,
.drawer-menu .ant-menu-item:not(.ant-menu-item-selected):not(.ant-menu-item-disabled):hover {
  background-color: rgba(255, 255, 255, 0.22) !important;
  color: var(--sidebar-accent) !important;
  transform: translateX(2px);
}

body.dark .ant-drawer .ant-drawer-content,
body.dark .ant-drawer .ant-drawer-body {
  background:
    radial-gradient(circle at top, rgba(22, 119, 255, 0.12), transparent 44%),
    linear-gradient(180deg, rgba(20, 24, 33, 0.98), rgba(15, 18, 26, 0.98)) !important;
}

html[data-theme='ultra-dark'] .ant-drawer .ant-drawer-content,
html[data-theme='ultra-dark'] .ant-drawer .ant-drawer-body {
  background:
    radial-gradient(circle at top, rgba(102, 170, 255, 0.08), transparent 40%),
    linear-gradient(180deg, rgba(8, 8, 10, 0.99), rgba(4, 4, 6, 0.99)) !important;
}
</style>
