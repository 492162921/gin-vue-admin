<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-select v-model="search.status" clearable placeholder="告警状态" class="w-32">
        <el-option label="待处理" value="open" />
        <el-option label="已确认" value="acknowledged" />
        <el-option label="已关闭" value="closed" />
      </el-select>
      <el-select v-model="search.level" clearable placeholder="告警级别" class="w-32">
        <el-option label="严重" value="critical" />
        <el-option label="警告" value="warning" />
      </el-select>
      <el-input v-model="search.ruleType" clearable class="w-40" placeholder="规则类型" @keyup.enter="loadAlerts" />
      <el-button icon="search" @click="loadAlerts">查询</el-button>
      <el-button :disabled="selectedIDs.length === 0" type="danger" plain @click="removeSelected">批量删除</el-button>
    </div>

    <el-table :data="alerts" row-key="ID" @selection-change="(rows) => { selectedIDs = rows.map((row) => row.ID) }">
      <el-table-column type="selection" width="48" />
      <el-table-column prop="ID" label="编号" width="78" />
      <el-table-column prop="ruleType" label="规则类型" min-width="150" />
      <el-table-column label="级别" width="90">
        <template #default="{ row }"><el-tag :type="row.level === 'critical' ? 'danger' : 'warning'">{{ row.level }}</el-tag></template>
      </el-table-column>
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-select :model-value="row.status" size="small" @change="(status) => setStatus(row, status)">
            <el-option label="待处理" value="open" />
            <el-option label="已确认" value="acknowledged" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="content" label="告警内容" min-width="280" show-overflow-tooltip />
      <el-table-column prop="collectedAt" label="采集时间" min-width="170" />
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }"><el-button link type="danger" @click="remove(row)">删除</el-button></template>
      </el-table-column>
    </el-table>
    <div class="gva-pagination">
      <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" />
    </div>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { batchDeleteAlerts, deleteAlert, getAlertList, updateAlertStatus } from '@/api/inspection'

  defineOptions({ name: 'InspectionAlert' })

  const alerts = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = 10
  const search = ref({ status: '', level: '', ruleType: '' })
  const selectedIDs = ref([])

  const loadAlerts = async () => {
    const result = await getAlertList({ page: page.value, pageSize, ...search.value })
    if (result.code === 0) {
      alerts.value = result.data.list
      total.value = result.data.total
    }
  }
  const changePage = (value) => {
    page.value = value
    loadAlerts()
  }
  const setStatus = async (alert, status) => {
    const result = await updateAlertStatus(alert.ID, { status })
    if (result.code === 0) {
      ElMessage.success('告警状态已更新')
      alert.status = status
    }
  }
  const remove = (alert) => ElMessageBox.confirm(`删除告警 #${alert.ID}？`, '确认删除', { type: 'warning' }).then(async () => {
    const result = await deleteAlert({ id: alert.ID })
    if (result.code === 0) {
      ElMessage.success('告警已删除')
      loadAlerts()
    }
  })
  const removeSelected = () => ElMessageBox.confirm(`删除已选的 ${selectedIDs.value.length} 条告警？`, '确认删除', { type: 'warning' }).then(async () => {
    const result = await batchDeleteAlerts({ ids: selectedIDs.value })
    if (result.code === 0) {
      ElMessage.success('告警已删除')
      selectedIDs.value = []
      loadAlerts()
    }
  })

  loadAlerts()
</script>
