<script>
    import websocketStore from "svelte-websocket-store";
    import {createEventDispatcher, onDestroy, onMount} from "svelte";
    import isEqual from "lodash.isequal";

    export let resource;
    export let id;
    let prevId;
    export let url;
    let prevUrl;
    export let socketOptions = [];
    let prevSocketOptions

    const opDelete = "delete";
    const opUpdate = "update";
    const opGet = "get"
    const dispatch = createEventDispatcher();
    let methodID = 0;
    let unsubscribe;
    let store = websocketStore(url, socketOptions);
    let item = {};
    let operations = new Map();

    const call = (resource, method, params) => {
        methodID += 1
        operations.set(methodID, method);
        return {
            jsonrpc: "2.0",
            method: resource ? `${resource}/${method}` : method,
            id: methodID,
            params: params
        }
    }

    const ref = {
        delete: (item) => $store = call(resource, opDelete, item),
        update: (item) => $store = call(resource, opUpdate, item),
    }

    // Props changed
    $: if (url != prevUrl || !isEqual(socketOptions, prevSocketOptions)){
        prevUrl = url;
        prevSocketOptions = socketOptions;
        if (unsubscribe) {
            unsubscribe();
            store = websocketStore(url, []);
        }
        unsubscribe = store.subscribe(data => {
            if (data.result) {
                const op = operations.get(data.id)
                operations.delete(data.id)
                switch (op) {
                    case opGet:
                        item = data.result;
                        break;
                    case opUpdate:
                        item = {...item, ...data.result}
                        break;
                    case opDelete:
                        dispatch("deleted", item)
                        item = {}
                        break;
                    default:
                        console.log(`orphan response: ${data.id}`)
                }
            }
        });
    }
    $: if (id != prevId){
        prevId = id;
        $store = call(resource, opGet,{id: id})
    }
    onMount(() => $store = call(resource, opGet,{id: id}))
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