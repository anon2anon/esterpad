var state = {
  padId: '',
  padList: [],

  userName: 'Guest',
  get userColor () {
    return '#' + ('000000' + this.userColorNum.toString(16)).slice(-6)
  },
  userColorNum: 16761095,
  get sessId () {
    return localStorage.getItem('sessId') || ''
  },
  set sessId (val) {
    localStorage.setItem('sessId', val)
  },

  isLoggedIn: false,
  perms: {
    view: true,
    chat: true,
    edit: true,
    whitewash: true,
    notGuest: false,
    admin: true
  }
}

import Vue from 'vue'
var bus = new Vue()

export { state, bus }
