<script>
    import TodoItem from "../jsonrpc2-websocket-spa/TodoItem.svelte";
    import {Datamap} from "../../swell";

    export let id;
    let url = "ws://localhost:3000/samples/ws2"
    if (process.env.ENV === "production") {
        url = `wss://${process.env.HOST}/samples/ws2`
    }

    const handleDeleted = (event) => {
        window.location.href = "/samples/svelte_ws2_todos_multi"
    }

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <Datamap resource="todos"
                     id={id} url={url}
                     let:item={todo}
                     let:ref={ref}
                     on:deleted={handleDeleted}>
                <TodoItem todo={todo} ref={ref}/>
            </Datamap>
        </div>
    </div>
</main>