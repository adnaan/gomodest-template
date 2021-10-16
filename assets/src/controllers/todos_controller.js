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

    addTodo(e) {
        e.preventDefault()
        let formData = new FormData(e.currentTarget);
        let json = {};
        formData.forEach((value, key) => json[key] = value);
        if (this.dispatcher) {
            this.dispatcher("todos/insert", "todos", json)
        }
    }

    updateTodo(e) {
        e.preventDefault()
        console.log(e.params)
        let formData = new FormData(e.currentTarget);
        let json = {
            id: e.params.id,
        };
        formData.forEach((value, key) => json[key] = value);
        if (this.dispatcher) {
            this.dispatcher("todos/update", `todo-${e.params.id}`, json)
        }
    }

    deleteTodo(e) {
        if (this.dispatcher) {
            this.dispatcher("todos/delete", `todo-${e.params.id}`, e.params)
        }
    }

}