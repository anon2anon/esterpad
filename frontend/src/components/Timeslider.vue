<template>
  <div class="editor flex">
    <vue-slider ref="slider" v-model="revision"
                :min="0" :max="maxRevision"
                tooltip="always" tooltip-dir="bottom"
                style="z-index: 90" @callback="revChange">
    </vue-slider>
    <div class="toolbar">
      <div class="toolbar-buttons">
        <div class="button-group">
          <div class="button"><span>B</span></div>
          <div class="button"><span>U</span></div>
          <div class="button"><span>T</span></div>
          <div class="button"><span>T</span></div>
          <div class="button"><span>O</span></div>
          <div class="button"><span>N</span></div>
          <div class="button"><span>S</span></div>
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
import CodemirrorAdapter from '@/ot/CodemirrorAdapter.js'
import TextOperation from '@/ot/TextOperation.js'
import CSSManager from '@/lib/cssmanager.js'
import { textColor } from '@/helpers'
import vueSlider from 'vue-slider-component'

export default {
  data () {
    return {
      cma: null,
      revision: 0,
      maxRevision: 0,
      cssManager: null,
      state: state
    }
  },
  components: {
    vueSlider
  },
  mounted () {
    log.debug('timeslider mounted')

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
  beforeDestroy () {
    log.debug('timeslider destroy')

    bus.$off('pad-id-changed', this.reinitCM)
    bus.$off('document', this.recvDocument)
    bus.$off('new-delta', this.newDelta)
    bus.$off('user-leave', this.userLeave)
    bus.$off('color-update', this.updateColor)
  },
  methods: {
    reinitCM (padId) {
      log.debug('reinitCM', padId)
      bus.$emit('send', 'EnterPad', {name: padId})

      if (this.cma) {
        log.debug('Clearing editor')
        this.cma.clear()
      } else {
        log.debug('Creating editor')

        let cm = CodeMirror(this.$refs.cm, {
          value: '', // (TODO: make cool spinner here)
          tabSize: 4,
          mode: 'text/plain',
          lineNumbers: true,
          lineWrapping: true,
          readOnly: 'nocursor'
        })

        this.cma = new CodemirrorAdapter(cm)
      }
    },
    revChange (val) {
      log.debug(val)
      bus.$emit('send', 'RevisionRequest', {revision: val})
    },
    recvDocument (doc) {
      log.debug('recv doc', doc)

      if (this.revision === 0) this.revision = doc.revision
      this.maxRevision = Math.max(this.maxRevision, doc.revision)

      if (this.cma) {
        log.debug('Clearing editor')
        this.cma.clear()
      }

      let to = (new TextOperation()).fromProtobuf(doc)
      log.debug('Converted doc', to)

      this.cma.applyOperation(to)
    },
    newDelta (delta) {
      this.maxRevision = Math.max(this.maxRevision, delta.id)
    },
    updateColor (userId, newColor) {
      this.cssManager.selectorStyle('.author-' + userId).background = newColor
      let fgColor = textColor(newColor)
      this.cssManager.selectorStyle('.author-' + userId).color = fgColor
    },
    userLeave (info) {
      log.warn('TODO: handle user leave in editor')
    }
  }
}
</script>

<style>

  .editor{
    display: grid;
    grid-template-rows: 50px 45px 1fr;
    height: 100%;
  }

 .CodeMirror, .flex {
   min-width: 100%;
   min-height: 100%;
 }
 .CodeMirror {
   font-family: Arial, sans-serif !important;
 }

 /* TODO: move to another CSS file */
 .padtext-bold {
   font-weight: bold;
 }
 .padtext-italic {
   font-style: italic;
 }
 .padtext-underline {
   text-decoration: underline;
 }
 .padtext-strike {
   text-decoration: line-through;
 }
 .padtext-underline.padtext-strike {
   text-decoration: underline line-through;
 }
</style>
