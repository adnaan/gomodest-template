import { Controller } from '@hotwired/stimulus'
import { useHover } from 'stimulus-use'

export default class extends Controller {
    static targets = ["tools"]
    connect() {
        useHover(this, { element: this.element });
    }

    mouseEnter() {
        this.toolsTarget.classList.remove('is-hidden')
    }

    mouseLeave() {
        // ...
        this.toolsTarget.classList.add('is-hidden')
    }
}
