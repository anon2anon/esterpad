/* eslint-disable import/no-mutable-exports */
import Vue from 'vue'

var state = {
  padId: '',
  padList: [],

  userName: 'Guest',
  userId: 0,
  userColor: '#ffffff',
  get sessId () {
    return localStorage.getItem('sessId') || ''
  },
  set sessId (val) {
    localStorage.setItem('sessId', val)
  },

  isLoggedIn: false,
  perms: {
    notGuest: true,
    chat: true,
    write: true,
    edit: true,
    whitewash: true,
    mod: true,
    admin: true
  },

  pushQueue: '',
  loading: true
}
var bus = new Vue()

export { state, bus }
