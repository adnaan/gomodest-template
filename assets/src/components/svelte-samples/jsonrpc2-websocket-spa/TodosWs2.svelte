<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import TodoItem from "./TodoItem.svelte";
    import {todosReducers, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell/";

    const socket = createJsonrpc2Socket(todosURL, []);
    const todos = socket.newStore([], todosReducers, "todos");
    let input = "";
    const pageSize = 3;
    let query = {offset: 0, limit: pageSize}

    let todosListStatus = todos.dispatch("todos/list");
    let todosInsertStatus;

    const handleCreateTodo = async () => {
        if (!input) return;
        todosInsertStatus = todos.dispatch("todos/insert", {text: input})
        input = "";
    }

    const sortTodosRecent = (a, b) => {
        return new Date(b.updated_at) - new Date(a.updated_at)
    }

    let page;
    let currentPageSize = 0;
    $: if (todos) {
        $todos.sort(sortTodosRecent)
        page = $todos.slice(query.offset, query.offset + query.limit)
        currentPageSize = page.length
    }

    const nextPage = () => {
        query = {...query, offset: query.offset += pageSize}
        if (query.offset >= $todos.length) {
            todosListStatus = todos.dispatch("todos/list", query)
        }
    }

    const prevPage = () => {
        query = {...query, offset: query.offset -= pageSize}
    }

</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">todos</h1>
            <form class="field has-addons mb-3"
                  style="justify-content: center"
                  on:submit|preventDefault={handleCreateTodo}>
                <div class="control">
                    <input bind:value={input}
                           class="input"
                           type="text"
                           placeholder="a todo"
                           disabled={$todosInsertStatus && ($todosInsertStatus.pending)}>
                </div>
                <div class="control">
                    <button class="button is-primary">
                        <span class="icon is-small">
                          <i class="fas fa-plus"></i>
                        </span>
                    </button>
                </div>
            </form>
            {#if $todosInsertStatus && $todosInsertStatus.rejected}
                <p class="has-text-centered help is-danger mb-3">
                    error creating todo: {$todosInsertStatus.rejected.message}
                </p>
            {/if}
            {#if $todosListStatus.pending}
                <p class="has-text-centered">
                    Loading ...
                </p>
            {/if}
            {#if $todosListStatus.rejected}
                <li class="box has-text-centered has-text-danger">
                    error fetching todos
                </li>
            {/if}
            {#if $todosListStatus.fulfilled}
                {#if $todos}
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
                                    disabled="{$todos && ($todos.length <= query.offset)
                            || (currentPageSize < query.limit
                            && (query.offset + currentPageSize === $todos.length))}">
                              <span class="icon is-small">
                                <i class="fas fa-arrow-right"></i>
                              </span>
                                <span>Next</span>
                            </button>
                        </p>
                    </div>
                    {#each page as todo (todo.id)}
                        <TodoItem todo={todo} dispatchTodos={todos.dispatch}/>
                    {:else}
                        <li class="has-text-centered"
                            transition:slide="{{delay: 1000, duration: 300, easing: elasticInOut}}">
                            Nothing here!
                        </li>
                    {/each}
                {/if}
            {/if}
        </div>
    </div>
</main>