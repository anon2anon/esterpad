import Vue from 'vue'
import Router from 'vue-router'
import Pad from '@/components/Pad'
import Login from '@/components/Login'
import Register from '@/components/Register'
import PadList from '@/components/PadList'
import Options from '@/components/Options'
import Admin from '@/components/Admin'

import { state, bus } from '@/globs'

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
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) next('/.login?go=' + to.path)
        next()
      }
    },
    {
      path: '/.admin',
      name: 'Admin',
      component: Admin,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) next('/.login?go=' + to.path)
        if (!state.perms.admin) {
          bus.$emit('auth-error', 'Hey, you\'re not admin!')
          next('/')
        }
        next()
      }
    },
    {
      path: '/.padlist',
      name: 'Pad List',
      component: PadList,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) next('/.login?go=' + to.path)
        next()
      }
    },
    {
      path: '/',
      redirect: '/.padlist'
    },
    {
      path: '/:padId',
      name: 'Pad',
      component: Pad,
      props: true,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) next('/.login?go=' + to.path)
        state.padId = to.params.padId
        bus.$emit('pad-id-changed', to.params.padId)
        next()
      }
    }
  ]
})
