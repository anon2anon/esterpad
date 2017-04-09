<template>
  <div ref="cm" class="flex"></div>
</template>

<script>
import { state, bus } from '@/globs'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import CodemirrorAdapter from '@/ot/CodemirrorAdapter.js'
import TextOperation from '@/ot/TextOperation.js'

export default {
  name: 'esterpad-editor',
  data () {
    return {
      cma: null,
      synchronized: true,
      outgoing: null,
      buffer: null,
      revision: 0
    }
  },
  mounted () {
    // isn't it a RC?
    this.reinitCM(state.padId)
    bus.$on('pad-id-changed', this.reinitCM)

    bus.$on('new-delta', this.newDelta)
  },
  methods: {
    sendTextOperation (textOp) {
      console.log('sending textOp', textOp)
      var ops = []
      for (var i in textOp.ops) {
        var op = textOp.ops[i]
        if (TextOperation.isInsert(op)) {
          ops.push({
            insert: {text: op},
            op: 'insert'
          })
        }
        if (TextOperation.isRetain(op)) {
          ops.push({
            retain: {len: op},
            op: 'retain'
          })
        }
        if (TextOperation.isDelete(op)) {
          ops.push({
            delete: {len: -op},
            op: 'delete'
          })
        }
      }
      bus.$emit('send', 'Delta', {
        revision: this.revision,
        ops: ops
      })
    },
    cmChangeCallback (textOp, inverse) {
      console.log('cmChangeCallback', textOp, inverse)
      if (this.synchronized) {
        this.synchronized = false
        this.sendTextOperation(textOp)
      } else if (this.buffer === null) {
        this.buffer = textOp
      } else {
        this.buffer = this.buffer.compose(textOp)
      }
    },
    reinitCM (padId) {
      console.log('reinitCM', padId)
      bus.$emit('send', 'EnterPad', {name: padId})

      var cm = CodeMirror(this.$refs.cm, {
        value: '', // (TODO: make cool spinner here)
        tabSize: 4,
        mode: 'text/plain',
        lineNumbers: true
      })

      this.cma = new CodemirrorAdapter(cm)
      this.cma.registerCallbacks({'change': this.cmChangeCallback})
    },
    newDelta (delta) {
      this.revision = delta.id

      var to = new TextOperation()
      for (var i in delta.ops) {
        var op = delta.ops[i]
        if (op.insert !== null) {
          to = to.insert(op.insert.text)
        }
        if (op.retain !== null) {
          to = to.retain(op.retain.len)
        }
        if (op.delete !== null) {
          to = to.delete(op.delete.len)
        }
      }
      console.log('Converted delta', to)

      if (state.userId === delta.userId) {
        this.synchronized = true
        if (this.buffer !== null) {
          this.synchronized = false
          this.sendTextOperation(this.buffer)
          this.buffer = null
        }
        return
      }

      if (this.synchronized) {
        this.cma.applyOperation(to)
      } else {
        if (this.buffer !== null) {
          var pair1 = TextOperation.transform(this.buffer, to)
          var pair2 = TextOperation.transform(this.outgoing, pair1[1])
          this.buffer = pair1[0]
          this.cma.applyOperation(pair2[1])
        } else {
          var pair = TextOperation.transform(this.outgoing, to)
          this.cma.applyOperation(pair[1])
        }
      }
    }
  }
}
</script>

<style>
 .CodeMirror, .flex {
   min-width: 100%;
   min-height: 100%;
 }
</style>
