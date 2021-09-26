<script>
    import websocketStore from "svelte-websocket-store";
    import {onDestroy, onMount} from "svelte";
    import isEqual from 'lodash.isequal';

    export let resource;
    export let url;
    let prevUrl
    export let query;
    let prevQuery;
    export let socketOptions = [];
    let prevSocketOptions;

    const opInsert = "insert";
    const opDelete = "delete";
    const opUpdate = "update";
    const opList = "list"
    let methodID = 0;
    let unsubscribe;
    let store = websocketStore(url, socketOptions);
    let items = [];
    const call = (method, params) => {
        methodID += 1
        return {
            jsonrpc: "2.0",
            method: method,
            id: methodID,
            params: params
        }
    }

    const method = (resource, op) => {
        return  resource ? `${resource}/${op}` : op
    }

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
        unsubscribe = store.subscribe(data => {
            if (data.result) {
                const op = data.result.method
                switch (op) {
                    case method(resource, opList):
                        if (data.result.data.length > 0) {
                            items = data.result.data;
                        }
                        break;
                    case method(resource, opInsert):
                        items = [...items, data.result.data]
                        break;
                    case method(resource, opUpdate):
                        items = items.map(item => (item.id === data.result.data.id) ? data.result.data : item)
                        break;
                    case method(resource, opDelete):
                        items = items.filter(item => item.id !== data.result.data.id)
                        break;
                    default:
                        console.log(`orphan response: ${JSON.stringify(data.result)}`)
                }
            }
        });
    }
    $: if (!isEqual(query,prevQuery)) {
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