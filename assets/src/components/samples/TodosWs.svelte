<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {onMount} from 'svelte';
    import websocketStore from "svelte-websocket-store";
    let url = "ws://localhost:3000/samples/ws"
    if (process.env.ENV === "production"){
        url = `wss://${process.env.HOST}/samples/ws`
    }
    export const todos = websocketStore(url, []);
    let id = 1;

    function rpcRequest(method, params) {
        id += 1
        return {
            jsonrpc: "1.0",
            method: method,
            id: id,
            params: params
        }
    }


    $todos = rpcRequest("Todos.List", [])
    let input = "";
    onMount(async () => {
    });

    const addTodo = async () => {
        if (!input) {
            return
        }
        $todos = rpcRequest("Todos.Add", [{text: input}])
        input = "";
    }

    const removeTodo = async (id) => {
        $todos = rpcRequest("Todos.Delete", [{id: id}])
    }

</script>

<!-- below code from https://areknawo.com/making-a-todo-app-in-svelte/ -->
<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">todos</h1>
            <form class="field has-addons" style="justify-content: center" on:submit|preventDefault={addTodo}>
                <div class="control">
                    <input bind:value={input} class="input" type="text" placeholder="a todo">
                </div>
                <div class="control">
                    <button class="button is-primary">
                        <span class="icon is-small">
                          <i class="fas fa-plus"></i>
                        </span>
                    </button>
                </div>
            </form>
            {#if $todos.result}
                <ul class:list={$todos.result.payload.length > 0}>
                    {#each $todos.result.payload as todo (todo.id)}
                        <li class="list-item" transition:slide="{{duration: 300, easing: elasticInOut}}">
                            <div class="is-flex" style="align-items: center">
                                <span class="is-pulled-left">{todo.text}</span>
                                <div style="flex: 1"></div>
                                <button class="button is-text is-pulled-right is-small"
                                        on:click={()=> removeTodo(todo.id)}>
                                <span class="icon">
                                  <i class="fas fa-trash"></i>
                                </span>
                                </button>
                            </div>
                        </li>
                    {:else}
                        <li class="has-text-centered"
                            transition:slide="{{delay: 600, duration: 300, easing: elasticInOut}}">
                            Nothing here!
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>
    </div>
</main>