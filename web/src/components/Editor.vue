<template>
  <div class="editor flex">
    <div class="toolbar">
      <div class="toolbar-buttons">
        <div class="button-group">
          <div class="button" @click="toggleMeta('bold')">
            <i class="material-icons">format_bold</i>
          </div>
          <div class="button" @click="toggleMeta('italic')">
            <i class="material-icons">format_italic</i>
          </div>
          <div class="button" @click="toggleMeta('underline')">
            <i class="material-icons">format_underlined</i>
          </div>
          <div class="button" @click="toggleMeta('strike')">
            <i class="material-icons">strikethrough_s</i>
          </div>
        </div>
        <div class="button-group">
          <router-link to="timeslider" class="link-black" append>
            <div class="button"><i class="material-icons">history</i></div>
          </router-link>
        </div>
      </div>
    </div>
    <div ref="cm"></div>
  </div>
</template>

<script>
import { state, bus } from '@/globs'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import CodemirrorAdapter from '@/ot/CodemirrorAdapter'
import TextOperation from '@/ot/TextOperation'
import CSSManager from '@/lib/cssmanager'
import { textColor } from '@/helpers'

export default {
  data () {
    return {
      cma: null,
      cm: null,
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
    log.debug('editor mounted')
    // isn't it a RC?
    this.reinitCM(state.padId)
    this.cssManager = new CSSManager()

    bus.$on('pad-id-changed', this.reinitCM)
    bus.$on('document', this.recvDocument)
    bus.$on('new-delta', this.newDelta)
    bus.$on('user-leave', this.userLeave)
    bus.$on('color-update', this.updateColor)

    if (state.userId !== 0) {
      this.updateColor(state.userId, state.userColor)
    }
  },
  beforeDestroy () {
    log.debug('editor destroy')

    bus.$off('pad-id-changed', this.reinitCM)
    bus.$off('document', this.recvDocument)
    bus.$off('new-delta', this.newDelta)
    bus.$off('user-leave', this.userLeave)
    bus.$off('color-update', this.updateColor)
  },
  methods: {
    sendTextOperation (textOp) {
      log.debug('sending textOp', textOp)
      this.synchronized = false
      this.outgoing = textOp
      let ops = textOp.ops.map(i => i.getProtobufData())
      bus.$emit('send', 'Delta', {
        revision: this.revision,
        ops: ops
      })
    },
    cmChangeCallback (textOp, inverse) {
      log.debug('cmChangeCallback', textOp, inverse)

      // TODO: check possibilty of delete and apply inverse

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
      log.debug('processDebounceBuffer', textOp)

      if (this.synchronized) {
        this.sendTextOperation(textOp)
      } else if (this.buffer === null) {
        this.buffer = textOp
      } else {
        this.buffer = this.buffer.compose(textOp)
      }
      this.debounceBuffer = null
    },
    toggleMeta (meta) {
      let from = this.cm.getCursor('from')
      let to = this.cm.getCursor('to')
      log.debug('toggleMeta', from, to, meta)

      let tmp = this.cma.toggleMeta(from, to, meta, state.perms.edit, state.userId)
      if (!tmp.ok) {
        bus.$emit('snack-msg', 'Sorry, you don\'t have permission for that, some edits dropped')
      }

      if (!tmp.delta.isNoop()) {
        this.cma.applyOperation(tmp.delta)
        this.processDelta(tmp.delta)
      }
    },
    reinitCM (padId) {
      log.debug('reinitCM', padId)
      bus.$emit('send', 'EnterPad', {name: padId})

      if (this.cma) {
        log.debug('Clearing editor')
        this.cma.clear()
      } else {
        log.debug('Creating editor')

        let that = this
        this.cm = CodeMirror(this.$refs.cm, {
          value: '', // (TODO: make cool spinner here)
          tabSize: 4,
          mode: 'text/plain',
          lineNumbers: true,
          lineWrapping: true,
          showCursorWhenSelecting: true,
          extraKeys: {
            'Ctrl-B': function () {
              that.toggleMeta('bold')
            },
            'Ctrl-I': function () {
              that.toggleMeta('italic')
            },
            'Ctrl-U': function () {
              that.toggleMeta('underline')
            },
            'Ctrl-S': function () {
              that.toggleMeta('strike')
            },
            'Ctrl-M': function (cm) {
              if (!state.perms.whitewash) {
                bus.$emit('snack-msg', 'Sorry, you don\'t have permission for that')
                return
              }

              let from = cm.indexFromPos(cm.getCursor('from'))
              let to = cm.indexFromPos(cm.getCursor('to'))
              if (from === to) return
              let docLen = cm.indexFromPos({ line: cm.lastLine(), ch: 0 }) +
                           cm.getLine(cm.lastLine()).length

              let whiteMeta = new TextOperation().retain(from)
                                                 .retain(to - from, {userId: 0})
                                                 .retain(docLen - to)

              that.cma.applyOperation(whiteMeta)
              that.processDelta(whiteMeta)
            }
          }
        })

        this.cma = new CodemirrorAdapter(this.cm)
        this.cma.registerCallbacks({'change': this.cmChangeCallback})
      }
    },
    recvDocument (doc) {
      log.debug('recv doc', doc)
      this.revision = doc.revision

      let to = (new TextOperation()).fromProtobuf(doc)
      log.debug('Converted doc', to)

      this.cma.clear()
      this.cma.applyOperation(to)
    },
    newDelta (delta) {
      if (delta.id !== this.revision + 1 && this.revision !== 0) {
        log.debug('too new delta', delta.id, 'saving to queue')
        this.incomingQueue[delta.id] = delta
        while ((this.revision + 1) in this.incomingQueue) {
          log.debug('applying delta from queue', this.revision + 1)
          this.newDelta(this.incomingQueue[this.revision + 1])
          delete this.incomingQueue[this.revision + 1]
        }
        return
      }
      this.revision = delta.id

      let to = (new TextOperation()).fromProtobuf(delta)
      log.debug('Converted delta', to)

      if (state.userId === delta.userId) {
        if (this.synchronized) {
          // TODO: fixme
          // delta undo
          this.cma.applyOperation(to)
        } else {
          this.synchronized = true
          if (this.buffer !== null) {
            this.sendTextOperation(this.buffer)
            this.buffer = null
          }
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
      let fgColor = textColor(newColor)
      this.cssManager.selectorStyle('.author-' + userId).color = fgColor
    },
    userLeave (info) {
      log.warn('TODO: handle user leave in editor', info)
    }
  }
}
</script>

<style scoped>
  .editor {
    display: grid;
    grid-template-rows: 45px 1fr;
    height: 100%;
  }
</style>
