<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import TodoItem from "./TodoItem.svelte";
    import {Datalist} from "../../swell"
    import {todosURL} from "../utils";

    let input = "";
    let query = {offset: 0}
    const handleCreateTodo = async (createTodo) => {
        if (!input) {
            return
        }
        createTodo({text: input})
        input = "";
    }
    const sortTodos = (a, b) => {
        return new Date(b.updated_at) - new Date(a.updated_at)
    }
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">todos</h1>
            <Datalist resource="todos"
                      query={query}
                      url={todosURL}
                      sort={sortTodos}
                      let:items={todos}
                      let:ref={ref}>
                <form class="field has-addons mb-6" style="justify-content: center"
                      on:submit|preventDefault={handleCreateTodo(ref.insert)}>
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
                {#if todos}
                    {#each todos as todo (todo.id)}
                        <TodoItem todo={todo} ref={ref}/>
                    {:else}
                        <li class="has-text-centered"
                            transition:slide="{{delay: 600, duration: 300, easing: elasticInOut}}">
                            Nothing here!
                        </li>
                    {/each}
                {/if}
            </Datalist>
        </div>
    </div>
</main>