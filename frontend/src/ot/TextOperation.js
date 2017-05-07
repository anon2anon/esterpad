import Op from './Op'

// Constructor for new operations.
function TextOperation () {
  if (!this || this.constructor !== TextOperation) {
    // => function was called without 'new'
    return new TextOperation()
  }

  // When an operation is applied to an input string, you can think of this as
  // if an imaginary cursor runs over the entire string and skips over some
  // parts, deletes some parts and inserts characters at some positions. These
  // actions (skip/delete/insert) are stored as an array in the 'ops' property.
  this.ops = []
  // An operation's baseLength is the length of every string the operation
  // can be applied to.
  this.baseLength = 0
  // The targetLength is the length of every string that results from applying
  // the operation on a valid input string.
  this.targetLength = 0
}

TextOperation.prototype.equals = function (other) {
  if (this.baseLength !== other.baseLength) return false
  if (this.targetLength !== other.targetLength) return false
  if (this.ops.length !== other.ops.length) return false
  for (var i = 0; i < this.ops.length; i++) {
    if (this.ops[i].op !== other.ops[i].op) return false
    if (this.ops[i].data !== other.ops[i].data) return false
    if (this.ops[i].metaEquals(other.ops[i])) return false
  }
  return true
}

// Operation are essentially lists of ops. There are three types of ops:
//
// * Retain ops: Advance the cursor position by a given number of characters.
//   Represented by positive ints.
// * Insert ops: Insert a given string at the current cursor position.
//   Represented by strings.
// * Delete ops: Delete the next n characters. Represented by negative ints.

// After an operation is constructed, the user of the library can specify the
// actions of an operation (skip/insert/delete) with these three builder
// methods. They all return the operation for convenient chaining.

// Skip over a given number of characters.
TextOperation.prototype.retain = function (n, meta) {
  if (n instanceof Op) {
    this.baseLength += n.len
    this.targetLength += n.len
    if (this.ops.length > 0 && this.ops[this.ops.length - 1].isRetain() &&
        this.ops[this.ops.length - 1].metaEquals(n)) {
      this.ops[this.ops.length - 1].len += n
    } else {
      this.ops.push(n)
    }
    return this
  }
  if (typeof n !== 'number') {
    throw new Error('retain expects an integer or Op')
  }
  if (n === 0) return this
  this.baseLength += n
  this.targetLength += n
  if (this.ops.length > 0 && this.ops[this.ops.length - 1].isRetain() &&
      this.ops[this.ops.length - 1].metaEquals(meta)) {
    // The last op is a retain op with same meta => we can merge them into one op.
    this.ops[this.ops.length - 1].len += n
  } else {
    // Create a new op.
    this.ops.push(new Op(n, meta))
  }
  return this
}

// Insert a string at the current position.
TextOperation.prototype.insert = function (str, meta) {
  if (str instanceof Op) {
    this.targetLength += str.len
    if (this.ops.length > 0 && this.ops[this.ops.length - 1].isInsert() &&
        this.ops[this.ops.length - 1].metaEquals(str)) {
      this.ops[this.ops.length - 1].data += str
    } else {
      this.ops.push(str)
    }
    return this
  }
  if (typeof str !== 'string') {
    throw new Error('insert expects a string')
  }
  if (str === '') { return this }
  this.targetLength += str.length
  let ops = this.ops
  if (ops.length > 0 && ops[ops.length - 1].isInsert() &&
      ops[ops.length - 1].metaEquals(meta)) {
    // Merge insert op.
    ops[ops.length - 1].data += str
  } else if (ops.length >= 2 && ops[ops.length - 1].isDelete()) {
    // It doesn't matter when an operation is applied whether the operation
    // is delete(3), insert('something') or insert('something'), delete(3).
    // Here we enforce that in this case, the insert op always comes first.
    // This makes all operations that have the same effect when applied to
    // a document of the right length equal in respect to the `equals` method.
    if (ops[ops.length - 2].isInsert() && ops[ops.length - 2].metaEquals(meta)) {
      ops[ops.length - 2].data += str
    } else {
      ops.push(ops[ops.length - 1])
      ops[ops.length - 2] = new Op(str, meta) // -2 is basically original delete
    }
  } else {
    ops.push(new Op(str, meta))
  }
  return this
}

// Delete a string at the current position.
TextOperation.prototype['delete'] = function (n) {
  if (n instanceof Op) {
    this.baseLength += n.len
    if (this.ops.length > 0 && this.ops[this.ops.length - 1].isDelete()) {
      this.ops[this.ops.length - 1].len += n.len
    } else {
      this.ops.push(n)
    }
    return this
  }
  if (typeof n !== 'number') {
    throw new Error('delete expects an integer or a string')
  }
  if (n === 0) return this
  if (n < 0) n = -n
  this.baseLength += n
  if (this.ops.length > 0 && this.ops[this.ops.length - 1].isDelete()) {
    this.ops[this.ops.length - 1].len += n
  } else {
    this.ops.push(new Op(-n))
  }
  return this
}

// Tests whether this operation has no effect.
TextOperation.prototype.isNoop = function () {
  return this.ops.length === 0
}

