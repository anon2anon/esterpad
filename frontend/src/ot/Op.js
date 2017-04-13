export default class {
  constructor (op, meta) {
    if (typeof op === 'string') {
      this.op = 'insert'
      this.data = op
      this.meta = meta || {}
    } else if (typeof op === 'number') {
      if (op > 0) {
        this.op = 'retain'
        this.data = op
        this.meta = meta || {}
      } else {
        this.op = 'delete'
        this.data = -op
      }
    }
  }

  get len () {
    if (this.op === 'insert') {
      return this.data.length
    }
    return this.data
  }

  set len (newLen) {
    if (this.op !== 'insert') {
      this.data = newLen
    } else {
      throw new Error('Trying to set string len');
    }
  }

  getProtobufData () {
    let tmp = {}
    tmp.op = this.op
    tmp[this.op] = {}
    if (this.meta) {
      tmp[this.op].meta = this.meta
    }
    if (this.op === 'insert') {
      tmp[this.op].text = this.data
    } else {
      tmp[this.op].len = this.data
    }
    return tmp
  }

  isInsert () {
    return this.op === 'insert'
  }

  isRetain () {
    return this.op === 'retain'
  }

  isDelete () {
    return this.op === 'delete'
  }
}
