<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {Datamap} from "../../swell";
    import {todosURL} from "../utils";
    export let id; // hydrated from the server

    const handleDeleted = (event) => {
        window.location.href = "/samples/svelte_ws2_todos_multi"
    }

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <Datamap resource="todos"
                     id={id} url={todosURL}
                     let:item={todo}
                     let:ref={ref}
                     on:deleted={handleDeleted}>
                <TodoItem todo={todo} ref={ref}/>
            </Datamap>
        </div>
    </div>
</main>