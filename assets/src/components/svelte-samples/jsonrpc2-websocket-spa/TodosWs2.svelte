<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import TodoItem from "./TodoItem.svelte";
    import {todosChangeEventHandlers, todosConn} from "../utils";
    import {createJsonrpc2SocketStore} from "../../swell/";

    const todos = createJsonrpc2SocketStore(todosConn, [], todosChangeEventHandlers)
    let input = "";
    const pageSize = 3;
    let query = {offset: 0, limit: pageSize}

    todos.change("todos/list",query)

    const handleCreateTodo = async () => {
        if (!input) {
            return
        }

        todos.change("todos/insert",{text: input})
        input = "";
    }
    const sortTodos = (a, b) => {
        return new Date(a.updated_at) - new Date(b.updated_at)
    }



    const nextPage = () => {
        query = {...query, offset: query.offset += pageSize}
        todos.change("todos/list",query)
    }


    const prevPage = () => {
        query = {...query, offset: query.offset -= pageSize}
        todos.change("todos/list",query)
    }
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">todos</h1>
                <form class="field has-addons mb-6"
                      style="justify-content: center"
                      on:submit|preventDefault={handleCreateTodo}>
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
                <div class="field has-addons"
                     style="justify-content: center">
                    <p class="control">
                        <button class="button"
                                on:click={prevPage}
                                disabled="{query.offset === 0}">
                                  <span class="icon is-small">
                                    <i class="fas fa-arrow-left"></i>
                                  </span>
                            <span>Previous</span>
                        </button>
                    </p>
                    <p class="control">
                        <button class="button" on:click={nextPage}
                                disabled="{query.offset > query.limit && todos && todos.length === 0}">
                      <span class="icon is-small">
                        <i class="fas fa-arrow-right"></i>
                      </span>
                            <span>Next</span>
                        </button>
                    </p>
                </div>
                {#if $todos}
                    {#each $todos as todo (todo.id)}
                        <TodoItem todo={todo} changeTodos={todos.change}/>
                    {:else}
                        <li class="has-text-centered"
                            transition:slide="{{delay: 1000, duration: 300, easing: elasticInOut}}">
                            Nothing here!
                        </li>
                    {/each}
                {/if}
        </div>
    </div>
</main>