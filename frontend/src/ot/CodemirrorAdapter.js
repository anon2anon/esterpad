/* eslint no-unexpected-multiline: 0 */

import TextOperation from './TextOperation.js'
import Selection from './Selection.js'

function CodeMirrorAdapter (cm) {
  this.cm = cm
  this.changeInProgress = false
  this.selectionChanged = false

  bind(this, 'onChanges')
  bind(this, 'onChange')
  bind(this, 'onCursorActivity')
  bind(this, 'onFocus')
  bind(this, 'onBlur')

  cm.on('changes', this.onChanges)
  cm.on('change', this.onChange)
  cm.on('cursorActivity', this.onCursorActivity)
  cm.on('focus', this.onFocus)
  cm.on('blur', this.onBlur)
}

// Removes all event listeners from the CodeMirror instance.
CodeMirrorAdapter.prototype.detach = function () {
  this.cm.off('changes', this.onChanges)
  this.cm.off('change', this.onChange)
  this.cm.off('cursorActivity', this.onCursorActivity)
  this.cm.off('focus', this.onFocus)
  this.cm.off('blur', this.onBlur)
}

function cmpPos (a, b) {
  if (a.line < b.line) return -1
  if (a.line > b.line) return 1
  if (a.ch < b.ch) return -1
  if (a.ch > b.ch) return 1
  return 0
}

// var posEq = (a, b) => cmpPos(a, b) === 0
var posLe = (a, b) => cmpPos(a, b) <= 0

var minPos = (a, b) => posLe(a, b) ? a : b
var maxPos = (a, b) => posLe(a, b) ? b : a

function codemirrorDocLength (doc) {
  return doc.indexFromPos({ line: doc.lastLine(), ch: 0 }) +
    doc.getLine(doc.lastLine()).length
}

// Converts a CodeMirror change array (as obtained from the 'changes' event
// in CodeMirror v4) or single change or linked list of changes (as returned
// by the 'change' event in CodeMirror prior to version 4) into a
// TextOperation and its inverse and returns them as a two-element array.
CodeMirrorAdapter.operationFromCodeMirrorChanges = function (changesDirty, doc) {
  // Approach: Replay the changes, beginning with the most recent one, and
  // construct the operation and its inverse. We have to convert the position
  // in the pre-change coordinate system to an index. We have a method to
  // convert a position in the coordinate system after all changes to an index,
  // namely CodeMirror's `indexFromPos` method. We can use the information of
  // a single change object to convert a post-change coordinate system to a
  // pre-change coordinate system. We can now proceed inductively to get a
  // pre-change coordinate system for all changes in the linked list.
  // A disadvantage of this approach is its complexity `O(n^2)` in the length
  // of the linked list of changes.

  let changes = []
  // changesDirty contains not-user generated changes, we should remove them.
  // Looks like change.origin undefined in this case.
  for (let c of changesDirty) {
    log.debug('change with origin', c.origin)
    if (typeof c.origin !== 'undefined') {
      changes.push(c)
    }
  }

  var docEndLength = codemirrorDocLength(doc)
  var operation = new TextOperation().retain(docEndLength)
  var inverse = new TextOperation().retain(docEndLength)

  var indexFromPos = function (pos) {
    return doc.indexFromPos(pos)
  }

  var last = arr => arr[arr.length - 1]

  function sumLengths (strArr) {
    if (strArr.length === 0) return 0
    var sum = 0
    for (var i = 0; i < strArr.length; i++) sum += strArr[i].length
    return sum + strArr.length - 1
  }

  function updateIndexFromPos (indexFromPos, change) {
    return function (pos) {
      if (posLe(pos, change.from)) return indexFromPos(pos)
      if (posLe(change.to, pos)) {
        return indexFromPos({
          line: pos.line + change.text.length - 1 - (change.to.line - change.from.line),
          ch: (change.to.line < pos.line)
            ? pos.ch
            : (change.text.length <= 1)
              ? pos.ch - (change.to.ch - change.from.ch) + sumLengths(change.text)
              : pos.ch - change.to.ch + last(change.text).length
        }) + sumLengths(change.removed) - sumLengths(change.text)
      }
      if (change.from.line === pos.line) {
        return indexFromPos(change.from) + pos.ch - change.from.ch
      }
      return indexFromPos(change.from) +
        sumLengths(change.removed.slice(0, pos.line - change.from.line)) +
        1 + pos.ch
    }
  }

  for (var i = changes.length - 1; i >= 0; i--) {
    var change = changes[i]
    indexFromPos = updateIndexFromPos(indexFromPos, change)

    var fromIndex = indexFromPos(change.from)
    var restLength = docEndLength - fromIndex - sumLengths(change.text)

    operation = new TextOperation()
      .retain(fromIndex)
     ['delete'](sumLengths(change.removed))
      .insert(change.text.join('\n'))
      .retain(restLength)
      .compose(operation)

    inverse = inverse.compose(new TextOperation()
                              .retain(fromIndex)
                             ['delete'](sumLengths(change.text))
                              .insert(change.removed.join('\n'))
                              .retain(restLength)
                             )

    docEndLength += sumLengths(change.removed) - sumLengths(change.text)
  }

  return [operation, inverse]
}

