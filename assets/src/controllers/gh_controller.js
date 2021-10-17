import {Controller} from "@hotwired/stimulus"
import {createEventDispatcher} from "./gohotwired";

export default class extends Controller {
    initialize() {
        let todosURL = "ws://localhost:3000/samples/gh/todos"
        if (process.env.ENV === "production") {
            todosURL = `wss://${process.env.HOST}/samples/gh/todos`
        }
        this.dispatcher = createEventDispatcher(todosURL, [])
    }

    submit(e){
        e.preventDefault()
        const {eventId, action, target, content, ...rest} = e.params
        if (!eventId || !action || !target || !content){
            console.error("gh controller requires event-id, action, target and content params")
            return
        }
        let json = {...rest};
        let formData = new FormData(e.currentTarget);
        formData.forEach((value, key) => json[key] = value);
        if (this.dispatcher) {
            this.dispatcher(eventId, action, target, content, json)
        }
    }

    change(e){
        e.preventDefault()
        const {eventId, action, target, content, ...rest} = e.params
        if (!eventId || !action || !target || !content){
            console.error("gh controller requires event-id, action, target and content params")
            return
        }
        let json = {...rest};
        if (this.dispatcher) {
            this.dispatcher(eventId, action, target, content, json)
        }
    }

}