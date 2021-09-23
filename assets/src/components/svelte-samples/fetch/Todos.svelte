<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {onMount} from 'svelte';

    let todos = [];
    let input = "";
    const todosAPI = '/samples/api/todos'
    onMount(async () => {
        const res = await fetch(todosAPI);
        todos = await res.json();
    });

    const addTodo = async() =>{
        if (!input){
            return
        }

        const res = await fetch(todosAPI, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                text: input
            })
        });
        const todo = await res.json();
        todos = [...todos, todo]
        input = "";
    }

    const removeTodo = async(id) => {
        const res = await fetch(todosAPI +"/" + id, {
            method: 'DELETE',
            headers: {'Content-Type': 'application/json'}
        });
        const index = todos.findIndex(todo => todo.id === id);
        todos.splice(index, 1);
        todos = todos
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
            <ul class:list={todos.length > 0}>
                {#each todos as todo (todo.id)}
                    <li class="list-item" transition:slide="{{duration: 300, easing: elasticInOut}}">
                        <div class="is-flex" style="align-items: center">
                            <span class="is-pulled-left">{todo.text}</span>
                            <div style="flex: 1"></div>
                            <button class="button is-text is-pulled-right is-small" on:click={()=> removeTodo(todo.id)}>
                                <span class="icon">
                                  <i class="fas fa-trash"></i>
                                </span>
                            </button>
                        </div>
                    </li>
                {:else}
                    <li class="has-text-centered"
                        transition:slide="{{delay: 600, duration: 300, easing: elasticInOut}}">Nothing here!
                    </li>
                {/each}
            </ul>
        </div>
    </div>
</main>