let metaToCMClasses = function (meta) {
  let classes = ''
  // TODO: grab actual list from proto
  for (let key of ['bold', 'italic', 'underline', 'strike']) {
    if (meta.hasOwnProperty(key)) {
      classes += ' padtext-' + (meta[key] ? '' : '-') + key
    } else {
      classes += ' padtext--' + key
    }
  }
  if (meta.hasOwnProperty('fontSize')) classes += ' padtext-fontSize' + meta.fontSize
  else classes += ' padtext-fontSize0'
  if (meta.hasOwnProperty('userId')) classes += ' author-' + meta.userId
  else classes += ' author-0'
  return classes
}

let CMClassesToMeta = function (classes) {
  let meta = {}
  // TODO: grab actual list from proto
  for (let klass of classes.split(' ').slice(1)) {
    if (klass.startsWith('padtext-')) {
      let tmp = klass.slice('padtext-'.length)

      if (tmp.startsWith('fontSize')) {
        meta.fontSize = +(tmp.slice('fontSize'.length))
        continue
      }

      if (tmp.startsWith('-')) {
        meta[tmp.slice(1)] = false
      } else {
        meta[tmp] = true
      }
    } else if (klass.startsWith('author-')) {
      meta.userId = +(klass.slice('author-'.length))
    }
  }
  return meta
}

// Apply an operation to a CodeMirror instance.
CodeMirrorAdapter.applyOperationToCodeMirror = function (operation, cm) {
  cm.operation(function () {
    var index = 0 // holds the current index into CodeMirror's content
    for (let op of operation.ops) {
      let from = cm.posFromIndex(index)

      if (op.isInsert() || op.isRetain()) {
        if (op.isInsert()) cm.replaceRange(op.data, from)

        if (Object.keys(op.meta).length > 0) { // have meta, merge marks classes
          let to = cm.posFromIndex(index + op.len)
          let classes = metaToCMClasses(op.meta)

          if (cm.findMarks(from, to).length === 0) { // this happens sometimes
            cm.markText(from, to, {
              className: classes,
              inclusiveRight: true
            })
          } else {
            for (let mark of cm.findMarks(from, to)) { // we need to update existing marks
              // generally, they should fully cover range once, but
              // TODO: check range coverage

              let oldPos = mark.find()
              let oldClass = mark.className

              // this comparison is okay since metaToCMClasses converts with static order
              if (oldClass === classes) { // it's same mark
                continue
              }
              mark.clear()
              // breaking mark
              if (index > cm.indexFromPos(oldPos.from)) { // adding mark before new
                cm.markText(oldPos.from, from, {
                  className: oldClass,
                  inclusiveRight: true
                })
              }
              if (index + op.len < cm.indexFromPos(oldPos.to)) { // adding after
                cm.markText(to, oldPos.to, {
                  className: oldClass,
                  inclusiveRight: true
                })
              }
              let newfrom = cm.posFromIndex(Math.max(index, cm.indexFromPos(oldPos.from)))
              let newto = cm.posFromIndex(Math.min(index + op.len, cm.indexFromPos(oldPos.to)))
              let newMeta = Object.assign({}, CMClassesToMeta(oldClass), op.meta)
              if (op.isInsert()) newMeta = op.meta

              cm.markText(newfrom, newto, {
                className: metaToCMClasses(newMeta),
                inclusiveRight: true
              })
            }
          }
        }

        index += op.len
      } else if (op.isDelete()) {
        let to = cm.posFromIndex(index + op.len)
        cm.replaceRange('', from, to)
      }
    }
  })
}

