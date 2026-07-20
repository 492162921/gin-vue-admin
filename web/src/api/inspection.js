import service from '@/utils/request'

const request = (url, method, data, params) =>
  service({ url, method, data, params })

export const createCluster = (data) =>
  request('/inspection/cluster', 'post', data)
export const updateCluster = (id, data) =>
  request('/inspection/cluster', 'put', data, { id })
export const deleteCluster = (data) =>
  request('/inspection/cluster', 'delete', data)
export const getCluster = (params) =>
  request('/inspection/cluster', 'get', undefined, params)
export const getClusterList = (params) =>
  request('/inspection/cluster/list', 'get', undefined, params)
export const refreshCluster = (data) =>
  request('/inspection/cluster/refresh', 'post', data)
export const getClusterNodes = (params) =>
  request('/inspection/cluster/nodes', 'get', undefined, params)
export const getClusterNamespaces = (params) =>
  request('/inspection/cluster/namespaces', 'get', undefined, params)

export const createRule = (data) => request('/inspection/rule', 'post', data)
export const updateRule = (id, data) =>
  request('/inspection/rule', 'put', data, { id })
export const deleteRule = (data) => request('/inspection/rule', 'delete', data)
export const getRuleList = (params) =>
  request('/inspection/rule/list', 'get', undefined, params)
export const setRuleEnabled = (data) =>
  request('/inspection/rule/enabled', 'post', data)

export const createTask = (data) => request('/inspection/task', 'post', data)
export const updateTask = (id, data) =>
  request('/inspection/task', 'put', data, { id })
export const deleteTask = (data) => request('/inspection/task', 'delete', data)
export const getTask = (params) =>
  request('/inspection/task', 'get', undefined, params)
export const getTaskList = (params) =>
  request('/inspection/task/list', 'get', undefined, params)
export const runTask = (data) => request('/inspection/task/run', 'post', data)
export const getTaskInspections = (params) =>
  request('/inspection/task/inspections', 'get', undefined, params)

export const getAlertList = (params) =>
  request('/inspection/alert/list', 'get', undefined, params)
export const updateAlertStatus = (id, data) =>
  request('/inspection/alert/status', 'patch', data, { id })
export const deleteAlert = (data) => request('/inspection/alert', 'delete', data)
export const batchDeleteAlerts = (data) =>
  request('/inspection/alert/batch-delete', 'delete', data)
