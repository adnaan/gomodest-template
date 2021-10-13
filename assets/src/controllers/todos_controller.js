import {Controller} from "@hotwired/stimulus"
import {createTurboSocket} from "./turbo-socket";

export default class extends Controller {
    static values = {streamTarget: String};
    initialize() {
        let todosURL = "ws://localhost:3000/samples/streams/todos/ws"
        this.dispatcher = createTurboSocket(todosURL, [])
    }

    connect() {
        // if (this.dispatcher) {
        //     this.dispatcher("todos/connect",this.streamTargetValue)
        // }
    }

    addTodo(e) {
        e.preventDefault()
        let formData = new FormData(e.currentTarget);
        let json = {};
        formData.forEach((value, key) => json[key] = value);
        console.log(JSON.stringify(json))
        if (this.dispatcher) {
            this.dispatcher("todos/insert", this.streamTargetValue, json)
        }
    }

    deleteTodo(e) {
        console.log(e.params)
        if (this.dispatcher) {
            this.dispatcher("todos/delete", `todo-${e.params.id}`, e.params)
        }
    }

}