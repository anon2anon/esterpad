var state = {
  padId: '',
  padList: [],

  userName: 'Guest',
  userId: 0,
  userColor: '#ffc107',
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

import Vue from 'vue'
var bus = new Vue()

export { state, bus }
