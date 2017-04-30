export default class {
  constructor (styleTitle) {
    let styleElement = document.createElement('style')
    document.documentElement.getElementsByTagName('head')[0].appendChild(styleElement)
    this.sheet = styleElement.sheet
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
