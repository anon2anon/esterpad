<template>
  <div ref="cm" class="flex"></div>
</template>

<script>
import state from '@/state'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import CodemirrorAdapter from '@/ot/CodemirrorAdapter.js'
import TextOperation from '@/ot/TextOperation.js'

export default {
  name: 'esterpad-editor',
  data () {
    return {
      state: state,
      cma: null
    }
  },
  mounted () {
    this.reinitCM(state.padId)
  },
  watch: {
    'state.padId' (to, from) {
      this.reinitCM(to)
    }
  },
  methods: {
    reinitCM (padId) {
      console.log('reinitCM', padId)
      state.sendMessage('EnterPad', {name: padId})

      var cm = CodeMirror(this.$refs.cm, {
        value: '', // (TODO: make cool spinner here)
        tabSize: 4,
        mode: 'text/plain',
        lineNumbers: true
      }) // this.cm?
      this.cma = new CodemirrorAdapter(cm)
      this.cma.registerCallbacks({'change': function (textOp, inverse) {
        console.log('change', textOp, inverse)
        var ops = []
        for (var i in textOp.ops) {
          var op = textOp.ops[i]
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
          if (TextOperation.isInsert(op)) {
            ops.push({
              insert: {text: op},
              op: 'insert'
            })
          }
        }
        state.sendMessage('Delta', {
          revision: 0,
          ops: ops
        })
      }})
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
