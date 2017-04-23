export default class {
  getSheetByTitle (title) {
    for (let s of document.styleSheets) {
      if (s.title === title) {
        return s
      }
    }
    return null
  }

  constructor (styleTitle) {
    this.sheet = this.getSheetByTitle(styleTitle)
    this.selectorList = []
  }

  indexOfSelector (selector) {
    for (let i = 0; i < this.selectorList.length; ++i) {
      if (this.selectorList[i] === selector) {
        return i
      }
    }
    return -1
  }

  selectorStyle (selector) {
    let i = this.indexOfSelector(selector)
    if (i < 0) {
      this.sheet.insertRule(selector + ' {}', 0)
      this.selectorList.splice(0, 0, selector)
      i = 0
    }
    return this.sheet.cssRules.item(i).style
  }

  removeSelectorStyle (selector) {
    let i = this.indexOfSelector(selector)
    if (i >= 0) {
      this.sheet.deleteRule(i)
      this.selectorList.splice(i, 1)
    }
  }
}
