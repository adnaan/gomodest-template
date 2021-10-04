<script>
    import {todoReducer, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell/";

    const socket = createJsonrpc2Socket(todosURL, []);
    const todo = socket.newStore([], todoReducer, "todos");

    let input = "";
    let todosInsertStatus;
    const handleCreateTodo = async () => {
        if (!input) return;
        todosInsertStatus = todo.dispatch("todos/insert", {text: input})
        input = "";
    }

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">create new todo</h1>
            {#if $todosInsertStatus && $todosInsertStatus.rejected}
                <p class="help is-danger has-text-centered">
                   error creating todo: {$todosInsertStatus.rejected.message}
                </p>
            {/if}

            <form class="field has-addons mb-6" style="justify-content: center"
                  on:submit|preventDefault={handleCreateTodo}>
                <div class="control">
                    <input bind:value={input}
                           class="input"
                           type="text"
                           placeholder="a todo"
                           disabled={$todosInsertStatus && ($todosInsertStatus.pending)}>
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