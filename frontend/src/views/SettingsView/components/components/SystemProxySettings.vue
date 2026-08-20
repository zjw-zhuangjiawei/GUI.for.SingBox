<script lang="ts" setup>
import { computed } from 'vue'

import { OS } from '@/enums/app'
import { useAppSettingsStore, useEnvStore } from '@/stores'

const appSettings = useAppSettingsStore()
const envStore = useEnvStore()

// Always bypassed so the app's own Clash API calls never loop back into the proxy.
const ALWAYS_BYPASS = ['localhost', '127.0.0.1', '::1']

const bypassItems = computed({
  get: () =>
    appSettings.app.proxyBypassList
      .split(';')
      .map((v) => v.trim())
      .filter(Boolean),
  set: (val: string[]) => {
    appSettings.app.proxyBypassList = val.join(';')
  },
})
</script>

<template>
  <div class="px-8 py-12 text-18 font-bold">{{ $t('settings.systemProxy') }}</div>

  <Card>
    <div class="px-8 py-12 flex items-center justify-between">
      <div class="text-16 font-bold">{{ $t('settings.autoSetSystemProxy') }}</div>
      <Switch v-model="appSettings.app.autoSetSystemProxy" />
    </div>
    <div
      v-platform="[OS.Darwin, OS.Linux]"
      :class="appSettings.app.systemProxyServices.length ? 'items-start' : 'items-center'"
      class="px-8 py-12 flex justify-between"
    >
      <div class="text-16 font-bold">{{ $t('settings.systemProxyServices') }}</div>
      <InputList
        v-model="appSettings.app.systemProxyServices"
        :placeholder="envStore.env.os === OS.Darwin ? 'Ethernet、Wi-Fi' : 'Wired connection 1'"
      />
    </div>
    <div v-platform="[OS.Darwin, OS.Linux]" class="px-8 py-12 flex items-center justify-between">
      <div class="text-16 font-bold">
        {{ $t('settings.autoSetSystemDNS') }}
        <span class="font-normal text-12">({{ $t('settings.autoSetSystemDNSTips') }})</span>
      </div>
      <Switch v-model="appSettings.app.autoSetSystemDNS" />
    </div>
    <div
      v-platform="[OS.Darwin, OS.Linux]"
      class="px-8 py-12 flex items-center justify-between gap-16"
    >
      <div class="text-16 font-bold">{{ $t('settings.systemProxyDNS') }}</div>
      <Input
        v-model="appSettings.app.systemProxyDNS"
        clearable
        editable
        placeholder="1.1.1.1,8.8.8.8"
      />
    </div>
    <div
      v-platform="[OS.Darwin, OS.Linux]"
      class="px-8 py-12 flex items-center justify-between gap-16"
    >
      <div class="text-16 font-bold">{{ $t('settings.systemDefaultDNS') }}</div>
      <Input
        v-model="appSettings.app.systemDefaultDNS"
        editable
        clearable
        placeholder="192.168.1.1,10.0.0.1"
      />
    </div>
    <div class="px-8 pt-12 pb-8 flex flex-col gap-12">
      <div class="text-16 font-bold">
        {{ $t('settings.proxyBypassList') }}
        <span class="font-normal text-12">({{ $t('settings.proxyBypassListTips') }})</span>
      </div>
      <TagInput
        v-model="bypassItems"
        :fixed="ALWAYS_BYPASS"
        :placeholder="$t('settings.proxyBypassListPlaceholder')"
        class="min-w-256"
      />
    </div>
  </Card>
</template>
