<template>
  <div class="chat-container">
    <div class="messages" ref="messages">
      <div v-for="message in messageList" key="message">
        {{ message }}
      </div>
    </div>
    <div @keydown.prevent.enter="enterPressed">
      <md-input-container>
        <md-textarea v-model="msg" placeholder="Write a message...">
        </md-textarea>
      </md-input-container>
    </div>
  </div>
</template>

<script>
import { state, bus } from '@/globs'

export default {
  name: 'esterpad-chat',
  data () {
    return {
      revision: 0,
      messageList: [],
      msg: ''
    }
  },
  mounted () {
    var that = this
    bus.$on('new-chat-msg', function (msg) {
      console.log('chat message', msg)
      that.revision = msg.id
      that.messageList.push(msg.text)
    })
    setTimeout(function () {
      that.$refs.messages.scrollTop = that.$refs.messages.scrollHeight
    }, 100) // TODO: fix me please
  },
  methods: {
    enterPressed (e) {
      if (e.ctrlKey || e.shiftKey) {
        this.msg += '\n'
        return
      }
      bus.$emit('send', 'Chat', {
        revision: this.revision,
        text: this.msg
      })
      this.messageList.push(state.userName + ': ' + this.msg)
      var that = this
      setTimeout(function () {
        that.$refs.messages.scrollTop = that.$refs.messages.scrollHeight
      }, 100)
      this.msg = ''
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
 }

 .messages {
   flex: 1 1 auto;
   overflow-y: scroll;
 }

 .md-input-container.md-input-placeholder {
   margin-bottom: 2px;
   margin-top: -12px;
 }
</style>
