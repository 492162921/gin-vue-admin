<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-button type="primary" icon="plus" @click="openForm()">新增规则</el-button>
      <el-input v-model="search.name" class="w-56" clearable placeholder="规则名称" @change="loadRules" />
      <el-button icon="search" @click="loadRules">查询</el-button>
    </div>
    <el-table :data="rules" row-key="ID">
      <el-table-column prop="name" label="规则名称" min-width="160" />
      <el-table-column prop="ruleType" label="规则类型" min-width="190" />
      <el-table-column prop="threshold" label="阈值" width="100" />
      <el-table-column label="范围" min-width="160"><template #default="{ row }">{{ row.scope === 'namespace' ? `命名空间：${row.namespace}` : '整个集群' }}</template></el-table-column>
      <el-table-column label="启用" width="100"><template #default="{ row }"><el-switch :model-value="row.enabled" @change="toggle(row, $event)" /></template></el-table-column>
      <el-table-column label="操作" width="140"><template #default="{ row }"><el-button link type="primary" @click="openForm(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
    </el-table>
    <div class="gva-pagination"><el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" /></div>

    <el-dialog v-model="formVisible" :title="form.ID ? '编辑规则' : '新增规则'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="规则名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="规则类型" required><el-select v-model="form.ruleType" class="w-full" filterable><el-option v-for="type in ruleTypes" :key="type" :label="type" :value="type" /></el-select></el-form-item>
        <el-form-item label="阈值"><el-input-number v-model="form.threshold" :min="0" /></el-form-item>
        <el-form-item label="作用域"><el-radio-group v-model="form.scope"><el-radio value="cluster">整个集群</el-radio><el-radio value="namespace">命名空间</el-radio></el-radio-group></el-form-item>
        <el-form-item v-if="form.scope === 'namespace'" label="命名空间" required><el-input v-model="form.namespace" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { createRule, deleteRule, getRuleList, setRuleEnabled, updateRule } from '@/api/inspection'

  defineOptions({ name: 'InspectionRule' })

  const ruleTypes = ['node_cpu', 'node_memory', 'pod_restart', 'pod_restart_high', 'pod_not_running', 'node_not_ready', 'node_disk_pressure', 'node_memory_pressure', 'node_pid_pressure', 'node_network_unavailable', 'node_unschedulable', 'pod_pending', 'pod_failed', 'pod_unknown', 'pod_oomkilled', 'pod_image_pull_backoff', 'pod_crash_loop', 'pod_not_ready', 'pod_evicted', 'container_not_ready']
  const rules = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = 10
  const search = ref({ name: '' })
  const formVisible = ref(false)
  const form = ref({})

  const loadRules = async () => {
    const res = await getRuleList({ page: page.value, pageSize, ...search.value })
    if (res.code === 0) { rules.value = res.data.list; total.value = res.data.total }
  }
  const changePage = (value) => { page.value = value; loadRules() }
  const openForm = (row) => {
    form.value = row ? { ...row } : { name: '', ruleType: 'node_cpu', threshold: 0, scope: 'cluster', namespace: '' }
    formVisible.value = true
  }
  const save = async () => {
    if (!form.value.name || !form.value.ruleType || (form.value.scope === 'namespace' && !form.value.namespace)) return ElMessage.warning('请填写规则必填项')
    const res = form.value.ID ? await updateRule(form.value.ID, form.value) : await createRule(form.value)
    if (res.code === 0) { ElMessage.success('已保存'); formVisible.value = false; loadRules() }
  }
  const toggle = async (row, enabled) => {
    const res = await setRuleEnabled({ id: row.ID, enabled })
    if (res.code === 0) { ElMessage.success('状态已更新'); loadRules() }
  }
  const remove = (row) => ElMessageBox.confirm(`删除规则“${row.name}”？绑定到任务的规则不能删除。`, '确认删除', { type: 'warning' }).then(async () => {
    const res = await deleteRule({ id: row.ID })
    if (res.code === 0) { ElMessage.success('已删除'); loadRules() }
  })

  loadRules()
</script>
