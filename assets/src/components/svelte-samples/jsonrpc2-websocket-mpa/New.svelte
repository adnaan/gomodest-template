<script>
import {todoChangeEventHandlers, todosConn} from "../utils";
import {createJsonrpc2SocketStore} from "../../swell/";
const todo = createJsonrpc2SocketStore(todosConn, [], todoChangeEventHandlers)

let input = "";
const handleCreateTodo = async () => {
    if (!input) {
        return
    }
    todo.change("todos/insert",{text: input})
    input = "";
}

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">create new todo</h1>
                <form class="field has-addons mb-6" style="justify-content: center"
                      on:submit|preventDefault={handleCreateTodo}>
                    <div class="control">
                        <input bind:value={input} class="input" type="text" placeholder="a todo">
                    </div>
                    <div class="control">
                        <button class="button is-primary">
                        <span class="icon is-small">
                          <i class="fas fa-plus"></i>
                        </span>
                        </button>
                    </div>
                </form>
        </div>
    </div>
</main>