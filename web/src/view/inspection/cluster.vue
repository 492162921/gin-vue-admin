<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-button type="primary" icon="plus" @click="openForm()">接入集群</el-button>
      <el-input v-model="search.name" class="w-56" clearable placeholder="按集群名称筛选" @change="loadClusters" />
      <el-button icon="search" @click="loadClusters">查询</el-button>
    </div>
    <el-table :data="clusters" row-key="ID">
      <el-table-column label="集群名称" prop="name" min-width="180" />
      <el-table-column label="状态" width="110">
        <template #default="{ row }"><el-tag :type="row.status === 'available' ? 'success' : 'danger'">{{ row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="Kubernetes 版本" prop="k8sVersion" min-width="160" />
      <el-table-column label="节点数" prop="nodeCount" width="100" />
      <el-table-column label="操作" width="260">
        <template #default="{ row }">
          <el-button link type="primary" @click="refresh(row)">刷新</el-button>
          <el-button link type="primary" @click="showResources(row)">资源</el-button>
          <el-button link type="primary" @click="openForm(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="gva-pagination">
      <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" />
    </div>

    <el-dialog v-model="formVisible" :title="form.ID ? '编辑集群' : '接入集群'" width="580px">
      <el-form :model="form" label-width="105px">
        <el-form-item label="集群名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="kubeconfig" required>
          <el-upload :auto-upload="false" :show-file-list="false" accept=".yaml,.yml,.conf" @change="readKubeconfig">
            <el-button>选择 kubeconfig 文件</el-button>
          </el-upload>
          <el-input v-model="form.kubeconfig" class="mt-3" type="textarea" :rows="6" placeholder="也可直接粘贴 kubeconfig 内容；内容不会在列表或接口回包中展示" />
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="formVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>

    <el-drawer v-model="resourceVisible" :title="`${selected?.name || ''} 的资源`" size="560px">
      <el-tabs>
        <el-tab-pane label="节点">
          <el-table :data="nodes"><el-table-column prop="name" label="名称" /><el-table-column prop="status" label="状态" /><el-table-column prop="CPUUsage" label="CPU %" /><el-table-column prop="MemUsage" label="内存 %" /></el-table>
        </el-tab-pane>
        <el-tab-pane label="命名空间">
          <el-table :data="namespaces"><el-table-column prop="name" label="名称" /><el-table-column prop="podCount" label="Pod 数量" /></el-table>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { createCluster, deleteCluster, getClusterList, getClusterNamespaces, getClusterNodes, refreshCluster, updateCluster } from '@/api/inspection'

  defineOptions({ name: 'InspectionCluster' })

  const clusters = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = 10
  const search = ref({ name: '' })
  const formVisible = ref(false)
  const form = ref({ name: '', kubeconfig: '' })
  const resourceVisible = ref(false)
  const selected = ref(null)
  const nodes = ref([])
  const namespaces = ref([])

  const loadClusters = async () => {
    const res = await getClusterList({ page: page.value, pageSize, ...search.value })
    if (res.code === 0) {
      clusters.value = res.data.list
      total.value = res.data.total
    }
  }
  const changePage = (value) => { page.value = value; loadClusters() }
  const openForm = (row) => {
    form.value = row ? { ID: row.ID, name: row.name, kubeconfig: '' } : { name: '', kubeconfig: '' }
    formVisible.value = true
  }
  const readKubeconfig = (file) => {
    const reader = new FileReader()
    reader.onload = () => { form.value.kubeconfig = reader.result }
    reader.readAsText(file.raw)
  }
  const save = async () => {
    if (!form.value.name || (!form.value.ID && !form.value.kubeconfig)) return ElMessage.warning('请填写名称并提供 kubeconfig')
    const res = form.value.ID ? await updateCluster(form.value.ID, form.value) : await createCluster(form.value)
    if (res.code === 0) { ElMessage.success('已保存'); formVisible.value = false; loadClusters() }
  }
  const refresh = async (row) => {
    const res = await refreshCluster({ id: row.ID })
    if (res.code === 0) { ElMessage.success('已刷新'); loadClusters() }
  }
  const showResources = async (row) => {
    selected.value = row
    const [nodeRes, namespaceRes] = await Promise.all([getClusterNodes({ id: row.ID }), getClusterNamespaces({ id: row.ID })])
    if (nodeRes.code === 0) nodes.value = nodeRes.data
    if (namespaceRes.code === 0) namespaces.value = namespaceRes.data
    resourceVisible.value = true
  }
  const remove = (row) => ElMessageBox.confirm(`删除集群“${row.name}”及其本地 kubeconfig？`, '确认删除', { type: 'warning' }).then(async () => {
    const res = await deleteCluster({ id: row.ID })
    if (res.code === 0) { ElMessage.success('已删除'); loadClusters() }
  })

  loadClusters()
</script>