CodeMirrorAdapter.prototype.toggleMeta = function (from, to, meta, perm, uid) {
  if (this.cm.indexFromPos(from) === this.cm.indexFromPos(to)) {
    return {
      ok: true,
      delta: new TextOperation()
    }
  }

  let allMeta = true
  let allOur = true

  for (let mark of this.cm.findMarks(from, to)) { // we need to check all marks
    let tmp = CMClassesToMeta(mark.className)
    if (!tmp.hasOwnProperty(meta) || !tmp[meta]) allMeta = false
    if (tmp.userId !== uid) allOur = false
    if (!allMeta && !allOur) break
  }

  // now creating delta
  let delta = new TextOperation().retain(this.cm.indexFromPos(from))
  let metaObj = {}
  metaObj[meta] = !allMeta

  for (let mark of this.cm.findMarks(from, to)) {
    // generally, makrs should cover all range, but
    // TODO: check coverage
    let tmp = CMClassesToMeta(mark.className)
    let pos = mark.find()
    let newfrom = Math.max(this.cm.indexFromPos(from),
                           this.cm.indexFromPos(pos.from))
    let newto = Math.min(this.cm.indexFromPos(to),
                         this.cm.indexFromPos(pos.to))
    let len = newto - newfrom

    if (perm || tmp.userId === uid) {
      delta = delta.retain(len, metaObj)
    } else {
      delta = delta.retain(len)
    }
  }

  let docLen = codemirrorDocLength(this.cm)
  delta = delta.retain(docLen - this.cm.indexFromPos(to))

  return {
    ok: perm || allOur,
    delta: delta
  }
}

CodeMirrorAdapter.prototype.registerCallbacks = function (cb) {
  this.callbacks = cb
}

CodeMirrorAdapter.prototype.onChange = function () {
  // By default, CodeMirror's event order is the following:
  // 1. 'change', 2. 'cursorActivity', 3. 'changes'.
  // We want to fire the 'selectionChange' event after the 'change' event,
  // but need the information from the 'changes' event. Therefore, we detect
  // when a change is in progress by listening to the change event, setting
  // a flag that makes this adapter defer all 'cursorActivity' events.
  this.changeInProgress = true
}

CodeMirrorAdapter.prototype.onChanges = function (_, changes) {
  let pair = CodeMirrorAdapter.operationFromCodeMirrorChanges(changes, this.cm)
  if (!pair[0].isNoop()) {
    this.trigger('change', pair[0], pair[1])
  }
  if (this.selectionChanged) this.trigger('selectionChange')
  this.changeInProgress = false
}

CodeMirrorAdapter.prototype.onCursorActivity =
  CodeMirrorAdapter.prototype.onFocus = function () {
    if (this.changeInProgress) {
      this.selectionChanged = true
    } else {
      this.trigger('selectionChange')
    }
  }

CodeMirrorAdapter.prototype.onBlur = function () {
  if (!this.cm.somethingSelected()) this.trigger('blur')
}

CodeMirrorAdapter.prototype.getValue = function () {
  return this.cm.getValue()
}

CodeMirrorAdapter.prototype.clear = function () {
  this.cm.setValue('')
  this.cm.clearHistory()
  this.cm.getAllMarks().forEach(mark => {
    mark.clear()
  })
}

CodeMirrorAdapter.prototype.getSelection = function () {
  var cm = this.cm

  var selectionList = cm.listSelections()
  var ranges = []
  for (var i = 0; i < selectionList.length; i++) {
    ranges[i] = new Selection.Range(
      cm.indexFromPos(selectionList[i].anchor),
      cm.indexFromPos(selectionList[i].head)
    )
  }

  return new Selection(ranges)
}

