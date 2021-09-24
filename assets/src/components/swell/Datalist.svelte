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
        insert: (item) => $store = call(resource, opInsert, item),
        delete: (item) => $store = call(resource, opDelete, item),
        update: (item) => $store = call(resource, opUpdate, item),
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
                const op = operations.get(data.id)
                operations.delete(data.id)
                switch (op) {
                    case opList:
                        if (data.result.length > 0) {
                            items = data.result;
                        }
                        break;
                    case opInsert:
                        items = [...items, data.result]
                        break;
                    case opUpdate:
                        items = items.map(item => (item.id === data.result.id) ? data.result : item)
                        break;
                    case opDelete:
                        items = items.filter(item => item.id !== data.result.id)
                        break;
                    default:
                        console.log(`orphan response: ${data.id}`)
                }

            }
        });
    }
    $: if (!isEqual(query,prevQuery)) {
        prevQuery = query
        $store = call(resource, opList, query)
    }
    onMount(() => $store = call(resource, opList, query))
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