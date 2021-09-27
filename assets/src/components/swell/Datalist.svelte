<script>
    import websocketStore from "svelte-websocket-store";
    import {createEventDispatcher, onDestroy, onMount} from "svelte";
    import isEqual from 'lodash.isequal';
    import {call, method, opDelete, opInsert, opList, opUpdate} from "./ops";

    export let resource;
    export let url;
    export let sort;
    let prevUrl
    export let query;
    let prevQuery;
    export let socketOptions = [];
    let prevSocketOptions;
    const dispatch = createEventDispatcher();
    let unsubscribe;
    let store = websocketStore(url, socketOptions);
    let items;

    const ref = {
        insert: (item) => $store = call(method(resource, opInsert), item),
        delete: (item) => $store = call(method(resource, opDelete), item),
        update: (item) => $store = call(method(resource, opUpdate), item),
    }

    // Props changed
    $: if (url != prevUrl || !isEqual(socketOptions, prevSocketOptions)) {
        prevUrl = url;
        prevSocketOptions = socketOptions;
        if (unsubscribe) {
            unsubscribe();
            store = websocketStore(url, socketOptions);
        }
        unsubscribe = store.subscribe(message => {
            if (message.error){
                console.error(message.error)
                dispatch("error", message.error)
                return;
            }
            if (message.result) {
                const op = message.result.method
                switch (op) {
                    case method(resource, opList):
                        if (!Array.isArray(message.result.data)) {
                            console.error(`method ${method(resource, opList)} must return an array`);
                            return;
                        }
                        items = message.result.data;
                        items.sort(sort);
                        break;
                    case method(resource, opInsert):
                        items = [...items, message.result.data];
                        dispatch("inserted", message.result.data);
                        items.sort(sort);
                        break;
                    case method(resource, opUpdate):
                        items = items.map(item => (item.id === message.result.data.id) ? message.result.data : item);
                        dispatch("updated", message.result.data);
                        items.sort(sort);
                        break;
                    case method(resource, opDelete):
                        items = items.filter(item => item.id !== message.result.data.id);
                        dispatch("deleted", message.result.data);
                        items.sort(sort);
                        break;
                    default:
                        console.error(`orphan response: ${JSON.stringify(message.result)}`)
                }
            }
        });
    }
    $: if (!isEqual(query, prevQuery)) {
        prevQuery = query
        $store = call(method(resource, opList), query)
    }
    onMount(() => $store = call(method(resource, opList), query))
    onDestroy(() => unsubscribe());
</script>

<div>
    <slot items={items} ref={ref}/>
    {#if store.loading}
        <slot name="loading"/>
    {:else}
        <slot name="fallback"/>
    {/if}
</div>