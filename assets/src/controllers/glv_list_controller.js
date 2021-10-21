import {Controller} from "@hotwired/stimulus";

export default class extends Controller {
    static targets = ["item"]
    static values = {
        order: {type: String, default: 'desc'},
        offset: {type: Number, default: 0},
        limit: {type: Number, default: 3},
    }

    itemTargetConnected(element) {
        this.sortItems(this.itemTargets)
    }

    itemTargetDisconnected(element) {
        this.sortItems(this.itemTargets)
    }

    orderValueChanged(value, previousValue){
        if (value === previousValue){
            return;
        }
        this.sortItems(this.itemTargets)
    }

    order(e) {
        this.orderValue = e.target.value;
    }

    orderAsc(){
        this.orderValue = 'asc'
    }

    orderDesc(){
        this.orderValue = 'desc'
    }

    // Private
    sortItems(itemTargets) {
        let compareItems = this.orderValue === 'asc'?  compareItemsAsc: compareItemsDesc
        if (itemsAreSorted(compareItems, itemTargets)) return;
        itemTargets.sort(compareItems).forEach(this.append)
    }

    append = child => this.element.append(child)
}

function itemsAreSorted(compareItems, [left, ...rights]) {
    for (const right of rights) {
        if (compareItems(left, right) > 0) return false
        left = right
    }
    return true
}

function compareItemsDesc(a, b) {
    return getSortByVal(b) - getSortByVal(a)
}

function compareItemsAsc(a, b) {
    return getSortByVal(a) - getSortByVal(b)
}

function getSortByVal(item) {
    return item.getAttribute("data-glv-sort-by");
}