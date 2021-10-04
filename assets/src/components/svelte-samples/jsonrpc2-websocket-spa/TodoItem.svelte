<script>
    import {slide} from "svelte/transition";
    import {elasticInOut} from "svelte/easing";
    import {createEventDispatcher} from 'svelte';

    const dispatch = createEventDispatcher();
    export let todo;
    export let dispatchTodos;
    export let mode = "view";
    let oldTodo;
    let showTool = false;
    let deleteTodoStatus;
    let updateTodoStatus;
    const toggleTool = () => {
        showTool = !showTool;
    }
    const handleDeleteTodo = async () => {
        deleteTodoStatus = dispatchTodos("todos/delete", {id: todo.id})
        dispatch("message", "deleted")
    }
    const confirmDelete = async () => {
        mode = "delete"
    }
    const edit = async () => {
        oldTodo = Object.assign({}, todo);
        mode = "edit";
    }
    const save = async () => {
        if (oldTodo.text != todo.text) {
            updateTodoStatus = dispatchTodos("todos/update", {id: todo.id, text: todo.text})
            dispatch("message", "updated")
           // mode = "view";
        } else if (oldTodo.text === todo.text) {
            mode = "view";
        }
    }

    $: if (updateTodoStatus) {
        if ($updateTodoStatus.fulfilled){
            if (mode === "edit") {
                mode = 'view';
            }
        }
    }

</script>

<li on:mouseenter={toggleTool} on:mouseleave={toggleTool}
    class="box" transition:slide="{{duration: 300, easing: elasticInOut}}">
    {#if mode === "view"}
        <div class="is-flex" style="align-items: center;position: relative">
            <span class="is-pulled-left">{todo.text}</span>
            <div style="flex: 1"></div>
            <div class="card has-background-white-ter has-shadow is-hoverable {showTool ? '':'is-hidden'}"
                 style="position: absolute;top: -30px;right: 0px;">
                <button class="button is-text is-small"
                        on:click={edit}>
                            <span class="icon">
                              <i class="fas fa-edit"></i>
                            </span>
                </button>
                <button class="button is-text is-small"
                        on:click={confirmDelete}>
                            <span class="icon">
                              <i class="fas fa-trash"></i>
                            </span>
                </button>
            </div>
        </div>
    {:else if mode === "edit"}
        {#if $updateTodoStatus && $updateTodoStatus.rejected}
            <p class="has-text-centered has-text-danger">
                error updating todo: {$updateTodoStatus.rejected.message}
            </p>
        {/if}
        <div class="is-flex" style="align-items: center">
            <input bind:value={todo.text} class="input is-small" type="text" placeholder="a todo">
            <div style="flex: 1"></div>
            <button on:click={save}
                    class="button is-primary is-small ml-2 {$updateTodoStatus && $updateTodoStatus.pending ? 'is-loading':''}">
                    <span class="icon">
                      <i class="fas fa-check"></i>
                    </span>
            </button>
        </div>
    {:else if mode === "delete"}
        <div class="is-flex" style="align-items: center">
            <p>Are you sure ? </p>
            <div style="flex: 1"></div>
            <button on:click={handleDeleteTodo}
                    class="button is-danger is-small ml-2 {$deleteTodoStatus && $deleteTodoStatus.pending ? 'is-loading':''}">
                    <span class="icon">
                      <i class="fas fa-check"></i>
                    </span>
            </button>
        </div>
    {/if}
</li>
