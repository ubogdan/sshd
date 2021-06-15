import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

const routes = [
    {
        path: '/swagger',
        name: 'Swagger',
        component: () => import('../views/Swagger.vue'),
        props: true
    },
    {
        path: '/sftp/:token',
        name: 'WebSftp',
        component: () => import('../views/Sftp.vue'),
        props: true

    },
    {
        path: '/guacamole/:token',
        name: 'Guacamole',
        component: () => import('../views/Guacamole.vue'),
        props: true
    },
    {
        path: '/ssh/:token',
        name: 'ssh',
        component: () => import('../views/Ssh.vue'),
        props: true
    },
]

const router = new VueRouter({
    mode: 'hash',
    base: process.env.BASE_URL,
    routes
})

export default router
