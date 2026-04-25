<template>
  <div>
    <n-h3>Dashboard</n-h3>
    <n-grid :cols="4" :x-gap="12" :y-gap="12">
      <n-gi>
        <n-card size="small">
          <n-statistic label="Users">
            {{ stats.users }}
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="Images">
            {{ stats.images }}
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="Groups">
            {{ stats.groups }}
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="Strategies">
            {{ stats.strategies }}
          </n-statistic>
        </n-card>
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
    stats.value = {
      users: users.data.pagination?.total || 0,
      images: images.data.pagination?.total || 0,
      groups: Array.isArray(groups.data.data) ? groups.data.data.length : 0,
      strategies: Array.isArray(strategies.data.data) ? strategies.data.data.length : 0,
    }
  } catch { /* */ }
})
</script>
