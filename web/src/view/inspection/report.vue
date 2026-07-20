<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-input v-model.number="inspectionID" class="w-44" clearable placeholder="巡检ID" @keyup.enter="generate" />
      <el-button type="primary" icon="magic-stick" :loading="generating" @click="generate">生成报告</el-button>
      <el-input v-model="search.title" clearable class="w-48" placeholder="报告标题" @keyup.enter="loadReports" />
      <el-button icon="search" @click="loadReports">查询</el-button>
    </div>

    <el-table :data="reports" row-key="ID">
      <el-table-column prop="ID" label="编号" width="80" />
      <el-table-column prop="title" label="标题" min-width="180" />
      <el-table-column prop="inspectionId" label="巡检ID" width="100" />
      <el-table-column prop="digest" label="摘要" min-width="260" show-overflow-tooltip />
      <el-table-column label="来源" width="100">
        <template #default="{ row }"><el-tag :type="row.source === 'ai' ? 'success' : 'info'">{{ row.source }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="CreatedAt" label="生成时间" min-width="170" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="showDetail(row)">详情</el-button>
          <el-button link type="success" @click="push(row)">推送</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="gva-pagination">
      <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" />
    </div>

    <el-drawer v-model="detailVisible" :title="detail?.title" size="55%">
      <pre class="whitespace-pre-wrap break-words text-sm leading-6">{{ detail?.contentMd }}</pre>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { deleteReport, generateReport, getReport, getReportList, pushReport } from '@/api/inspection'

  defineOptions({ name: 'InspectionReport' })

  const reports = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = 10
  const search = ref({ title: '' })
  const inspectionID = ref()
  const generating = ref(false)
  const detail = ref()
  const detailVisible = ref(false)

  const loadReports = async () => {
    const result = await getReportList({ page: page.value, pageSize, ...search.value })
    if (result.code === 0) {
      reports.value = result.data.list
      total.value = result.data.total
    }
  }
  const changePage = (value) => {
    page.value = value
    loadReports()
  }
  const generate = async () => {
    if (!inspectionID.value) {
      ElMessage.warning('请输入巡检ID')
      return
    }
    generating.value = true
    try {
      const result = await generateReport({ id: inspectionID.value })
      if (result.code === 0) {
        ElMessage.success('报告已生成')
        loadReports()
      }
    } finally {
      generating.value = false
    }
  }
  const showDetail = async (report) => {
    const result = await getReport({ id: report.ID })
    if (result.code === 0) {
      detail.value = result.data
      detailVisible.value = true
    }
  }
  const push = async (report) => {
    const result = await pushReport({ id: report.ID })
    if (result.code === 0) ElMessage.success('报告摘要已推送')
  }
  const remove = (report) => ElMessageBox.confirm(`删除报告 #${report.ID}？`, '确认删除', { type: 'warning' }).then(async () => {
    const result = await deleteReport({ id: report.ID })
    if (result.code === 0) {
      ElMessage.success('报告已删除')
      loadReports()
    }
  })

  loadReports()
</script>