// Apply an operation to a string, returning a new string. Throws an error if
// there's a mismatch between the input string and the operation.
/* TextOperation.prototype.apply = function (str) {
  var operation = this
  if (str.length !== operation.baseLength) {
    throw new Error('The operation's base length must be equal to the string's length.')
  }
  var newStr = [], j = 0
  var strIndex = 0
  var ops = this.ops
  for (var i = 0, l = ops.length; i < l; i++) {
    var op = ops[i]
    if (op.isRetain()) {
      if (strIndex + op > str.length) {
        throw new Error('Operation can't retain more characters than are left in the string.')
      }
      // Copy skipped part of the old string.
      newStr[j++] = str.slice(strIndex, strIndex + op)
      strIndex += op
    } else if (op.isInsert()) {
      // Insert string.
      newStr[j++] = op
    } else { // delete op
      strIndex -= op
    }
  }
  if (strIndex !== str.length) {
    throw new Error('The operation didn't operate on the whole string.')
  }
  return newStr.join('')
} */

// Computes the inverse of an operation. The inverse of an operation is the
// operation that reverts the effects of the operation, e.g. when you have an
// operation 'insert('hello '); skip(6);' then the inverse is 'delete('hello ')
// skip(6);'. The inverse should be used for implementing undo.
TextOperation.prototype.invert = function (str) {
  var strIndex = 0
  var inverse = new TextOperation()
  var ops = this.ops
  for (let op of ops) {
    if (op.isRetain()) {
      inverse.retain(op)
      strIndex += op.len
    } else if (op.isInsert()) {
      inverse['delete'](op.len)
    } else { // delete op
      // TODO: meta somehow
      inverse.insert(str.slice(strIndex, strIndex - op.len))
      strIndex -= op.len
    }
  }
  return inverse
}

