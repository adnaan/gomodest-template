<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {Datalist} from "../../swell"

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

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">todos</h1>
            <button class="button is-primary"
                    on:click={()=>window.location.href = '/samples/svelte_ws2_todos_multi/new'}>
                New
            </button>
            <Datalist resource="todos" url={url} let:items={todos} let:ref={ref}>
                {#each todos as todo (todo.id)}
                    <li on:click="{() => window.location.href = '/samples/svelte_ws2_todos_multi/' + todo.id}"
                        class="box is-clickable" transition:slide="{{duration: 300, easing: elasticInOut}}">
                        <div class="is-flex" style="align-items: center;position: relative">
                            <span class="is-pulled-left">{todo.text}</span>
                        </div>
                    </li>
                {:else}
                    <li class="has-text-centered"
                        transition:slide="{{delay: 600, duration: 300, easing: elasticInOut}}">
                        Nothing here!
                    </li>
                {/each}
            </Datalist>
        </div>
    </div>
</main>