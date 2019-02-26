import Vue from 'vue'
import Router from 'vue-router'
import Pad from '@/components/Pad.vue'
import Login from '@/components/Login.vue'
import Register from '@/components/Register.vue'
import PadList from '@/components/PadList.vue'
import Options from '@/components/Options.vue'
import Admin from '@/components/Admin.vue'
import Users from '@/components/Users.vue'

import Editor from '@/components/Editor.vue'
import Timeslider from '@/components/Timeslider.vue'

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
      path: '/:padId',
      component: Pad,
      children: [
        {path: '', component: Editor},
        {path: 'timeslider', component: Timeslider}
      ],
      meta: {
        requiresLogin: true,
        updatesPadId: true
      }
    }
  ]
})