// Compose merges two consecutive operations into one operation, that
// preserves the changes of both. Or, in other words, for each input string S
// and a pair of consecutive operations A and B,
// apply(apply(S, A), B) = apply(S, compose(A, B)) must hold.
TextOperation.prototype.compose = function (operation2) {
  var operation1 = this
  if (operation1.targetLength !== operation2.baseLength) {
    throw new Error('The base length of the second operation has to be the target length of the first operation')
  }

  let operation = new TextOperation() // the combined operation
  let ops1 = operation1.ops
  let ops2 = operation2.ops // for fast access
  let i1 = 0
  let i2 = 0 // current index into ops1 respectively ops2
  let op1 = ops1[i1++]
  let op2 = ops2[i2++] // current ops
  while (true) {
    // Dispatch on the type of op1 and op2
    if (typeof op1 === 'undefined' && typeof op2 === 'undefined') {
      // end condition: both ops1 and ops2 have been processed
      break
    }

    if (op1 && op1.isDelete()) {
      operation['delete'](op1)
      op1 = ops1[i1++]
      continue
    }
    if (op2 && op2.isInsert()) {
      operation.insert(op2)
      op2 = ops2[i2++]
      continue
    }

    if (typeof op1 === 'undefined') {
      throw new Error('Cannot compose operations: first operation is too short.')
    }
    if (typeof op2 === 'undefined') {
      throw new Error('Cannot compose operations: first operation is too long.')
    }

    if (op1.isRetain() && op2.isRetain()) {
      if (op1.len > op2.len) {
        operation.retain(op2)
        op1.len -= op2.len
        // merge metas
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        // merge metas
        operation.retain(op2)
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        let op3 = op1
        // op3.meta = op1.meta.merge(op2.meta)
        operation.retain(op3)
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
    } else if (op1.isInsert() && op2.isDelete()) {
      if (op1.len > op2.len) {
        op1.data = op1.data.slice(op2.len)
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
    } else if (op1.isInsert() && op2.isRetain()) {
      if (op1.len > op2.len) {
        // merge metas
        operation.insert(op1.data.slice(0, op2.len))
        op1.data = op1.data.slice(op2)
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        // merge metas
        operation.insert(op1)
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        // merge metas
        operation.insert(op1)
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
    } else if (op1.isRetain() && op2.isDelete()) {
      if (op1.len > op2.len) {
        operation['delete'](op2)
        op1.len -= op2.len
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        operation['delete'](op2)
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        operation['delete'](op1.len)
        op2 -= op1
        op1 = ops1[i1++]
      }
    } else {
      throw new Error(
        'This shouldn\'t happen: op1: ' +
          JSON.stringify(op1) + ', op2: ' +
          JSON.stringify(op2)
      )
    }
  }
  return operation
}

function getSimpleOp (operation, fn) {
  let ops = operation.ops
  switch (ops.length) {
    case 1:
      return ops[0]
    case 2:
      return ops[0].isRetain() ? ops[1] : (ops[1].isRetain() ? ops[0] : null)
    case 3:
      if (ops[0].isRetain() && ops[2].isRetain()) return ops[1]
  }
  return null
}

function getStartIndex (operation) {
  if (operation.ops[0].isRetain()) return operation.ops[0].len
  return 0
}

// When you use ctrl-z to undo your latest changes, you expect the program not
// to undo every single keystroke but to undo your last sentence you wrote at
// a stretch or the deletion you did by holding the backspace key down. This
// This can be implemented by composing operations on the undo stack. This
// method can help decide whether two operations should be composed. It
// returns true if the operations are consecutive insert operations or both
// operations delete text at the same position. You may want to include other
// factors like the time since the last change in your decision.
TextOperation.prototype.shouldBeComposedWith = function (other) {
  if (this.isNoop() || other.isNoop()) return true

  let startA = getStartIndex(this)
  let startB = getStartIndex(other)
  let simpleA = getSimpleOp(this)
  let simpleB = getSimpleOp(other)
  if (!simpleA || !simpleB) return false

  if (simpleA.isInsert() && simpleB.isInsert()) {
    return startA + simpleA.len === startB
  }

  if (simpleA.isDelete() && simpleB.isDelete()) {
    // there are two possibilities to delete: with backspace and with the
    // delete key.
    return (startB + simpleB.len === startA) || startA === startB
  }

  return false
}

// Decides whether two operations should be composed with each other
// if they were inverted, that is
// `shouldBeComposedWith(a, b) = shouldBeComposedWithInverted(b^{-1}, a^{-1})`.
TextOperation.prototype.shouldBeComposedWithInverted = function (other) {
  if (this.isNoop() || other.isNoop()) return true

  let startA = getStartIndex(this)
  let startB = getStartIndex(other)
  let simpleA = getSimpleOp(this)
  let simpleB = getSimpleOp(other)
  if (!simpleA || !simpleB) return false

  if (simpleA.isInsert() && simpleB.isInsert()) {
    return startA + simpleA.len === startB || startA === startB
  }

  if (simpleA.isDelete() && simpleB.isDelete()) {
    return startB + simpleB.len === startA
  }

  return false
}

// Transform takes two operations A and B that happened concurrently and
// produces two operations A' and B' (in an array) such that
// `apply(apply(S, A), B') = apply(apply(S, B), A')`. This function is the
// heart of OT.
TextOperation.transform = function (operation1, operation2) {
  if (operation1.baseLength !== operation2.baseLength) {
    throw new Error('Both operations have to have the same base length')
  }

  let operation1prime = new TextOperation()
  let operation2prime = new TextOperation()
  let ops1 = operation1.ops
  let ops2 = operation2.ops
  let i1 = 0
  let i2 = 0
  let op1 = ops1[i1++]
  let op2 = ops2[i2++]
  while (true) {
    // At every iteration of the loop, the imaginary cursor that both
    // operation1 and operation2 have that operates on the input string must
    // have the same position in the input string.

    if (typeof op1 === 'undefined' && typeof op2 === 'undefined') {
      // end condition: both ops1 and ops2 have been processed
      break
    }

    // next two cases: one or both ops are insert ops
    // => insert the string in the corresponding prime operation, skip it in
    // the other one. If both op1 and op2 are insert ops, prefer op1.
    if (op1 && op1.isInsert()) {
      operation1prime.insert(op1)
      operation2prime.retain(op1.len)
      op1 = ops1[i1++]
      continue
    }
    if (op2 && op2.isInsert()) {
      operation1prime.retain(op2.len)
      operation2prime.insert(op2)
      op2 = ops2[i2++]
      continue
    }

    if (typeof op1 === 'undefined') {
      throw new Error('Cannot compose operations: first operation is too short.')
    }
    if (typeof op2 === 'undefined') {
      throw new Error('Cannot compose operations: first operation is too long.')
    }

    var minl
    if (op1.isRetain() && op2.isRetain()) {
      // Simple case: retain/retain
      let meta1 = op1.meta
      let meta2 = op2.meta
      if (op1.len > op2.len) {
        minl = op2.len
        op1.len -= op2.len
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        minl = op2.len
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        minl = op1.len
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
      operation1prime.retain(minl, meta1)
      operation2prime.retain(minl, meta2)
    } else if (op1.isDelete() && op2.isDelete()) {
      // Both operations delete the same string at the same position. We don't
      // need to produce any operations, we just skip over the delete ops and
      // handle the case that one operation deletes more than the other.
      if (op1.len > op2.len) {
        op1.len -= op2.len
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
      // next two cases: delete/retain and retain/delete
    } else if (op1.isDelete() && op2.isRetain()) {
      if (op1.len > op2.len) {
        minl = op2.len
        op1.len -= op2.len
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        minl = op2.len
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        minl = op1.len
        op2.len -= op1.len
        op1 = ops1[i1++]
      }
      operation1prime['delete'](minl)
    } else if (op1.isRetain() && op2.isDelete()) {
      if (op1.len > op2.len) {
        minl = op2.len
        op1.len -= op2.len
        op2 = ops2[i2++]
      } else if (op1.len === op2.len) {
        minl = op1.len
        op1 = ops1[i1++]
        op2 = ops2[i2++]
      } else {
        minl = op1.len
        op2 -= op1
        op1 = ops1[i1++]
      }
      operation2prime['delete'](minl)
    } else {
      throw new Error('The two operations aren\'t compatible')
    }
  }

  return [operation1prime, operation2prime]
}

export default TextOperation
