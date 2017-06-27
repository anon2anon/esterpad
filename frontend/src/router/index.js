import Vue from 'vue'
import Router from 'vue-router'
import Pad from '@/components/Pad'
import Login from '@/components/Login'
import Register from '@/components/Register'
import PadList from '@/components/PadList'

import state from '@/state'

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
        next()
      }
    }
  ]
})
