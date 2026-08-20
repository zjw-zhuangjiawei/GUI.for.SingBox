<script setup lang="ts">
import { computed, ref } from 'vue'

import Input from '@/components/Input/index.vue'
import Tag from '@/components/Tag/index.vue'

interface Props {
  modelValue?: string[]
  /** items always shown, non-removable, rendered weakened */
  fixed?: string[]
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => [],
  fixed: () => [],
  placeholder: '',
})

const emit = defineEmits(['change', 'update:modelValue'])

const draft = ref('')

const userItems = computed(() => {
  const seen = new Set<string>()
  return props.modelValue.filter(
    (v) => v && !props.fixed.includes(v) && !seen.has(v) && seen.add(v),
  )
})

const add = (value: string) => {
  const v = String(value ?? '').trim()
  draft.value = ''
  if (!v || props.fixed.includes(v) || userItems.value.includes(v)) return
  const next = [...userItems.value, v]
  emit('update:modelValue', next)
  emit('change', next)
}

const remove = (index: number) => {
  const next = [...userItems.value]
  next.splice(index, 1)
  emit('update:modelValue', next)
  emit('change', next)
}

const onKeydown = (e: KeyboardEvent) => {
  if (e.isComposing || e.keyCode === 229) return
  if (e.key === 'Backspace' && !draft.value && userItems.value.length) {
    remove(userItems.value.length - 1)
  }
}
</script>

<template>
  <Input
    v-model="draft"
    v-bind="$attrs"
    class="gui-tag-input"
    :placeholder="placeholder"
    @submit="add"
    @keydown="onKeydown"
  >
    <Tag v-for="item in fixed" :key="'fixed-' + item" size="small" class="gui-tag-input__fixed">
      {{ item }}
    </Tag>
    <Tag
      v-for="(item, index) in userItems"
      :key="'user-' + item"
      size="small"
      closeable
      @close="remove(index)"
    >
      {{ item }}
    </Tag>
  </Input>
</template>

<style lang="less" scoped>
.gui-tag-input {
  :deep(.gui-tag) {
    border-radius: 4px;
    margin: 2px;
    padding: 0 6px;
  }
}
.gui-tag-input__fixed {
  opacity: 0.5;
  pointer-events: none;
}
</style>
