import Vue from 'vue'
import Router from 'vue-router'
import Pad from '@/components/Pad'
import Login from '@/components/Login'
import Register from '@/components/Register'
import PadList from '@/components/PadList'
import Options from '@/components/Options'
import Admin from '@/components/Admin'
import Users from '@/components/Users'
import Timeslider from '@/components/Timeslider'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/.login',
      name: 'Login',
      component: Login
    },
    {
      path: '/.register',
      name: 'Register',
      component: Register
    },
    {
      path: '/.options',
      name: 'Options',
      component: Options,
      meta: {
        requiresLogin: true
      }
    },
    {
      path: '/.users',
      name: 'Users',
      component: Users,
      meta: {
        requiresLogin: true,
        requiresMod: true
      }
    },
    {
      path: '/.admin',
      name: 'Admin',
      component: Admin,
      meta: {
        requiresLogin: true,
        requiresAdmin: true
      }
    },
    {
      path: '/.padlist',
      name: 'Pad List',
      component: PadList,
      meta: {
        requiresLogin: true
      }
    },
    {
      path: '/',
      redirect: '/.padlist'
    },
    {
      path: '/.timeslider/:padId',
      name: 'Timeslider',
      component: Timeslider,
      meta: {
        requiresLogin: true,
        updatesPadId: true
      }
    },
    {
      path: '/:padId',
      name: 'Pad',
      component: Pad,
      meta: {
        requiresLogin: true,
        updatesPadId: true
      }
    }
  ]
})
