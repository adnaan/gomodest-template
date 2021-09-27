<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {Datalist} from "../../swell"
    import {todosURL} from "../utils";

    const sortTodos = (a,b) => {
        return new Date(b.updated_at) - new Date(a.updated_at)
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
            <Datalist resource="todos"
                      url={todosURL}
                      sort={sortTodos}
                      let:items={todos}
                      let:ref={ref}>
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