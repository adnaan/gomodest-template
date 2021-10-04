<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {todosReducers, todosURL} from "../utils";
    import {createJsonrpc2Socket} from "../../swell/";

    const socket = createJsonrpc2Socket(todosURL, []);
    const todos = socket.newStore([], todosReducers, "todos");
    const pageSize = 3;
    let query = {offset: 0, limit: pageSize}
    let todosListStatus = todos.dispatch("todos/list");

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
            <h1 class="title">todos</h1>
            <button class="button is-primary"
                    on:click={()=>window.location.href = '/samples/svelte_ws2_todos_multi/new'}>
                New
            </button>
            <hr/>

            {#if $todosListStatus.pending}
                <p class="has-text-centered">
                    Loading ...
                </p>
            {/if}
            {#if $todosListStatus.rejected}
                <li class="box has-text-centered has-text-danger">
                    error fetching todos {$todosListStatus.rejected}
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
                        <li on:click="{() => window.location.href = '/samples/svelte_ws2_todos_multi/' + todo.id}"
                            class="box is-clickable" transition:slide="{{duration: 300, easing: elasticInOut}}">
                            <div class="is-flex" style="align-items: center;position: relative">
                                <span class="is-pulled-left">{todo.text}</span>
                            </div>
                        </li>
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