<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {todoChangeEventHandlers, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell";

    export let id; // hydrated from the server
    const socket = createJsonrpc2Socket(`${todosURL}/${id}`, []);
    const todo = socket.newStore([], todoChangeEventHandlers, "todos");
    const todosGetStatus = todo.dispatch("todos/get", {id: id})
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            {#if $todosGetStatus.loading}
                <p class="has-text-centered">
                    Loading ...
                </p>
            {:else}
                {#if $todosGetStatus.error}
                    <p class="has-text-centered has-text-danger">
                        error fetching todo
                    </p>
                {:else}
                    {#if $todo}
                        <TodoItem todo={$todo} changeTodo={todo.dispatch}/>
                    {:else}
                        Not found
                    {/if}
                {/if}
            {/if}
        </div>
    </div>
</main>