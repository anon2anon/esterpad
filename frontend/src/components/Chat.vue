<template>
  <div class="chat-container">
    <div class="messages" ref="messages">
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
      msg: ''
    }
  },
  mounted () {
    var that = this
    bus.$on('new-chat-msg', function (msg) {
      console.log('chat message', msg)
      that.appendMsg(msg)
    })
    bus.$on('color-update', this.updateColor)
  },
  methods: {
    enterPressed (e) {
      if (e.ctrlKey || e.shiftKey) {
        this.msg += '\n'
        return
      }
      if (this.msg.trim() === '') return
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
      var msgdiv = document.createElement('div')
      var msgtext = document.createTextNode(msg.text)
      msgdiv.appendChild(msgtext)
      msgdiv.className = 'chat-message chat-author-' + msg.userId
      msgdiv.style = 'background: ' + state.colorMap[msg.userId]
      this.$refs.messages.appendChild(msgdiv)
      this.$refs.messages.scrollTop = this.$refs.messages.scrollHeight
    },
    updateColor (userId, newColor) {
      var tmp = this.$refs.messages.getElementsByClassName('chat-author-' + userId)
      for (let div of tmp) {
        div.style = 'background: ' + newColor
        console.log(div)
      }
    },
    autoGrow () {
      this.$refs.msgbox.style.height = '5px'
      this.$refs.msgbox.style.height = this.$refs.msgbox.scrollHeight + 'px'
      if (this.$refs.msgbox.scrollHeight >= 100) {
        this.$refs.msgbox.scrollTop = this.$refs.msgbox.scrollHeight
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
   white-space: pre;
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