CodeMirrorAdapter.prototype.setSelection = function (selection) {
  var ranges = []
  for (var i = 0; i < selection.ranges.length; i++) {
    var range = selection.ranges[i]
    ranges[i] = {
      anchor: this.cm.posFromIndex(range.anchor),
      head: this.cm.posFromIndex(range.head)
    }
  }
  this.cm.setSelections(ranges)
}

var addStyleRule = (function () {
  var added = {}
  var styleElement = document.createElement('style')
  document.documentElement.getElementsByTagName('head')[0].appendChild(styleElement)
  var styleSheet = styleElement.sheet

  return function (css) {
    if (added[css]) return
    added[css] = true
    styleSheet.insertRule(css, (styleSheet.cssRules || styleSheet.rules).length)
  }
}())

CodeMirrorAdapter.prototype.setOtherCursor = function (position, color, clientId) {
  var cursorPos = this.cm.posFromIndex(position)
  var cursorCoords = this.cm.cursorCoords(cursorPos)
  var cursorEl = document.createElement('span')
  cursorEl.className = 'other-client'
  cursorEl.style.display = 'inline-block'
  cursorEl.style.padding = '0'
  cursorEl.style.marginLeft = cursorEl.style.marginRight = '-1px'
  cursorEl.style.borderLeftWidth = '2px'
  cursorEl.style.borderLeftStyle = 'solid'
  cursorEl.style.borderLeftColor = color
  cursorEl.style.height = (cursorCoords.bottom - cursorCoords.top) * 0.9 + 'px'
  cursorEl.style.zIndex = 0
  cursorEl.setAttribute('data-clientid', clientId)
  return this.cm.setBookmark(cursorPos, { widget: cursorEl, insertLeft: true })
}

CodeMirrorAdapter.prototype.setOtherSelectionRange = function (range, color, clientId) {
  var match = /^#([0-9a-fA-F]{6})$/.exec(color)
  if (!match) { throw new Error('only six-digit hex colors are allowed.') }
  var selectionClassName = 'selection-' + match[1]
  var rule = '.' + selectionClassName + ' { background: ' + color + ' }'
  addStyleRule(rule)

  var anchorPos = this.cm.posFromIndex(range.anchor)
  var headPos = this.cm.posFromIndex(range.head)

  return this.cm.markText(
    minPos(anchorPos, headPos),
    maxPos(anchorPos, headPos),
    { className: selectionClassName }
  )
}

CodeMirrorAdapter.prototype.setOtherSelection = function (selection, color, clientId) {
  var selectionObjects = []
  for (var i = 0; i < selection.ranges.length; i++) {
    var range = selection.ranges[i]
    if (range.isEmpty()) {
      selectionObjects[i] = this.setOtherCursor(range.head, color, clientId)
    } else {
      selectionObjects[i] = this.setOtherSelectionRange(range, color, clientId)
    }
  }
  return {
    clear: function () {
      for (var i = 0; i < selectionObjects.length; i++) {
        selectionObjects[i].clear()
      }
    }
  }
}

CodeMirrorAdapter.prototype.trigger = function (event) {
  var args = Array.prototype.slice.call(arguments, 1)
  var action = this.callbacks && this.callbacks[event]
  if (action) { action.apply(this, args) }
}

CodeMirrorAdapter.prototype.applyOperation = function (operation) {
  if (operation.isNoop()) return
  CodeMirrorAdapter.applyOperationToCodeMirror(operation, this.cm)
}

CodeMirrorAdapter.prototype.registerUndo = function (undoFn) {
  this.cm.undo = undoFn
}

CodeMirrorAdapter.prototype.registerRedo = function (redoFn) {
  this.cm.redo = redoFn
}

// Throws an error if the first argument is falsy. Useful for debugging.
// function assert (b, msg) {
//   if (!b) {
//     throw new Error(msg || 'assertion error')
//   }
// }

// Bind a method to an object, so it doesn't matter whether you call
// object.method() directly or pass object.method as a reference to another
// function.
function bind (obj, method) {
  var fn = obj[method]
  obj[method] = function () {
    fn.apply(obj, arguments)
  }
}

export default CodeMirrorAdapter
