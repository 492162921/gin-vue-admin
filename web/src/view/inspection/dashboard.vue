<template>
  <div class="dashboard">
    <el-row :gutter="16" class="summary-row">
      <el-col v-for="card in summaryCards" :key="card.label" :xs="12" :sm="12" :md="6">
        <el-card shadow="never" class="summary-card">
          <div class="summary-label">{{ card.label }}</div>
          <div class="summary-value">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="chart-card">
          <template #header>近 7 天异常趋势</template>
          <Chart height="320px" :option="trendOption" />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="10">
        <el-card shadow="never" class="chart-card">
          <template #header>告警级别分布</template>
          <Chart height="320px" :option="levelOption" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="chart-card">
          <template #header>集群 CPU / 内存使用率</template>
          <Chart height="320px" :option="usageOption" />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="10">
        <el-card shadow="never" class="chart-card">
          <template #header>集群告警分布</template>
          <Chart height="320px" :option="clusterOption" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import Chart from '@/components/charts/index.vue'
  import {
    getDashboardAlertDistribution,
    getDashboardInspectionTrend,
    getDashboardResourceUsage,
    getDashboardSummary
  } from '@/api/inspection'

  defineOptions({ name: 'InspectionDashboard' })

  const summary = ref({
    clusterCount: 0,
    nodeCount: 0,
    todayInspectionCount: 0,
    openAlertCount: 0
  })
  const usage = ref([])
  const trend = ref([])
  const distribution = ref({ byLevel: [], byCluster: [] })

  const summaryCards = computed(() => [
    { label: '集群数量', value: summary.value.clusterCount },
    { label: '节点数量', value: summary.value.nodeCount },
    { label: '今日巡检次数', value: summary.value.todayInspectionCount },
    { label: '未关闭告警', value: summary.value.openAlertCount }
  ])

  const trendOption = computed(() => ({
    tooltip: { trigger: 'axis' },
    grid: { left: 48, right: 24, top: 28, bottom: 36 },
    xAxis: { type: 'category', data: trend.value.map((item) => item.date.slice(5)) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      name: '异常数',
      type: 'line',
      smooth: true,
      data: trend.value.map((item) => item.anomalyCount),
      areaStyle: { opacity: 0.16 },
      lineStyle: { width: 3 },
      itemStyle: { color: '#f56c6c' }
    }]
  }))

  const usageOption = computed(() => ({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: ['CPU', '内存'] },
    grid: { left: 48, right: 24, top: 48, bottom: 60 },
    xAxis: {
      type: 'category',
      data: usage.value.map((item) => item.clusterName),
      axisLabel: { interval: 0, rotate: 20 }
    },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      { name: 'CPU', type: 'bar', data: usage.value.map((item) => item.cpuUsage), itemStyle: { color: '#409eff' } },
      { name: '内存', type: 'bar', data: usage.value.map((item) => item.memoryUsage), itemStyle: { color: '#67c23a' } }
    ]
  }))

  const pieOption = (data) => ({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, type: 'scroll' },
    series: [{
      type: 'pie',
      radius: ['42%', '70%'],
      avoidLabelOverlap: true,
      label: { formatter: '{b}: {c}' },
      data: data.map((item) => ({ name: item.name, value: item.count }))
    }]
  })
  const levelOption = computed(() => pieOption(distribution.value.byLevel))
  const clusterOption = computed(() => pieOption(distribution.value.byCluster))

  const loadDashboard = async () => {
    try {
      const [summaryResult, usageResult, trendResult, distributionResult] = await Promise.all([
        getDashboardSummary(),
        getDashboardResourceUsage(),
        getDashboardInspectionTrend({ days: 7 }),
        getDashboardAlertDistribution()
      ])
      if (summaryResult.code === 0) summary.value = summaryResult.data
      if (usageResult.code === 0) usage.value = usageResult.data
      if (trendResult.code === 0) trend.value = trendResult.data
      if (distributionResult.code === 0) distribution.value = distributionResult.data
    } catch (error) {
      ElMessage.error('加载巡检总览数据失败')
    }
  }

  onMounted(loadDashboard)
</script>

<style scoped lang="scss">
  .dashboard {
    padding: 16px;
  }

  .summary-row,
  .chart-card {
    margin-bottom: 16px;
  }

  .summary-card {
    margin-bottom: 16px;
  }

  .summary-label {
    color: var(--el-text-color-secondary);
    font-size: 14px;
  }

  .summary-value {
    margin-top: 10px;
    color: var(--el-text-color-primary);
    font-size: 30px;
    font-weight: 600;
  }
</style>
