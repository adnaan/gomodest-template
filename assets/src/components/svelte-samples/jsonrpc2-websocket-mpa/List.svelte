<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {todosChangeEventHandlers, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell/";

    const socket = createJsonrpc2Socket(todosURL,[]);
    const todos = socket.newStore([],todosChangeEventHandlers,"todos");

    const sortTodos = (a, b) => {
        return new Date(b.updated_at) - new Date(a.updated_at)
    }
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="title">todos</h1>
            <button class="button is-primary"
                    on:click={()=>window.location.href = '/samples/svelte_ws2_todos_multi/new'}>
                New
            </button>
            <hr/>

                {#if $todos}
                    {#each $todos as todo (todo.id)}
                        <li on:click="{() => window.location.href = '/samples/svelte_ws2_todos_multi/' + todo.id}"
                            class="box is-clickable" transition:slide="{{duration: 300, easing: elasticInOut}}">
                            <div class="is-flex" style="align-items: center;position: relative">
                                <span class="is-pulled-left">{todo.text}</span>
                            </div>
                        </li>
                    {:else}
                        <li class=""
                            transition:slide="{{delay: 600, duration: 300, easing: elasticInOut}}">
                            Nothing here!
                        </li>
                    {/each}
                {/if}
        </div>
    </div>
</main>