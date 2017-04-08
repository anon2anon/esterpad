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
      cma: null
    }
  },
  mounted () {
    this.reinitCM(state.padId)
    // isn't it a RC?
    bus.$on('pad-id-changed', this.reinitCM)
    bus.$on('new-delta', this.newDelta)
  },
  methods: {
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
      this.cma.registerCallbacks({'change': function (textOp, inverse) {
        console.log('change', textOp, inverse)
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
          revision: 0,
          ops: ops
        })
      }})
    },
    newDelta (delta) {
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
      console.log(to)
      this.cma.applyOperation(to)
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
