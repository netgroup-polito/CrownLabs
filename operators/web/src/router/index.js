import Vue from 'vue'
import VueRouter from 'vue-router'
import ManagementCluster from '../views/ManagementCluster.vue'
import DescribeCluster from '../views/DescribeCluster.vue'
import ResourceLogs from '../views/ResourceLogs.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'ManagementCluster',
    component: ManagementCluster
  },
  {
    path: '/cluster/',
    name: 'DescribeCluster',
    component: DescribeCluster,
    props: true
  },
  {
    path: '/logs/',
    name: 'ResourceLogs',
    component: ResourceLogs,
    props: true
  },
  {
    path: '*',
    component: ManagementCluster
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
