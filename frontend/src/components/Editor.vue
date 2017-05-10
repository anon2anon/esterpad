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
    bus.$on('document', this.recvDocument)
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
      if (needUpdate) this.cma.applyOperation(ourMeta, false)

      this.processDelta(textOp)
    },
    processDelta (delta) {
      if (this.debounceBuffer !== null) {
        this.debounceBuffer = this.debounceBuffer.compose(delta)
      } else {
        this.debounceBuffer = delta
      }

      let that = this
      clearTimeout(this.debounceTimer)
      this.debounceTimer = setTimeout(function () {
        that.processDebounceBuffer(delta)
      }, 150) // maybe we should tune this
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

      let that = this

      let toggleMeta = function (cm, meta) {
        let from = cm.indexFromPos(cm.getCursor('from'))
        let to = cm.indexFromPos(cm.getCursor('to'))
        console.log('toggleMeta', from, to, meta)

        // TODO: check permissions
      }

      let cm = CodeMirror(this.$refs.cm, {
        value: '', // (TODO: make cool spinner here)
        tabSize: 4,
        mode: 'text/plain',
        lineNumbers: true,
        lineWrapping: true,
        extraKeys: {
          'Ctrl-B': function (cm) {
            toggleMeta(cm, 'bold')
          },
          'Ctrl-I': function (cm) {
            toggleMeta(cm, 'italic')
          },
          'Ctrl-U': function (cm) {
            toggleMeta(cm, 'underline')
          },
          'Ctrl-S': function (cm) {
            toggleMeta(cm, 'strike')
          },
          'Ctrl-M': function (cm) {
            if (!state.perms.mod) {
              bus.$emit('auth-error', 'Sorry, you don\'t have permission for that')
              return
            }

            let from = cm.indexFromPos(cm.getCursor('from'))
            let to = cm.indexFromPos(cm.getCursor('to'))
            if (from === to) return
            let docLen = cm.indexFromPos({ line: cm.lastLine(), ch: 0 }) +
                         cm.getLine(cm.lastLine()).length

            let whiteMeta = new TextOperation()
            whiteMeta = whiteMeta.retain(from)
            whiteMeta = whiteMeta.retain(to - from, {userId: 0})
            whiteMeta = whiteMeta.retain(docLen - to)

            that.cma.applyOperation(whiteMeta, false)
            that.processDelta(whiteMeta)
          }
        }
      })

      this.cma = new CodemirrorAdapter(cm)
      this.cma.registerCallbacks({'change': this.cmChangeCallback})
    },
    recvDocument (doc) {
      console.log('recv doc', doc)
      this.revision = doc.revision

      let to = (new TextOperation()).fromProtobuf(doc)
      console.log('Converted doc', to)

      this.cma.applyOperation(to)
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

      let to = (new TextOperation()).fromProtobuf(delta)
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
