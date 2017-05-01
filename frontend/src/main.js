// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import router from './router'

import VueMaterial from 'vue-material'
import 'vue-material/dist/vue-material.css'
Vue.use(VueMaterial)

import MyUser from '@/components/MyUser'
import Editor from '@/components/Editor'
import UserList from '@/components/UserList'
import Chat from '@/components/Chat'

Vue.component('esterpad-myuser', MyUser)
Vue.component('esterpad-editor', Editor)
Vue.component('esterpad-userlist', UserList)
Vue.component('esterpad-chat', Chat)

Vue.config.productionTip = false

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  template: '<App/>',
  components: { App }
})

import * as protobuf from 'protobufjs'
import * as jsonDescr from './assets/proto.json'
var proto = protobuf.Root.fromJSON(jsonDescr)

import { state, bus } from './globs'
window['_state'] = state
window['_bus'] = bus

import Push from 'push.js'

if (!Push.Permission.has()) {
  Push.Permission.request()
}

bus.$on('push', function (header, body) {
  Push.create(header, {
    body: body,
    onClick: function () {
      window.focus()
      this.close()
      Push.clear()
    }
  })
})

var SMessages = proto.lookup('esterpad.SMessages')
var CMessages = proto.lookup('esterpad.CMessages')

var wsUrl = 'ws://' + window.location.host + '/.ws'
if (window.location.hostname === 'localhost') {
  wsUrl = 'ws://localhost:9000/.ws'
}
var conn = new WebSocket(wsUrl)
conn.binaryType = 'arraybuffer'

bus.$on('send', function () {
  var args = [] // accepts any number of messages
  for (var i = 0; 2 * i < arguments.length; i++) {
    var tmp = {}
    tmp[arguments[i]] = arguments[i + 1]
    tmp['CMessages'] = arguments[i]
    console.log('send', arguments[i], arguments[i + 1])
    args.push(tmp)
  }
  var buffer = CMessages.encode({
    cm: args
  }).finish()
  conn.send(buffer)
})

conn.onopen = function (evt) {
  console.log('WS connected')
  if (state.sessId) {
    bus.$emit('send', 'Session', {sessId: state.sessId})
  }
}

conn.onclose = function (evt) {
  console.log('WS closed')
  // TODO: reconnect
}

conn.onmessage = function (evt) {
  var messages = SMessages.decode(new Uint8Array(evt.data)).sm
  if (!messages) return // ping
  console.log('messages', messages)
  messages.forEach(function (message) {
    console.log(message)
    if (message.Auth !== null) { // Our info
      state.isLoggedIn = true
      state.userName = message.Auth.nickname
      state.userId = message.Auth.userId
      state.userColorNum = message.Auth.color
      state.colorMap[message.Auth.userId] = state.userColor
      if (message.Auth.sessId) {
        state.sessId = message.Auth.sessId
      }
      state.perms = {
        chat: Boolean(message.Auth.perms & 1),
        edit: Boolean(message.Auth.perms & (1 << 1)),
        whitewash: Boolean(message.Auth.perms & (1 << 2)),
        notGuest: Boolean(message.Auth.perms & (1 << 3)),
        admin: Boolean(message.Auth.perms & (1 << 4))
      }
      if ('go' in router.currentRoute.query) {
        router.push(router.currentRoute.query['go'])
      } else {
        router.push('/.padlist')
      }
    } else if (message.UserInfo !== null) { // User connected/updated
      let color = '#' + ('000000' + message.UserInfo.color.toString(16)).slice(-6)
      state.colorMap[message.UserInfo.userId] = color
      bus.$emit('color-update', message.UserInfo.userId, color)
      bus.$emit('user-info', message.UserInfo)
    } else if (message.UserLeave !== null) {
      bus.$emit('user-leave', message.UserLeave)
    } else if (message.Chat !== null) { // Chat message
      bus.$emit('new-chat-msg', message.Chat)
    } else if (message.Delta !== null) { // New delta
      bus.$emit('new-delta', message.Delta)
    } else if (message.AuthError) {
      bus.$emit('auth-error', 'Login error #' + message.AuthError.error)
    } else if (message.PadList !== null) {
      state.padList = state.padList.concat(message.PadList.pads)
    } else {
      console.error('Unknown message type')
    }
  })
}
