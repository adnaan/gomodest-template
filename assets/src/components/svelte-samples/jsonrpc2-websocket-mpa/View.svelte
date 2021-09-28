<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {todoChangeEventHandlers, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell";

    export let id; // hydrated from the server
    const socket = createJsonrpc2Socket(`${todosURL}/${id}`, []);
    const todo = socket.newStore([], todoChangeEventHandlers, "todos");

    todo.change("todos/get", {id: id})
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            {#if $todo}
                <TodoItem todo={$todo} changeTodo={todo.change}/>
            {:else}
                Not found
            {/if}
        </div>
    </div>
</main>