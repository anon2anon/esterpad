<template>
  <div ref="cm" class="flex"></div>
</template>

<script>
import { state, bus } from '@/globs'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import CodemirrorAdapter from '@/ot/CodemirrorAdapter.js'
import TextOperation from '@/ot/TextOperation.js'
import CSSManager from '@/lib/cssmanager.js'

export default {
  name: 'esterpad-editor',
  data () {
    return {
      cma: null,
      synchronized: true,
      outgoing: null,
      buffer: null,
      revision: 0,
      incomingQueue: {},
      debounceBuffer: null,
      debounceTimer: null,
      cssManager: null
    }
  },
  mounted () {
    // isn't it a RC?
    this.reinitCM(state.padId)
    this.cssManager = new CSSManager()

    bus.$on('pad-id-changed', this.reinitCM)
    bus.$on('new-delta', this.newDelta)
    bus.$on('user-leave', this.userLeave)
    bus.$on('color-update', this.updateColor)

    this.updateColor(state.userId, state.userColor)
  },
  methods: {
    sendTextOperation (textOp) {
      console.log('sending textOp', textOp)
      this.synchronized = false
      this.outgoing = textOp
      let ops = textOp.ops.map(i => i.getProtobufData())
      bus.$emit('send', 'Delta', {
        revision: this.revision,
        ops: ops
      })
    },
    cmChangeCallback (textOp, inverse) {
      console.log('cmChangeCallback', textOp, inverse)

      let ourMeta = new TextOperation()
      let needUpdate = false
      for (let op of textOp.ops) {
        if (op.isRetain()) ourMeta = ourMeta.retain(op.len)
        if (op.isInsert()) {
          ourMeta = ourMeta.retain(op.len, {userId: state.userId})
          needUpdate = true
        }
      }
      if (needUpdate) this.cma.applyOperation(ourMeta)

      if (this.debounceBuffer !== null) {
        this.debounceBuffer = this.debounceBuffer.compose(textOp)
      } else {
        this.debounceBuffer = textOp
      }

      let that = this
      clearTimeout(this.debounceTimer)
      this.debounceTimer = setTimeout(function () {
        that.processDebounceBuffer(textOp)
      }, 300) // maybe we should tune this
      // maybe send after each whitespace or something
    },
    processDebounceBuffer () {
      let textOp = this.debounceBuffer
      console.log('processDebounceBuffer', textOp)

      if (this.synchronized) {
        this.sendTextOperation(textOp)
      } else if (this.buffer === null) {
        this.buffer = textOp
      } else {
        this.buffer = this.buffer.compose(textOp)
      }
      this.debounceBuffer = null
    },
    reinitCM (padId) {
      console.log('reinitCM', padId)
      bus.$emit('send', 'EnterPad', {name: padId})

      let cm = CodeMirror(this.$refs.cm, {
        value: '', // (TODO: make cool spinner here)
        tabSize: 4,
        mode: 'text/plain',
        lineNumbers: true,
        lineWrapping: true,
        extraKeys: {
          'Ctrl-B': function (cm) {
            alert('bold')
          }
        }
      })

      this.cma = new CodemirrorAdapter(cm)
      this.cma.registerCallbacks({'change': this.cmChangeCallback})
    },
    newDelta (delta) {
      if (delta.id !== this.revision + 1 && this.revision !== 0) {
        console.log('too new delta', delta.id, 'saving to queue')
        this.incomingQueue[delta.id] = delta
        while ((this.revision + 1) in this.incomingQueue) {
          console.log('applying delta from queue', this.revision + 1)
          this.newDelta(this.incomingQueue[this.revision + 1])
          delete this.incomingQueue[this.revision + 1]
        }
        return
      }
      this.revision = delta.id

      let convertMeta = function (meta) {
        let res = {}
        // TODO: process all meta
        if (meta.changemask & (1 << 5)) res.userId = meta.userId
        return res
      }

      let to = new TextOperation()
      // TODO: move to TextOperation.js
      for (let op of delta.ops) {
        if (op.insert !== null) {
          to = to.insert(op.insert.text, convertMeta(op.insert.meta))
        }
        if (op.retain !== null) {
          to = to.retain(op.retain.len, convertMeta(op.retain.meta))
        }
        if (op.delete !== null) {
          to = to.delete(op.delete.len)
        }
      }
      console.log('Converted delta', to)

      if (state.userId === delta.userId) {
        this.synchronized = true
        if (this.buffer !== null) {
          this.sendTextOperation(this.buffer)
          this.buffer = null
        }
      } else {
        if (this.synchronized) {
          this.cma.applyOperation(to)
        } else {
          if (this.buffer !== null) {
            let pair1 = TextOperation.transform(this.outgoing, to)
            let pair2 = TextOperation.transform(this.buffer, pair1[1])
            this.outgoing = pair1[0]
            this.buffer = pair2[0]
            this.cma.applyOperation(pair2[1])
          } else {
            let pair = TextOperation.transform(this.outgoing, to)
            this.outgoing = pair[0]
            this.cma.applyOperation(pair[1])
          }
        }
      }
    },
    updateColor (userId, newColor) {
      this.cssManager.selectorStyle('.author-' + userId).background = newColor
    },
    userLeave (info) {
      console.warn('TODO: handle user leave in editor')
    }
  }
}
</script>

<style>
 .CodeMirror, .flex {
   min-width: 100%;
   min-height: 100%;
 }
 .CodeMirror {
   font-family: Arial, sans-serif !important;
 }
</style>
