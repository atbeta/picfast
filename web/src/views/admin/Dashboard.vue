<template>
  <div>
    <n-h3>仪表盘</n-h3>
    <n-grid :cols="4" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
      <n-gi span="4 sm:2 md:1">
        <n-card size="small"><n-statistic label="用户数">{{ stats.users }}</n-statistic></n-card>
      </n-gi>
      <n-gi span="4 sm:2 md:1">
        <n-card size="small"><n-statistic label="图片数">{{ stats.images }}</n-statistic></n-card>
      </n-gi>
      <n-gi span="4 sm:2 md:1">
        <n-card size="small"><n-statistic label="分组数">{{ stats.groups }}</n-statistic></n-card>
      </n-gi>
      <n-gi span="4 sm:2 md:1">
        <n-card size="small"><n-statistic label="存储策略">{{ stats.strategies }}</n-statistic></n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NH3, NGrid, NGi, NCard, NStatistic } from 'naive-ui'
import { adminGetUsers, adminGetImages, adminGetGroups, adminGetStrategies } from '../../api/admin'

const stats = ref({ users: 0, images: 0, groups: 0, strategies: 0 })

onMounted(async () => {
  try {
    const [users, images, groups, strategies] = await Promise.all([
      adminGetUsers(),
      adminGetImages(),
      adminGetGroups(),
      adminGetStrategies(),
    ])
    const ug = groups.data.data
    const sg = strategies.data.data
    stats.value = {
      users: users.data.data?.total || users.data.pagination?.total || (users.data.data?.items || users.data.data || []).length,
      images: images.data.data?.total || images.data.pagination?.total || (images.data.data?.items || images.data.data || []).length,
      groups: Array.isArray(ug) ? ug.length : (ug?.items || []).length,
      strategies: Array.isArray(sg) ? sg.length : (sg?.items || []).length,
    }
  } catch { /* */ }
})
</script>
