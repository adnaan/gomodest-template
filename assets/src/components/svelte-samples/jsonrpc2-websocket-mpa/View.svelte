<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {todoReducer, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell";

    export let id; // hydrated from the server
    const socket = createJsonrpc2Socket(`${todosURL}/${id}`, []);
    const todo = socket.newStore([], todoReducer, "todos");
    const todosGetStatus = todo.dispatch("todos/get", {id: id})
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            {#if $todosGetStatus.pending}
                <p class="has-text-centered">
                    Loading ...
                </p>
            {/if}
            {#if $todosGetStatus.rejected}
                <p class="has-text-centered has-text-danger">
                    error fetching todo: {$todosGetStatus.rejected.message}
                </p>
            {/if}
            {#if $todosGetStatus.fulfilled}
                {#if $todo}
                    <TodoItem todo={$todo} dispatchTodos={todo.dispatch}/>
                {:else}
                    Not found
                {/if}
            {/if}
        </div>
    </div>
</main>