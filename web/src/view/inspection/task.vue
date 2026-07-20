<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-button type="primary" icon="plus" @click="openForm()">新增任务</el-button>
      <el-input v-model="search.name" class="w-56" clearable placeholder="任务名称" @change="loadTasks" />
      <el-button icon="search" @click="loadTasks">查询</el-button>
    </div>
    <el-table :data="tasks" row-key="ID">
      <el-table-column prop="name" label="任务名称" min-width="150" />
      <el-table-column label="集群" min-width="140"><template #default="{ row }">{{ row.cluster?.name || '-' }}</template></el-table-column>
      <el-table-column prop="cronExpr" label="Cron 表达式" min-width="130"><template #default="{ row }">{{ row.cronExpr || '仅手动执行' }}</template></el-table-column>
      <el-table-column label="规则数" width="90"><template #default="{ row }">{{ row.rules?.length || 0 }}</template></el-table-column>
      <el-table-column prop="status" label="状态" width="100"><template #default="{ row }"><el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ row.status }}</el-tag></template></el-table-column>
      <el-table-column label="操作" min-width="250" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :loading="runningID === row.ID" @click="run(row)">立即执行</el-button>
          <el-button link type="primary" @click="showHistory(row)">历史</el-button>
          <el-button link type="primary" @click="openForm(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="gva-pagination"><el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" /></div>

    <el-dialog v-model="formVisible" :title="form.ID ? '编辑任务' : '新增任务'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="任务名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="目标集群" required><el-select v-model="form.clusterId" class="w-full"><el-option v-for="cluster in clusters" :key="cluster.ID" :label="cluster.name" :value="cluster.ID" /></el-select></el-form-item>
        <el-form-item label="巡检规则" required><el-select v-model="form.ruleIds" class="w-full" multiple filterable><el-option v-for="rule in rules" :key="rule.ID" :label="`${rule.name} (${rule.ruleType})`" :value="rule.ID" /></el-select></el-form-item>
        <el-form-item label="Cron 表达式"><el-input v-model="form.cronExpr" placeholder="留空仅支持手动执行" /></el-form-item>
        <el-form-item v-if="form.ID" label="状态"><el-select v-model="form.status" class="w-full"><el-option label="active" value="active" /><el-option label="inactive" value="inactive" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>

    <el-drawer v-model="historyVisible" :title="`${historyTask?.name || ''} · 巡检历史`" size="70%">
      <el-table :data="inspections" row-key="ID">
        <el-table-column prop="ID" label="编号" width="80" />
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column prop="anomalyCount" label="异常数" width="100"><template #default="{ row }"><el-tag :type="row.anomalyCount ? 'danger' : 'success'">{{ row.anomalyCount }}</el-tag></template></el-table-column>
        <el-table-column prop="summary" label="摘要" min-width="220" />
        <el-table-column prop="startedAt" label="开始时间" min-width="170" />
        <el-table-column label="明细" width="80"><template #default="{ row }"><el-button link type="primary" @click="selectedDetails = row.details || []">查看</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination :current-page="historyPage" :page-size="pageSize" :total="historyTotal" layout="total, prev, pager, next" @current-change="changeHistoryPage" /></div>
      <el-divider content-position="left">巡检明细</el-divider>
      <el-table :data="selectedDetails" max-height="300">
        <el-table-column prop="ruleType" label="规则" width="180" />
        <el-table-column prop="resourceName" label="资源" min-width="180" />
        <el-table-column prop="message" label="结果" min-width="180" />
        <el-table-column prop="isAnomaly" label="异常" width="80"><template #default="{ row }"><el-tag :type="row.isAnomaly ? 'danger' : 'success'">{{ row.isAnomaly ? '是' : '否' }}</el-tag></template></el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { createTask, deleteTask, getClusterList, getRuleList, getTaskInspections, getTaskList, runTask, updateTask } from '@/api/inspection'

  defineOptions({ name: 'InspectionTask' })

  const tasks = ref([])
  const clusters = ref([])
  const rules = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = 10
  const search = ref({ name: '' })
  const formVisible = ref(false)
  const form = ref({})
  const runningID = ref()
  const historyVisible = ref(false)
  const historyTask = ref()
  const inspections = ref([])
  const historyTotal = ref(0)
  const historyPage = ref(1)
  const selectedDetails = ref([])

  const loadOptions = async () => {
    const [clusterRes, ruleRes] = await Promise.all([getClusterList({ page: 1, pageSize: 999 }), getRuleList({ page: 1, pageSize: 999 })])
    if (clusterRes.code === 0) clusters.value = clusterRes.data.list
    if (ruleRes.code === 0) rules.value = ruleRes.data.list
  }
  const loadTasks = async () => {
    const res = await getTaskList({ page: page.value, pageSize, ...search.value })
    if (res.code === 0) { tasks.value = res.data.list; total.value = res.data.total }
  }
  const changePage = (value) => { page.value = value; loadTasks() }
  const openForm = (row) => {
    form.value = row
      ? { ...row, clusterId: row.clusterId, ruleIds: row.rules?.map((rule) => rule.ID) || [] }
      : { name: '', clusterId: undefined, ruleIds: [], cronExpr: '', status: 'active' }
    formVisible.value = true
  }
  const save = async () => {
    if (!form.value.name || !form.value.clusterId || !form.value.ruleIds?.length) return ElMessage.warning('请填写任务、集群和规则')
    const res = form.value.ID ? await updateTask(form.value.ID, form.value) : await createTask(form.value)
    if (res.code === 0) { ElMessage.success('已保存'); formVisible.value = false; loadTasks() }
  }
  const run = async (row) => {
    runningID.value = row.ID
    try {
      const res = await runTask({ id: row.ID })
      if (res.code === 0) ElMessage.success(`巡检完成，发现 ${res.data.anomalyCount} 项异常`)
    } finally {
      runningID.value = undefined
    }
  }
  const remove = (row) => ElMessageBox.confirm(`删除任务“${row.name}”？`, '确认删除', { type: 'warning' }).then(async () => {
    const res = await deleteTask({ id: row.ID })
    if (res.code === 0) { ElMessage.success('已删除'); loadTasks() }
  })
  const loadHistory = async () => {
    if (!historyTask.value) return
    const res = await getTaskInspections({ taskId: historyTask.value.ID, page: historyPage.value, pageSize })
    if (res.code === 0) { inspections.value = res.data.list; historyTotal.value = res.data.total; selectedDetails.value = [] }
  }
  const showHistory = (row) => { historyTask.value = row; historyPage.value = 1; historyVisible.value = true; loadHistory() }
  const changeHistoryPage = (value) => { historyPage.value = value; loadHistory() }

  loadOptions()
  loadTasks()
</script>
