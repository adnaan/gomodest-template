<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {todoChangeEventHandlers, todosURL} from "../utils";

    export let id; // hydrated from the server
    const todosConn = {
        url: `${todosURL}/${id}`,
        socketOptions: []
    };

    import {createJsonrpc2SocketStore} from "../../swell/";
    const todo = createJsonrpc2SocketStore(todosConn, [], todoChangeEventHandlers)

    todo.change("todos/get",{id: id})
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