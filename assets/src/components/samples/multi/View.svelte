<script>
    import {call} from "../utils";
    import {todos} from "../todos_ws2_store";
    import TodoItem from "../TodoItem.svelte";
    export let id;
    $todos = call("get",{id: id})
    const handleMessage = (event) => {
        if (event.detail === "deleted") {
            window.location.href = "/samples/svelte_ws2_todos_multi"
        }
        if (event.detail === "updated") {
            $todos = call("get",{id: id})
        }
    }

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            {#if $todos.result}
                <TodoItem todo={$todos.result} on:message={handleMessage}/>
            {/if }
        </div>
    </div>
</main>