import * as log from 'loglevel'
window.log = log
log.setLevel('debug')

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
let proto = protobuf.Root.fromJSON(jsonDescr)

import { num2color } from './helpers'
import { state, bus } from './globs'
window['_state'] = state
window['_bus'] = bus

import Push from 'push.js'

if (!Push.Permission.has()) {
  Push.Permission.request()
}

bus.$on('push', function (header, body) {
  if (state.pushQueue) {
    Push.clear()
  }
  state.pushQueue = body + '\n' + state.pushQueue
  Push.create(header, {
    body: state.pushQueue,
    onClick: function () {
      window.focus()
      this.close()
      Push.clear()
      state.pushQueue = ''
    }
  })
})

let SMessages = proto.lookup('esterpad.SMessages')
let CMessages = proto.lookup('esterpad.CMessages')

let wsUrl = window.location.protocol === 'https:' ? 'wss' : 'ws'
wsUrl += '://'
if (window.location.hostname === 'localhost') {
  wsUrl += 'localhost:9000'
} else {
  wsUrl += window.location.host
}
wsUrl += '/.ws'
let conn = new WebSocket(wsUrl)
conn.binaryType = 'arraybuffer'

bus.$on('send', function () {
  let args = [] // accepts any number of messages
  for (let i = 0; 2 * i < arguments.length; i++) {
    let tmp = {}
    tmp[arguments[i]] = arguments[i + 1]
    tmp['CMessages'] = arguments[i]
    log.debug('send', arguments[i], arguments[i + 1])
    args.push(tmp)
  }
  let buffer = CMessages.encode({
    cm: args
  }).finish()
  conn.send(buffer)
})

conn.onopen = function (evt) {
  log.debug('WS connected')
  if (state.sessId) {
    bus.$emit('send', 'Session', {sessId: state.sessId})
  }
}

conn.onclose = function (evt) {
  log.debug('WS closed')
  // TODO: reconnect
}

conn.onmessage = function (evt) {
  let messages = SMessages.decode(new Uint8Array(evt.data)).sm
  if (!messages) return // ping
  log.debug('messages', messages)
  messages.forEach(function (message) {
    log.debug(message)
    if (message.Auth !== null) { // Our info
      state.isLoggedIn = true
      state.userName = message.Auth.nickname
      state.userId = message.Auth.userId
      state.userColor = num2color(message.Auth.color)
      if (message.Auth.sessId) {
        state.sessId = message.Auth.sessId
      }
      state.perms = {
        notGuest: Boolean(message.Auth.perms & 1),
        chat: Boolean(message.Auth.perms & (1 << 1)),
        write: Boolean(message.Auth.perms & (1 << 2)),
        edit: Boolean(message.Auth.perms & (1 << 3)),
        whitewash: Boolean(message.Auth.perms & (1 << 4)),
        mod: Boolean(message.Auth.perms & (1 << 5)),
        admin: Boolean(message.Auth.perms & (1 << 6))
      }
      state.padList = []
      if ('go' in router.currentRoute.query) {
        router.push(router.currentRoute.query['go'])
      } else {
        router.push('/.padlist')
      }
    } else if (message.UserInfo !== null) { // User connected/updated
      let color = num2color(message.UserInfo.color)
      bus.$emit('color-update', message.UserInfo.userId, color)
      bus.$emit('user-info', message.UserInfo)
    } else if (message.UserLeave !== null) {
      bus.$emit('user-leave', message.UserLeave)
    } else if (message.Chat !== null) { // Chat message
      bus.$emit('new-chat-msg', message.Chat)
    } else if (message.Delta !== null) { // New delta
      bus.$emit('new-delta', message.Delta)
    } else if (message.Document !== null) { // Document revision
      bus.$emit('document', message.Document)
    } else if (message.AuthError) {
      bus.$emit('auth-error', message.AuthError.error)
    } else if (message.PadList !== null) {
      state.padList = state.padList.concat(message.PadList.pads)
    } else {
      log.error('Unknown message type', message)
    }
  })
}
