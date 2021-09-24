<script>
import {Datamap} from "../../swell";

let url = "ws://localhost:3000/samples/ws2"
if (process.env.ENV === "production") {
    url = `wss://${process.env.HOST}/samples/ws2`
}

let input = "";
const handleCreateTodo = async (createTodo) => {
    if (!input) {
        return
    }
    createTodo({text: input})
    input = "";
}

const handleCreated = (event) => {
    window.location.href = "/samples/svelte_ws2_todos_multi"
}

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">create new todo</h1>
            <Datamap resource="todos" url={url} let:ref={ref} on:created={handleCreated}>
                <form class="field has-addons mb-6" style="justify-content: center"
                      on:submit|preventDefault={handleCreateTodo(ref.insert)}>
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
            </Datamap>
        </div>
    </div>
</main>