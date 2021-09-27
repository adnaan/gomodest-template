<script>
    import websocketStore from "svelte-websocket-store";
    import {createEventDispatcher, onDestroy, onMount} from "svelte";
    import isEqual from "lodash.isequal";
    import {call, method, opDelete, opGet, opInsert, opList, opUpdate} from "./ops";

    export let resource;
    export let id;
    let prevId;
    export let url;
    let prevUrl;
    export let socketOptions = [];
    let prevSocketOptions

    const dispatch = createEventDispatcher();
    let unsubscribe;
    let store = websocketStore(url, socketOptions);
    let item;

    const ref = {
        insert: (item) => $store = call(method(resource, opInsert), item),
        delete: (item) => $store = call(method(resource, opDelete), item),
        update: (item) => $store = call(method(resource, opUpdate), item),
    }

    // Props changed
    $: if (url != prevUrl || !isEqual(socketOptions, prevSocketOptions)){
        prevUrl = url;
        prevSocketOptions = socketOptions;
        if (unsubscribe) {
            unsubscribe();
            store = websocketStore(url, []);
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
                        break;
                    case method(resource, opGet):
                        item = message.result.data;
                        break;
                    case method(resource, opInsert):
                        dispatch("inserted", item)
                        item = message.result.data
                        break;
                    case method(resource, opUpdate):
                        dispatch("updated", item)
                        item = {...item, ...message.result.data}
                        break;
                    case method(resource, opDelete):
                        dispatch("deleted", item)
                        item = {}
                        break;
                    default:
                        console.error(`orphan response: ${message.id}`)
                }
            }
        });
    }
    $: if (id != prevId){
        prevId = id;
        $store = call(method(resource, opGet),{id: id})
    }
    onMount(() => $store = call(method(resource, opGet),{id: id}))
    onDestroy(() => unsubscribe());
</script>

<div>
    <slot item={item} ref={ref}/>
    {#if store.loading}
        <slot name="loading"/>
    {:else}
        <slot name="fallback"/>
    {/if}
</div>