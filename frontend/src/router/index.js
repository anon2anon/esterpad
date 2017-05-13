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
        if (!state.isLoggedIn) {
          next('/.login?go=' + to.path)
          return
        }
        next()
      }
    },
    {
      path: '/.users',
      name: 'Users',
      component: Users,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) {
          next('/.login?go=' + to.path)
          return
        }
        if (!state.perms.mod) {
          bus.$emit('snack-msg', 'Hey, you\'re not mod!')
          next('/')
          return
        }
        next()
      }
    },
    {
      path: '/.admin',
      name: 'Admin',
      component: Admin,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) {
          next('/.login?go=' + to.path)
          return
        }
        if (!state.perms.admin) {
          bus.$emit('snack-msg', 'Hey, you\'re not admin!')
          next('/')
          return
        }
        next()
      }
    },
    {
      path: '/.padlist',
      name: 'Pad List',
      component: PadList,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) {
          next('/.login?go=' + to.path)
          return
        }
        next()
      }
    },
    {
      path: '/',
      redirect: '/.padlist'
    },
    {
      path: '/.timeslider',
      name: 'Timeslider',
      component: Timeslider,
      props: true,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) {
          next('/')
          return
        }
        next()
      }
    },
    {
      path: '/:padId',
      name: 'Pad',
      component: Pad,
      props: true,
      beforeEnter: (to, from, next) => {
        if (!state.isLoggedIn) {
          next('/.login?go=' + to.path)
          return
        }
        let pid = to.params.padId
        if (pid.indexOf('.') !== -1 || pid.indexOf('/') !== -1) {
          bus.$emit('snack-msg', 'Error 404, redirecting you to main page')
          next('/')
          return
        }
        state.padId = pid
        bus.$emit('pad-id-changed', pid)
        next()
      }
    }
  ]
})
