<template>
  <div class="chat-container">
    <div class="messages" ref="messages" @scroll="onScroll">
    </div>
    <div @keydown.prevent.enter="enterPressed">
      <textarea v-model="msg" id="chat-input" ref="msgbox"
                placeholder="Write a message..."
                @keyup="autoGrow(this)">
      </textarea>
    </div>
  </div>
</template>

<script>
import { state, bus } from '@/globs'

export default {
  name: 'esterpad-chat',
  data () {
    return {
      state: state,
      messageList: [],
      msg: '',
      lastId: -1,
      waitingForHistory: false
    }
  },
  mounted () {
    bus.$on('new-chat-msg', this.appendMsg)
  },
  methods: {
    enterPressed (e) {
      if (e.ctrlKey || e.shiftKey) {
        this.msg += '\n'
        return
      }
      if (this.msg.trim() === '') return

      // TODO: move to MyUser
      if (this.msg.startsWith('/color')) {
        let colorName = this.msg.substr(6).trim()
        // parse CSS colors
        state.userColorNum = parseInt(colorName, 16)
        state.colorMap[state.userId] = state.userColor
        bus.$emit('color-update', state.userId, state.userColor)
        bus.$emit('send', 'EditUser', {
          changemask: 2,
          color: state.userColorNum
        })
        this.msg = ''
        return
      }

      bus.$emit('send', 'Chat', {
        text: this.msg
      })

      this.appendMsg({
        text: state.userName + ': ' + this.msg,
        userId: state.userId
      })
      this.msg = ''
    },
    appendMsg (msg) {
      console.log('chat message', msg)

      if (document.hidden) { // msg.needPush &&
        bus.$emit('push', state.padId, msg.text)
      }

      let append = true
      if (this.lastId === -1) {
        this.lastId = msg.id
      } else {
        if (msg.id < this.lastId) append = false
        this.lastId = Math.min(this.lastId, msg.id)
      }
      // min because lastId is id of first message in history

      let msgdiv = document.createElement('div')
      let msgtext = document.createTextNode(msg.text)
      msgdiv.appendChild(msgtext)
      msgdiv.className = 'author-' + msg.userId
      if (append) {
        let needScroll = this.$refs.messages.scrollTop + this.$refs.messages.offsetHeight >=
          this.$refs.messages.scrollHeight - 1
        this.$refs.messages.appendChild(msgdiv)
        if (needScroll) {
          this.$refs.messages.scrollTop = this.$refs.messages.scrollHeight
        }
      } else {
        this.$refs.messages.insertBefore(msgdiv,
                                         this.$refs.messages.childNodes[0])

        if (this.$refs.messages.scrollTop === 0) {
          this.$refs.messages.scrollTop = msgdiv.clientHeight / 2
        }

        this.waitingForHistory = false
      }
    },
    autoGrow () {
      this.$refs.msgbox.style.height = '5px'
      this.$refs.msgbox.style.height = this.$refs.msgbox.scrollHeight + 'px'
      if (this.$refs.msgbox.scrollHeight >= 100) {
        this.$refs.msgbox.scrollTop = this.$refs.msgbox.scrollHeight
      }
    },
    onScroll () {
      if (!this.waitingForHistory && this.lastId !== 1 &&
          this.$refs.messages.scrollTop < 5) {
        this.waitingForHistory = true
        bus.$emit('send', 'ChatRequest', {
          from: this.lastId - 1,
          count: 50
        })
      }
    }
  }
}
</script>

<style scoped>
 .chat-container {
   min-height: 100%;
   display: flex;
   flex-flow: column nowrap;
   flex: 1;
   white-space: pre-wrap;
 }

 .messages {
   flex: 1 1 0;
   overflow-y: scroll;
   word-break: break-all;
 }

 #chat-input {
   width: 100%;
   resize: none;
   background-color: transparent;
   border-style: solid;
   border-width: 0px 0px 1px 0px;
   border-color: darkred;
   outline: 0;
   overflow: hidden;

   height: 25px;
   min-height: 25px;
   max-height: 100px;
   margin-bottom: -6px;
 }
</style>
